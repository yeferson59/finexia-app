<script lang="ts">
	/**
	 * Cuánto efectivo hay, y en qué monedas.
	 *
	 * El total va en la moneda de la pantalla porque es lo único que se puede
	 * sumar. Debajo, una barra partida por moneda con la cifra de cada una
	 * escrita bajo su tramo: el ancho dice cuánto pesa en el total y la cifra,
	 * cuánto hay en su propia moneda, que es la que coincide con el extracto del
	 * banco. Un saldo en pesos convertido a dólares es una cifra que el usuario
	 * no reconoce, así que la convertida va en pequeño.
	 */
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import type { CashCurrencyTotal, CashSummary } from '../cash';
	import { formatAnnualRate, type CashYield } from '../rates';

	interface Props {
		summary: CashSummary;
		/** Lo que rinde el efectivo; `null` si no hay ninguna cuenta con dinero. */
		yielding: CashYield | null;
	}

	let { summary, yielding }: Props = $props();

	const money = (amount: number, currency: string) =>
		privacy.money(formatCurrency(amount, currency));

	const percent = new Intl.NumberFormat('es-CO', { style: 'percent', maximumFractionDigits: 0 });

	/* Una sola moneda, y la misma de la pantalla: el reparto repetiría el total. */
	const showCurrencies = $derived(
		summary.byCurrency.length > 1 || summary.byCurrency[0]?.currency !== summary.currency
	);

	/* Sin tasa una moneda no está en el total, así que tampoco pesa en él. */
	const share = (group: CashCurrencyTotal) => (summary.total > 0 ? group.value / summary.total : 0);

	const unrated = (group: CashCurrencyTotal) => !group.value && group.balance !== 0;

	/*
	 * Cada tramo ocupa lo que pesa, con un mínimo para que su cifra quepa debajo:
	 * una moneda con el 2 % sigue siendo dinero que el usuario tiene que ver.
	 */
	const columns = $derived(
		summary.byCurrency
			.map((group) => `minmax(8.5rem, ${Math.max(share(group) * 100, 6).toFixed(2)}fr)`)
			.join(' ')
	);

	/* De la moneda que más pesa a la que menos, el ámbar se va apagando. */
	function tint(index: number): number {
		const count = summary.byCurrency.length;
		return count <= 1 ? 1 : 1 - (index / (count - 1)) * 0.6;
	}
</script>

