/**
 * Helpers puros, constantes y tipos de la feature `portfolio`, compartidos por
 * los componentes de detalle y de alta. Sin dependencias de Svelte ni de red.
 *
 * Los contratos del backend NO se redeclaran aquí: se importan de
 * `$lib/api/types` (única fuente de verdad) y se reexportan para que los
 * componentes de la feature no tengan que conocer la capa de API.
 */
import type {
	GrowthDataPoint,
	GrowthSummary,
	Holding,
	PortfolioSummary,
	TopTransaction
} from '$lib/api/types';
import { formatSignedPercent } from '$lib/shared/format/percent';
import { timeWeightedReturn } from '$lib/shared/finance/returns';
import { assetTypeColor, formatAssetType } from '$lib/shared/format/asset-type';
import { formatPortfolioType } from '$lib/shared/format/portfolio-type';

export type { GrowthDataPoint, GrowthSummary, PortfolioSummary, TopTransaction };

export const PORTFOLIO_TYPES: { value: string; label: string }[] = [
	{ value: 'stocks_etfs', label: 'Acciones y ETF' },
	{ value: 'stocks', label: 'Solo Acciones' },
	{ value: 'etfs', label: 'Solo ETFs' },
	{ value: 'cryptos', label: 'Criptomonedas' },
	{ value: 'bonds', label: 'Bonos y Renta Fija' },
	{ value: 'diversified', label: 'Portafolio Diverso' },
	{ value: 'forex', label: 'Divisas y Forex' },
	{ value: 'commodities', label: 'Commodities' },
	{ value: 'cash', label: 'Efectivo' }
];

/**
 * Subconjunto de `Holding` que necesita la agregación por ticker. Se deriva del
 * contrato para que un cambio de nombre/tipo en la API rompa aquí.
 */
export type RawHolding = Pick<
	Holding,
	| 'ticker'
	| 'name'
	| 'assetType'
	| 'quantity'
	| 'price'
	| 'marketPrice'
	| 'currency'
	| 'costBasisBase'
	| 'marketValueBase'
	| 'fxConverted'
>;

/** Holding agregado por ticker, listo para pintar. */
export interface HoldingView {
	symbol: string;
	name: string;
	assetType: string;
	quantity: number;
	/** Precio de mercado por unidad, en la moneda del activo (`currency`). */
	marketPrice: number;
	currency: string;
	/** Coste y valor en la moneda base del portafolio: sumables entre filas. */
	costBasis: number;
	value: number;
	gainLoss: number;
	gainLossPct: number;
	allocation: number;
	/**
	 * `false` si alguna de las posiciones agrupadas no pudo convertirse por
	 * falta de tasa: sus importes están en su moneda nativa y el total que las
	 * incluya mezcla monedas.
	 */
	fxConverted: boolean;
}

/**
 * Importe en moneda base que envía el backend, con vuelta al cálculo nativo
 * (`cantidad × precio`) cuando el campo no viene. Solo coinciden si la posición
 * ya estaba en la moneda base; en cualquier otro caso el fallback es la mezcla
 * de monedas que estos campos vinieron a resolver, y `fxConverted` lo delata.
 */
function baseAmount(raw: string | undefined, nativeFallback: number): number {
	if (raw === undefined || raw === '') return nativeFallback;
	const parsed = parseFloat(raw);
	return Number.isFinite(parsed) ? parsed : nativeFallback;
}

export interface TypeBreakdownSlice {
	type: string;
	label: string;
	value: number;
	color: string;
	pct: number;
}

export interface DonutSegment extends TypeBreakdownSlice {
	dasharray: string;
	dashoffset: number;
}

/**
 * Agrupa las entradas por ticker: el mismo activo en varias plataformas
 * aparece como una sola fila con cantidad y coste base acumulados.
 */
