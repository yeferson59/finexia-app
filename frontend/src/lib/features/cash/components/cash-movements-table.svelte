<script lang="ts">
	/**
	 * Los movimientos de efectivo, del más reciente al más antiguo.
	 *
	 * El importe lleva signo pero no color de ganancia: pintar un depósito de
	 * verde diría justo lo que esta pantalla explica que no es. Solo los intereses
	 * van en verde, porque son lo único de aquí que es rendimiento.
	 */
	import Button from '$lib/ui/button.svelte';
	import DataTable from '$lib/ui/data-table.svelte';
	import Pagination from '$lib/ui/pagination.svelte';
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import { formatCalendarDate } from '$lib/shared/format/date';
	import { cashAccountLabel, cashKindSign, formatCashKind, type CashMovement } from '../cash';
	import CashDeleteConfirm from './cash-delete-confirm.svelte';

	interface Props {
		movements: CashMovement[];
		/** Cuántos hay en total; puede ser más de los que se trajeron. */
		total: number;
		onEdit: (movement: CashMovement) => void;
	}

	let { movements, total, onEdit }: Props = $props();

	const PER_PAGE = 15;
	let page = $state(1);
	const paged = $derived(movements.slice((page - 1) * PER_PAGE, page * PER_PAGE));

	let deleting = $state<CashMovement | null>(null);

	function amount(movement: CashMovement): string {
		const sign = cashKindSign(movement.kind);
		const value = formatCurrency(Math.abs(parseFloat(movement.amount) || 0), movement.currency);
		const prefix = sign > 0 ? '+' : sign < 0 ? '−' : '';

		return privacy.money(`${prefix}${value}`);
	}

	function fees(movement: CashMovement): string {
		const value = parseFloat(movement.fees) || 0;
		return value > 0 ? privacy.money(formatCurrency(value, movement.feesCurrency)) : '';
	}

	function date(iso: string): string {
		return formatCalendarDate(iso, { day: 'numeric', month: 'short', year: 'numeric' });
	}
</script>

<DataTable caption="Movimientos de efectivo, del más reciente al más antiguo">
	<thead>
		<tr>
			<th scope="col" class="col-date">Fecha</th>
			<th scope="col">Movimiento</th>
			<th scope="col" class="col-account">Cuenta</th>
			<th scope="col" class="num">Importe</th>
			<th scope="col" class="col-actions" aria-label="Acciones"></th>
		</tr>
	</thead>
	<tbody>
		{#each paged as movement (movement.id)}
			<tr>
				<td class="col-date date">{date(movement.date)}</td>
				<th scope="row" class="kind">
					{formatCashKind(movement.kind)}
					<span class="detail narrow-only">{cashAccountLabel(movement)}</span>
					{#if movement.notes}
						<span class="detail">{movement.notes}</span>
					{/if}
				</th>
				<td class="col-account account">{cashAccountLabel(movement)}</td>
				<td class="num figure">
					<span class:income={movement.kind === 'interest'}>{amount(movement)}</span>
					{#if fees(movement)}
						<span class="detail">comisión {fees(movement)}</span>
					{/if}
				</td>
				<td class="col-actions">
					{#if movement.editable}
						<Button type="button" variant="ghost" size="sm" onclick={() => onEdit(movement)}>
							Editar
						</Button>
					{/if}
					<Button type="button" variant="ghost" size="sm" onclick={() => (deleting = movement)}>
						Borrar
					</Button>
				</td>
			</tr>
		{/each}
	</tbody>
</DataTable>

<div class="pager">
	<Pagination bind:page total={movements.length} perPage={PER_PAGE} label="movimientos" />
	{#if total > movements.length}
		<p class="hint">Se muestran los {movements.length} movimientos más recientes de {total}.</p>
	{/if}
</div>

<CashDeleteConfirm movement={deleting} onClose={() => (deleting = null)} />

<style>
	.date {
		white-space: nowrap;
		color: var(--text-muted);
	}

	.kind {
		font-weight: 500;
		text-align: left;
		color: var(--text);
	}

	.detail {
		display: block;
		margin-top: 0.15rem;
		max-width: 36ch;
		font-size: 0.78rem;
		font-weight: 400;
		color: var(--text-dim);
		white-space: normal;
	}

	.account {
		color: var(--text-muted);
	}

	.figure {
		font-variant-numeric: tabular-nums;
		white-space: nowrap;
	}

	.income {
		color: var(--green);
	}

	.col-actions {
		text-align: right;
		white-space: nowrap;
	}

	.narrow-only {
		display: none;
	}

	.pager {
		padding: 0 1.25rem;
	}

	.pager .hint {
		padding-bottom: 1rem;
	}

	@media (max-width: 720px) {
		.col-account,
		.col-date {
			display: none;
		}

		.narrow-only {
			display: block;
		}

		.pager {
			padding: 0 0.9rem;
		}
	}
</style>
