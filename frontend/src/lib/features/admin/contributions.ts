/**
 * Los aportes de `/apoyar` vistos desde la administración: cómo se nombra cada
 * estado y medio de pago, y la frase que abre la pantalla.
 *
 * Aparte de `admin.ts` y `desk.ts` porque es otro dominio con su propio
 * vocabulario: aquí no hay nada que envejezca, solo pagos y su estado.
 */

import type { SupportStatus, SupportSummary } from '$lib/api/types';
import { formatCurrency } from '$lib/shared/format/money';
import type { BadgeTone } from './admin';
import { plural } from './desk';

export type { AdminContribution, SupportStatus, SupportSummary } from '$lib/api/types';

/** Los estados, en el orden en que los recorre una orden, con su nombre. */
export const CONTRIBUTION_STATUSES: { value: SupportStatus; label: string }[] = [
	{ value: 'approved', label: 'Aprobados' },
	{ value: 'pending', label: 'En proceso' },
	{ value: 'rejected', label: 'Rechazados' },
	{ value: 'voided', label: 'Anulados' },
	{ value: 'created', label: 'Sin pagar' }
];

const STATUS_LABELS: Record<SupportStatus, string> = {
	created: 'Sin pagar',
	pending: 'En proceso',
	rejected: 'Rechazado',
	approved: 'Aprobado',
	voided: 'Anulado'
};

/*
 * Solo lo aprobado va en verde, que es lo que se cuenta. Un rechazo no es un
 * error del sistema —la persona puede reintentar—, así que va en el ámbar de lo
 * que está a medias y no en rojo; el rojo queda para lo que se deshizo después
 * de cobrarse.
 */
const STATUS_TONES: Record<SupportStatus, BadgeTone> = {
	created: 'neutral',
	pending: 'amber',
	rejected: 'warning',
	approved: 'success',
	voided: 'danger'
};

export function contributionStatusLabel(status: SupportStatus): string {
	return STATUS_LABELS[status] ?? status;
}

export function contributionStatusTone(status: SupportStatus): BadgeTone {
	return STATUS_TONES[status] ?? 'neutral';
}

/** Lee un `?status=` de la URL; cualquier otra cosa es «todos». */
export function parseContributionStatus(raw: string | null): SupportStatus | undefined {
	return CONTRIBUTION_STATUSES.find((s) => s.value === raw)?.value;
}

/*
 * Los medios que Bold escribe en `payment_method`. El webhook y la consulta no
 * usan exactamente el mismo vocabulario (`CARD` frente a `CREDIT_CARD`), así que
 * los dos llevan al mismo nombre; uno que no está aquí se enseña tal cual.
 */
const PAYMENT_METHODS: Record<string, string> = {
	CARD: 'Tarjeta',
	CREDIT_CARD: 'Tarjeta',
	DEBIT_CARD: 'Tarjeta débito',
	PSE: 'PSE',
	NEQUI: 'Nequi',
	DAVIPLATA: 'Daviplata',
	BOTON_BANCOLOMBIA: 'Botón Bancolombia',
	BANCOLOMBIA_TRANSFER: 'Botón Bancolombia'
};

/** El medio de pago con su nombre; `—` si aún no hay pago. */
export function paymentMethodLabel(method: string): string {
	if (!method) return '—';
	return PAYMENT_METHODS[method.toUpperCase()] ?? method;
}

const cop = (value: number) => formatCurrency(value, 'COP');

/**
 * La frase que abre la pantalla: cuánto se ha aportado y si hay algo en vuelo.
 *
 * Lo sin pagar no se menciona: son pasarelas que alguien abrió y cerró, y la
 * conciliación con Bold las borra a los siete días. Contarlas aquí haría que la frase
 * hablara sobre todo de ellas.
 */
export function describeContributions(summary: SupportSummary): string {
	const approved = summary.counts.approved ?? 0;
	const pending = summary.counts.pending ?? 0;

	const head =
		approved === 0
			? 'Todavía no hay aportes aprobados.'
			: `${cop(summary.approvedTotal)} en ${plural(approved, 'aporte aprobado', 'aportes aprobados')}` +
				(summary.approvedRecent > 0
					? `, ${cop(summary.approvedRecent)} de ellos en los últimos 30 días.`
					: ', ninguno en los últimos 30 días.');

	if (pending === 0) return head;

	return `${head} ${plural(pending, 'pago sigue', 'pagos siguen')} en proceso en Bold.`;
}
