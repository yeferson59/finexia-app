<script lang="ts">
	import { resolve } from '$app/paths';
	import CardHeader from '$lib/ui/card-header.svelte';
	import Stat from '$lib/ui/stat.svelte';
	import CurrencyToggle from './currency-toggle.svelte';
	import ExchangeRateNote from './exchange-rate-note.svelte';
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import { FALLBACK_CURRENCY, partitionByCurrency } from '$lib/shared/currency';
	import type { ExchangeRate } from '$lib/api/types';

	interface PortfolioSummary {
		id: string;
		name: string;
		baseCurrency?: string;
		displayCurrency?: string;
		totalMarketValue: string;
		totalCostBase: string;
		totalGainLoss: string;
		totalGainLossPct: string;
		totalPositions: number;
	}

	const {
		summaries = [],
		currency = FALLBACK_CURRENCY,
		displayRate = null
	}: {
		summaries: PortfolioSummary[];
		currency?: string;
		/** Tasa con la que están convertidas estas cifras; `null` si no hay conversión que enseñar. */
		displayRate?: ExchangeRate | null;
	} = $props();

	function fmtMoney(value: number): string {
		return privacy.money(formatCurrency(value, currency));
	}

	// Un portafolio que el backend no pudo convertir viene en su propia moneda.
	// Sumarlo aquí daría un patrimonio que no está en ninguna: se deja fuera y
	// se dice cuántos faltan, que es un total honesto en vez de uno inventado.
	const split = $derived(partitionByCurrency(summaries, currency));
	const counted = $derived(split.converted);
	const excluded = $derived(split.unconverted.length);

	const netWorth = $derived(
		counted.reduce((acc, s) => acc + parseFloat(s.totalMarketValue || '0'), 0)
	);
	const totalGainLoss = $derived(
		counted.reduce((acc, s) => acc + parseFloat(s.totalGainLoss || '0'), 0)
	);
	const totalCostBase = $derived(
		counted.reduce((acc, s) => acc + parseFloat(s.totalCostBase || '0'), 0)
	);
	const totalPositions = $derived(counted.reduce((acc, s) => acc + (s.totalPositions ?? 0), 0));
	const gainLossPct = $derived(totalCostBase > 0 ? (totalGainLoss / totalCostBase) * 100 : 0);
	const isIncreasing = $derived(totalGainLoss >= 0);
</script>

<div class="net-worth-card">
	<CardHeader eyebrow="Patrimonio total" title="Patrimonio Neto">
		{#snippet action()}
			<!--
				Apilados y alineados a la derecha: `.card-header-action` es un div
				normal, así que sin este contenedor el selector se estiraría al ancho
				que ocupe la línea de la tasa.
			-->
			<div class="currency-controls">
				<CurrencyToggle {currency} />
				<ExchangeRateNote rate={displayRate} />
			</div>
		{/snippet}
	</CardHeader>

	<div class="net-worth-content">
		<div class="main-metric">
			<h1 class="amount">
				{fmtMoney(netWorth)}
			</h1>
			{#if counted.length > 0}
				<p class="amount-delta" class:positive={isIncreasing} class:negative={!isIncreasing}>
					{isIncreasing ? '+' : '−'}{fmtMoney(Math.abs(totalGainLoss))}
					· {new Intl.NumberFormat('es-CO', {
						minimumFractionDigits: 2,
						maximumFractionDigits: 2
					}).format(Math.abs(gainLossPct))}% total
				</p>
			{:else if summaries.length > 0}
				<p class="amount-delta neutral">Ningún portafolio se puede convertir a {currency}</p>
			{:else}
				<p class="amount-delta neutral">Sin portafolios registrados</p>
			{/if}
			{#if excluded > 0 && counted.length > 0}
				<p class="fx-note">
					{excluded === 1
						? 'Falta 1 portafolio: no hay tasa para pasarlo a'
						: `Faltan ${excluded} portafolios: no hay tasa para pasarlos a`}
					{currency}.
				</p>
			{/if}
		</div>

		<div class="metric-stats">
			<Stat label="Portafolios" value={String(counted.length)} />
			<Stat label="Posiciones" tone="highlight" value={String(totalPositions)} />
			<Stat
				label="Ganancia"
				tone={isIncreasing ? 'positive' : 'negative'}
				value="{isIncreasing ? '+' : ''}{new Intl.NumberFormat('es-CO', {
					minimumFractionDigits: 2,
					maximumFractionDigits: 2
				}).format(gainLossPct)}%"
			/>
		</div>
	</div>

	<div class="card-footer">
		<a href={resolve('/dashboard/portfolios')} class="action-button secondary">Ver portafolios</a>
	</div>
