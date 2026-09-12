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
 */

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
