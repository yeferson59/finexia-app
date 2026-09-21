<script lang="ts">
	/**
	 * Una fila del extracto de efectivo: el día, qué pasó, en qué cuenta, cuánto y
	 * lo que se le puede hacer.
	 *
	 * La rejilla la marca `--ledger-columns`, que pone `cash-movements` para que
	 * todas las filas —y las de sus grupos de abonos— compartan columnas y los
	 * importes caigan en la misma vertical de mes a mes.
	 *
	 * `compact` es un abono dentro de un grupo desplegado: la cuenta y el tipo ya
	 * los dice el grupo, así que solo lleva la fecha, el importe y las acciones.
	 *
	 * Un dividendo, lo recibido por una venta y lo que se pagó por una compra no
	 * se editan ni se borran aquí: son dinero que entró o salió por una operación
	 * sobre una acción, y cambian o se van con ella. La fila lleva a ella en su
	 * lugar.
	 */
	import { resolve } from '$app/paths';
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import { formatCalendarDate } from '$lib/shared/format/date';
	import { cashKindSign, formatCashKind, type CashMovement } from '../cash';

	interface Props {
		movement: CashMovement;
		showPortfolio: boolean;
		compact?: boolean;
		onEdit: () => void;
		onDelete: () => void;
	}

	let { movement, showPortfolio, compact = false, onEdit, onDelete }: Props = $props();

	/** A dónde lleva la fila de un movimiento que pertenece a otra transacción. */
	const LINKED_LABELS: Record<string, string> = {
		sale: 'Ver venta',
		dividend: 'Ver dividendo',
		purchase: 'Ver compra'
	};

	const day = $derived(formatCalendarDate(movement.date, { day: 'numeric' }));
	const weekday = $derived(
		formatCalendarDate(movement.date, { weekday: 'short' }).replace('.', '')
	);
	const fullDate = $derived(
		formatCalendarDate(movement.date, { day: 'numeric', month: 'long', year: 'numeric' })
	);

	const amount = $derived.by(() => {
		const sign = cashKindSign(movement.kind);
		const prefix = sign > 0 ? '+' : sign < 0 ? '−' : '';
		const value = Math.abs(parseFloat(movement.amount) || 0);
		return privacy.money(`${prefix}${formatCurrency(value, movement.currency)}`);
	});

	const fee = $derived.by(() => {
		const value = parseFloat(movement.fees) || 0;
		return value > 0 ? privacy.money(formatCurrency(value, movement.feesCurrency)) : '';
	});

	/* Lo que distingue un «Editar» de los otros catorce para quien no ve la fila. */
	const described = $derived(`${formatCashKind(movement.kind).toLowerCase()} del ${fullDate}`);

	/* Las filas que pertenecen a otra transacción: no se editan ni se borran
	   aquí, y llevan enlace a la operación de la que salieron. Un dividendo y
	   una venta traen el dinero; una compra se lo lleva. */
	const isLinked = $derived(
		movement.kind === 'dividend' || movement.kind === 'sale' || movement.kind === 'purchase'
	);

	/* La acción de la que viene, en el mismo portafolio que el saldo. */
	const originHref = $derived(
		isLinked && movement.originTicker
			? resolve('/dashboard/portfolios/[id]/assets/[symbol]', {
					id: movement.portfolioId,
					symbol: movement.originTicker
				})
			: null
	);
</script>

