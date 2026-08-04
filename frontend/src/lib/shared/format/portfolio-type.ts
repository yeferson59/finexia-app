/**
 * Etiquetas legibles del `type` de un portafolio.
 *
 * Vivían en `features/portfolio`, pero el resumen del dashboard también las
 * necesita —pintaba el valor crudo, `stocks_etfs`— y una feature no importa de
 * otra. Es presentación pura de una cadena del backend, así que baja a shared.
 */

/** Combinaciones de clases de activo que devuelve el backend. */
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