export function groupHoldings(list: RawHolding[]): HoldingView[] {
	const grouped: Record<string, HoldingView> = {};

	for (const h of list) {
		const quantity = parseFloat(h.quantity) || 0;
		const costPrice = parseFloat(h.price) || 0;
		const marketPrice = parseFloat(h.marketPrice) || costPrice;
		// Los totales se toman convertidos: sumar `cantidad × precio` de una
		// posición en EUR con una en USD da un número sin significado.
		const costBasis = baseAmount(h.costBasisBase, quantity * costPrice);
		const value = baseAmount(h.marketValueBase, quantity * marketPrice);
		// Ausente (backend anterior) no es lo mismo que `false`: solo se avisa
		// cuando el backend afirma que faltó la tasa, para no marcar como
		// sospechoso un portafolio de una sola moneda.
		const fxConverted = h.fxConverted ?? true;

		const existing = grouped[h.ticker];
		if (existing) {
			existing.quantity += quantity;
			existing.costBasis += costBasis;
			existing.value += value;
			existing.gainLoss = existing.value - existing.costBasis;
			existing.gainLossPct =
				existing.costBasis > 0 ? (existing.gainLoss / existing.costBasis) * 100 : 0;
			existing.fxConverted &&= fxConverted;
		} else {
			grouped[h.ticker] = {
				symbol: h.ticker,
				name: h.name,
				assetType: h.assetType,
				quantity,
				marketPrice,
				currency: h.currency,
				costBasis,
				value,
				gainLoss: value - costBasis,
				gainLossPct: costBasis > 0 ? ((value - costBasis) / costBasis) * 100 : 0,
				allocation: 0,
				fxConverted
			};
		}
	}

	const rows = Object.values(grouped);
	const total = rows.reduce((sum, h) => sum + h.value, 0);
	return rows.map((h) => ({ ...h, allocation: total > 0 ? (h.value / total) * 100 : 0 }));
}

export function computeTypeBreakdown(holdings: HoldingView[]): TypeBreakdownSlice[] {
	const grouped: Record<string, { label: string; value: number; color: string }> = {};
	for (const h of holdings) {
		const key = h.assetType;
		if (!grouped[key]) {
			grouped[key] = {
				label: formatAssetType(key),
				value: 0,
				color: assetTypeColor(key)
			};
		}
		grouped[key].value += h.value;
	}
	const total = Object.values(grouped).reduce((s, v) => s + v.value, 0);
	return Object.entries(grouped)
		.map(([type, data]) => ({
			type,
			...data,
			pct: total > 0 ? (data.value / total) * 100 : 0
		}))
		.sort((a, b) => b.value - a.value);
}

export const DONUT_RADIUS = 60;
const DONUT_CIRCUMFERENCE = 2 * Math.PI * DONUT_RADIUS;
const DONUT_GAP = 3;

export function computeDonutSegments(typeBreakdown: TypeBreakdownSlice[]): DonutSegment[] {
	const gap = typeBreakdown.length > 1 ? DONUT_GAP : 0;
	let acc = 0;
	return typeBreakdown.map((slice) => {
		const sliceLen = (slice.pct / 100) * DONUT_CIRCUMFERENCE;
		const dash = Math.max(sliceLen - gap, 0);
		const segment: DonutSegment = {
			...slice,
			dasharray: `${dash} ${DONUT_CIRCUMFERENCE - dash}`,
			dashoffset: -acc
		};
		acc += sliceLen;
		return segment;
	});
}

/**
 * Porcentaje con signo y dos decimales, «+12,35%».
 *
 * Delega en el formateador compartido, que escribe la coma decimal de es-CO: en
 * una misma tarjeta convivían «+12.35%» y «$1.234,50», con dos convenciones
 * distintas para el mismo número.
 */
export function formatPct(value: number): string {
	return formatSignedPercent(value, 2);
}

/**
 * Rentabilidad real del portafolio, en porcentaje, sobre su serie de
 * crecimiento.
 *
 * Es la ponderada por tiempo: descuenta aportes y retiros, así que mide cómo se
 * comportó el dinero y no cuánto dinero entró. La ganancia sobre coste que
 * enseñan las otras tarjetas contesta a otra pregunta —cuánto vale hoy lo que
 * costó— y por eso las dos cifras pueden no parecerse: un aporte grande hecho
 * después de una subida hunde la segunda sin que la cartera haya perdido nada.
 *
 * `null` mientras la serie no dé ni un tramo que medir, que es el estado de un
 * portafolio recién creado.
 */
export function realReturnPct(points: GrowthDataPoint[] | undefined): number | null {
	const twr = timeWeightedReturn(points ?? []);
	return twr === null ? null : twr * 100;
}

/** Tono del badge de riesgo, a partir del nombre del nivel. */
export function riskTone(name: string): 'success' | 'warning' | 'danger' | 'neutral' {
	const n = name.toLowerCase();
	if (n.includes('bajo')) return 'success';
	if (n.includes('moderado')) return 'warning';
	if (n.includes('alto')) return 'danger';
	return 'neutral';
}

// ---------------------------------------------------------------------------
// Listado de portafolios (`dashboard/portfolios`)
// ---------------------------------------------------------------------------

