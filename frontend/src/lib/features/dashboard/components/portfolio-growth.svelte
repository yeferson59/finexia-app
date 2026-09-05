<script lang="ts">
	import CardHeader from '$lib/ui/card-header.svelte';
	import EmptyState from '$lib/ui/empty-state.svelte';
	import Stat from '$lib/ui/stat.svelte';
	import GrowthChart from './growth-chart.svelte';
	import GrowthControls from './growth-controls.svelte';
	import GrowthReadout from './growth-readout.svelte';
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCompactCurrency, formatCurrency } from '$lib/shared/format/money';
	import { formatPercent, formatSignedPercent } from '$lib/shared/format/percent';
	import { timeWeightedReturn } from '$lib/shared/finance/returns';
	import {
		GROWTH_LABELS,
		filterByPeriod,
		growthScale,
		toGrowthPoints,
		type GrowthView,
		type Period
	} from '../dashboard';

	import type { GrowthDataPoint, GrowthSummary } from '$lib/api/types';

	const {
		data = [],
		summary = { firstDate: '', initialValue: '0', currentValue: '0', totalGrowthPct: '0' },
		bare = false,
		formatMoney = undefined
	}: {
		data: GrowthDataPoint[];
		summary: GrowthSummary;
		/**
		 * Sin la caja: en el panel la gráfica va sobre el fondo de la página, que
		 * ya no apila tarjetas. El detalle de portafolio la sigue queriendo
		 * enmarcada, porque allí sí convive con otras.
		 */
		bare?: boolean;
		/**
		 * Cómo escribir un importe. Quien la incruste puede imponer el suyo para
		 * que la tarjeta no discrepe de las cifras que tiene al lado: el detalle
		 * de portafolio escribe «US$ 45.035,10» en sus tarjetas y le pasa ese
		 * mismo formateador. Sin él manda el ayudante compartido, que es lo que
		 * usa el panel.
		 */
		formatMoney?: (value: number) => string;
	} = $props();

	let selectedPeriod = $state<Period>('Todo');
	/*
	 * En qué unidad se dibuja. Arranca en dinero porque es la pregunta que
	 * primero se hace quien abre el panel —cuánto tengo—; el porcentaje contesta
	 * la segunda, que es cuánto de eso se ha ganado.
	 */
	let view = $state<GrowthView>('value');
	/** Punto bajo el cursor (ratón o teclado); `null` cuando no hay ninguno. */
	let activeIndex = $state<number | null>(null);

	const isPercent = $derived(view === 'percent');
	const labels = $derived(GROWTH_LABELS[view]);

	const filteredData = $derived(filterByPeriod(data, selectedPeriod));

	const points = $derived(toGrowthPoints(filteredData, view));

	/*
	 * En porcentaje el cero entra siempre en la escala: sin él una racha entera
	 * en negativo se dibujaba como una curva que sube, con el eje empezando en
	 * −12 % y ninguna referencia que dijera dónde estaba el equilibrio.
	 */
	const scale = $derived(
		growthScale(
			isPercent ? [...points.flatMap((p) => [p.mv, p.cb]), 0] : points.flatMap((p) => [p.mv, p.cb])
		)
	);

	const activePoint = $derived(activeIndex === null ? null : (points[activeIndex] ?? null));

	/*
	 * La rentabilidad del tramo a la vista, limpia de aportes y retiros.
	 *
	 * Es la cifra que el «Rendimiento» de al lado no puede dar: aquel divide la
	 * ganancia de hoy entre lo invertido hoy, así que un aporte grande hecho
	 * después de una subida lo hunde sin que la cartera haya perdido nada. Esta
	 * encadena tramos y solo se mueve con el mercado. `null` mientras la serie
	 * no dé ni un tramo que medir.
	 */
	const realReturnPct = $derived.by(() => {
		const twr = timeWeightedReturn(filteredData);
		return twr === null ? null : twr * 100;
	});

	// Métricas del resumen (la foto completa, no la del punto señalado).
	//
	// La ganancia es mercado menos capital invertido, no la diferencia entre el
	// valor de hoy y el del primer snapshot: esa segunda cuenta convertía en
	// «ganancia» el dinero que entraba, y un portafolio nuevo o una posición
	// añadida disparaban la cifra mientras la cartera perdía. Es también la que
	// contradecía al punto del gráfico, que siempre restó bien.
	const currentVal = $derived(parseFloat(summary.currentValue || '0'));
	// Del dato crudo y no de `points`: en la vista de porcentaje aquellos ya no
	// son importes, y el respaldo habría restado dos rentabilidades.
	const lastRaw = $derived(filteredData.at(-1) ?? null);
	// Ausente (backend anterior) no es cero: se cae al último punto de la serie.
	const absoluteGain = $derived(
		summary.gainLoss !== undefined
			? parseFloat(summary.gainLoss || '0')
			: parseFloat(lastRaw?.totalValue || '0') - parseFloat(lastRaw?.totalCostBase || '0')
	);
	const investedVal = $derived(currentVal - absoluteGain);
	const gainPct = $derived(
		summary.gainLossPct !== undefined
			? parseFloat(summary.gainLossPct || '0')
			: investedVal > 0
				? (absoluteGain / investedVal) * 100
				: 0
	);
	const isPositive = $derived(absoluteGain >= 0);

	/*
	 * La serie viene en una sola moneda: el backend convierte cada portafolio a
	 * la del perfil, o a la que pida el panel. Antes el signo de dólar estaba
	 * fijo en la tarjeta, así que unas cifras en pesos se leían como dólares.
	 */
	const currency = $derived((summary.currency || 'USD').trim().toUpperCase());

	/*
	 * Fechas cuyo total incluye algún portafolio que no se pudo convertir: sus
	 * importes entran a valor nominal, así que el total no está del todo en la
	 * moneda que dice. Se avisa en vez de callarlo.
	 */
	const unconvertedDates = $derived(
		filteredData.filter((d) => (d.portfoliosUnconverted ?? 0) > 0).length
	);

	/** Al cambiar de periodo el índice anterior ya no señala al mismo día. */
	function selectPeriod(period: Period) {
		selectedPeriod = period;
		activeIndex = null;
	}

	function fmt(v: number): string {
		return new Intl.NumberFormat('es-CO', {
			minimumFractionDigits: 2,
			maximumFractionDigits: 2
		}).format(v);
	}

	/*
	 * Por `formatCurrency` y no componiendo símbolo + número a mano: esta
	 * tarjeta escribía «$89.406,10» mientras la cifra grande de la misma página
	 * decía «$89,406.10», porque una pasaba por el ayudante compartido —que da a
	 * cada moneda el locale en el que se lee— y la otra formateaba en es-CO
	 * fuera cual fuera la moneda. Dos formatos para el mismo número.
	 */
	function fmtMoney(v: number): string {
		return formatMoney ? formatMoney(v) : privacy.money(formatCurrency(v, currency));
	}

	/** Etiquetas del eje: el importe abreviado, que es lo que cabe entre marcas. */
	function fmtMoneyAbbrev(v: number): string {
		return privacy.money(formatCompactCurrency(v, currency));
	}

	/*
	 * Un porcentaje no es un importe: el modo privado no lo tapa. Enmascarar
	 * «+8,4 %» no esconde cuánto tiene nadie y sí deja la gráfica ilegible justo
	 * cuando alguien la enseña en una pantalla ajena, que es para lo que está
	 * ese modo.
	 */
	const fmtPct = (v: number) => formatSignedPercent(v, 1);

	/*
	 * Decimales del eje según lo que abarque: una serie que se mueve dos puntos
	 * necesita el decimal —sin él las cinco marcas dicen «2%»— y una que se mueve
	 * ochenta no lo quiere.
	 */
	const fmtPctAbbrev = (v: number) => formatPercent(v, scale.yRange < 5 ? 1 : 0);

	const formatValue = $derived(isPercent ? fmtPct : fmtMoney);
	const formatAbbrev = $derived(isPercent ? fmtPctAbbrev : fmtMoneyAbbrev);

	/*
	 * Con más de un año a la vista, día y mes no bastan: el eje repetía "1 de
	 * jun" para dos junios distintos. El año solo aparece cuando hace falta,
	 * para no recargar los rangos cortos.
	 */
	const spansYears = $derived(
		points.length > 1 && points[0].date.slice(0, 4) !== points[points.length - 1].date.slice(0, 4)
	);

	/*
	 * En rangos largos el eje pasa a "mes año" y suelta el día: con
	 * "01 de jun de 25" las seis etiquetas se pisaban unas a otras, y a esa
	 * escala el día no aporta nada —el detalle exacto lo da el cursor—.
	 */
	function fmtDate(iso: string): string {
		const d = new Date(iso + 'T00:00:00');
		return spansYears
			? d.toLocaleDateString('es-CO', { month: 'short', year: '2-digit' })
			: d.toLocaleDateString('es-CO', { day: '2-digit', month: 'short' });
	}

	function fmtLongDate(iso: string): string {
		const d = new Date(iso + 'T00:00:00');
		return d.toLocaleDateString('es-CO', { day: '2-digit', month: 'long', year: 'numeric' });
	}
