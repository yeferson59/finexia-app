<script lang="ts">
	import CardHeader from '$lib/ui/card-header.svelte';
	import Stat from '$lib/ui/stat.svelte';
	import { privacy } from '$lib/stores/privacy.svelte';

	interface GrowthDataPoint {
		date: string;
		totalValue: string;
		totalCostBase: string;
		gainLoss: string;
		gainLossPct: string;
	}

	interface GrowthSummary {
		initialValue: string;
		currentValue: string;
		totalGrowthPct: string;
	}

	const {
		data = [],
		summary = { initialValue: '0', currentValue: '0', totalGrowthPct: '0' }
	}: { data: GrowthDataPoint[]; summary: GrowthSummary } = $props();

	type Period = '1M' | '3M' | '6M' | '1Y' | 'Todo';
	let selectedPeriod = $state<Period>('Todo');

	const periods: Period[] = ['1M', '3M', '6M', '1Y', 'Todo'];

	function filterByPeriod(points: GrowthDataPoint[], period: Period): GrowthDataPoint[] {
		if (period === 'Todo') return points;
		const monthsMap: Record<Exclude<Period, 'Todo'>, number> = {
			'1M': 1,
			'3M': 3,
			'6M': 6,
			'1Y': 12
		};
		const now = new Date();
		const cutoff = new Date(now.getFullYear(), now.getMonth() - monthsMap[period], now.getDate());
		return points.filter((d) => new Date(d.date) >= cutoff);
	}

	const filteredData = $derived(filterByPeriod(data, selectedPeriod));

	// SVG dimensions
	const padL = 52;
	const padR = 20;
	const padT = 20;
	const padB = 32;
	const svgW = 600;
	const svgH = 240;
	const plotW = svgW - padL - padR;
	const plotH = svgH - padT - padB;

	const values = $derived(
		filteredData.map((d) => ({
			mv: parseFloat(d.totalValue || '0'),
			cb: parseFloat(d.totalCostBase || '0'),
			date: d.date
		}))
	);

	const allNums = $derived(values.flatMap((v) => [v.mv, v.cb]));
	const yMin = $derived(allNums.length > 0 ? Math.min(...allNums) * 0.97 : 0);
	const yMax = $derived(allNums.length > 0 ? Math.max(...allNums) * 1.03 : 1);
	const yRange = $derived(yMax - yMin || 1);

	function toX(i: number, n: number): number {
		if (n <= 1) return padL + plotW / 2;
		return padL + (i / (n - 1)) * plotW;
	}

	function toY(v: number): number {
		return padT + plotH - ((v - yMin) / yRange) * plotH;
	}

	const mvPoints = $derived(
		values.map((v, i) => `${toX(i, values.length)},${toY(v.mv)}`).join(' ')
	);
	const cbPoints = $derived(
		values.map((v, i) => `${toX(i, values.length)},${toY(v.cb)}`).join(' ')
	);
	const mvFill = $derived(
		values.length < 2
			? ''
			: `${mvPoints} ${toX(values.length - 1, values.length)},${padT + plotH} ${toX(0, values.length)},${padT + plotH}`
	);

	// Y-axis ticks
	const yTicks = $derived(
		Array.from({ length: 5 }, (_, i) => {
			const frac = i / 4;
			const val = yMax - frac * yRange;
			const y = padT + frac * plotH;
			return { y, label: fmtAbbrev(val) };
		})
	);

	// X-axis labels (max 6 evenly spaced)
	const xLabels = $derived(
		values.length === 0
			? []
			: values
					.map((v, i) => ({ i, date: v.date }))
					.filter(({ i }) => {
						const n = values.length;
						if (n <= 6) return true;
						const step = Math.ceil(n / 6);
						return i % step === 0 || i === n - 1;
					})
	);

	// Summary metrics
	const currentVal = $derived(parseFloat(summary.currentValue || '0'));
	const initialVal = $derived(parseFloat(summary.initialValue || '0'));
	const totalGrowthPct = $derived(parseFloat(summary.totalGrowthPct || '0'));
	const absoluteGain = $derived(currentVal - initialVal);
	const isPositive = $derived(absoluteGain >= 0);

	function fmt(v: number): string {
		return new Intl.NumberFormat('es-CO', {
			minimumFractionDigits: 2,
			maximumFractionDigits: 2
		}).format(v);
	}

	function fmtMoney(v: number): string {
		return privacy.money('$' + fmt(v));
	}

	function fmtAbbrev(v: number): string {
		const abs = Math.abs(v);
		if (abs >= 1_000_000) return privacy.money(`$${(v / 1_000_000).toFixed(1)}M`);
		if (abs >= 1_000) return privacy.money(`$${(v / 1_000).toFixed(0)}k`);
		return privacy.money(`$${v.toFixed(0)}`);
	}

	function fmtDate(iso: string): string {
		const d = new Date(iso + 'T00:00:00');
		return d.toLocaleDateString('es-CO', { day: '2-digit', month: 'short' });
	}
</script>