</div>

<style>
	.net-worth-card {
		position: relative;
		overflow: hidden;
		background: var(--surface);
		border: 1px solid var(--border-strong);
		border-radius: 14px;
		padding: 2rem;
		backdrop-filter: blur(10px);
	}

	/* Warm amber wash anchoring the hero figure */
	.net-worth-card::before {
		content: '';
		position: absolute;
		inset: 0;
		background: radial-gradient(
			ellipse 60% 90% at 0% 0%,
			rgba(212, 145, 42, 0.07),
			transparent 55%
		);
		pointer-events: none;
	}

	.net-worth-card > * {
		position: relative;
	}

	.currency-controls {
		display: flex;
		flex-direction: column;
		align-items: flex-end;
	}

	.net-worth-content {
		display: grid;
		grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
		gap: 3rem;
		margin-bottom: 2rem;
		align-items: end;
	}

	.main-metric {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
		min-width: 0;
	}

	.amount {
		font-family: var(--font-mono);
		font-size: clamp(1.65rem, 7vw, 3rem);
		overflow-wrap: anywhere;
		font-weight: 600;
		color: var(--text);
		margin: 0;
		line-height: 1;
		letter-spacing: -0.03em;
		font-variant-numeric: tabular-nums;
	}

	.amount-delta {
		font-size: 0.85rem;
		font-weight: 400;
		margin: 0;
	}

	.amount-delta.positive {
		color: var(--green);
	}

	.amount-delta.negative {
		color: var(--red);
	}

	.amount-delta.neutral {
		color: rgba(236, 234, 229, 0.5);
	}

	/* Mismo aviso que la tarjeta de portafolio, para que "falta una tasa" se
	   reconozca igual en el panel y en el listado. */
	.fx-note {
		margin: 0.6rem 0 0;
		padding: 0.5rem 0.7rem;
		border: 1px solid rgba(212, 145, 42, 0.3);
		border-radius: 8px;
		background: rgba(212, 145, 42, 0.08);
		color: rgba(236, 234, 229, 0.75);
		font-size: 0.78rem;
		line-height: 1.4;
	}

	.metric-stats {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 1.5rem;
		min-width: 0;
	}

	.card-footer {
		display: flex;
		gap: 0.75rem;
	}

	.action-button {
		padding: 0.75rem 1.5rem;
		border-radius: 6px;
		font-weight: 600;
		font-size: 0.85rem;
		cursor: pointer;
		text-decoration: none;
		display: inline-block;
		transition:
			background 0.2s ease,
			border-color 0.2s ease,
			color 0.2s ease;
		font-family: var(--font-body);
	}

	.action-button.secondary {
		background: transparent;
		border: 1px solid var(--border-strong);
		color: var(--text);
	}

	.action-button.secondary:hover {
		background: rgba(212, 145, 42, 0.06);
		border-color: rgba(212, 145, 42, 0.4);
		color: var(--amber-light);
	}

	@media (max-width: 1024px) {
		.net-worth-content {
			grid-template-columns: 1fr;
			gap: 2rem;
		}

		.metric-stats {
			grid-template-columns: repeat(3, minmax(0, 1fr));
		}
	}

	@media (max-width: 768px) {
		.net-worth-card {
			padding: 1.5rem;
		}

		.net-worth-content {
			gap: 1.5rem;
		}

		.action-button {
			padding: 0.75rem 1rem;
			font-size: 0.85rem;
		}
	}

	@media (max-width: 480px) {
		.metric-stats {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}
	}
</style>
