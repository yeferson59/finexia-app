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
			'Qué rindió tu dinero mes a mes, con qué riesgo, y adónde llega si el ritmo se mantiene. Todo descargable en XLSX, porque los datos son tuyos.',
		points: [
			'Rentabilidad mes a mes',
			'Volatilidad y máxima caída',
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

/** Abreviaturas de los meses, como en la matriz del dashboard. */
export const TOUR_MONTHS = [
	'Ene',
	'Feb',
	'Mar',
	'Abr',
	'May',
	'Jun',
	'Jul',
	'Ago',
	'Sep',
	'Oct',
	'Nov',
	'Dic'
];

/**
 * La cifra de cabecera de la ficha de reportes: lo que rindió el dinero en todo
 * el historial.
 *
 * Es el mismo +12,4 % que enseña el resumen, porque sale de encadenar los dos
 * años de la matriz de abajo: las cuatro vistas del recorrido cuentan la misma
 * cuenta.
 */
export const TOUR_RECORD = {
	label: 'Lo que rindió tu dinero',
	value: '+12,4%',
	span: 'Del 1 de enero de 2025 al 30 de noviembre de 2026, 23 meses de historial.',
	/* El patrimonio y la ganancia son los del resumen: es la misma cuenta. */
	money: 'Hoy la cuenta vale $248.500,00 sobre los $221.100,00 que has puesto: +$27.400,00.'
};

/**
 * Reportes: rentabilidad mes a mes, un año por fila y el total cerrando cada
 * uno. Diciembre de 2026 va sin dato porque el año sigue en curso.
 */
export const TOUR_RETURNS = [
	{
		year: '2026',
		values: [2.4, 1.1, -0.8, 1.6, 0.6, 1.2, -1.4, 2.6, 0.3, -1.1, -1.0, null],
		total: 5.5
	},
	{
		year: '2025',
		values: [0.8, 1.2, -1.1, 1.9, 0.5, -0.7, 1.4, 0.6, -1.8, 1.5, 0.9, 1.2],
		total: 6.5
	}
];

/** Las medidas de «Cómo se movió», sacadas de la matriz de arriba. */
export const TOUR_KEY_STATS = [
	{ label: 'Mejor mes', detail: 'agosto de 2026', value: '+2,6%' },
	{ label: 'Peor mes', detail: 'septiembre de 2025', value: '−1,8%' },
	{ label: 'Volatilidad anualizada', detail: '', value: '4,3%' },
	{ label: 'Máxima caída', detail: '', value: '−2,1%' },
	{ label: 'Ratio de Sharpe', detail: '', value: '1,43' }
];

/**
 * La proyección: la tasa anual del historial y adónde llega el saldo si se
 * mantiene. Sale de encadenar los dos años de la matriz sobre los 23 meses que
 * cubren, así que es el mismo +12,4 % visto por año.
 */
export const TOUR_PROJECTION = {
	rate: '+6,3% anual',
	columns: [
		{ period: '2026, hoy', value: '$248.500', accrued: '0,0%' },
		{ period: '2028', value: '$280.800', accrued: '+13,0%' },
		{ period: '2031', value: '$337.300', accrued: '+35,7%' }
	]
};

/** Los archivos descargables, cada uno con lo que trae dentro. */
export const TOUR_REPORTS = [
	{ title: 'Resumen mensual', description: 'Valor, capital y ganancia, mes a mes', format: 'XLSX' },
	{ title: 'Transacciones', description: 'Cada compra, venta y dividendo', format: 'XLSX' },
	{
		title: 'Riesgo y volatilidad',
		description: 'Las medidas y la serie diaria',
		format: 'XLSX'
	}
];

/**
 * Fondo de una celda de la matriz: verde lo que subió, rojo lo que bajó, y más
 * intenso cuanto mayor fue el movimiento.
 *
 * Espeja `returnBackground` de `features/reports`, con la misma escala y el
 * mismo punto de saturación. Está copiado y no importado porque una feature no
 * importa de otra (docs/FRONTEND_ARCHITECTURE.md); si allí cambia la escala,
 * aquí también.
 */
export function tourReturnBackground(value: number | null): string {
	if (value === null) return '';

	const intensity = Math.min(Math.abs(value) / 2.5, 1);
	const alpha = (0.05 + intensity * 0.2).toFixed(3);

	return value < 0 ? `rgba(224, 90, 90, ${alpha})` : `rgba(34, 201, 126, ${alpha})`;
}
