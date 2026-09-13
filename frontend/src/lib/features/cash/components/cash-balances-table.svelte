<script lang="ts">
	/**
	 * Los saldos, uno por plataforma, portafolio y moneda.
	 *
	 * La cuenta se nombra por la plataforma —es donde está el dinero— y el
	 * portafolio va debajo, porque es dónde cuenta. El botón de cada fila abre
	 * el formulario con ese saldo ya elegido: casi siempre se anota un movimiento
	 * sobre una cuenta que ya existe.
	 */
	import Button from '$lib/ui/button.svelte';
	import DataTable from '$lib/ui/data-table.svelte';
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import { formatCalendarDate } from '$lib/shared/format/date';
	import type { CashBalance } from '../cash';

	interface Props {
		balances: CashBalance[];
		onRecord: (balance: CashBalance) => void;
	}

	let { balances, onRecord }: Props = $props();

	const money = (amount: string, currency: string) =>
		privacy.money(formatCurrency(parseFloat(amount) || 0, currency));

	const displayCurrency = $derived(balances[0]?.displayCurrency ?? '');

	/* Si todo está en la moneda de la pantalla, la columna repetiría el saldo. */
	const showValue = $derived(balances.some((b) => b.currency !== b.displayCurrency));

	function lastMovement(balance: CashBalance): string {
		if (!balance.lastMovementDate) return 'sin movimientos';
		return `último movimiento el ${formatCalendarDate(balance.lastMovementDate, {
			day: 'numeric',
			month: 'short',
			year: 'numeric'
		})}`;
	}
</script>

<DataTable caption="Saldos de efectivo por plataforma, portafolio y moneda">
	<thead>
		<tr>
			<th scope="col">Cuenta</th>
			<th scope="col" class="col-currency">Moneda</th>
			<th scope="col" class="num">Saldo</th>
			{#if showValue}
				<th scope="col" class="num col-value">En {displayCurrency}</th>
			{/if}
			<th scope="col" class="col-actions" aria-label="Acciones"></th>
		</tr>
	</thead>
	<tbody>
		{#each balances as balance (balance.entryId)}
			<tr class="row" class:emptied={(parseFloat(balance.balance) || 0) === 0}>
				<td class="cell-account">
					<span class="source">{balance.sourceName || 'Sin plataforma'}</span>
					<span class="meta">{balance.portfolioName} · {lastMovement(balance)}</span>
				</td>
				<td class="col-currency"><span class="code">{balance.currency}</span></td>
				<td class="num figure">{money(balance.balance, balance.currency)}</td>
				{#if showValue}
					<td class="num figure col-value muted">
						{#if balance.fxConverted}
							{money(balance.value, balance.displayCurrency)}
						{:else}
							<span title="No hay tasa de cambio guardada para convertir este saldo">—</span>
						{/if}
					</td>
				{/if}
				<td class="col-actions">
					<Button type="button" variant="secondary" size="sm" onclick={() => onRecord(balance)}>
						Movimiento
					</Button>
				</td>
			</tr>
		{/each}
	</tbody>
</DataTable>

<style>
	.row :global(td) {
		padding-top: 0.85rem;
		padding-bottom: 0.85rem;
	}

	.cell-account {
		width: 100%;
		min-width: 12rem;
	}

	.source {
		display: block;
		font-family: var(--font-display);
		font-size: 1.02rem;
		font-weight: 500;
		color: var(--text);
	}

	.meta {
		display: block;
		margin-top: 0.2rem;
		font-size: 0.8rem;
		color: var(--text-dim);
	}

	.code {
		font-family: var(--font-mono);
		font-size: 0.75rem;
		letter-spacing: 0.06em;
		color: var(--amber);
	}

	.figure {
		font-variant-numeric: tabular-nums;
		white-space: nowrap;
	}

	.muted {
		color: var(--text-muted);
	}

	.col-actions {
		text-align: right;
		white-space: nowrap;
	}

	/* Una cuenta vaciada sigue en la lista —es donde cae el próximo depósito—,
	   pero sin competir con las que tienen dinero. */
	.emptied .source,
	.emptied .figure {
		color: var(--text-dim);
	}

	@media (max-width: 720px) {
		.col-currency,
		.col-value {
			display: none;
		}
	}
</style>
