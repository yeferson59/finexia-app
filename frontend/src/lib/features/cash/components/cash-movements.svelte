<script lang="ts">
	/**
	 * Los movimientos de efectivo, como un extracto: por meses, del más reciente
	 * al más antiguo.
	 *
	 * El mes va de cabecera y cada movimiento solo lleva su día, que es como se
	 * lee el extracto de un banco y lo que deja sitio a la nota.
	 *
	 * El importe lleva signo pero no color de ganancia: pintar un depósito de
	 * verde diría justo lo que esta pantalla explica que no es. Solo los intereses
	 * van en verde, porque son lo único de aquí que es rendimiento.
	 *
	 * Los intereses que abona sola la tasa de una cuenta van en una fila por saldo
	 * y mes, que se despliega desde su nombre: una cuenta que rinde abona cada
	 * día, y treinta filas iguales taparían los depósitos y retiros.
	 *
	 * Editar y borrar aparecen al pasar por la fila o al llegar con el teclado:
	 * dos botones por fila en quince filas eran treinta, compitiendo con los
	 * importes. En una pantalla táctil, donde no hay «pasar por encima», están
	 * siempre.
	 */
	import Pagination from '$lib/ui/pagination.svelte';
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import { formatCalendarDate } from '$lib/shared/format/date';
	import type { CashMovement } from '../cash';
	import {
		groupAutomaticInterest,
		groupCashLedgerByMonth,
		type CashInterestGroup
	} from '../interest';
	import CashDeleteConfirm from './cash-delete-confirm.svelte';
	import CashLedgerEntry from './cash-ledger-entry.svelte';

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

	/* Se agrupa antes de paginar: un mes de abonos es una fila, no dos páginas. */
	const rows = $derived(groupAutomaticInterest(movements));
	const months = $derived(
		groupCashLedgerByMonth(rows.slice((page - 1) * PER_PAGE, page * PER_PAGE))
	);

	let deleting = $state<CashMovement | null>(null);

	/* Los grupos de abonos desplegados, por su clave. */
	let open = $state<Record<string, boolean>>({});

	const day = (iso: string) => formatCalendarDate(iso, { day: 'numeric' });

	const fullDate = (iso: string) =>
		formatCalendarDate(iso, { day: 'numeric', month: 'long', year: 'numeric' });

	/* Los días que cubre un grupo, dichos para quien no ve la columna del día. */
	const span = (group: CashInterestGroup) =>
		group.since === group.date
			? fullDate(group.date)
			: `del ${day(group.since)} al ${fullDate(group.date)}`;

	const creditsId = (group: CashInterestGroup) => `cash-credits-${group.key.replaceAll(':', '-')}`;

	const income = (amount: number, currency: string) =>
		privacy.money(`+${formatCurrency(amount, currency)}`);

	const count = (n: number) => (n === 1 ? '1 movimiento' : `${n} movimientos`);
</script>

