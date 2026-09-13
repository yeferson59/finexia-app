<script lang="ts">
	/**
	 * Cuánto efectivo hay, y en qué monedas.
	 *
	 * El total va en la moneda de la pantalla porque es lo único que se puede
	 * sumar; debajo, cada moneda con su propio importe, que es el que coincide
	 * con el extracto del banco. Un saldo en pesos convertido a dólares es una
	 * cifra que el usuario no reconoce.
	 */
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import type { CashSummary } from '../cash';

	let { summary }: { summary: CashSummary } = $props();

	const money = (amount: number, currency: string) =>
		privacy.money(formatCurrency(amount, currency));

	/* Una sola moneda, y la misma de la pantalla: el reparto repetiría el total. */
	const showCurrencies = $derived(
		summary.byCurrency.length > 1 || summary.byCurrency[0]?.currency !== summary.currency
	);
</script>

<section class="summary" aria-labelledby="cash-total">
	<p class="total" id="cash-total">
		<span class="amount">{money(summary.total, summary.currency)}</span>
		<span class="caption">
			en efectivo, en {summary.funded}
			{summary.funded === 1 ? 'cuenta con saldo' : 'cuentas con saldo'}
		</span>
	</p>

	{#if showCurrencies}
		<ul class="currencies" aria-label="Efectivo por moneda">
			{#each summary.byCurrency as group (group.currency)}
				<li>
					<span class="code">{group.currency}</span>
					<span class="native">{money(group.balance, group.currency)}</span>
					{#if group.currency !== summary.currency && group.value > 0}
						<span class="converted">≈ {money(group.value, summary.currency)}</span>
					{/if}
				</li>
			{/each}
		</ul>
	{/if}

	{#if summary.unconverted > 0}
		<p class="fx-note">
			{summary.unconverted}
			{summary.unconverted === 1 ? 'saldo no tiene' : 'saldos no tienen'} tasa de cambio a {summary.currency}
			y {summary.unconverted === 1 ? 'queda' : 'quedan'} fuera del total; {summary.unconverted === 1
				? 'aparece'
				: 'aparecen'} abajo en su propia moneda.
		</p>
	{/if}

	<!-- La regla que distingue esta pantalla de un simple saldo: no todo lo que
	     sube el efectivo es rendimiento. -->
	<p class="how">
		Los depósitos y retiros no cuentan como ganancia ni como pérdida. Los intereses sí: suman al
		saldo y a tu rentabilidad.
	</p>
</section>

<style>
	.summary {
		display: flex;
		flex-direction: column;
		gap: 0.85rem;
		margin-bottom: 1.5rem;
	}

	.total {
		display: flex;
		flex-wrap: wrap;
		align-items: baseline;
		gap: 0.35rem 0.75rem;
		margin: 0;
	}

	.amount {
		font-family: var(--font-display);
		font-size: clamp(1.9rem, 4vw, 2.6rem);
		font-weight: 500;
		letter-spacing: -0.02em;
		color: var(--text);
		font-variant-numeric: tabular-nums;
	}

	.caption {
		font-size: 0.9rem;
		color: var(--text-muted);
	}

	.currencies {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem 1.5rem;
		margin: 0;
		padding: 0;
		list-style: none;
	}

	.currencies li {
		display: flex;
		align-items: baseline;
		gap: 0.5rem;
		font-size: 0.88rem;
	}

	.code {
		font-family: var(--font-mono);
		font-size: 0.72rem;
		letter-spacing: 0.06em;
		color: var(--amber);
	}

	.native {
		color: var(--text);
		font-variant-numeric: tabular-nums;
	}

	.converted {
		color: var(--text-dim);
		font-variant-numeric: tabular-nums;
	}

	.fx-note,
	.how {
		max-width: 62ch;
		margin: 0;
		font-size: 0.83rem;
		line-height: 1.55;
		color: var(--text-muted);
	}

	.fx-note {
		padding-left: 0.75rem;
		border-left: 2px solid var(--amber);
	}
</style>
