/**
 * Schemas Zod de los formularios de fondos (`routes/dashboard/funds`) y los
 * mensajes con que se explica un rechazo del backend.
 *
 * Un fondo se sigue de una de dos formas, elegida al crearlo: por unidades (el
 * extracto trae unidades y valor de unidad) o por saldo (la app solo enseña el
 * saldo). Los formularios de un fondo por saldo hablan solo de dinero; las
 * unidades las calcula el backend.
 */
import { z } from 'zod';
import { SUPPORTED_CURRENCIES } from '$lib/shared/currency';

/** Un número positivo que cabe en las columnas del backend (NUMERIC(20, 8)). */
function positive(messages: { type: string; positive: string; tooBig: string }) {
	return z.coerce.number(messages.type).positive(messages.positive).lt(1e12, messages.tooBig);
}

const unitsField = positive({
	type: 'Escribe las unidades con números.',
	positive: 'Las unidades tienen que ser más que cero.',
	tooBig: 'Son demasiadas unidades.'
});

const unitValueField = positive({
	type: 'Escribe el valor de unidad con números.',
	positive: 'El valor de unidad tiene que ser mayor que cero.',
	tooBig: 'El valor de unidad es demasiado grande.'
});

const amountField = positive({
	type: 'Escribe el importe con números.',
	positive: 'El importe tiene que ser mayor que cero.',
	tooBig: 'El importe es demasiado grande.'
});

const balanceField = positive({
	type: 'Escribe el saldo con números.',
	positive: 'El saldo tiene que ser mayor que cero.',
	tooBig: 'El saldo es demasiado grande.'
});

/** Un campo opcional: vacío o ausente es que no se dijo. */
const blankToUndefined = (v: unknown) => (v === '' || v === null ? undefined : v);

const notesField = z
	.string()
	.max(500, 'La nota no puede pasar de 500 caracteres.')
	.nullish()
	.transform((v) => (v ?? '').trim());

/** Una casilla: marcada llega como «on», sin marcar no llega. */
const checkboxField = z.preprocess((v) => v === 'on' || v === 'true' || v === true, z.boolean());

const fundIdentity = {
	portfolioId: z.uuid('Elige el portafolio.'),
	sourceId: z.uuid('Elige la plataforma.'),
	currency: z.enum(SUPPORTED_CURRENCIES, 'Elige la moneda.'),
	name: z
		.string('Ponle un nombre.')
		.trim()
		.min(1, 'Ponle un nombre.')
		.max(100, 'El nombre no puede pasar de 100 caracteres.')
};

/** Un fondo del catálogo de la Superfinanciera: sus cinco códigos, «5-31-3644-1-501». */
const publicFundIdField = z
	.string('Elige el fondo de la Superfinanciera.')
	.regex(/^\d+(-\d+){4}$/, 'Elige el fondo de la Superfinanciera.');

/**
 * Alta de un fondo por unidades: qué es, dónde está y la primera compra. Lo que
 * vale hoy una unidad es opcional; sin eso el fondo vale lo que costó hasta la
 * primera marca.
 *
 * Enlazado a un fondo de la Superfinanciera (`publicFundId`), el valor de unidad
 * de la compra también es opcional: el backend toma el publicado ese día.
 */
export const fundCreateSchema = z
	.object({
		...fundIdentity,
		date: z.iso.date('Elige el día en que compraste las unidades.'),
		units: unitsField,
		unitValue: z.preprocess(blankToUndefined, unitValueField.optional()),
		currentUnitValue: z.preprocess(blankToUndefined, unitValueField.optional()),
		currentDate: z.preprocess(blankToUndefined, z.iso.date().optional()),
		publicFundId: z.preprocess(blankToUndefined, publicFundIdField.optional())
	})
	.refine((v) => v.unitValue !== undefined || v.publicFundId !== undefined, {
		path: ['unitValue'],
		error: 'Escribe el valor de unidad de la compra.'
	})
	.refine((v) => !v.publicFundId || v.currency === 'COP', {
		path: ['currency'],
		error: 'Los valores de la Superfinanciera están en pesos: el fondo tiene que ser en COP.'
	})
	.refine((v) => !v.currentUnitValue || !v.currentDate || v.currentDate >= v.date, {
		path: ['currentDate'],
		error: 'El valor de hoy no puede ser de antes de la compra.'
	});

/** Enlazar un fondo a uno de la Superfinanciera. */
export const fundLinkSchema = z.object({
	id: z.uuid('No sabemos qué fondo enlazar.'),
	publicFundId: publicFundIdField
});

/** Deshacer el enlace de un fondo. */
export const fundUnlinkSchema = z.object({
	id: z.uuid('No sabemos qué fondo desenlazar.')
});

