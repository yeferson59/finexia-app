/**
 * Schemas Zod de los formularios de fondos (`routes/dashboard/funds`) y los
 * mensajes con que se explica un rechazo del backend.
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

/** Un campo opcional: vacío o ausente es que no se dijo. */
const blankToUndefined = (v: unknown) => (v === '' || v === null ? undefined : v);

/**
 * Alta de un fondo: qué es, dónde está y la primera compra de unidades. Lo que
 * vale hoy una unidad es opcional; sin eso el fondo vale lo que costó hasta la
 * primera marca.
 */
export const fundCreateSchema = z
	.object({
		portfolioId: z.uuid('Elige el portafolio.'),
		sourceId: z.uuid('Elige la plataforma.'),
		currency: z.enum(SUPPORTED_CURRENCIES, 'Elige la moneda.'),
		name: z
			.string('Ponle un nombre.')
			.trim()
			.min(1, 'Ponle un nombre.')
			.max(100, 'El nombre no puede pasar de 100 caracteres.'),
		date: z.iso.date('Elige el día en que compraste las unidades.'),
		units: unitsField,
		unitValue: unitValueField,
		currentUnitValue: z.preprocess(blankToUndefined, unitValueField.optional()),
		currentDate: z.preprocess(blankToUndefined, z.iso.date().optional())
	})
	.refine((v) => !v.currentUnitValue || !v.currentDate || v.currentDate >= v.date, {
		path: ['currentDate'],
		error: 'El valor de hoy no puede ser de antes de la compra.'
	});

/** Lo que valía una unidad un día, leído del extracto. */
export const fundMarkSchema = z.object({
	id: z.uuid('No sabemos qué fondo actualizar.'),
	date: z.iso.date('Elige el día del extracto.'),
	unitValue: unitValueField,
	notes: z
		.string()
		.max(500, 'La nota no puede pasar de 500 caracteres.')
		.nullish()
		.transform((v) => (v ?? '').trim())
});

export const fundMarkDeleteSchema = z.object({
	id: z.uuid('No sabemos de qué fondo es la marca.'),
	date: z.iso.date('No sabemos qué marca borrar.')
});

export const fundDeleteSchema = z.object({
	id: z.uuid('No sabemos qué fondo quitar.')
});

/** Un día del selector como lo guarda el backend: medianoche UTC de ese día. */
export function toFundDateTime(date: string): string {
	return `${date}T00:00:00Z`;
}

/** Lo que se dice cuando un fondo no se puede crear ni quitar. */
export function fundErrorMessage(status: number, details = ''): string {
	if (details.includes('still has positions')) {
		return 'Ese fondo todavía está en un portafolio. Borra primero su posición desde el portafolio.';
	}
	if (details.includes('date cannot be in the future')) {
		return 'La fecha no puede ser futura.';
	}
	if (details.includes('before the purchase')) {
		return 'El valor de hoy no puede ser de antes de la compra.';
	}
	if (details.includes('not available yet')) {
		return 'Por ahora los fondos se siguen por unidades.';
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
	if (status === 404) return 'Ese fondo o esa marca ya no existen. Recarga la página.';

	return details || 'No pudimos guardar el valor. Vuelve a intentarlo en un momento.';
}
