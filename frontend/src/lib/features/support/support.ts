/**
 * Los aportes de `/apoyar`: los montos que se ofrecen y cómo se lee el
 * resultado al volver de la pasarela.
 *
 * La firma, las llaves, el guardado de cada orden y el webhook de Bold viven en
 * el backend (`docs/API.md` §2.13); aquí solo lo que también necesita el
 * navegador.
 */
import type { SupportStatus } from '$lib/api/types';

/** Bold no cobra menos de $1.000 COP. Los dos topes repiten los del backend. */
export const MIN_AMOUNT = 1_000;
/** Un techo propio, por debajo de los límites de la cuenta de Bold. */
export const MAX_AMOUNT = 2_000_000;

export const AMOUNT_PRESETS = [10_000, 20_000, 50_000, 100_000] as const;
export const DEFAULT_AMOUNT = 20_000;

/** «20.000», «$ 20 000» o «20000» → 20000. Sin dígitos, `NaN`. */
export function parsePesos(raw: string): number {
	const digits = raw.replace(/\D/g, '');
	return digits ? Number(digits) : Number.NaN;
}

export type PaymentOutcome = 'approved' | 'pending' | 'failed';

export interface SupportResult {
	outcome: PaymentOutcome;
	/** Lo cobrado, en pesos, si Bold lo confirmó. */
	total: number | null;
}

/** Los estados que guarda el backend. `created` no está: aún no hay pago. */
const STORED: Partial<Record<SupportStatus, PaymentOutcome>> = {
	approved: 'approved',
	pending: 'pending',
	rejected: 'failed',
	voided: 'failed'
};

/** Lo que Bold escribe en `bold-tx-status` al devolver a quien paga. */
const CLAIMED: Record<string, PaymentOutcome> = {
	APPROVED: 'approved',
	PROCESSING: 'pending',
	PENDING: 'pending',
	REJECTED: 'failed',
	FAILED: 'failed',
	VOIDED: 'failed'
};

/**
 * El resultado que se le enseña a quien vuelve de Bold.
 *
 * Manda el estado que guarda el backend (`stored`), que ya viene de Bold: del
 * webhook o de preguntarle al volver. Si el backend no contesta o la orden
 * sigue en `created`, se usa el `bold-tx-status` de la URL, pero nunca para dar
 * un pago por aprobado: esa URL la puede escribir cualquiera.
 */
export function supportResult(
	stored: SupportStatus | null,
	total: number | null,
	fromUrl: string | null
): SupportResult | null {
	const outcome = stored ? STORED[stored] : undefined;
	if (outcome) return { outcome, total };

	const claimed = CLAIMED[(fromUrl ?? '').toUpperCase()];
	if (!claimed) return null;
	return { outcome: claimed === 'failed' ? 'failed' : 'pending', total: null };
}
