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

/** Una plataforma con su peso dentro de la cuenta y su puesto en el reparto. */
export interface PlatformShare {
	platform: Platform;
	/** Parte del total invertido de la cuenta, de 0 a 100. */
	share: number;
	/** Puesto en el reparto: 0 es la mayor. Fija el tono de su tramo. */
	rank: number;
}

/**
 * Ordena las plataformas por lo que guardan, de mayor a menor, y le pone a cada
 * una su parte de la cuenta.
 *
 * El listado enseñaba seis tarjetas del mismo tamaño para seis plataformas que
 * no tienen el mismo tamaño, y el `percent` —lo que el schema llama «lo que
 * hace legible el orden»— iba de nota al pie. Aquí pasa a ser el orden.
 *
 * `percent` lo informa el backend contra su propio total; cuando no viene se
 * calcula sobre la suma del listado, que son todas las plataformas de la
 * cuenta y por tanto el mismo denominador.
 */
export function rankByShare(platforms: Platform[]): PlatformShare[] {
	const amount = (p: Platform) => parseFloat(p.totalValue) || 0;
	const total = platforms.reduce((sum, p) => sum + amount(p), 0);

	return [...platforms]
		.sort((a, b) => amount(b) - amount(a))
		.map((platform, rank) => ({
			platform,
			share: platform.percent ?? (total > 0 ? (amount(platform) / total) * 100 : 0),
			rank
		}));
}

/**
 * Opacidad del ámbar con la que se pinta un tramo, según su puesto.
 *
 * Un color por plataforma pediría una paleta categórica que esta interfaz no
 * tiene; escalonar el mismo ámbar mantiene la barra como un solo objeto y deja
 * el orden legible, que es lo que se está contando. El suelo de 0.38 es para
 * que el tramo más pequeño siga viéndose sobre el fondo.
 */
export function shareTint(rank: number, count: number): number {
	if (count <= 1) return 1;
	return 1 - (Math.min(rank, count - 1) / (count - 1)) * 0.62;
}
