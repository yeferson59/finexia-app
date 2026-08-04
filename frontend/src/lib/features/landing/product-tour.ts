/**
 * Datos del recorrido "Dentro de Finexia" de la landing.
 *
 * Son cifras de ejemplo, no datos reales: viven aquí en vez de dentro de los
 * componentes para que las cuatro maquetas cuadren entre sí (el patrimonio del
 * resumen es la suma de los portafolios, y el rendimiento es el mismo en todas
 * las vistas). Si cambia una cifra, cambia en un solo sitio.
 */

/** Entradas del menú lateral del dashboard, en el mismo orden que la app. */
export const TOUR_NAV = [
	'Dashboard',
	'Portafolios',
	'Plataformas',
	'Transacciones',
	'Reportes',
	'Notificaciones',
	'Configuración'
] as const;

export type TourViewId = 'resumen' | 'portafolios' | 'transacciones' | 'reportes';

export interface TourView {
	id: TourViewId;
	/** Etiqueta de la pestaña. */
	tab: string;
	/** Entrada del menú lateral que queda marcada como activa. */
	nav: (typeof TOUR_NAV)[number];
	/** Migas de pan de la barra superior de la maqueta. */
	crumb: string;
	/** Titular del pie que describe la vista. */
	title: string;
	/** Qué resuelve la vista, en una frase. */
	description: string;
	/** Capacidades concretas, como etiquetas. */
	points: string[];
}

export const TOUR_VIEWS: TourView[] = [
	{
		id: 'resumen',
		tab: 'Resumen',
		nav: 'Dashboard',
		crumb: 'Dashboard / Resumen',
		title: 'Tu patrimonio agregado, en una pantalla',
		description:
			'El total de todos tus portafolios, cuánto has ganado o perdido sobre el capital invertido y cómo ha evolucionado desde que empezaste a registrar.',
		points: [
			'Patrimonio neto',
			'Valor de mercado vs. capital invertido',
			'Rangos 1M · 6M · 1A · Todo',
			'Ver en USD o COP',
			'Modo oculto'
		]
	},
	{
		id: 'portafolios',
		tab: 'Portafolios',
		nav: 'Portafolios',
		crumb: 'Dashboard / Portafolios',
		title: 'Los portafolios que tú defines',
		description:
			'Agrupa activos de plataformas distintas bajo el portafolio que tengas en mente, con su meta, su perfil de riesgo y el peso de cada posición.',
		points: [
			'Metas y perfil de riesgo',
			'Peso por posición',
			'Ganancia por portafolio',
			'Activos de varias plataformas',
			'Divisa propia por portafolio'
		]
	},
	{
		id: 'transacciones',
		tab: 'Transacciones',
		nav: 'Transacciones',
		crumb: 'Dashboard / Transacciones',
		title: 'Cada movimiento, registrado por ti',
		description:
			'Compras, ventas, dividendos, intereses, traspasos y comisiones. Uno a uno o importando un CSV, con vista previa antes de confirmar nada.',
		points: [
			'8 tipos de operación',
			'Importación CSV',
			'Mapeo de columnas',
			'Vista previa antes de confirmar',
			'Historial por activo'
		]
	},
	{
		id: 'reportes',
		tab: 'Reportes',
		nav: 'Reportes',
		crumb: 'Dashboard / Reportes',
		title: 'Reportes que puedes llevarte',
		description:
			'Rentabilidad mes a mes, estadísticas de tu cartera y proyección a cinco años. Todo descargable en XLSX, porque los datos son tuyos.',
		points: [
			'Rentabilidad mensual',
			'Estadísticas clave',
			'Proyección a 5 años',
			'Descarga en XLSX',
			'Sin bloqueos de exportación'
		]
	}
];

/** Resumen: cifras de la tarjeta de patrimonio neto. */
export const TOUR_NET_WORTH = {
	total: '$248.500,00',
	delta: '+$27.400,00 · 12,40% total',
	stats: [
		{ label: 'Portafolios', value: '3' },
		{ label: 'Posiciones', value: '18' },
		{ label: 'Ganancia', value: '+12,40%' }
	]
};