<li class="entry" class:compact>
	{#if compact}
		<p class="date">{fullDate}</p>
	{:else}
		<p class="day">
			<span class="day-number" aria-hidden="true">{day}</span>
			<span class="weekday" aria-hidden="true">{weekday}</span>
			<span class="sr-only">{fullDate}</span>
		</p>

		<div class="what">
			<p class="kind">
				{formatCashKind(movement.kind)}
				{#if isLinked && movement.originTicker}
					<span class="tag">de {movement.originTicker}</span>
				{/if}
				{#if movement.automatic}
					<span class="tag">automático</span>
				{/if}
			</p>
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
	{/if}

	<p class="figures">
		<!-- Lo recibido por una venta no va en verde: es capital que vuelve a la
		     cuenta, y su ganancia, si la hubo, ya estaba en la acción. -->
		<span class="amount" class:income={movement.kind === 'interest' || movement.kind === 'dividend'}
			>{amount}</span
		>
		{#if fee}
			<span class="fee">comisión {fee}</span>
		{/if}
	</p>

	<div class="actions">
		{#if isLinked}
			{#if originHref}
				<a class="action" href={originHref}>
					{LINKED_LABELS[movement.kind]}<span class="sr-only"> {described}</span>
				</a>
			{/if}
		{:else}
			{#if movement.editable}
				<button type="button" class="action" onclick={onEdit}>
					Editar<span class="sr-only"> {described}</span>
				</button>
			{/if}
			<button type="button" class="action danger" onclick={onDelete}>
				Borrar<span class="sr-only"> {described}</span>
			</button>
		{/if}
	</div>
</li>

<style>
	.entry {
		display: grid;
		grid-template-columns: var(--ledger-columns);
		grid-template-areas: 'day what where figures actions';
		align-items: start;
		column-gap: 1.25rem;
		padding: 0.85rem 0.75rem;
		transition: background-color 0.15s ease;
	}

	.entry.compact {
		grid-template-columns: minmax(0, 1fr) 10rem 8.5rem;
		grid-template-areas: 'date figures actions';
		align-items: center;
		padding: 0.3rem 0.75rem 0.3rem 1.25rem;
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
		font-size: 1.45rem;
		font-weight: 300;
		font-variant-numeric: lining-nums;
		color: var(--text);
	}

	.weekday {
		margin-top: 0.3rem;
		font-size: 0.7rem;
		white-space: nowrap;
		color: var(--text-dim);
	}

	.date {
		grid-area: date;
		font-size: 0.82rem;
		color: var(--text-muted);
	}

	.what {
		grid-area: what;
		min-width: 0;
	}

	.kind {
		font-size: 0.9rem;
		font-weight: 500;
		color: var(--text);
	}

	.tag {
		margin-left: 0.35rem;
		font-size: 0.72rem;
		font-weight: 400;
		color: var(--text-dim);
	}

	.note {
		max-width: 46ch;
		margin-top: 0.15rem;
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
		font-size: 0.84rem;
		color: var(--text-muted);
	}

	.code {
		font-size: 0.68rem;
		font-weight: 600;
		letter-spacing: 0.06em;
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
	}

	.amount {
		font-family: var(--font-mono);
		font-size: 0.95rem;
		font-variant-numeric: tabular-nums;
		white-space: nowrap;
		color: var(--text);
	}

	.compact .amount {
		font-size: 0.85rem;
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
		margin-top: -0.3rem;
	}

	.compact .actions {
		margin-top: 0;
	}

	.action {
		padding: 0.35rem 0.6rem;
		border: none;
		border-radius: 6px;
		background: transparent;
		font: inherit;
		font-size: 0.8rem;
		color: var(--text-muted);
		cursor: pointer;
	}

	a.action {
		display: inline-block;
		text-decoration: none;
		white-space: nowrap;
	}

	.action:hover {
		background: var(--surface-3);
		color: var(--text);
	}

	.action.danger:hover {
		color: var(--red);
	}

	/*
	 * Con ratón, las acciones esperan a que se pase por la fila o se llegue a ella
	 * con el teclado; la fila se ilumina para que se vea de quién son. Con el dedo
	 * no hay «pasar por encima», así que se quedan siempre.
	 */
	@media (hover: hover) and (pointer: fine) {
		.actions {
			opacity: 0;
			transition: opacity 0.15s ease;
		}

		.entry:hover,
		.entry:focus-within {
			background: var(--surface);
		}

		.entry:hover .actions,
		.entry:focus-within .actions {
			opacity: 1;
		}
	}

	@media (max-width: 760px) {
		.entry {
			grid-template-areas:
				'day what figures'
				'day where where'
				'day actions actions';
			row-gap: 0.35rem;
			column-gap: 0.75rem;
			padding: 0.85rem 0.25rem;
		}

		.entry.compact {
			grid-template-columns: minmax(0, 1fr) auto;
			grid-template-areas:
				'date figures'
				'actions actions';
			padding: 0.4rem 0.25rem 0.4rem 0.9rem;
		}

		.actions {
			justify-content: flex-start;
			margin: 0 0 0 -0.6rem;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.entry,
		.actions {
			transition: none;
		}
	}
</style>