/**
 * Alta de un fondo por saldo: lo que metiste desde un día y, si lo sabes, lo que
 * la app dice que tiene hoy. Es el arranque rápido; los aportes y retiros
 * siguientes se anotan después.
 */
export const fundBalanceCreateSchema = z
	.object({
		...fundIdentity,
		date: z.iso.date('Elige desde cuándo tienes el fondo.'),
		amount: amountField,
		currentBalance: z.preprocess(blankToUndefined, balanceField.optional()),
		currentDate: z.preprocess(blankToUndefined, z.iso.date().optional())
	})
	.refine((v) => !v.currentBalance || !v.currentDate || v.currentDate >= v.date, {
		path: ['currentDate'],
		error: 'El saldo de hoy no puede ser de antes del primer aporte.'
	});

/**
 * Lo que valía un fondo un día, leído del extracto: el valor de unidad de uno
 * por unidades, o el saldo de uno por saldo. Uno de los dos, nunca ambos.
 */
export const fundMarkSchema = z
	.object({
		id: z.uuid('No sabemos qué fondo actualizar.'),
		date: z.iso.date('Elige el día del extracto.'),
		unitValue: z.preprocess(blankToUndefined, unitValueField.optional()),
		balance: z.preprocess(blankToUndefined, balanceField.optional()),
		notes: notesField
	})
	.refine((v) => (v.unitValue === undefined) !== (v.balance === undefined), {
		path: ['unitValue'],
		error: 'Escribe el valor que trae tu extracto.'
	});

/**
 * Varias marcas de una vez: la tabla ya leída en el navegador
 * (`parseMarksTable`), que viaja como JSON en un campo oculto.
 */
export const fundMarksBulkSchema = z.object({
	id: z.uuid('No sabemos qué fondo actualizar.'),
	tracking: z.enum(['units', 'balance']),
	marks: z.preprocess(
		(v) => {
			try {
				return typeof v === 'string' ? JSON.parse(v) : v;
			} catch {
				return undefined;
			}
		},
		z
			.array(
				z.object({ date: z.iso.date(), value: z.number().positive().lt(1e12) }),
				'Pega al menos un valor.'
			)
			.min(1, 'Pega al menos un valor.')
			.max(400, 'Son demasiados valores para una vez: pega como mucho 400.')
	)
});

export const fundMarkDeleteSchema = z.object({
	id: z.uuid('No sabemos de qué fondo es la marca.'),
	date: z.iso.date('No sabemos qué marca borrar.')
});

export const fundDeleteSchema = z.object({
	id: z.uuid('No sabemos qué fondo quitar.')
});

/**
 * Un aporte a un fondo por saldo. `balanceBefore`, opcional, es lo que tenía el
 * fondo al cierre del día anterior: con él, el aporte entra al valor exacto de
 * ese día y no al del último saldo anotado.
 */
export const fundContributionSchema = z.object({
	id: z.uuid('No sabemos a qué fondo aportar.'),
	portfolioId: z.uuid('Elige el portafolio.'),
	sourceId: z.uuid('Elige la plataforma.'),
	date: z.iso.date('Elige el día del aporte.'),
	amount: amountField,
	balanceBefore: z.preprocess(blankToUndefined, balanceField.optional()),
	notes: notesField
});

/**
 * Un retiro de una posición de un fondo por saldo. `fees` es lo que se quedó la
 * entidad —una penalidad, un impuesto— y cuenta como pérdida.
 */
export const fundWithdrawalSchema = z
	.object({
		id: z.uuid('No sabemos de qué fondo retirar.'),
		entryId: z.uuid('Elige de qué portafolio sale.'),
		date: z.iso.date('Elige el día del retiro.'),
		amount: amountField,
		fees: z.coerce
			.number('Escribe la comisión con números.')
			.nonnegative('La comisión no puede ser negativa.')
			.default(0),
		all: checkboxField,
		notes: notesField
	})
	.refine((v) => v.fees < v.amount, {
		path: ['fees'],
		error: 'La comisión no puede ser mayor que el retiro.'
	});

export const fundMovementDeleteSchema = z.object({
	txnId: z.uuid('No sabemos qué movimiento borrar.')
});

/** Un día del selector como lo guarda el backend: medianoche UTC de ese día. */
export function toFundDateTime(date: string): string {
	return `${date}T00:00:00Z`;
}

/** La Superfinanciera no respondió: se puede volver a intentar. */
const SFC_UNAVAILABLE =
	'La Superintendencia Financiera no respondió. Vuelve a intentarlo en unos minutos.';