/** Resumen: serie del gráfico de crecimiento (porcentaje de altura, 0–100). */
export const TOUR_GROWTH_SERIES = [18, 24, 21, 32, 38, 35, 46, 52, 49, 61, 68, 74, 82];

/** Portafolios y transacciones: filas de ejemplo de las tablas. */
export const TOUR_PORTFOLIOS = [
	{ name: 'Jubilación', kind: 'Largo plazo', value: '$132.400', delta: '+14,20%', up: true },
	{ name: 'Cripto', kind: 'Alto riesgo', value: '$68.900', delta: '+21,80%', up: true },
	{ name: 'Reserva', kind: 'Conservador', value: '$47.200', delta: '−1,40%', up: false }
];

export const TOUR_HOLDINGS = [
	{
		symbol: 'VWCE',
		name: 'Vanguard FTSE All-World',
		weight: 38,
		value: '$94.430',
		delta: '+11,2%'
	},
	{ symbol: 'BTC', name: 'Bitcoin', weight: 29, value: '$72.065', delta: '+24,6%' },
	{ symbol: 'AAPL', name: 'Apple Inc.', weight: 23, value: '$57.155', delta: '+6,8%' },
	{ symbol: 'CASH', name: 'Efectivo USD', weight: 10, value: '$24.850', delta: '0,0%' }
];

export const TOUR_TRANSACTIONS = [
	{ type: 'Compra', asset: 'VWCE', platform: 'Degiro', date: '12 mar', amount: '−$2.400,00' },
	{ type: 'Dividendo', asset: 'AAPL', platform: 'Degiro', date: '08 mar', amount: '+$184,20' },
	{ type: 'Compra', asset: 'BTC', platform: 'Binance', date: '02 mar', amount: '−$1.100,00' },
	{ type: 'Venta', asset: 'TSLA', platform: 'Revolut', date: '27 feb', amount: '+$3.260,00' },
	{ type: 'Interés', asset: 'CASH', platform: 'Revolut', date: '25 feb', amount: '+$46,80' }
];

/** Reportes: rentabilidad mensual del año en curso, en porcentaje. */
export const TOUR_CALENDAR = [
	{ month: 'Ene', value: 2.4 },
	{ month: 'Feb', value: 1.1 },
	{ month: 'Mar', value: -0.8 },
	{ month: 'Abr', value: 3.2 },
	{ month: 'May', value: 0.6 },
	{ month: 'Jun', value: 1.9 },
	{ month: 'Jul', value: -1.4 },
	{ month: 'Ago', value: 2.8 },
	{ month: 'Sep', value: 0.3 },
	{ month: 'Oct', value: 1.6 },
	{ month: 'Nov', value: 2.2 },
	{ month: 'Dic', value: null }
];

export const TOUR_KEY_STATS = [
	{ label: 'Mejor mes', value: '+3,20%' },
	{ label: 'Peor mes', value: '−1,40%' },
	{ label: 'Meses en positivo', value: '9 de 11' },
	{ label: 'Crecimiento anual', value: '+12,40%' }
];

export const TOUR_REPORTS = [
	{ title: 'Resumen mensual', format: 'XLSX' },
	{ title: 'Estado de resultados', format: 'XLSX' },
	{ title: 'Riesgo y volatilidad', format: 'XLSX' }
];

/**
 * Tramo de color de una celda del calendario, con los mismos cortes que usa el
 * dashboard en `features/reports`.
 */
export function tourPerformanceClass(value: number | null): string {
	if (value === null) return 'mk-cal-empty';
	if (value >= 2) return 'mk-cal-strong-up';
	if (value >= 0) return 'mk-cal-up';
	if (value > -1) return 'mk-cal-down';
	return 'mk-cal-strong-down';
}