/**
 * Una fila del listado: el resumen del backend con sus cifras ya en números.
 *
 * El listado compara portafolios entre sí, así que lo que necesita de cada uno
 * es su tamaño y cómo le va, más lo que dice qué es: el nombre que le puso su
 * dueño, lo que escribió sobre él y su nivel de riesgo.
 */
export interface PortfolioRow {
	id: string;
	name: string;
	/** Lo que escribió el dueño. Vacía si no escribió nada. */
	description: string;
	/** Etiqueta de la combinación de clases: la reserva por si no hay descripción. */
	typeLabel: string;
	riskName: string;
	isDefault: boolean;
	positions: number;
	/** Moneda del importe: la de visualización si se pudo convertir, si no la suya. */
	currency: string;
	value: number;
	cost: number;
	gain: number;
	gainPct: number;
	/**
	 * `false` cuando no había tasa hacia la moneda pedida: los importes se
	 * quedaron en la moneda base del portafolio, así que ni se suman con los
	 * demás ni se comparan con ellos.
	 */
	converted: boolean;
	/** Posiciones que el propio portafolio suma sin convertir. */
	unconverted: number;
}

const toNumber = (raw: string | undefined): number => parseFloat(raw ?? '') || 0;

/**
 * Convierte los resúmenes en filas, de mayor a menor valor.
 *
 * El orden es lo que hace legible una escalera: la primera barra es la más
 * larga y las demás se leen contra ella. Los portafolios que no se pudieron
 * convertir van al final, juntos: su importe está en otra moneda y ponerlos
 * entre los demás invitaría a comparar barras que miden cosas distintas.
 */
export function toPortfolioRows(
	summaries: PortfolioSummary[],
	displayCurrency: string
): PortfolioRow[] {
	const rows = summaries.map((summary): PortfolioRow => {
		const currency = summary.displayCurrency || summary.baseCurrency || displayCurrency;

		return {
			id: summary.id,
			name: summary.name,
			description: summary.description ?? '',
			typeLabel: formatPortfolioType(summary.type),
			riskName: summary.riskName,
			isDefault: summary.isDefault ?? false,
			positions: summary.totalPositions,
			currency,
			value: toNumber(summary.totalMarketValue),
			cost: toNumber(summary.totalCostBase),
			gain: toNumber(summary.totalGainLoss),
			gainPct: toNumber(summary.totalGainLossPct),
			converted: currency === displayCurrency,
			unconverted: summary.positionsUnconverted ?? 0
		};
	});

	return rows.sort((a, b) => {
		if (a.converted !== b.converted) return a.converted ? -1 : 1;
		return b.value - a.value;
	});
}

/**
 * Escala del carril de las barras: el mayor importe que hay que dibujar.
 *
 * Es el máximo entre los valores y los costes, no solo entre los valores: en un
 * portafolio en pérdida el coste queda por fuera del extremo de la barra, y con
 * el carril escalado solo a los valores esa parte se salía del ancho.
 *
 * Sale de la lista entera y no de la hoja que se esté viendo: si se reescalara
 * por hoja, la primera barra de cada una llegaría al final y parecería el mayor
 * portafolio de todos.
 */
export function portfolioBarScale(rows: PortfolioRow[]): number {
	return rows
		.filter((row) => row.converted)
		.reduce((top, row) => Math.max(top, row.value, row.cost), 0);
}

/** Lo que suman las filas que están en la misma moneda. */
export interface PortfolioTotals {
	value: number;
	cost: number;
	gain: number;
	gainPct: number;
	positions: number;
	/** Filas sumadas. */
	counted: number;
	/** Filas que se listan pero no se suman: están en otra moneda. */
	excluded: number;
}

/**
 * Totales del listado.
 *
 * Solo suma lo convertido. Un portafolio en otra moneda se sigue listando —es
 * suyo y tiene que verlo— pero no entra en el total: sumarlo daría una cifra
 * que no está en ninguna moneda.
 */
export function portfolioTotals(rows: PortfolioRow[]): PortfolioTotals {
	const counted = rows.filter((row) => row.converted);
	const value = counted.reduce((sum, row) => sum + row.value, 0);
	const cost = counted.reduce((sum, row) => sum + row.cost, 0);

	return {
		value,
		cost,
		gain: value - cost,
		// Sobre lo que costó, que es la misma cuenta que hace el backend por
		// portafolio. Sin coste no hay porcentaje que calcular, y un 0 se leería
		// como un portafolio que no se movió.
		gainPct: cost > 0 ? ((value - cost) / cost) * 100 : 0,
		positions: counted.reduce((sum, row) => sum + row.positions, 0),
		counted: counted.length,
		excluded: rows.length - counted.length
	};
}
