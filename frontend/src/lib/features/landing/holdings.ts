/**
 * Las posiciones del ejemplo, una a una, para el mapa del hero.
 *
 * El hero enseña lo que hace Finexia: las mismas dieciocho posiciones, vistas
 * por la plataforma donde están, por el portafolio al que tú las asignas o por
 * su tipo. Para que la reagrupación no mienta, cada posición existe una sola
 * vez y los grupos se calculan de ella: las cinco plataformas suman lo que dice
 * `TOUR_BREAKDOWN`, los tres portafolios lo que dice `TOUR_PORTFOLIOS`, y todo
 * cierra en los $248,500 del resumen. Lo comprueba `holdings.spec.ts`.
 */

import { TOUR_BREAKDOWN, TOUR_PORTFOLIOS } from './product-tour';

export type HoldingLens = 'platform' | 'portfolio' | 'type';

export type AssetType = 'Acciones' | 'ETFs' | 'Cripto' | 'Bonos' | 'Efectivo';

export interface Holding {
	/** Lo que cabe dentro del bloque: el símbolo, o una palabra. */
	short: string;
	name: string;
	platform: string;
	portfolio: string;
	type: AssetType;
	/** En dólares. */
	value: number;
}

export const HOLDINGS: Holding[] = [
	// Degiro
	{
		short: 'VWCE',
		name: 'Vanguard FTSE All-World',
		platform: 'Degiro',
		portfolio: 'Jubilación',
		type: 'ETFs',
		value: 38400
	},
	{
		short: 'CSPX',
		name: 'iShares Core S&P 500',
		platform: 'Degiro',
		portfolio: 'Jubilación',
		type: 'ETFs',
		value: 18200
	},
	{
		short: 'ASML',
		name: 'ASML Holding',
		platform: 'Degiro',
		portfolio: 'Jubilación',
		type: 'Acciones',
		value: 9800
	},
	{
		short: 'MSFT',
		name: 'Microsoft',
		platform: 'Degiro',
		portfolio: 'Jubilación',
		type: 'Acciones',
		value: 8300
	},
	{
		short: 'EIMI',
		name: 'iShares Core MSCI Emerging Markets',
		platform: 'Degiro',
		portfolio: 'Jubilación',
		type: 'ETFs',
		value: 5400
	},
	{
		short: 'IB01',
		name: 'iShares $ Treasury Bond 0-1yr',
		platform: 'Degiro',
		portfolio: 'Reserva',
		type: 'ETFs',
		value: 4100
	},
	// Binance
	{
		short: 'BTC',
		name: 'Bitcoin',
		platform: 'Binance',
		portfolio: 'Cripto',
		type: 'Cripto',
		value: 32400
	},
	{
		short: 'ETH',
		name: 'Ethereum',
		platform: 'Binance',
		portfolio: 'Cripto',
		type: 'Cripto',
		value: 12800
	},
	{
		short: 'SOL',
		name: 'Solana',
		platform: 'Binance',
		portfolio: 'Cripto',
		type: 'Cripto',
		value: 6900
	},
	{
		short: 'BNB',
		name: 'BNB',
		platform: 'Binance',
		portfolio: 'Cripto',
		type: 'Cripto',
		value: 3800
	},
	// Revolut
	{
		short: 'GOOGL',
		name: 'Alphabet',
		platform: 'Revolut',
		portfolio: 'Jubilación',
		type: 'Acciones',
		value: 18800
	},
	{
		short: 'ETH',
		name: 'Ethereum',
		platform: 'Revolut',
		portfolio: 'Cripto',
		type: 'Cripto',
		value: 13000
	},
	{
		short: 'USD',
		name: 'Efectivo en dólares',
		platform: 'Revolut',
		portfolio: 'Reserva',
		type: 'Efectivo',
		value: 9500
	},
	// IBKR
	{
		short: 'NVDA',
		name: 'NVIDIA',
		platform: 'IBKR',
		portfolio: 'Jubilación',
		type: 'Acciones',
		value: 21300
	},
	{
		short: 'AAPL',
		name: 'Apple',
		platform: 'IBKR',
		portfolio: 'Jubilación',
		type: 'Acciones',
		value: 12200
	},
	{
		short: 'UST',
		name: 'Bono del Tesoro de EE. UU. a 2 años',
		platform: 'IBKR',
		portfolio: 'Reserva',
		type: 'Bonos',
		value: 6400
	},
	// Banco
	{
		short: 'CDT',
		name: 'CDT a 360 días',
		platform: 'Banco',
		portfolio: 'Reserva',
		type: 'Bonos',
		value: 18000
	},
	{
		short: 'Ahorro',
		name: 'Cuenta de ahorros',
		platform: 'Banco',
		portfolio: 'Reserva',
		type: 'Efectivo',
		value: 9200
	}
];

