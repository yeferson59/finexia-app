<script lang="ts">
	/**
	 * Los aportes y retiros ya anotados de un fondo, del más reciente, con lo
	 * necesario para corregirlos o borrarlos.
	 *
	 * En un fondo por unidades cada fila dice también cuántas unidades y a qué
	 * valor de unidad, que es lo que trae el extracto; en uno por saldo, solo el
	 * dinero: sus unidades son de Finexia.
	 *
	 * Borrar pide confirmación en la misma fila: un movimiento borrado cambia el
	 * costo, la ganancia y, en un fondo por saldo, todo lo que viene después.
	 */
	import { enhance } from '$app/forms';
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import { formatCalendarDate } from '$lib/shared/format/date';
	import type { Fund, FundMovement } from '../funds';

	interface Props {
		fund: Fund;
		movements: FundMovement[];
		/** El movimiento que se está corrigiendo arriba, para marcarlo. */
		editingId: string | null;
		onEdit: (movement: FundMovement) => void;
	}

	let { fund, movements, editingId, onEdit }: Props = $props();

	const byBalance = $derived(fund.tracking === 'balance');

	/* La fila cuyo borrado se está confirmando. */
	let confirmingId = $state<string | null>(null);

	const money = (value: string) =>
		privacy.money(formatCurrency(parseFloat(value) || 0, fund.currency));

	const unitValue = (value: string) =>
		privacy.money(formatCurrency(parseFloat(value) || 0, fund.currency, 6));

	const units = (value: string) =>
		new Intl.NumberFormat('es-CO', { maximumFractionDigits: 8 }).format(parseFloat(value) || 0);

	const day = (iso: string) =>
		formatCalendarDate(iso, { day: 'numeric', month: 'short', year: 'numeric' });

	const label = (m: FundMovement) =>
		m.kind === 'contribution' ? 'Aporte' : m.all ? 'Retiro total' : 'Retiro';
</script>

<section class="history" aria-labelledby="fund-movements-title">
	<h3 id="fund-movements-title">Movimientos</h3>
	{#if movements.length === 0}
		<p class="hint">Todavía no hay ninguno.</p>
	{:else}
		<ul>
			{#each movements as m (m.txnId)}
				<li class:editing={m.txnId === editingId}>
					<span class="when">{day(m.date)}</span>
					<span class="what" class:out={m.kind === 'withdrawal'}>
						{label(m)}
						{#if !byBalance}
							<small>{units(m.units)} u × {unitValue(m.unitValue)}</small>
						{/if}
					</span>
					<span class="amount">
						{m.kind === 'withdrawal' ? '−' : '+'}{money(m.amount)}
						{#if parseFloat(m.fees) > 0}
							<small>comisión {money(m.fees)}</small>
						{/if}
					</span>
					<span class="row-actions">
						{#if confirmingId === m.txnId}
							<form
								method="POST"
								action="?/deleteMovement"
								use:enhance={() =>
									async ({ update }) => {
										confirmingId = null;
										await update();
									}}
							>
								<input type="hidden" name="txnId" value={m.txnId} />
								<button type="submit" class="remove danger">Sí, borrar</button>
							</form>
							<button type="button" class="remove" onclick={() => (confirmingId = null)}>
								No
							</button>
						{:else}
							<button
								type="button"
								class="remove"
								aria-label="Corregir el movimiento del {day(m.date)}"
								onclick={() => onEdit(m)}
							>
								Corregir
							</button>
							<button
								type="button"
								class="remove"
								aria-label="Borrar el movimiento del {day(m.date)}"
								onclick={() => (confirmingId = m.txnId)}
							>
								Borrar
							</button>
						{/if}
					</span>
				</li>
			{/each}
		</ul>
	{/if}
</section>

<style>
	.history {
		margin-top: 1.5rem;
		padding-top: 1.25rem;
		border-top: 1px solid var(--border);
	}

	.history h3 {
		margin: 0 0 0.75rem;
		font-size: 0.78rem;
		font-weight: 500;
		letter-spacing: 0.04em;
		text-transform: uppercase;
		color: var(--text-dim);
	}

	ul {
		display: grid;
		gap: 0.35rem;
		max-height: 18rem;
		margin: 0;
		padding: 0;
		overflow-y: auto;
		list-style: none;
	}

	li {
		display: grid;
		grid-template-columns: minmax(6rem, 1fr) minmax(6rem, 1.2fr) minmax(6rem, 1.2fr) auto;
		align-items: center;
		gap: 0.75rem;
		padding: 0.4rem 0.35rem;
		border-radius: 6px;
		font-size: 0.85rem;
	}

	li.editing {
		background: rgba(212, 145, 42, 0.08);
	}

	.when {
		color: var(--text-muted);
	}

	/* Sin verde: un aporte es dinero que metiste, no ganancia. */
	.what {
		display: grid;
		color: var(--text);
	}

	.what.out {
		color: var(--text-muted);
	}

	small {
		font-size: 0.72rem;
		color: var(--text-dim);
	}

	.amount {
		display: grid;
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
		text-align: right;
		overflow-wrap: anywhere;
	}

	.row-actions {
		display: flex;
		gap: 0.35rem;
		justify-content: flex-end;
	}

	.remove {
		padding: 0.2rem 0.5rem;
		border: 1px solid var(--border);
		border-radius: 6px;
		background: transparent;
		font: inherit;
		font-size: 0.75rem;
		color: var(--text-dim);
		cursor: pointer;
	}

	.remove:hover,
	.remove.danger {
		border-color: var(--red);
		color: var(--red);
	}

	@media (max-width: 480px) {
		li {
			grid-template-columns: 1fr auto;
		}

		.row-actions {
			grid-column: 1 / -1;
		}
	}
</style>
