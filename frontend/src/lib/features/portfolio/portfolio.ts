/**
 * Helpers puros, constantes y tipos de la feature `portfolio`, compartidos por
 * los componentes de detalle y de alta. Sin dependencias de Svelte ni de red.
 *
 * Los contratos del backend NO se redeclaran aquí: se importan de
 * `$lib/api/types` (única fuente de verdad) y se reexportan para que los
 * componentes de la feature no tengan que conocer la capa de API.
 */
import type { GrowthDataPoint, GrowthSummary, Holding, TopTransaction } from '$lib/api/types';

export type { GrowthDataPoint, GrowthSummary, TopTransaction };

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

export const ASSET_TYPE_LABELS: Record<string, string> = {
	stock: 'Acciones',
	etf: 'ETFs',
	crypto: 'Cripto',
	bond: 'Bonos',
	cash: 'Efectivo',
	real_estate: 'Inmobiliario',
	commodity: 'Materias primas',
	other: 'Otros'
};

export const ASSET_TYPE_COLORS: Record<string, string> = {
	stock: '#4f8ef7',
	etf: '#d4912a',
	crypto: '#f97316',
	bond: '#22c55e',
	cash: '#94a3b8',
	real_estate: '#a855f7',
	commodity: '#b45309',
	other: '#6b7280'
};

/**
 * Subconjunto de `Holding` que necesita la agregación por ticker. Se deriva del
 * contrato para que un cambio de nombre/tipo en la API rompa aquí.
 */
export type RawHolding = Pick<
	Holding,
	'ticker' | 'name' | 'assetType' | 'quantity' | 'price' | 'marketPrice'
>;

/** Holding agregado por ticker, listo para pintar. */
export interface HoldingView {
	symbol: string;
	name: string;
	assetType: string;
	quantity: number;
	marketPrice: number;
	costBasis: number;
	value: number;
	gainLoss: number;
	gainLossPct: number;
	allocation: number;
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
		const costBasis = quantity * costPrice;
		const value = quantity * marketPrice;

		const existing = grouped[h.ticker];
		if (existing) {
			existing.quantity += quantity;
			existing.costBasis += costBasis;
			existing.value += value;
			existing.gainLoss = existing.value - existing.costBasis;
			existing.gainLossPct =
				existing.costBasis > 0 ? (existing.gainLoss / existing.costBasis) * 100 : 0;
		} else {
			grouped[h.ticker] = {
				symbol: h.ticker,
				name: h.name,
				assetType: h.assetType,
				quantity,
				marketPrice,
				costBasis,
				value,
				gainLoss: value - costBasis,
				gainLossPct: costBasis > 0 ? ((value - costBasis) / costBasis) * 100 : 0,
				allocation: 0
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
				label: ASSET_TYPE_LABELS[key] ?? key,
				value: 0,
				color: ASSET_TYPE_COLORS[key] ?? '#6b7280'
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

export function formatPct(value: number): string {
	return `${value >= 0 ? '+' : ''}${value.toFixed(2)}%`;
}

/**
 * Etiquetas legibles del `type` de un portafolio (combinaciones de clases de
 * activo que devuelve el backend).
 */
export const PORTFOLIO_TYPE_LABELS: Record<string, string> = {
	stocks: 'Acciones',
	etfs: 'ETFs',
	cryptos: 'Criptomonedas',
	bonds: 'Bonos',
	cash: 'Efectivo',
	forex: 'Forex',
	real_estates: 'Inmobiliario',
	commodities: 'Materias primas',
	stocks_etfs: 'Acciones & ETFs',
	stocks_cryptos: 'Acciones & Cripto',
	stocks_bonds: 'Acciones & Bonos',
	stocks_cash: 'Acciones & Efectivo',
	stocks_real_estates: 'Acciones & Inmobiliario',
	stocks_commodities: 'Acciones & Materias primas',
	etfs_cryptos: 'ETFs & Cripto',
	etfs_bonds: 'ETFs & Bonos',
	etfs_cash: 'ETFs & Efectivo',
	etfs_real_estates: 'ETFs & Inmobiliario',
	etfs_commodities: 'ETFs & Materias primas',
	cryptos_bonds: 'Cripto & Bonos',
	cryptos_cash: 'Cripto & Efectivo',
	cryptos_real_estates: 'Cripto & Inmobiliario',
	cryptos_commodities: 'Cripto & Materias primas',
	bonds_cash: 'Bonos & Efectivo',
	bonds_real_estates: 'Bonos & Inmobiliario',
	bonds_commodities: 'Bonos & Materias primas',
	cash_real_estates: 'Efectivo & Inmobiliario',
	cash_commodities: 'Efectivo & Materias primas',
	real_estates_commodities: 'Inmobiliario & Materias primas',
	forex_stocks: 'Forex & Acciones',
	forex_etfs: 'Forex & ETFs',
	forex_cryptos: 'Forex & Cripto',
	forex_bonds: 'Forex & Bonos',
	forex_cash: 'Forex & Efectivo',
	forex_real_states: 'Forex & Inmobiliario',
	forex_commodities: 'Forex & Materias primas',
	diversified: 'Diversificado'
};

export function formatPortfolioType(type: string): string {
	return PORTFOLIO_TYPE_LABELS[type] ?? type.replace(/_/g, ' ');
}

/** Tono del badge de riesgo, a partir del nombre del nivel. */
export function riskTone(name: string): 'success' | 'warning' | 'danger' | 'neutral' {
	const n = name.toLowerCase();
	if (n.includes('bajo')) return 'success';
	if (n.includes('moderado')) return 'warning';
	if (n.includes('alto')) return 'danger';
	return 'neutral';
}
