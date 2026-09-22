import { createHash, randomBytes } from 'node:crypto';
import { env } from '$env/dynamic/private';

/**
 * La pasarela de Bold para los aportes de `/apoyar` (botón de pagos,
 * integración personalizada: developers.bold.co/pagos-en-linea).
 *
 * No es el backend de Finexia, así que no vive en `lib/api`: es un servicio
 * externo que solo se toca desde el servidor, porque la firma de integridad se
 * calcula con la llave secreta, que nunca sale de aquí.
 *
 * - `BOLD_API_KEY`: la llave de identidad. Bold la considera pública (viaja
 *   al navegador dentro de la configuración del checkout).
 * - `BOLD_SECRET_KEY`: la llave secreta. Solo firma.
 *
 * Las dos tienen versión de pruebas y de producción en el panel de Bold; el
 * entorno lo decide la pareja que se configure. Si falta alguna, los aportes
 * quedan apagados.
 */

const VOUCHER_URL = 'https://payments.api.bold.co/v2/payment-voucher';
export const BOLD_CURRENCY = 'COP';

/** Lo que el navegador le pasa a `new BoldCheckout(…)`. */
export interface BoldCheckoutOptions {
	orderId: string;
	currency: string;
	/** Entero en pesos, como texto: así lo pide Bold. */
	amount: string;
	apiKey: string;
	integritySignature: string;
	redirectionUrl: string;
	description: string;
}

/** Estados de `payment_status` que documenta la consulta de transacciones. */
export type BoldPaymentStatus =
	'APPROVED' | 'PROCESSING' | 'PENDING' | 'REJECTED' | 'FAILED' | 'VOIDED' | 'NO_TRANSACTION_FOUND';

export interface BoldPayment {
	status: BoldPaymentStatus;
	/** Total cobrado en pesos, si Bold lo devuelve. */
	total: number | null;
}

interface BoldKeys {
	apiKey: string;
	secretKey: string;
}

export function boldKeys(): BoldKeys | null {
	const apiKey = env.BOLD_API_KEY?.trim();
	const secretKey = env.BOLD_SECRET_KEY?.trim();
	return apiKey && secretKey ? { apiKey, secretKey } : null;
}

/** SHA-256 en hex de `{orderId}{amount}{currency}{secretKey}`, en ese orden. */
export function integritySignature(
	orderId: string,
	amount: number,
	currency: string,
	secretKey: string
): string {
	return createHash('sha256').update(`${orderId}${amount}${currency}${secretKey}`).digest('hex');
}

/**
 * Identificador único de la orden. Bold admite hasta 60 caracteres
 * alfanuméricos, `-` y `_`, y recomienda un timestamp para no repetir.
 */
export function newOrderId(now = Date.now(), suffix = randomBytes(4).toString('hex')): string {
	return `FNX-APOYO-${now}-${suffix}`;
}

/** Solo órdenes nuestras: el id llega por la URL de vuelta y va a un path. */
export function isOrderId(value: string | null): value is string {
	return value !== null && /^FNX-APOYO-\d{13}-[0-9a-f]{8}$/.test(value);
}

export function createCheckout(
	keys: BoldKeys,
	amount: number,
	redirectionUrl: string
): BoldCheckoutOptions {
	const orderId = newOrderId();
	return {
		orderId,
		currency: BOLD_CURRENCY,
		amount: String(amount),
		apiKey: keys.apiKey,
		integritySignature: integritySignature(orderId, amount, BOLD_CURRENCY, keys.secretKey),
		redirectionUrl,
		description: 'Aporte a Finexia'
	};
}

const STATUSES = new Set<string>([
	'APPROVED',
	'PROCESSING',
	'PENDING',
	'REJECTED',
	'FAILED',
	'VOIDED',
	'NO_TRANSACTION_FOUND'
]);

/**
 * El estado real de una orden, preguntado a Bold con la llave de identidad.
 * El `bold-tx-status` de la URL de vuelta lo puede escribir cualquiera; este
 * no. `null` si Bold no responde o responde algo que no se entiende.
 */
export async function fetchPayment(
	fetch: typeof globalThis.fetch,
	keys: BoldKeys,
	orderId: string
): Promise<BoldPayment | null> {
	try {
		const response = await fetch(`${VOUCHER_URL}/${encodeURIComponent(orderId)}`, {
			headers: { Authorization: `x-api-key ${keys.apiKey}` },
			signal: AbortSignal.timeout(8000)
		});
		if (!response.ok) return null;
		const body: unknown = await response.json();
		if (typeof body !== 'object' || body === null) return null;
		const { payment_status: status, total } = body as Record<string, unknown>;
		if (typeof status !== 'string' || !STATUSES.has(status)) return null;
		return {
			status: status as BoldPaymentStatus,
			total: typeof total === 'number' ? total : null
		};
	} catch {
		return null;
	}
}
