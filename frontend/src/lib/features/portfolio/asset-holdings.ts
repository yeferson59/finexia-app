/**
 * Vista consolidada de activos: lo que el usuario tiene de cada uno sumando
 * todos sus portafolios (`GET /portfolios/holdings`).
 *
 * Responde a una pregunta que ninguna otra pantalla contesta —«¿cuánto tengo de
 * X?»—: las posiciones del detalle solo suman dentro de su portafolio y el
 * donut del panel pliega todo a ocho clases de activo, así que un activo
 * repartido entre tres portafolios no tenía una sola fila en ninguna parte.
 *
 * Helpers puros: sin Svelte y sin red. Los contratos vienen de `$lib/api/types`.
 */
import type { AssetHolding } from '$lib/api/types';
import { formatAssetType } from '$lib/shared/format/asset-type';
import { buildSlices } from '$lib/shared/chart/pie';

export type { AssetHolding };

/**
 * Colores de las porciones, en orden fijo.
 *
 * No son los de `ASSET_TYPE_COLORS`: allí el color **es** la clase de activo, y
 * aquí cada porción es un activo distinto, así que dos acciones saldrían del
 * mismo color y la gráfica no distinguiría nada. Son los mismos tonos de la
 * marca reescalonados a la banda de luminosidad que pide un fondo oscuro
 * (OKLCH L≈0.63): a la claridad original, siete tonos seguidos sobre este fondo
 * se empastan entre sí. Validados para daltonismo por pares contiguos.
 *
 * Se reparten por posición en el ranking, no por activo: la torta es un «los N
 * mayores» y su leyenda va al lado, así que el orden es el dato. Como no hay
 * ningún filtro que cambie el conjunto, no hay repintado que confunda.
 */
export const ASSET_SERIES_COLORS = [
	'#bb7900',
	'#6383e3',
	'#00a363',
	'#a170c6',
	'#c56f41',
	'#3794bf',
	'#a48609'
];

/** Gris del agregado «Otros»: no es una serie más, es el resto. */
export const OTHERS_COLOR = '#8a8780';

/**
 * Cuántos activos se dibujan por separado antes de agrupar en «Otros».
 *
 * Uno menos que colores hay, porque la regla de abajo admite una porción extra
 * cuando sobra un solo activo: así el índice nunca da la vuelta al arreglo y
 * dos porciones no pueden salir del mismo color.
 */
export const PIE_MAX_SLICES = ASSET_SERIES_COLORS.length - 1;

/** Fila de la tabla: el holding del backend ya convertido a números. */
export interface AssetHoldingRow {
	assetId: string;
	ticker: string;
	name: string;
	assetType: string;
	/** Etiqueta legible de la clase de activo. */
	typeLabel: string;
	/** Unidades, sumadas entre portafolios. Solo significan algo en su fila. */
	quantity: number;
	/** Precio por unidad en `currency`, o `null` si la posición va a coste. */
	marketPrice: number | null;
	/** Moneda en la que cotiza el activo (la de `marketPrice`). */
	currency: string;
	/** Valor de la posición, en la moneda de visualización. */
	value: number;
	percent: number;
	portfolios: number;
	priceSource: string;
	/**
	 * `false` si alguna posición del activo no pudo convertirse por falta de
	 * tasa: su importe va en su moneda nativa y el total la mezcla.
	 */
	fxConverted: boolean;
}

/**
 * Precio por unidad, o `null` cuando no hay ninguno.
 *
 * El backend manda cadena vacía para una posición valorada a coste: cada
 * entrada pagó lo suyo y ningún número representa al activo. Vacío no es cero
 * —un 0 se leería como un activo que no vale nada— y por eso no se parsea.
 */
function toPrice(raw: string): number | null {
	if (raw === '') return null;
	const parsed = parseFloat(raw);

	return Number.isFinite(parsed) ? parsed : null;
}

export function toAssetHoldingRows(holdings: AssetHolding[]): AssetHoldingRow[] {
	return holdings.map((h) => ({
		assetId: h.assetId,
		ticker: h.ticker,
		name: h.name,
		assetType: h.assetType,
		typeLabel: formatAssetType(h.assetType),
		quantity: parseFloat(h.quantity) || 0,
		// Vacío no es cero: es «no hay precio que represente al activo», y un 0
		// se leería como un activo que no vale nada.
		marketPrice: toPrice(h.marketPrice),
		currency: h.currency,
		value: parseFloat(h.marketValue) || 0,
		percent: h.percent,
		portfolios: h.portfolios,
		priceSource: h.priceSource,
		fxConverted: h.positionsUnconverted === 0
	}));
}

/** Porción de la torta de concentración. */
export interface ConcentrationSlice {
	/** Clave de la porción: el ticker, o `__others__` para el agregado. */
	key: string;
	label: string;
	value: number;
	percent: number;
	color: string;
	/** Cuántos activos hay detrás: 1, salvo en «Otros». */
	assets: number;
}

/**
 * Reparte las filas en porciones: los `max` mayores por separado y el resto
 * agrupado en «Otros».
 *
 * Agrupar no es un adorno: pasado un puñado de porciones la torta deja de
 * leerse, y las colas de una cartera diversificada son decenas de rebanadas de
 * un grado. El detalle no se pierde —está en la tabla de al lado, entera.
 *
 * Los porcentajes son los que calculó el backend sobre el total convertido; no
 * se recalculan aquí para que la torta y la columna «peso» no puedan discrepar.
 * «Otros» es la suma de los suyos.
 */
export function buildConcentration(
	rows: AssetHoldingRow[],
	max = PIE_MAX_SLICES
): ConcentrationSlice[] {
	const ranked = [...rows].sort((a, b) => b.value - a.value).filter((row) => row.value > 0);

	// Un solo activo de más no merece una rebanada «Otros» que lo esconda y
	// ocupe lo mismo: con `max + 1` filas caben todas.
	const cut = ranked.length <= max + 1 ? ranked.length : max;
	const head = ranked.slice(0, cut);
	const tail = ranked.slice(cut);

	const slices: ConcentrationSlice[] = head.map((row, i) => ({
		key: row.ticker,
		label: row.ticker,
		value: row.value,
		percent: row.percent,
		color: ASSET_SERIES_COLORS[i],
		assets: 1
	}));

	if (tail.length > 0) {
		slices.push({
			key: '__others__',
			label: 'Otros',
			value: tail.reduce((sum, row) => sum + row.value, 0),
			percent: tail.reduce((sum, row) => sum + row.percent, 0),
			color: OTHERS_COLOR,
			assets: tail.length
		});
	}

	return slices;
}

/** Las porciones con su geometría, listas para pintar. */
export function buildConcentrationSlices(rows: AssetHoldingRow[], max = PIE_MAX_SLICES) {
	return buildSlices(buildConcentration(rows, max));
}

/**
 * Unidades de un activo, con los decimales que hagan falta.
 *
 * No es dinero y no lleva su formato: 0,00000123 BTC y 15 acciones son la misma
 * columna, y redondear a dos decimales convierte la primera en cero. Ocho es lo
 * que guarda la base de datos.
 */
export function formatQuantity(value: number): string {
	return new Intl.NumberFormat('es-CO', { maximumFractionDigits: 8 }).format(value);
}
