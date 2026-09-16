<script lang="ts">
	/**
	 * Cuánto efectivo hay, en qué monedas y cuánto está rindiendo.
	 *
	 * A la izquierda, las cifras: el total en la moneda de la pantalla —lo único
	 * que se puede sumar—, lo que abonaron los intereses este mes y el reparto por
	 * moneda con el importe de cada una en la suya, que es el que coincide con el
	 * extracto. A la derecha, el mapa de rendimiento, que es lo que esta pantalla
	 * tiene que decir y ninguna otra dice: dónde trabaja el dinero y dónde está
	 * parado.
	 */
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import type { CashCurrencyTotal, CashSummary } from '../cash';
	import { formatAnnualRate, type CashYield } from '../rates';
	import type { CashYieldBlock, CashYieldMap } from '../yield';
	import CashYieldChart from './cash-yield-chart.svelte';

	interface Props {
		summary: CashSummary;
		/** Lo que rinde el efectivo; `null` si no hay ninguna cuenta con dinero. */
		yielding: CashYield | null;
		map: CashYieldMap;
		onSelect: (block: CashYieldBlock) => void;
	}

	let { summary, yielding, map, onSelect }: Props = $props();

	const money = (amount: number, currency: string) =>
		privacy.money(formatCurrency(amount, currency));

	const percent = new Intl.NumberFormat('es-CO', { style: 'percent', maximumFractionDigits: 0 });

	/* Una sola moneda, y la misma de la pantalla: el reparto repetiría el total. */
	const showCurrencies = $derived(
		summary.byCurrency.length > 1 || summary.byCurrency[0]?.currency !== summary.currency
	);

	const share = (group: CashCurrencyTotal) => (summary.total > 0 ? group.value / summary.total : 0);

	/* Sin tasa de cambio una moneda no está en el total, así que tampoco pesa en él. */
	const unrated = (group: CashCurrencyTotal) => !group.value && group.balance !== 0;

	/* La frase que encabeza el mapa: la media y, si lo hay, lo que no rinde. */
	const headline = $derived.by(() => {
		if (!yielding) return '';
		if (yielding.pct <= 0)
			return 'Ninguna cuenta con saldo tiene tasa: todo tu efectivo está parado.';

		const average = `Rinde ${formatAnnualRate(yielding.pct.toFixed(2))} de media`;
		if (yielding.idle === 0) return `${average}, y no hay dinero parado.`;
		return `${average}, con ${yielding.idle} ${yielding.idle === 1 ? 'cuenta parada' : 'cuentas paradas'}.`;
	});
</script>