</script>

<div class="growth-card" class:bare>
	<div class="card-top">
		{#if bare}
			<h2 class="growth-title">Crecimiento</h2>
		{:else}
			<CardHeader eyebrow="Portafolio" title="Crecimiento del portafolio" divider={false} />
		{/if}
		<GrowthControls
			{view}
			period={selectedPeriod}
			onview={(next) => (view = next)}
			onperiod={selectPeriod}
		/>
	</div>

	<div class="divider"></div>

	<div class="metrics-row">
		<Stat
			label="Total ganancia"
			tone={isPositive ? 'positive' : 'negative'}
			value="{isPositive ? '+' : '−'}{fmtMoney(Math.abs(absoluteGain))}"
		/>
		<Stat
			label="Rendimiento"
			tone={isPositive ? 'positive' : 'negative'}
			value="{isPositive ? '+' : ''}{fmt(gainPct)}%"
		/>
		<!-- La cifra que el «Rendimiento» de al lado no puede dar: aquella divide
		     la ganancia entre lo invertido hoy y se hunde con un aporte hecho tras
		     una subida; esta encadena tramos y solo se mueve con el mercado. Va
		     con el periodo en la etiqueta porque, a diferencia de las otras tres,
		     mide el tramo que se está viendo. -->
		<Stat
			label="Rentabilidad real, {selectedPeriod}"
			tone={realReturnPct === null ? 'default' : realReturnPct >= 0 ? 'positive' : 'negative'}
			value={realReturnPct === null ? '—' : fmtPct(realReturnPct)}
		/>
		<!-- El código va en la etiqueta porque el símbolo no siempre distingue:
		     en es-CO el peso y el dólar comparten el "$". -->
		<Stat label="Valor actual en {currency}" tone="highlight" value={fmtMoney(currentVal)} />
	</div>

	{#if unconvertedDates > 0}
		<p class="fx-note" role="status">
			Faltan tasas para convertir algún portafolio a {currency}: en {unconvertedDates}
			{unconvertedDates === 1 ? 'fecha' : 'fechas'} sus importes se suman sin convertir.
		</p>
	{/if}

	{#if points.length < 2}
		<EmptyState
			bordered
			title="Aún no hay suficiente historial"
			description="El gráfico se dibuja a medida que el sistema registra las capturas diarias de tu portafolio."
		/>
	{:else}
		<GrowthReadout
			point={activePoint}
			primaryLabel={labels.primary}
			secondaryLabel={labels.secondary}
			{isPercent}
			{formatValue}
			formatDate={fmtLongDate}
		/>

		<GrowthChart
			{points}
			{scale}
			active={activeIndex}
			primaryLabel={labels.primary}
			secondaryLabel={labels.secondary}
			caption={labels.caption}
			baseline={isPercent ? 0 : null}
			{formatAbbrev}
			formatDate={fmtDate}
			formatFullDate={fmtLongDate}
			{formatValue}
			onactivate={(index) => (activeIndex = index)}
		/>

		<div class="legend">
			<div class="legend-item">
				<span class="legend-line amber"></span>
				<span>{labels.primary}</span>
			</div>
			<div class="legend-item">
				<span class="legend-line gray"></span>
				<span>{labels.secondary}</span>
			</div>
		</div>

		{#if isPercent}
			<p class="view-note">
				La rentabilidad descuenta aportes y retiros: solo se mueve con el mercado. La ganancia sobre
				coste sí depende de cuándo entró cada aporte, y por eso las dos líneas se separan.
			</p>
		{/if}
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

	.growth-card.bare {
		background: none;
		border: none;
		border-radius: 0;
		padding: 0;
		backdrop-filter: none;
		/* En el panel la gráfica comparte fila con el extracto, así que su ancho no
		   se deduce del de la ventana: las cuatro métricas se reparten según lo que
		   mida la columna, no la pantalla. */
		container-type: inline-size;
	}

	@container (max-width: 620px) {
		.growth-card.bare .metrics-row {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}
	}

	.growth-title {
		margin: 0;
		font-family: var(--font-body);
		font-size: 1.05rem;
		font-weight: 500;
		color: var(--text);
	}

	/*
	 * Las etiquetas de las métricas, en caja normal y con la tipografía del
	 * resto del panel. `Stat` las pinta en versalitas monoespaciadas para el
	 * resto de la aplicación, y aquí eso era el mismo antetítulo repetido cuatro
	 * veces bajo un titular que ya no lo lleva. Solo en modo `bare`, así que el
	 * detalle de portafolio no se entera.
	 */
	.growth-card.bare :global(.stat-label) {
		font-family: var(--font-body);
		font-size: 0.75rem;
		letter-spacing: normal;
		text-transform: none;
		color: var(--text-dim);
	}

	.growth-card.bare .divider {
		display: none;
	}

	.growth-card.bare .metrics-row {
		margin-top: 1.25rem;
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

	.metrics-row {
		display: grid;
		grid-template-columns: repeat(4, minmax(0, 1fr));
		gap: 1.5rem;
		margin-bottom: 1.5rem;
		padding-bottom: 1.5rem;
		border-bottom: 1px solid var(--border);
	}

	.fx-note {
		margin: -0.75rem 0 1.25rem;
		font-size: 0.8rem;
		color: rgba(212, 145, 42, 0.85);
	}

	/* Solo en la vista de porcentaje: las dos líneas se separan y hay que decir
	   por qué, o la diferencia se lee como un error de la gráfica. */
	.view-note {
		margin: 0.6rem 0 0;
		font-size: 0.72rem;
		line-height: 1.5;
		color: rgba(236, 234, 229, 0.42);
	}

	.legend {
		display: flex;
		gap: 1.5rem;
		margin-top: 0.75rem;
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
		border-top: 1.5px dashed var(--cost);
	}

	@media (max-width: 768px) {
		.growth-card:not(.bare) {
			padding: 1.5rem;
		}
	}

	@media (max-width: 900px) {
		.metrics-row {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}
	}

	@media (max-width: 600px) {
		.card-top {
			flex-direction: column;
			align-items: flex-start;
		}
	}
</style>