/** Solo un fondo por unidades, en pesos, toma el valor publicado. */
const NOT_LINKABLE =
	'Solo un fondo que se sigue por unidades, en pesos, puede tomar el valor que publica la Superfinanciera.';

/** Lo que se dice cuando un fondo no se puede crear ni quitar. */
export function fundErrorMessage(status: number, details = ''): string {
	if (status === 503) return SFC_UNAVAILABLE;
	if (details.includes('can be linked to a published fund')) return NOT_LINKABLE;
	if (details.includes('the SFC published none')) {
		return 'La Superfinanciera no publicó un valor de unidad para ese día. Escribe el de tu extracto.';
	}
	if (details.includes('still has positions')) {
		return 'Ese fondo todavía está en un portafolio. Borra primero su posición desde el portafolio.';
	}
	if (details.includes('date cannot be in the future')) {
		return 'La fecha no puede ser futura.';
	}
	if (details.includes('before the purchase')) {
		return 'El valor de hoy no puede ser de antes de la compra.';
	}
	if (details.includes('before the contribution')) {
		return 'El saldo de hoy no puede ser de antes del primer aporte.';
	}
	if (details.includes('too small against what went in')) {
		return 'Ese saldo es demasiado pequeño frente a lo que aportaste. Revisa las cifras.';
	}
	if (status === 404)
		return 'No encontramos ese portafolio, plataforma o fondo. Recarga la página.';

	return details || 'No pudimos guardar el fondo. Vuelve a intentarlo en un momento.';
}

/** Lo que se dice cuando una marca no se puede guardar ni borrar. */
export function fundMarkErrorMessage(status: number, details = ''): string {
	if (details.includes('date cannot be in the future')) {
		return 'La fecha no puede ser futura: pon el día del extracto.';
	}
	if (details.includes('years ago')) {
		return 'Esa fecha es demasiado antigua. Revisa el año.';
	}
	if (details.includes('unit value must be')) {
		return 'El valor de unidad tiene que ser mayor que cero.';
	}
	if (details.includes('held no units on that day')) {
		return 'Ese día el fondo no tenía nada: el saldo tiene que ser de un día después del primer aporte.';
	}
	if (status === 404) return 'Ese fondo o esa marca ya no existen. Recarga la página.';

	return details || 'No pudimos guardar el valor. Vuelve a intentarlo en un momento.';
}

/** Lo que se dice cuando un aporte o un retiro no se puede anotar ni borrar. */
export function fundMovementErrorMessage(status: number, details = ''): string {
	if (details.includes('does not hold that much')) {
		const held = details.match(/it held ([\d.]+)/)?.[1];
		return held
			? `Ese día la posición tenía unos ${held}: no se puede retirar más. Si sacaste todo, marca «Retirar todo».`
			: 'La posición no tenía tanto ese día. Si sacaste todo, marca «Retirar todo».';
	}
	if (details.includes('balance before the first contribution')) {
		return 'Antes del primer aporte el fondo no tenía nada: deja vacío el saldo de antes.';
	}
	if (details.includes('held no units on that day')) {
		return 'Ese día el fondo no tenía nada para valorar. Revisa la fecha.';
	}
	if (details.includes('followed by units')) {
		return 'Este fondo se sigue por unidades: anota sus compras y ventas desde el portafolio.';
	}
	if (details.includes('insufficient cash')) {
		return 'La cuenta de efectivo no tiene saldo suficiente para pagar el aporte.';
	}
	if (details.includes('date cannot be in the future')) {
		return 'La fecha no puede ser futura.';
	}
	if (status === 404) return 'Ese fondo o ese movimiento ya no existen. Recarga la página.';

	return details || 'No pudimos guardar el movimiento. Vuelve a intentarlo en un momento.';
}

/** Lo que se dice cuando un fondo no se puede enlazar ni desenlazar. */
export function fundLinkErrorMessage(status: number, details = ''): string {
	if (status === 503) return SFC_UNAVAILABLE;
	if (status === 409 || details.includes('can be linked to a published fund')) return NOT_LINKABLE;
	if (details.includes('public fund not found')) {
		return 'Ese fondo ya no está en el catálogo de la Superfinanciera. Búscalo de nuevo.';
	}
	if (status === 404) return 'Ese fondo ya no existe. Recarga la página.';

	return details || 'No pudimos enlazar el fondo. Vuelve a intentarlo en un momento.';
}

/** Lo que se dice cuando la búsqueda en la Superfinanciera no responde. */
export function publicFundSearchErrorMessage(status: number): string {
	if (status === 400) return 'Escribe al menos dos letras del nombre del fondo o de la entidad.';
	if (status === 503) return SFC_UNAVAILABLE;

	return 'No pudimos buscar en la Superfinanciera. Vuelve a intentarlo en un momento.';
}