<section class="summary" aria-labelledby="cash-total">
	<div class="figures">
		<p class="total" id="cash-total">
			<span class="amount">{money(summary.total, summary.currency)}</span>
			<span class="caption">
				en efectivo, entre {summary.funded}
				{summary.funded === 1 ? 'cuenta con saldo' : 'cuentas con saldo'}
			</span>
		</p>

		{#if summary.interestThisMonth > 0}
			<p class="earned">
				<span class="earned-amount">+{money(summary.interestThisMonth, summary.currency)}</span>
				en intereses este mes
			</p>
		{/if}

		{#if showCurrencies}
			<dl class="currencies" aria-label="Efectivo por moneda">
				{#each summary.byCurrency as group (group.currency)}
					<div class="currency" class:unrated={unrated(group)}>
						<dt>{group.currency}</dt>
						<dd class="native">{money(group.balance, group.currency)}</dd>
						<dd class="weight">
							{#if unrated(group)}
								fuera del total
							{:else}
								{percent.format(share(group))}
							{/if}
						</dd>
					</div>
				{/each}
			</dl>
		{/if}

		{#if summary.unconverted > 0}
			<p class="feedback warning fx-note">
				{summary.unconverted}
				{summary.unconverted === 1 ? 'cuenta no tiene' : 'cuentas no tienen'} tasa de cambio a {summary.currency}:
				{summary.unconverted === 1 ? 'queda' : 'quedan'} fuera del total y del mapa, y {summary.unconverted ===
				1
					? 'aparece'
					: 'aparecen'} abajo en su propia moneda.
			</p>
		{/if}
	</div>

	{#if map.blocks.length > 0}
		<figure class="yield" aria-labelledby="cash-yield-headline" aria-describedby="cash-yield-how">
			<figcaption>
				<p class="headline" id="cash-yield-headline">{headline}</p>
				<p class="how" id="cash-yield-how">
					Cada bloque es una cuenta: el ancho es cuánto dinero guarda y la altura, a qué tasa rinde.
				</p>
			</figcaption>
			<CashYieldChart {map} average={yielding?.pct ?? 0} {onSelect} />
		</figure>
	{/if}
</section>

<style>
	.summary {
		display: grid;
		grid-template-columns: minmax(16rem, 22rem) minmax(0, 1fr);
		align-items: start;
		gap: 2.5rem 4rem;
		margin-bottom: 3.5rem;
	}

	.figures {
		display: flex;
		flex-direction: column;
		min-width: 0;
	}

	.total {
		display: flex;
		flex-direction: column;
		gap: 0.55rem;
		margin: 0;
	}

	/* La letra de las cifras del panel, como el patrimonio del resumen, a un
	   tamaño que deja sitio al mapa en la misma fila. */
	.amount {
		font-family: var(--font-mono);
		font-size: clamp(2.1rem, 4.2vw, 2.9rem);
		font-weight: 500;
		line-height: 1;
		letter-spacing: -0.04em;
		color: var(--text);
		overflow-wrap: anywhere;
	}

	.caption {
		font-size: 0.9rem;
		font-weight: 300;
		color: var(--text-muted);
	}

	.earned {
		margin: 0.35rem 0 0;
		font-size: 0.9rem;
		font-weight: 300;
		color: var(--text-muted);
	}

	.earned-amount {
		font-family: var(--font-mono);
		font-weight: 400;
		color: var(--green);
	}

	/*
	 * El reparto por moneda es una tabla corta: el código, lo que hay en su
	 * propia moneda y lo que pesa en el total, cada columna en su vertical para
	 * que las cifras se comparen sin buscarlas.
	 */
	.currencies {
		display: grid;
		grid-template-columns: auto minmax(0, 1fr) auto;
		column-gap: 1rem;
		margin: 1.75rem 0 0;
		border-top: 1px solid var(--border);
	}

	.currency {
		display: grid;
		grid-template-columns: subgrid;
		grid-column: 1 / -1;
		align-items: baseline;
		padding: 0.55rem 0;
		border-bottom: 1px solid var(--border);
	}

	dt {
		font-size: 0.74rem;
		font-weight: 600;
		letter-spacing: 0.05em;
		color: var(--text-muted);
	}

	dd {
		margin: 0;
	}

	.native {
		font-family: var(--font-mono);
		font-size: 0.92rem;
		font-variant-numeric: tabular-nums;
		white-space: nowrap;
		color: var(--text);
	}

	.weight {
		justify-self: end;
		font-family: var(--font-mono);
		font-size: 0.78rem;
		font-variant-numeric: tabular-nums;
		color: var(--text-dim);
	}

	.unrated .native {
		color: var(--text-muted);
	}

	.unrated .weight {
		font-family: var(--font-body);
	}

	.fx-note {
		margin: 1.1rem 0 0;
	}

	.yield {
		min-width: 0;
		margin: 0;
	}

	figcaption {
		margin-bottom: 1.6rem;
	}

	.headline {
		margin: 0;
		font-family: var(--font-display);
		font-size: 1.3rem;
		font-weight: 300;
		line-height: 1.3;
		letter-spacing: -0.01em;
		color: var(--text);
		text-wrap: balance;
	}

	.how {
		max-width: 60ch;
		margin: 0.35rem 0 0;
		font-size: 0.8rem;
		line-height: 1.5;
		color: var(--text-dim);
	}

	@media (max-width: 960px) {
		.summary {
			grid-template-columns: minmax(0, 1fr);
			gap: 2.5rem;
		}

		.currencies {
			max-width: 26rem;
		}
	}
</style>
