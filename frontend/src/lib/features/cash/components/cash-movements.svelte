<script lang="ts">
	/**
	 * Los movimientos de efectivo, como un extracto: por meses, del más reciente
	 * al más antiguo.
	 *
	 * El mes va de cabecera y cada movimiento solo lleva su día, que es como se
	 * lee el extracto de un banco y lo que deja sitio a la nota. Como tabla de
	 * cinco columnas, en una pantalla estrecha la cuenta se partía en cuatro
	 * líneas y los botones quedaban cortados.
	 *
	 * El importe lleva signo pero no color de ganancia: pintar un depósito de
	 * verde diría justo lo que esta pantalla explica que no es. Solo los intereses
	 * van en verde, porque son lo único de aquí que es rendimiento.
	 */
	import Pagination from '$lib/ui/pagination.svelte';
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import { formatCalendarDate } from '$lib/shared/format/date';
	import {
		cashKindSign,
		formatCashKind,
		groupCashMovementsByMonth,
		type CashMovement
	} from '../cash';
	import CashDeleteConfirm from './cash-delete-confirm.svelte';

	interface Props {
		movements: CashMovement[];
		/** Cuántos hay en total; puede ser más de los que se trajeron. */
		total: number;
		/** Si se nombra también el portafolio. Con uno solo, sobra. */
		showPortfolio?: boolean;
		onEdit: (movement: CashMovement) => void;
	}

	let { movements, total, showPortfolio = true, onEdit }: Props = $props();

	const PER_PAGE = 15;
	let page = $state(1);
	const months = $derived(
		groupCashMovementsByMonth(movements.slice((page - 1) * PER_PAGE, page * PER_PAGE))
	);

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

	const day = (iso: string) => formatCalendarDate(iso, { day: 'numeric' });

	const weekday = (iso: string) => formatCalendarDate(iso, { weekday: 'short' }).replace('.', '');

	const fullDate = (iso: string) =>
		formatCalendarDate(iso, { day: 'numeric', month: 'long', year: 'numeric' });

	/* Lo que distingue un «Editar» de los otros catorce para quien no ve la fila. */
	const described = (movement: CashMovement) =>
		`${formatCashKind(movement.kind).toLowerCase()} del ${fullDate(movement.date)}`;
</script>