<div class="ledger">
	{#each months as month (month.key)}
		<section class="month" aria-labelledby="cash-month-{month.key}">
			<header class="month-head">
				<h3 id="cash-month-{month.key}">{month.label}</h3>
				<p class="month-count">
					{count(
						month.rows.reduce(
							(sum, row) => sum + (row.type === 'interest' ? row.movements.length : 1),
							0
						)
					)}
				</p>
			</header>

			<ul class="entries">
				{#each month.rows as row (row.key)}
					{#if row.type === 'interest'}
						{@const expanded = open[row.key] === true}
						<li class="entry group" class:expanded>
							<p class="day">
								<span class="day-number" aria-hidden="true">{day(row.date)}</span>
								<span class="weekday" aria-hidden="true">desde el {day(row.since)}</span>
								<span class="sr-only">{span(row)}</span>
							</p>

							<div class="what">
								<button
									type="button"
									class="toggle"
									aria-expanded={expanded}
									aria-controls={creditsId(row)}
									onclick={() => (open[row.key] = !expanded)}
								>
									<svg viewBox="0 0 16 16" aria-hidden="true"><path d="M6 4l4 4-4 4" /></svg>
									Intereses de la cuenta
								</button>
								<p class="note">{row.movements.length} abonos automáticos</p>
							</div>

							<p class="where">
								<span class="source">{row.sourceName || 'Sin plataforma'}</span>
								<span class="code">{row.currency}</span>
								{#if showPortfolio}
									<span class="portfolio">{row.portfolioName}</span>
								{/if}
							</p>

							<p class="figures">
								<span class="amount income">{income(row.amount, row.currency)}</span>
							</p>
						</li>

						<li class="credits" id={creditsId(row)} hidden={!expanded}>
							<ul>
								{#each row.movements as movement (movement.id)}
									<CashLedgerEntry
										{movement}
										{showPortfolio}
										compact
										onEdit={() => onEdit(movement)}
										onDelete={() => (deleting = movement)}
									/>
								{/each}
							</ul>
						</li>
					{:else}
						<CashLedgerEntry
							movement={row.movement}
							{showPortfolio}
							onEdit={() => onEdit(row.movement)}
							onDelete={() => (deleting = row.movement)}
						/>
					{/if}
				{/each}
			</ul>
		</section>
	{/each}
</div>

<div class="pager">
	<Pagination bind:page total={rows.length} perPage={PER_PAGE} label="movimientos" />
	{#if total > movements.length}
		<p class="hint">Se muestran los {movements.length} movimientos más recientes de {total}.</p>
	{/if}
</div>

<CashDeleteConfirm movement={deleting} {showPortfolio} onClose={() => (deleting = null)} />

<style>
	.ledger {
		display: grid;
		gap: 2rem;
	}

	/* El mes, como la cabecera de una hoja del extracto: su nombre y cuántos
	   movimientos lleva, sobre un filete más marcado que el de las filas. */
	.month-head {
		display: flex;
		align-items: baseline;
		justify-content: space-between;
		gap: 1rem;
		padding: 0 0.75rem 0.6rem;
		border-bottom: 1px solid var(--border-strong);
	}

	h3 {
		margin: 0;
		font-family: var(--font-display);
		font-size: 1.1rem;
		font-weight: 400;
		color: var(--text);
	}

	.month-count {
		margin: 0;
		font-size: 0.76rem;
		color: var(--text-dim);
	}

	.entries,
	.credits ul {
		margin: 0;
		padding: 0;
		list-style: none;
	}

	.entries > :global(li + li) {
		border-top: 1px solid var(--border);
	}

	/* La fila de un grupo de abonos: la rejilla y las piezas de
	   `cash-ledger-entry`, donde está explicada. */
	.entry {
		display: grid;
		grid-template-columns: var(--ledger-columns);
		grid-template-areas: 'day what where figures actions';
		align-items: start;
		column-gap: 1.25rem;
		padding: 0.85rem 0.75rem;
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

	/* «desde el 1» no cabe en la columna del día: baja a dos líneas en vez de
	   montarse sobre el nombre. */
	.weekday {
		margin-top: 0.3rem;
		font-size: 0.7rem;
		line-height: 1.2;
		color: var(--text-dim);
	}

	.what {
		grid-area: what;
		min-width: 0;
	}

	.toggle {
		display: inline-flex;
		align-items: center;
		gap: 0.35rem;
		margin: -0.2rem 0 0 -0.4rem;
		padding: 0.2rem 0.4rem;
		border: none;
		border-radius: 6px;
		background: transparent;
		font: inherit;
		font-size: 0.9rem;
		font-weight: 500;
		color: var(--text);
		cursor: pointer;
	}

	.toggle:hover {
		background: var(--surface-2);
	}

	.toggle svg {
		width: 0.8rem;
		height: 0.8rem;
		fill: none;
		stroke: currentColor;
		stroke-width: 1.75;
		stroke-linecap: round;
		stroke-linejoin: round;
		transition: transform 0.2s ease;
	}

	.expanded .toggle svg {
		transform: rotate(90deg);
	}

	.note {
		margin-top: 0.15rem;
		font-size: 0.8rem;
		color: var(--text-dim);
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
		text-align: right;
	}

	.amount {
		font-family: var(--font-mono);
		font-size: 0.95rem;
		font-variant-numeric: tabular-nums;
		white-space: nowrap;
	}

	.income {
		color: var(--green);
	}

	/* Los abonos de un grupo, sangrados bajo su fila con un filete a la izquierda
	   que dice de quién son. */
	.credits {
		padding: 0 0 0.6rem;
	}

	.credits ul {
		margin-left: calc(0.75rem + 3rem + 0.6rem);
		border-left: 1px solid var(--border-strong);
	}

	.pager {
		margin-top: 0.5rem;
	}

	.pager .hint {
		padding: 0 0.75rem 1rem;
	}

	.ledger {
		--ledger-columns: 3rem minmax(0, 1.35fr) minmax(0, 1fr) 10rem 8.5rem;
	}

	@media (max-width: 760px) {
		.ledger {
			--ledger-columns: 2.5rem minmax(0, 1fr) auto;
		}

		.entry {
			grid-template-areas:
				'day what figures'
				'day where where';
			row-gap: 0.35rem;
			column-gap: 0.75rem;
			padding: 0.85rem 0.25rem;
		}

		.month-head {
			padding: 0 0.25rem 0.6rem;
		}

		.credits ul {
			margin-left: 1.25rem;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.toggle svg {
			transition: none;
		}
	}
</style>
