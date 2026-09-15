/**
 * Schemas Zod de los formularios de efectivo (`routes/dashboard/cash`).
 */

import { z } from 'zod';
import { SUPPORTED_CURRENCIES } from '$lib/shared/currency';

const movementFields = {
	kind: z.enum(['deposit', 'withdrawal', 'interest'], 'Elige qué tipo de movimiento es.'),
	amount: z.coerce
		.number('Escribe el importe con números.')
		.positive('El importe tiene que ser mayor que cero.'),
	// Un campo vacío llega como cadena vacía, que es una comisión de cero.
	fees: z.coerce
		.number('Escribe la comisión con números.')
		.nonnegative('La comisión no puede ser negativa.')
		.default(0),
	// Lo que envía el selector de fecha: día, sin hora ni zona.
	date: z.iso.date('Elige la fecha del movimiento.'),
	notes: z
		.string()
		.max(500, 'La nota no puede pasar de 500 caracteres.')
		.nullish()
		.transform((v) => (v ?? '').trim())
};

/**
 * Las dos reglas que dependen de más de un campo, y que el backend también
 * aplica: dicen lo mismo aquí, antes del viaje, y en español.
 *
 * - Los intereses no llevan comisión. Se guardan sin flujo de dinero, así que
 *   una comisión sobre ellos no se restaría de nada y desaparecería.
 * - La comisión de un retiro no puede superar el retiro: pasado ese punto, el
 *   retiro se convertiría en dinero que metiste.
 */
function checkFees(v: { kind: string; amount: number; fees: number }, ctx: z.RefinementCtx) {
	if (v.kind === 'interest' && v.fees > 0) {
		ctx.addIssue({
			code: 'custom',
			path: ['fees'],
			message: 'Los intereses se anotan netos, sin comisión.'
		});
	}

	if (v.kind === 'withdrawal' && v.fees > v.amount) {
		ctx.addIssue({
			code: 'custom',
			path: ['fees'],
			message: 'La comisión no puede ser mayor que el retiro.'
		});
	}
}

/** Alta de un movimiento: también dice sobre qué saldo cae. */
export const cashMovementCreateSchema = z
	.object({
		...movementFields,
		portfolioId: z.uuid('Elige el portafolio.'),
		sourceId: z.uuid('Elige la plataforma.'),
		currency: z.enum(SUPPORTED_CURRENCIES, 'Elige la moneda.')
	})
	.superRefine(checkFees);

/** Edición: el saldo no cambia, así que no se piden portafolio, plataforma ni moneda. */
export const cashMovementUpdateSchema = z
	.object({ id: z.uuid('No sabemos qué movimiento editar.'), ...movementFields })
	.superRefine(checkFees);

export const cashMovementDeleteSchema = z.object({
	id: z.uuid('No sabemos qué movimiento borrar.')
});

type MovementFields = Pick<
	z.infer<typeof cashMovementUpdateSchema>,
	'kind' | 'amount' | 'fees' | 'date' | 'notes'
>;

/**
 * El cuerpo que espera el backend para los campos comunes.
 *
 * La fecha viaja como medianoche UTC de ese día, que es como el backend la
 * guarda: `new Date('2026-09-10')` daría lo mismo, pero en una zona al oeste de
 * Greenwich `toISOString` de una fecha local caería en el día anterior.
 */
export function toCashMovementBody(data: MovementFields) {
	return {
		kind: data.kind,
		amount: data.amount,
		fees: data.kind === 'interest' ? 0 : data.fees,
		date: `${data.date}T00:00:00Z`,
		notes: data.notes
	};
}

/** Si `value` no tiene más decimales de los que caben, sin tropezar con la coma flotante. */
function hasAtMostDecimals(value: number, decimals: number): boolean {
	const scaled = value * 10 ** decimals;
	return Math.abs(scaled - Math.round(scaled)) < 1e-6;
}

/**
 * Los valores de una versión de la tasa, en porcentaje como los publica la
 * entidad: `9.25` es 9,25 % E.A. Los decimales son los que guarda el backend.
 */
const rateValues = {
	annualRatePct: z.coerce
		.number('Escribe la tasa con números.')
		.positive('La tasa tiene que ser mayor que cero.')
		.max(100, 'La tasa no puede pasar de 100 %.')
		.refine((v) => hasAtMostDecimals(v, 4), 'Escribe la tasa con hasta cuatro decimales.'),
	// Vacía es que no hay retención.
	withholdingPct: z.coerce
		.number('Escribe la retención con números.')
		.min(0, 'La retención no puede ser negativa.')
		.lt(100, 'La retención tiene que ser menor que 100 %.')
		.refine((v) => hasAtMostDecimals(v, 2), 'Escribe la retención con hasta dos decimales.')
		.default(0),
	/** Cada cuánto la entidad abona lo que la cuenta rinde. */
	posting: z
		.enum(['daily', 'monthly'], 'Elige cada cuánto se abonan los intereses.')
		.default('daily'),
	/**
	 * Lo máximo sobre lo que rinde la cuenta. Vacío es que rinde sobre todo, y
	 * llega como cadena vacía desde el formulario, así que se lee aparte de
	 * `z.coerce.number`, que la convertiría en cero.
	 */
	maxBalance: z
		.union([z.string(), z.number(), z.null()])
		.optional()
		.transform((v) => (v === '' || v === null || v === undefined ? null : Number(v)))
		.refine((v) => v === null || Number.isFinite(v), 'Escribe el tope con números.')
		.refine((v) => v === null || v > 0, 'El tope tiene que ser mayor que cero.')
		.refine((v) => v === null || v < 1e12, 'El tope es demasiado grande.')
		.refine(
			(v) => v === null || hasAtMostDecimals(v, 8),
			'Escribe el tope con hasta ocho decimales.'
		)
};

/** Una tasa, o una versión nueva: la cuenta y el día desde el que rige. */
export const cashRateCreateSchema = z.object({
	sourceId: z.uuid('No sabemos a qué cuenta darle la tasa.'),
	currency: z.enum(SUPPORTED_CURRENCIES, 'No sabemos a qué cuenta darle la tasa.'),
	effectiveFrom: z.iso.date('Elige desde qué día rige la tasa.'),
	...rateValues
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
 * Una corrección dice la versión entera, así que un tope que se deja vacío lo
 * quita: viaja como `null`, no se omite.
 */
export function toCashRateBody(data: {
	annualRatePct: number;
	withholdingPct: number;
	posting: 'daily' | 'monthly';
	maxBalance: number | null;
}) {
	return {
		annualRatePct: data.annualRatePct,
		withholdingPct: data.withholdingPct,
		posting: data.posting,
		maxBalance: data.maxBalance
	};
}

/**
 * Recalcular los intereses de una cuenta desde un día: se tiran los que ya
 * estaban calculados y se vuelven a calcular sobre lo que guarda hoy.
 */
export const cashRecalculateSchema = z.object({
	sourceId: z.uuid('No sabemos de qué cuenta recalcular los intereses.'),
	currency: z.enum(SUPPORTED_CURRENCIES, 'No sabemos de qué cuenta recalcular los intereses.'),
	from: z.iso.date('Elige desde qué día recalcular.')
});

/** Un día del selector como lo guarda el backend: medianoche UTC de ese día. */
export function toCalendarDateTime(date: string): string {
	return `${date}T00:00:00Z`;
}