<div class="ledger">
	{#each months as month (month.key)}
		<section class="month" aria-labelledby="cash-month-{month.key}">
			<h3 class="month-title" id="cash-month-{month.key}">{month.label}</h3>

			<ul class="entries">
				{#each month.movements as movement (movement.id)}
					<li class="entry">
						<p class="day">
							<span class="day-number" aria-hidden="true">{day(movement.date)}</span>
							<span class="weekday" aria-hidden="true">{weekday(movement.date)}</span>
							<span class="sr-only">{fullDate(movement.date)}</span>
						</p>

						<div class="what">
							<p class="kind">{formatCashKind(movement.kind)}</p>
							{#if movement.notes}
								<p class="note">{movement.notes}</p>
							{/if}
						</div>

						<p class="where">
							<span class="source">{movement.sourceName || 'Sin plataforma'}</span>
							<span class="code">{movement.currency}</span>
							{#if showPortfolio}
								<span class="portfolio">{movement.portfolioName}</span>
							{/if}
						</p>

						<p class="figures">
							<span class="amount" class:income={movement.kind === 'interest'}>
								{amount(movement)}
							</span>
							{#if fees(movement)}
								<span class="fee">comisión {fees(movement)}</span>
							{/if}
						</p>

						<div class="actions">
							{#if movement.editable}
								<button type="button" class="action" onclick={() => onEdit(movement)}>
									Editar<span class="sr-only"> {described(movement)}</span>
								</button>
							{/if}
							<button type="button" class="action danger" onclick={() => (deleting = movement)}>
								Borrar<span class="sr-only"> {described(movement)}</span>
							</button>
						</div>
					</li>
				{/each}
			</ul>
		</section>
	{/each}
</div>

<div class="pager">
	<Pagination bind:page total={movements.length} perPage={PER_PAGE} label="movimientos" />
	{#if total > movements.length}
		<p class="hint">Se muestran los {movements.length} movimientos más recientes de {total}.</p>
	{/if}
</div>

<CashDeleteConfirm movement={deleting} {showPortfolio} onClose={() => (deleting = null)} />

<style>
	.month + .month {
		border-top: 1px solid var(--border-strong);
	}

	.month-title {
		margin: 0;
		padding: 1.1rem 1.5rem 0.3rem;
		font-family: var(--font-display);
		font-size: 1rem;
		font-weight: 500;
		color: var(--text-muted);
	}

	.entries {
		margin: 0;
		padding: 0 0 0.35rem;
		list-style: none;
	}

	/*
	 * Día, qué pasó, dónde, cuánto y las acciones. Las dos últimas columnas son
	 * fijas para que los importes caigan en la misma vertical de mes a mes.
	 */
	.entry {
		display: grid;
		grid-template-columns: 2.5rem minmax(0, 1.35fr) minmax(0, 1fr) 9.5rem 8.5rem;
		grid-template-areas: 'day what where figures actions';
		align-items: start;
		column-gap: 1.25rem;
		margin: 0 1.5rem;
		padding: 0.85rem 0;
	}

	.entry + .entry {
		border-top: 1px solid var(--border);
	}

	p {
		margin: 0;
	}

	.day {
		grid-area: day;
		display: flex;
		flex-direction: column;
		line-height: 1;
	}

	.day-number {
		font-family: var(--font-display);
		font-size: 1.4rem;
		font-weight: 400;
		font-variant-numeric: lining-nums;
		color: var(--text);
	}

	.weekday {
		margin-top: 0.3rem;
		font-size: 0.72rem;
		color: var(--text-dim);
	}

	.what {
		grid-area: what;
		min-width: 0;
		padding-top: 0.1rem;
	}

	.kind {
		font-size: 0.9rem;
		font-weight: 500;
		color: var(--text);
	}

	.note {
		margin-top: 0.15rem;
		max-width: 44ch;
		font-size: 0.8rem;
		line-height: 1.45;
		color: var(--text-dim);
		overflow-wrap: anywhere;
	}

	.where {
		grid-area: where;
		display: flex;
		flex-wrap: wrap;
		align-items: baseline;
		column-gap: 0.5rem;
		min-width: 0;
		padding-top: 0.1rem;
		font-size: 0.84rem;
		color: var(--text-muted);
	}

	.code {
		font-family: var(--font-mono);
		font-size: 0.72rem;
		letter-spacing: 0.04em;
		color: var(--text-dim);
	}

	.portfolio {
		flex-basis: 100%;
		margin-top: 0.15rem;
		font-size: 0.78rem;
		color: var(--text-dim);
	}

	.figures {
		grid-area: figures;
		display: flex;
		flex-direction: column;
		align-items: flex-end;
		gap: 0.2rem;
		padding-top: 0.05rem;
	}

	.amount {
		font-family: var(--font-mono);
		font-size: 0.95rem;
		font-variant-numeric: tabular-nums;
		white-space: nowrap;
		color: var(--text);
	}

	.income {
		color: var(--green);
	}

	.fee {
		font-family: var(--font-mono);
		font-size: 0.74rem;
		white-space: nowrap;
		color: var(--text-dim);
	}

	.actions {
		grid-area: actions;
		display: flex;
		justify-content: flex-end;
		gap: 0.15rem;
		margin-top: -0.25rem;
	}

	/* Texto, no botones con borde: dos por fila y quince filas eran treinta
	   cajas compitiendo con los importes. */
	.action {
		padding: 0.35rem 0.6rem;
		border: none;
		border-radius: 6px;
		background: transparent;
		font: inherit;
		font-size: 0.8rem;
		color: var(--text-dim);
		cursor: pointer;
		transition:
			background 0.2s ease,
			color 0.2s ease;
	}

	.action:hover {
		background: var(--surface-2);
		color: var(--text);
	}

	.action.danger:hover {
		color: var(--red);
	}

	.pager {
		padding: 0 1.5rem;
	}

	.pager .hint {
		padding-bottom: 1rem;
	}

	/* En estrecho el día se queda a la izquierda de todo y el resto se apila:
	   qué pasó con su importe, dónde, y las acciones. */
	@media (max-width: 720px) {
		.month-title {
			padding: 1rem 1rem 0.2rem;
		}

		.entry {
			grid-template-columns: 2.25rem minmax(0, 1fr) auto;
			grid-template-areas:
				'day what figures'
				'day where where'
				'day actions actions';
			row-gap: 0.35rem;
			column-gap: 0.75rem;
			margin: 0 1rem;
		}

		.actions {
			justify-content: flex-start;
			margin: 0 0 0 -0.6rem;
		}

		.pager {
			padding: 0 1rem;
		}
	}
</style>
