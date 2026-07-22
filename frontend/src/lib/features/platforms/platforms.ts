/**
 * Constantes y helpers de la feature `platforms`, compartidos por el detalle y
 * el formulario de alta (antes duplicados en cada página).
 */

/** Etiquetas legibles para el `sourceType` de una plataforma. */
export const PLATFORM_TYPES = new Map<string, string>([
	['broker', 'Bróker'],
	['investment_bank', 'Banco de Inversión'],
	['trading_platform', 'Plataforma de Trading'],
	['neobank', 'NeoBank'],
	['de_fi', 'DeFi'],
	['crypto_wallet', 'Billetera Cripto'],
	['mutual_funds', 'Fondos Mutuos'],
	['brokerage_house', 'Casa de Bolsa'],
	['other', 'Otro']
]);

export function formatSourceType(type: string): string {
	return PLATFORM_TYPES.get(type) ?? type;
}

/** Plataforma de inversión tal como la devuelve `lib/api/platforms`. */
export interface Platform {
	id: string;
	name: string;
	description: string;
	sourceType: string;
	isActive: boolean;
	createdAt: string;
	updatedAt: string;
	investments: number;
	totalValue: string;
}
