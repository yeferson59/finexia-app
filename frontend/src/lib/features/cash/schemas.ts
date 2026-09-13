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