<div class="growth-card">
	<div class="card-top">
		<CardHeader eyebrow="Portafolio" title="Crecimiento del portafolio" divider={false} />
		<div class="period-tabs" role="tablist" aria-label="Período">
			{#each periods as p (p)}
				<button
					role="tab"
					aria-selected={selectedPeriod === p}
					class="period-btn"
					class:active={selectedPeriod === p}
					onclick={() => (selectedPeriod = p)}>{p}</button
				>
			{/each}
		</div>
	</div>

	<div class="divider"></div>

	<div class="metrics-row">
		<Stat
			label="Total ganancia"
			tone={isPositive ? 'positive' : 'negative'}
			value="{isPositive ? '+' : '−'}{fmtMoney(Math.abs(absoluteGain))}"
		/>
		<Stat
			label="Desde creación"
			tone={isPositive ? 'positive' : 'negative'}
			value="{isPositive ? '+' : ''}{fmt(totalGrowthPct)}%"
		/>
		<Stat label="Valor actual" tone="highlight" value={fmtMoney(currentVal)} />
	</div>

	{#if filteredData.length < 2}
		<div class="empty-chart">
			<p>
				El gráfico se mostrará aquí a medida que el sistema registre capturas diarias del
				portafolio.
			</p>
		</div>
	{:else}
		<svg class="chart" viewBox="0 0 {svgW} {svgH}" preserveAspectRatio="xMidYMid meet">
			<defs>
				<linearGradient id="growthGradient" x1="0%" y1="0%" x2="0%" y2="100%">
					<stop offset="0%" style="stop-color: var(--amber); stop-opacity: 0.18" />
					<stop offset="100%" style="stop-color: var(--amber); stop-opacity: 0" />
				</linearGradient>
			</defs>

			<!-- Grid lines -->
			{#each yTicks as tick (tick.y)}
				<line
					x1={padL}
					y1={tick.y}
					x2={svgW - padR}
					y2={tick.y}
					stroke="var(--border)"
					stroke-width="1"
				/>
				<text
					x={padL - 5}
					y={tick.y + 4}
					text-anchor="end"
					fill="rgba(236,234,229,0.38)"
					font-size="9"
					font-family="var(--font-mono)">{tick.label}</text
				>
			{/each}

			<!-- Market value fill -->
			{#if mvFill}
				<polygon points={mvFill} fill="url(#growthGradient)" />
			{/if}

			<!-- Cost base line (dashed gray) -->
			<polyline
				points={cbPoints}
				fill="none"
				stroke="rgba(236,234,229,0.28)"
				stroke-width="1.5"
				stroke-dasharray="6 4"
				stroke-linecap="round"
				stroke-linejoin="round"
			/>

			<!-- Market value line (solid amber) -->
			<polyline
				points={mvPoints}
				fill="none"
				stroke="var(--amber)"
				stroke-width="2.5"
				stroke-linecap="round"
				stroke-linejoin="round"
			/>

			<!-- Last point circle -->
			{#if values.length > 0}
				{@const lx = toX(values.length - 1, values.length)}
				{@const ly = toY(values[values.length - 1].mv)}
				<circle
					cx={lx}
					cy={ly}
					r="4"
					fill="var(--amber-light)"
					stroke="rgba(0,0,0,0.35)"
					stroke-width="1.5"
				/>
			{/if}

			<!-- X-axis labels -->
			{#each xLabels as { i, date } (i)}
				<text
					x={toX(i, values.length)}
					y={padT + plotH + 20}
					text-anchor="middle"
					fill="rgba(236,234,229,0.38)"
					font-size="9"
					font-family="var(--font-mono)">{fmtDate(date)}</text
				>
			{/each}
		</svg>

		<div class="legend">
			<div class="legend-item">
				<span class="legend-line amber"></span>
				<span>Valor de mercado</span>
			</div>
			<div class="legend-item">
				<span class="legend-line gray"></span>
				<span>Capital invertido</span>
			</div>
		</div>
	{/if}
</div>

<style>
	.growth-card {
		background: var(--surface);
		border: 1px solid var(--border-strong);
		border-radius: 14px;
		padding: 2rem;
		backdrop-filter: blur(10px);
	}

	.card-top {
		display: flex;
		flex-wrap: wrap;
		align-items: flex-start;
		justify-content: space-between;
		gap: 1rem;
		margin-bottom: 1.25rem;
	}

	.divider {
		height: 1px;
		background: var(--border);
		margin-bottom: 1.5rem;
	}

	.period-tabs {
		display: flex;
		gap: 0.2rem;
		background: rgba(255, 255, 255, 0.03);
		border: 1px solid var(--border);
		border-radius: 8px;
		padding: 0.2rem;
		flex-shrink: 0;
	}

	.period-btn {
		padding: 0.3rem 0.65rem;
		border: none;
		background: transparent;
		color: var(--text-dim);
		border-radius: 6px;
		font-size: 0.72rem;
		font-weight: 600;
		font-family: var(--font-mono);
		cursor: pointer;
		transition:
			background 0.15s ease,
			color 0.15s ease;
	}

	.period-btn.active {
		background: rgba(212, 145, 42, 0.18);
		color: var(--amber-light);
	}

	.period-btn:hover:not(.active) {
		background: rgba(255, 255, 255, 0.05);
		color: var(--text);
	}

	.metrics-row {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 1.5rem;
		margin-bottom: 1.5rem;
		padding-bottom: 1.5rem;
		border-bottom: 1px solid var(--border);
	}

	.chart {
		width: 100%;
		display: block;
		margin-bottom: 0.75rem;
	}

	.legend {
		display: flex;
		gap: 1.5rem;
		font-size: 0.72rem;
		font-family: var(--font-mono);
		color: var(--text-dim);
	}

	.legend-item {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.legend-line {
		display: inline-block;
		width: 22px;
		height: 0;
	}

	.legend-line.amber {
		border-top: 2.5px solid var(--amber);
	}

	.legend-line.gray {
		border-top: 1.5px dashed rgba(236, 234, 229, 0.28);
	}

	.empty-chart {
		padding: 3rem 2rem;
		text-align: center;
		color: var(--text-dim);
		font-size: 0.82rem;
		border: 1px dashed var(--border);
		border-radius: 8px;
		line-height: 1.6;
	}

	@media (max-width: 768px) {
		.growth-card {
			padding: 1.5rem;
		}
	}

	@media (max-width: 600px) {
		.metrics-row {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}

		.card-top {
			flex-direction: column;
			align-items: flex-start;
		}
	}
</style>