<section class="summary" aria-labelledby="cash-total">
	<p class="total" id="cash-total">
		<span class="amount">{money(summary.total, summary.currency)}</span>
		<span class="caption">
			en efectivo, en {summary.funded}
			{summary.funded === 1 ? 'cuenta con saldo' : 'cuentas con saldo'}
		</span>
		{#if summary.interestThisMonth > 0}
			<!-- La parte de la cifra que es rendimiento, en el verde de los intereses. -->
			<span class="caption" style:color="var(--green)">
				+{money(summary.interestThisMonth, summary.currency)} en intereses este mes
			</span>
		{/if}
		<!-- Lo que rinde el dinero que rinde, y cuánto está parado: es el dato que
		     dice si vale la pena mover algo. -->
		{#if yielding}
			<span class="caption">
				{#if yielding.pct > 0}
					{formatAnnualRate(yielding.pct.toFixed(2))} de media
					{#if yielding.idle > 0}
						· {yielding.idle}
						{yielding.idle === 1 ? 'cuenta sin tasa' : 'cuentas sin tasa'}
					{/if}
				{:else}
					Ninguna cuenta con saldo tiene tasa
				{/if}
			</span>
		{/if}
	</p>

	{#if showCurrencies}
		<ul class="currencies" style:--columns={columns} aria-label="Efectivo por moneda">
			{#each summary.byCurrency as group, index (group.currency)}
				<li
					class:unrated={unrated(group)}
					style:--tint={tint(index)}
					style:--share={share(group)}
					style:--index={index}
				>
					<span class="bar" aria-hidden="true"></span>
					<span class="code">{group.currency}</span>
					<span class="native">{money(group.balance, group.currency)}</span>
					{#if unrated(group)}
						<span class="detail">Sin tasa a {summary.currency}, fuera del total</span>
					{:else}
						<span class="detail">
							{percent.format(share(group))} del total
							{#if group.currency !== summary.currency}
								<span class="converted">≈ {money(group.value, summary.currency)}</span>
							{/if}
						</span>
					{/if}
				</li>
			{/each}
		</ul>
	{/if}

	{#if summary.unconverted > 0}
		<p class="fx-note">
			{summary.unconverted}
			{summary.unconverted === 1 ? 'cuenta no tiene' : 'cuentas no tienen'} tasa de cambio a {summary.currency}
			y {summary.unconverted === 1 ? 'queda' : 'quedan'} fuera del total; {summary.unconverted === 1
				? 'aparece'
				: 'aparecen'} abajo en su propia moneda.
		</p>
	{/if}
</section>

<style>
	.summary {
		margin-bottom: 2.75rem;
	}

	.total {
		display: flex;
		flex-wrap: wrap;
		align-items: baseline;
		gap: 0.35rem 0.9rem;
		margin: 0;
	}

	/* Cifras proporcionales a este tamaño, como el patrimonio del resumen: las
	   tabulares dejan cada dígito con el ancho de un cero y el número se suelta. */
	.amount {
		font-family: var(--font-mono);
		font-size: clamp(2.25rem, 5.5vw, 3.25rem);
		font-weight: 500;
		line-height: 1;
		letter-spacing: -0.035em;
		color: var(--text);
		overflow-wrap: anywhere;
	}

	.caption {
		font-size: 0.95rem;
		font-weight: 300;
		color: var(--text-muted);
	}

	/*
	 * La barra y su leyenda son la misma rejilla: cada moneda es una columna con
	 * su tramo arriba y sus cifras debajo, así que la cifra no puede separarse
	 * del tramo que describe.
	 */
	.currencies {
		display: grid;
		grid-template-columns: var(--columns);
		column-gap: 3px;
		margin: 1.75rem 0 0;
		padding: 0;
		list-style: none;
	}

	.currencies li {
		display: flex;
		flex-direction: column;
		min-width: 0;
	}

	.bar {
		height: 12px;
		margin-bottom: 0.85rem;
		background: var(--amber);
		opacity: var(--tint);
	}

	.currencies li:first-child .bar {
		border-radius: 3px 0 0 3px;
	}

	.currencies li:last-child .bar {
		border-radius: 0 3px 3px 0;
	}

	/* Una moneda sin tasa no tiene peso que pintar: su tramo es un hueco con
	   borde, del ancho mínimo, para que se vea que existe pero no cuenta. */
	.unrated .bar {
		background: transparent;
		border: 1px dashed var(--text-dim);
		opacity: 1;
	}

	.code {
		font-family: var(--font-mono);
		font-size: 0.74rem;
		font-weight: 600;
		letter-spacing: 0.04em;
		color: var(--text-muted);
	}

	.native {
		margin-top: 0.2rem;
		padding-right: 0.75rem;
		font-family: var(--font-mono);
		font-size: 1.05rem;
		font-variant-numeric: tabular-nums;
		white-space: nowrap;
		color: var(--text);
	}

	.detail {
		display: flex;
		flex-direction: column;
		gap: 0.1rem;
		margin-top: 0.3rem;
		padding-right: 0.75rem;
		font-size: 0.78rem;
		color: var(--text-dim);
	}

	.converted {
		font-family: var(--font-mono);
		font-size: 0.74rem;
		font-variant-numeric: tabular-nums;
		white-space: nowrap;
	}

	/* El único movimiento de la pantalla: la barra se dibuja una vez al entrar,
	   tramo a tramo, y lleva la vista al reparto antes que a las cuentas. */
	@media (prefers-reduced-motion: no-preference) {
		.bar {
			transform-origin: left center;
			animation: draw 0.5s cubic-bezier(0.22, 1, 0.36, 1) both;
			animation-delay: calc(var(--index) * 90ms);
		}
	}

	@keyframes draw {
		from {
			transform: scaleX(0);
		}
		to {
			transform: scaleX(1);
		}
	}

	.fx-note {
		max-width: 62ch;
		margin: 1.25rem 0 0;
		padding-left: 0.75rem;
		border-left: 2px solid rgba(212, 145, 42, 0.45);
		font-size: 0.8rem;
		line-height: 1.5;
		color: var(--text-muted);
	}

	/*
	 * En estrecho las columnas no caben: cada moneda es una fila, con su tramo
	 * medido contra el ancho entero en vez de repartido.
	 */
	@media (max-width: 640px) {
		.currencies {
			grid-template-columns: minmax(0, 1fr);
			row-gap: 1.1rem;
		}

		.currencies li {
			display: grid;
			grid-template-columns: auto minmax(0, 1fr);
			grid-template-areas:
				'code native'
				'bar bar'
				'detail detail';
			align-items: baseline;
			column-gap: 0.6rem;
		}

		.code {
			grid-area: code;
		}

		.native {
			grid-area: native;
			margin: 0;
		}

		.bar {
			grid-area: bar;
			width: max(calc(var(--share) * 100%), 6px);
			height: 8px;
			margin: 0.45rem 0 0;
		}

		.unrated .bar {
			width: 2.5rem;
		}

		.currencies li:first-child .bar,
		.currencies li:last-child .bar {
			border-radius: 2px;
		}

		.detail {
			grid-area: detail;
			flex-direction: row;
			flex-wrap: wrap;
			gap: 0 0.6rem;
		}
	}
</style>
