/**
 * Schemas Zod de los formularios de efectivo (`routes/dashboard/cash`): los
 * movimientos, los bolsillos, los traslados y los depósitos. Los de la tasa
 * están en `rate-schemas.ts`, y los campos que comparten, en `form-fields.ts`.
 */

import { z } from 'zod';
import { SUPPORTED_CURRENCIES } from '$lib/shared/currency';
import { pocketField, rateValues } from './form-fields';

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
		currency: z.enum(SUPPORTED_CURRENCIES, 'Elige la moneda.'),
		pocketId: pocketField
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

/**
 * Lo que el formulario de un movimiento manda, tal cual, para dárselo a su
 * schema. Vive junto al schema porque es su otra mitad: los nombres de los
 * campos y las reglas que los juzgan cambian a la vez.
 */
export function cashMovementFields(formData: FormData) {
	return {
		kind: formData.get('kind'),
		amount: formData.get('amount'),
		fees: formData.get('fees'),
		date: formData.get('date'),
		notes: formData.get('notes')
	};
}

/** El nombre de un cajón de la cuenta, como lo guarda el backend: recortado. */
const cashPocketName = z
	.string('Ponle un nombre al bolsillo.')
	.transform((v) => v.trim())
	.refine((v) => v.length > 0, 'Ponle un nombre al bolsillo.')
	.refine((v) => v.length <= 100, 'El nombre no puede pasar de 100 caracteres.');

/** Abrir un bolsillo flexible en una cuenta. */
export const cashPocketCreateSchema = z.object({
	sourceId: z.uuid('Elige la plataforma.'),
	currency: z.enum(SUPPORTED_CURRENCIES, 'Elige la moneda.'),
	name: cashPocketName
});

/** Lo único que cambia de un bolsillo es el nombre. */
export const cashPocketRenameSchema = z.object({
	id: z.uuid('No sabemos qué bolsillo renombrar.'),
	name: cashPocketName
});

export const cashPocketDeleteSchema = z.object({
	id: z.uuid('No sabemos qué bolsillo borrar.')
});

/**
 * Lo que llega al otro lado de un traslado, en la moneda de destino.
 *
 * Vacío es «no lo dijeron», no «cero». Dentro de una moneda no hace falta
 * —llega lo mismo que sale— y entre dos es obligatorio: sin él el importe
 * viajaría tal cual con otra etiqueta, que es la diferencia entre 400.000 pesos
 * y 400.000 dólares. Cuál de los dos casos es lo decide el objeto entero.
 *
 * Se pide el importe y no la tasa porque los dos importes son los que están en
 * el extracto —lo que salió y lo que entró— y la tasa entre ellos no. Guardarlo
 * tal cual además deja el saldo exacto: un importe calculado desde una tasa cae
 * a unos centavos de lo que de verdad llegó, y esos centavos se leerían como
 * ganancia.
 */
const arrivingAmount = z
	.union([z.string(), z.number(), z.null()])
	.nullish()
	.transform((v) => (v === null || v === undefined || v === '' ? undefined : Number(v)))
	.refine((v) => v === undefined || (Number.isFinite(v) && v > 0), {
		error: 'Lo que llega tiene que ser mayor que cero.'
	});

/**
 * Mover dinero entre dos saldos de un portafolio: dos cajones de una cuenta, o
 * dos cuentas distintas —de la app donde está el ahorro al bróker que va a
 * gastarlo—. Los bolsillos vacíos son la cuenta principal de su plataforma.
 *
 * Origen y destino no pueden ser el mismo sitio, pero «el mismo sitio» es
 * plataforma, moneda y cajón a la vez: la cuenta principal de dos plataformas
 * son dos sitios, aunque las dos manden el bolsillo vacío.
 */
export const cashMoveSchema = z
	.object({
		portfolioId: z.uuid('Elige el portafolio.'),
		sourceId: z.uuid('Elige la plataforma.'),
		currency: z.enum(SUPPORTED_CURRENCIES, 'Elige la moneda.'),
		fromPocketId: pocketField,
		toSourceId: z.uuid('Elige la plataforma a la que va.'),
		toCurrency: z.enum(SUPPORTED_CURRENCIES, 'Elige la moneda a la que llega.'),
		toPocketId: pocketField,
		amount: z.coerce
			.number('Escribe el importe con números.')
			.positive('El importe tiene que ser mayor que cero.'),
		toAmount: arrivingAmount,
		date: z.iso.date('Elige la fecha del movimiento.'),
		notes: z
			.string()
			.max(500, 'La nota no puede pasar de 500 caracteres.')
			.nullish()
			.transform((v) => (v ?? '').trim())
	})
	.refine(
		(v) =>
			v.toSourceId !== v.sourceId ||
			v.toCurrency !== v.currency ||
			(v.fromPocketId ?? '') !== (v.toPocketId ?? ''),
		{ path: ['toPocketId'], error: 'Elige un destino distinto del origen.' }
	)
	.superRefine((v, ctx) => {
		if (v.toCurrency === v.currency) {
			if (v.toAmount !== undefined && v.toAmount !== v.amount) {
				ctx.addIssue({
					code: 'custom',
					path: ['toAmount'],
					message: `Dentro de ${v.currency} llega lo mismo que sale.`
				});
			}
			return;
		}

		if (v.toAmount === undefined) {
			ctx.addIssue({
				code: 'custom',
				path: ['toAmount'],
				message: `Escribe cuántos ${v.toCurrency} llegaron.`
			});
		}
	});

