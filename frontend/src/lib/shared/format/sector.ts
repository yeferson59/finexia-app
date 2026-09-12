/**
 * Etiquetas del `sector` de un activo.
 *
 * El vocabulario es el de `market.Sector` del backend —los once sectores GICS
 * en minúsculas y con guion bajo— más los dos cubos que el reparto por
 * industria deriva y el catálogo nunca guarda:
 *
 * - `unclassified`: el activo *puede* tener sector y nadie se lo ha puesto. Es
 *   trabajo pendiente, y su tamaño es justo lo que hay que enseñar: una cartera
 *   con el 60 % sin clasificar todavía no tiene respuesta por industria, y un
 *   gráfico que escondiera esas filas afirmaría que sí.
 * - `not_applicable`: no hay industria que rellenar —una cripto, un saldo en
 *   efectivo, un inmueble—. Nunca la habrá.
 *
 * Que sean dos y no uno es la única razón por la que este módulo existe en vez
 * de imprimir el valor crudo: «Sin clasificar» y «Sin industria» piden cosas
 * distintas al lector, y juntarlas convertiría una tarea pendiente en un hecho.
 *
 * Sin tabla de colores, a diferencia de `asset-type.ts`: el reparto se pinta con
 * una sola barra ámbar —el largo dice la magnitud y el nombre de la fila, de
 * quién es—, así que una paleta de trece tonos aquí sería código muerto.
 *
 * La otra mitad del módulo es el **desglose**: un activo puede estar repartido
 * entre varias industrias en vez de pertenecer a una (un ETF de mercado ancho),
 * y entonces trae `sectorWeights` y deja `sector` vacío. Las dos formas son
 * excluyentes, así que las funciones de abajo son las que deciden cuál enseña
 * cada pantalla.
 */

import { formatPercent } from './percent';

/**
 * Una industria y su peso, en porcentaje.
 *
 * Se declara aquí en vez de importar `SectorWeight` de `$lib/api/types` porque
 * `lib/shared` es la capa más baja y no importa de `api`
 * (docs/FRONTEND_ARCHITECTURE.md). Es estructuralmente el mismo tipo, así que
 * lo que llega de la API encaja sin conversión.
 */
type WeightedSector = { sector: string; weight: number };

/** Sectores que devuelve el backend (`market.Sector`), más sus dos cubos. */
export const SECTOR_LABELS: Record<string, string> = {
	technology: 'Tecnología',
	communication_services: 'Comunicaciones',
	healthcare: 'Salud',
	financials: 'Finanzas',
	consumer_discretionary: 'Consumo discrecional',
	consumer_staples: 'Consumo básico',
	industrials: 'Industria',
	energy: 'Energía',
	materials: 'Materiales',
	utilities: 'Servicios públicos',
	real_estate: 'Inmobiliario',
	unclassified: 'Sin clasificar',
	not_applicable: 'Sin industria'
};

/**
 * Los sectores que un activo puede llevar, en el orden en que se ofrecen.
 *
 * Deja fuera los dos cubos a propósito: `unclassified` y `not_applicable` los
 * deriva el backend al repartir y no se guardan nunca, así que ofrecerlos en un
 * desplegable sería ofrecer un valor que la API rechaza con 400. «Sin
 * clasificar» ya es lo que significa dejar el campo vacío.
 */
export const SECTOR_OPTIONS = [
	'technology',
	'communication_services',
	'healthcare',
	'financials',
	'consumer_discretionary',
	'consumer_staples',
	'industrials',
	'energy',
	'materials',
	'utilities',
	'real_estate'
].map((value) => ({ value, label: SECTOR_LABELS[value] }));

/**
 * Clases de activo que pueden llevar sector, que es la misma línea que traza
 * `market.AssetType.HasSector` en el backend: hay una empresa detrás de una
 * acción, un fondo o un bono, y no la hay detrás de una cripto, un saldo en
 * efectivo, un inmueble o un lingote.
 *
 * Se repite aquí para poder **no enseñar** el campo cuando no aplica, en vez de
 * dejar que el usuario lo rellene y que la API conteste 400. El backend sigue
 * siendo quien decide; esto solo evita ofrecer lo que va a rechazar.
 */
const CLASSIFIABLE_TYPES = new Set(['stock', 'etf', 'bond', 'other']);

/** Si una clase de activo admite sector. */
export function typeHasSector(assetType: string): boolean {
	return CLASSIFIABLE_TYPES.has(assetType);
}

/**
 * Etiqueta legible de un sector. Uno desconocido conserva su nombre crudo en
 * vez de desaparecer del gráfico: la fila representa dinero del usuario y
 * esconderla descuadraría el reparto.
 */
export function formatSector(sector: string): string {
	return SECTOR_LABELS[sector] ?? sector;
}

/**
 * Segunda línea de la fila, cuando la hay.
 *
 * Solo los dos cubos la llevan, y por el mismo motivo: sin ella, «Sin
 * clasificar, 24 %» se lee como una categoría de inversión y no como lo que es.
 */
export function sectorHint(sector: string): string {
	if (sector === 'unclassified') return 'Activos a los que les falta la industria';
	if (sector === 'not_applicable') return 'Cripto, efectivo e inmuebles no tienen industria';

	return '';
}

/**
 * Cuántas industrias hay detrás de un activo, en palabras.
 *
 * Es lo que va en la celda de un fondo de mercado ancho. Sin ella, un VOO se
 * vería igual que un activo sin clasificar —los dos traen `sector` vacío—, que
 * es justo el malentendido que el desglose existe para deshacer: uno es trabajo
 * pendiente y el otro es un fondo bien clasificado en once industrias.
 *
 * Una sola industria se escribe con su nombre: un desglose de una fila dice lo
 * mismo que el campo `sector` y no hay por qué contarlo.
 */
export function formatSectorBreakdown(weights: WeightedSector[]): string {
	if (weights.length === 0) return '';
	if (weights.length === 1) return formatSector(weights[0].sector);

	return `${weights.length} industrias`;
}

/**
 * El desglose entero, para el `title` de esa celda y para un panel de detalle:
 * «Tecnología 33,1 % · Finanzas 13,8 % · …».
 *
 * En el orden en que llega, que es el que manda el backend: de mayor a menor
 * peso, como la ficha del fondo.
 */
export function describeSectorBreakdown(weights: WeightedSector[]): string {
	return weights.map((w) => `${formatSector(w.sector)} ${formatPercent(w.weight)}`).join(' · ');
}

/**
 * La suma de los pesos, que **no** tiene por qué ser 100.
 *
 * La ficha de un fondo real deja unas décimas en caja y una transcripción a
 * medias deja más. El backend reparte la posición normalizando sobre este
 * total, así que enseñarlo es la forma de que quien lo está escribiendo vea si
 * el 97,3 % es la caja del fondo o una fila que se le olvidó.
 */
export function sectorWeightsTotal(weights: WeightedSector[]): number {
	return weights.reduce((total, w) => total + (Number.isFinite(w.weight) ? w.weight : 0), 0);
}
