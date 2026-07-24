/**
 * Constantes y helpers de la feature `platforms`, compartidos por el detalle y
 * el formulario de alta (antes duplicados en cada página).
 *
 * El contrato `Platform` no se redeclara: se reexporta el de `$lib/api/types`
 * (única fuente de verdad) para que los componentes no dependan de la capa de API.
 */
import type { Platform } from '$lib/api/types';

export type { Platform };

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