/**
 * Un depósito a tasa fija: el dinero, el día en que se abrió, el plazo y la
 * tasa, todo en una sola escritura, porque son una misma cosa. La tasa no se le
 * da después al bolsillo: es lo que el bolsillo es.
 *
 * `openedOn` puede estar en el pasado —«lo abrí hace dos semanas»— y el backend
 * calcula de una vez los días que ya ganó. `maturesOn` vacío es un depósito sin
 * plazo.
 */
export const cashDepositCreateSchema = z
	.object({
		portfolioId: z.uuid('Elige el portafolio.'),
		sourceId: z.uuid('Elige la plataforma.'),
		currency: z.enum(SUPPORTED_CURRENCIES, 'Elige la moneda.'),
		name: cashPocketName,
		amount: z.coerce
			.number('Escribe el importe con números.')
			.positive('El importe tiene que ser mayor que cero.')
			.lt(1e12, 'El importe es demasiado grande.'),
		openedOn: z.iso.date('Elige el día en que abriste el depósito.'),
		maturesOn: z
			.union([z.iso.date('Elige la fecha de vencimiento.'), z.literal('')])
			.nullish()
			.transform((v) => v || undefined),
		annualRatePct: rateValues.annualRatePct,
		withholdingPct: rateValues.withholdingPct,
		posting: z
			.enum(['daily', 'monthly', 'at_maturity'], 'Elige cada cuánto se abonan los intereses.')
			.default('daily')
	})
	.refine((v) => !v.maturesOn || v.maturesOn > v.openedOn, {
		path: ['maturesOn'],
		error: 'El vencimiento tiene que ser posterior al día en que lo abriste.'
	})
	// Abonar al vencer necesita un vencimiento en el que abonar.
	.refine((v) => v.posting !== 'at_maturity' || !!v.maturesOn, {
		path: ['maturesOn'],
		error: 'Para abonar al vencimiento hay que ponerle un plazo.'
	});

/**
 * Cancelar un depósito antes de su plazo: el día en que el dinero vuelve a la
 * cuenta principal, y lo que cobra la entidad por romperlo.
 */
export const cashDepositCloseSchema = z.object({
	id: z.uuid('No sabemos qué depósito cancelar.'),
	// Hoy o ayer: un día futuro calcularía días que el depósito no ha vivido, y
	// uno pasado querría recalcular días que ya están cerrados.
	closesOn: z.iso.date('Elige el día en que lo cancelas.'),
	penalty: z.coerce
		.number('Escribe la penalidad con números.')
		.nonnegative('La penalidad no puede ser negativa.')
		.default(0)
});

/** Lo que se dice cuando un depósito no se puede abrir ni cancelar. */
export function cashDepositErrorMessage(status: number, details = ''): string {
	if (details.includes('takes no movements or rate versions')) {
		return 'Un depósito a tasa fija no admite movimientos ni cambios de tasa. Cancélalo si quieres sacar el dinero antes.';
	}
	if (details.includes('already closed')) {
		return 'Ese depósito ya está cerrado. Recarga la página.';
	}
	if (details.includes('already earned interest')) {
		const from = details.match(/cancelled from (\d{4}-\d{2}-\d{2})/)?.[1];
		return from
			? `Ya hay intereses calculados más allá de ese día. Cancélalo el ${from} o después.`
			: 'Ya hay intereses calculados más allá de ese día. Elige un día posterior.';
	}
	if (details.includes('closesOn cannot be in the future')) {
		return 'No puedes cancelarlo en un día que aún no ha pasado. Elige hoy.';
	}
	if (details.includes('closesOn cannot be before')) {
		return 'Elige hoy o ayer: los días pasados no se recalculan.';
	}
	if (details.includes('openedOn cannot be in the future')) {
		return 'La fecha de apertura no puede ser futura: pon el día en que metiste el dinero.';
	}
	if (details.includes('more than 5 years ago')) {
		return 'La fecha de apertura no puede ser de hace más de cinco años.';
	}
	if (details.includes('maturesOn must be after openedOn')) {
		return 'El vencimiento tiene que ser posterior al día en que lo abriste.';
	}
	if (details.includes('maturesOn cannot be before')) {
		return 'Ese depósito ya venció. Anótalo como un depósito y unos intereses en la cuenta, que es lo que te pagaron.';
	}
	if (details.includes('needs a maturesOn')) {
		return 'Para abonar al vencimiento hay que ponerle un plazo.';
	}
	if (details.includes('platform is inactive')) {
		return 'Esa plataforma está inactiva. Actívala para abrir un depósito en ella.';
	}
	if (details.includes('already has a pocket with that name')) {
		return 'Esa cuenta ya tiene un bolsillo con ese nombre. Ponle otro.';
	}
	if (status === 404) return 'No encontramos esa cuenta. Recarga la página.';

	return details || 'No pudimos guardar el depósito. Vuelve a intentarlo en un momento.';
}

/** Lo que se dice cuando un bolsillo no se puede guardar ni borrar. */
export function cashPocketErrorMessage(status: number, details?: string): string {
	if (status === 404) return 'Ese bolsillo ya no existe. Recarga la página.';
	if (status === 409 && (details ?? '').includes('name')) {
		return 'Esa cuenta ya tiene un bolsillo con ese nombre. Ponle otro.';
	}
	if (status === 409) {
		return 'Ese bolsillo todavía tiene movimientos. Mueve el dinero a la cuenta principal y borra sus movimientos antes de quitarlo.';
	}

	return details || 'No pudimos guardar el bolsillo. Vuelve a intentarlo en un momento.';
}

/** Un día del selector como lo guarda el backend: medianoche UTC de ese día. */
export function toCalendarDateTime(date: string): string {
	return `${date}T00:00:00Z`;
}
