/**
 * Schemas Zod de la tasa de una cuenta de efectivo: anotarla, moverla,
 * corregirla, pausarla, borrarla y recalcular sus intereses.
 */

import { z } from 'zod';
import { SUPPORTED_CURRENCIES } from '$lib/shared/currency';
import { pocketField, rateValues } from './form-fields';

/**
 * Los tramos llegan como dos listas paralelas, una entrada por fila. Una fila en
 * blanco no es un tramo, y el resto se ordena por saldo: el orden en que se
 * escribieron no cambia lo que dicen.
 */
function cashRateTierFields(formData: FormData) {
	const from = formData.getAll('tierFromBalance').map(String);
	const pct = formData.getAll('tierAnnualRatePct').map(String);

	return from
		.map((fromBalance, i) => ({ fromBalance, annualRatePct: pct[i] ?? '' }))
		.filter((tier) => tier.fromBalance.trim() !== '' || tier.annualRatePct.trim() !== '')
		.sort((a, b) => (parseFloat(a.fromBalance) || 0) - (parseFloat(b.fromBalance) || 0));
}

/** Los valores de una versión de la tasa como los manda su formulario. */
export function cashRateFields(formData: FormData) {
	return {
		annualRatePct: formData.get('annualRatePct'),
		withholdingPct: formData.get('withholdingPct'),
		posting: formData.get('posting') ?? 'daily',
		tiers: cashRateTierFields(formData)
	};
}

/**
 * El permiso para rehacer los intereses ya calculados desde el primer día de la
 * tasa. El formulario lo manda cuando ese día cae en días ya calculados, después
 * de decir en pantalla lo que va a pasar; sin él, el backend lo rechaza.
 */
const recomputeField = z
	.literal('true')
	.nullish()
	.transform((v) => v === 'true');

/**
 * Una tasa, o una versión nueva: la cuenta y el día desde el que rige. El día
 * puede ser pasado —la tasa que la cuenta ya rendía antes de anotarla— y los
 * intereses desde entonces se calculan al guardarla.
 */
export const cashRateCreateSchema = z.object({
	sourceId: z.uuid('No sabemos a qué cuenta darle la tasa.'),
	currency: z.enum(SUPPORTED_CURRENCIES, 'No sabemos a qué cuenta darle la tasa.'),
	pocketId: pocketField,
	effectiveFrom: z.iso.date('Elige desde qué día rige la tasa.'),
	recompute: recomputeField,
	...rateValues
});

/**
 * Mover el día en que empieza la versión más reciente: un cambio de tasa que la
 * entidad anunció para un día y aplicó otro, o una tasa anotada desde el día
 * equivocado. Si ya generó intereses, pide `recompute` para rehacerlos.
 */
export const cashRateRescheduleSchema = z.object({
	id: z.uuid('No sabemos qué tasa mover.'),
	effectiveFrom: z.iso.date('Elige desde qué día rige la tasa.'),
	recompute: recomputeField
});

/** Corrección de la versión más reciente: sus valores, no sus fechas. */
export const cashRateUpdateSchema = z.object({
	id: z.uuid('No sabemos qué tasa corregir.'),
	...rateValues
});

/** Pausa: el primer día en que la cuenta ya no rinde. */
export const cashRateEndSchema = z.object({
	id: z.uuid('No sabemos qué tasa pausar.'),
	endsOn: z.iso.date('Elige desde qué día deja de rendir.')
});

export const cashRateDeleteSchema = z.object({
	id: z.uuid('No sabemos qué tasa borrar.')
});

/**
 * Los valores como los espera el backend.
 *
 * Una corrección dice la versión entera, así que los tramos viajan siempre, y
 * una lista vacía los quita.
 */
export function toCashRateBody(data: {
	annualRatePct: number;
	withholdingPct: number;
	posting: 'daily' | 'monthly';
	tiers: { fromBalance: number; annualRatePct: number }[];
}) {
	return {
		annualRatePct: data.annualRatePct,
		withholdingPct: data.withholdingPct,
		posting: data.posting,
		tiers: data.tiers
	};
}

/**
 * Recalcular los intereses de una cuenta desde un día: se tiran los que ya
 * estaban calculados y se vuelven a calcular sobre lo que guarda hoy.
 */
export const cashRecalculateSchema = z.object({
	sourceId: z.uuid('No sabemos de qué cuenta recalcular los intereses.'),
	currency: z.enum(SUPPORTED_CURRENCIES, 'No sabemos de qué cuenta recalcular los intereses.'),
	pocketId: pocketField,
	from: z.iso.date('Elige desde qué día recalcular.')
});