export const HOLDING_LENSES: { id: HoldingLens; label: string }[] = [
	{ id: 'platform', label: 'Plataforma' },
	{ id: 'portfolio', label: 'Portafolio' },
	{ id: 'type', label: 'Tipo de activo' }
];

/** El orden de las filas: el de las listas del recorrido, y los tipos por peso. */
const PLATFORM_ORDER = TOUR_BREAKDOWN.map((row) => row.name);
export const PORTFOLIO_ORDER = TOUR_PORTFOLIOS.map((p) => p.name);
const TYPE_ORDER: AssetType[] = ['Acciones', 'Cripto', 'ETFs', 'Bonos', 'Efectivo'];

export interface PlacedHolding extends Holding {
	/** Posición estable en `HOLDINGS`: la clave del bloque en las tres vistas. */
	id: number;
	row: number;
	/** Inicio y ancho del bloque, en fracción de la fila más larga posible. */
	x: number;
	w: number;
}

export interface HoldingGroup {
	label: string;
	value: number;
	count: number;
}

export interface HoldingLayout {
	groups: HoldingGroup[];
	blocks: PlacedHolding[];
}

function keyOf(holding: Holding, lens: HoldingLens): string {
	if (lens === 'platform') return holding.platform;
	if (lens === 'portfolio') return holding.portfolio;
	return holding.type;
}

function orderOf(lens: HoldingLens): readonly string[] {
	if (lens === 'platform') return PLATFORM_ORDER;
	if (lens === 'portfolio') return PORTFOLIO_ORDER;
	return TYPE_ORDER;
}

/**
 * El total del grupo más grande entre las tres vistas. Es la escala común: un
 * bloque mide lo mismo en todas, así que al cambiar de vista las piezas se
 * mueven pero no crecen ni encogen, que es justo lo que hay que ver.
 */
export function layoutScale(holdings: Holding[] = HOLDINGS): number {
	let max = 0;
	for (const lens of HOLDING_LENSES) {
		const totals = new Map<string, number>();
		for (const h of holdings) {
			const key = keyOf(h, lens.id);
			totals.set(key, (totals.get(key) ?? 0) + h.value);
		}
		for (const total of totals.values()) max = Math.max(max, total);
	}
	return max;
}

/**
 * Filas y bloques de una vista. Dentro de cada fila los bloques van agrupados
 * por portafolio —el color—, y dentro de cada portafolio de mayor a menor.
 */
export function layoutHoldings(lens: HoldingLens, holdings: Holding[] = HOLDINGS): HoldingLayout {
	const scale = layoutScale(holdings);
	const order = orderOf(lens);
	const rank = (value: string, list: readonly string[]) => {
		const i = list.indexOf(value);
		return i === -1 ? list.length : i;
	};

	const groups: HoldingGroup[] = order
		.map((label) => {
			const members = holdings.filter((h) => keyOf(h, lens) === label);
			return {
				label,
				value: members.reduce((sum, h) => sum + h.value, 0),
				count: members.length
			};
		})
		.filter((g) => g.count > 0);

	const blocks: PlacedHolding[] = [];
	groups.forEach((group, row) => {
		const members = holdings
			.map((h, id) => ({ ...h, id }))
			.filter((h) => keyOf(h, lens) === group.label)
			.sort(
				(a, b) =>
					rank(a.portfolio, PORTFOLIO_ORDER) - rank(b.portfolio, PORTFOLIO_ORDER) ||
					b.value - a.value
			);

		let offset = 0;
		for (const h of members) {
			blocks.push({ ...h, row, x: offset / scale, w: h.value / scale });
			offset += h.value;
		}
	});

	return { groups, blocks };
}

export const HOLDINGS_TOTAL = HOLDINGS.reduce((sum, h) => sum + h.value, 0);

/** Importe en dólares sin céntimos, como la columna estrecha del panel. */
export function formatUsd(value: number): string {
	return '$' + Math.round(value).toLocaleString('en-US');
}
