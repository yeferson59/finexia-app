/**
 * Los aportes de `/apoyar`: los montos que se ofrecen y cómo se lee el
 * resultado que Bold devuelve al volver de la pasarela.
 *
 * La firma, las llaves y la consulta a Bold viven en `$lib/server/bold`; aquí
 * solo lo que también necesita el navegador.
 */

/** Bold no cobra menos de $1.000 COP. */
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

const OUTCOMES: Record<string, PaymentOutcome> = {
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
 * Manda el estado que Bold responde a la consulta (`confirmed`). El
 * `bold-tx-status` de la URL solo se usa si Bold no contesta, y nunca para
 * dar un pago por aprobado: esa URL la puede escribir cualquiera.
 */
export function supportResult(
	confirmed: string | null,
	total: number | null,
	fromUrl: string | null
): SupportResult | null {
	if (confirmed) {
		const outcome = OUTCOMES[confirmed];
		return outcome ? { outcome, total } : null;
	}
	const claimed = OUTCOMES[(fromUrl ?? '').toUpperCase()];
	if (!claimed) return null;
	return { outcome: claimed === 'failed' ? 'failed' : 'pending', total: null };
}
