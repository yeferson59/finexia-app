/**
 * Schemas Zod de los formularios de administración.
 *
 * Las tres pantallas validaban a mano (`if (!email) …`), cada una con su
 * propio criterio. Aquí quedan las mismas reglas y **los mismos mensajes**,
 * escritos una vez.
 *
 * Los ids no se validan como UUID a propósito: el backend también emite ids de
 * invitación y de waitlist con otro formato, y rechazarlos aquí dejaría fuera
 * acciones que el servidor acepta.
 */

import { z } from 'zod';

/** Roles que se pueden asignar al invitar. */
export const inviteRoleSchema = z.enum(['customer', 'admin'], { error: 'Rol inválido' });

/**
 * Invitación de un usuario. El rol llega vacío desde el atajo de la lista de
 * espera, y en ese caso vale `customer`.
 */
export const inviteUserSchema = z.object({
	email: z.string().trim().min(1, 'El correo es requerido'),
	name: z.string().trim().default(''),
	role: z
		.string()
		.transform((v) => v.trim().toLowerCase() || 'customer')
		.pipe(inviteRoleSchema)
});

/** Identificador de una fila sobre la que actúa el admin. */
export const rowIdSchema = z.string().trim().min(1, 'ID requerido');

/**
 * Alta de un activo del catálogo.
 *
 * Los cuatro campos obligatorios comparten mensaje porque el formulario los
 * pide juntos: da igual cuál falte, el aviso es el mismo.
 */
const REQUIRED_ASSET_FIELDS = 'Ticker, nombre, tipo y moneda son requeridos';

/**
 * Una línea del desglose por industrias tal y como sale del formulario.
 *
 * El peso es un porcentaje (33.1, no 0.331) y tiene que ser positivo: una fila
 * en cero no dice nada que no diga una fila ausente.
 */
const sectorWeightFormSchema = z.object({
	sector: z.string().trim().min(1),
	weight: z.coerce.number().positive('Cada peso del desglose debe ser mayor que 0')
});

/**
 * Las dos reglas que un desglose tiene que cumplir antes de salir del navegador.
 *
 * No sustituyen a las del backend —la API es alcanzable sin este formulario—,
 * pero son las que convierten un 400 en un aviso junto al campo que lo causó,
 * que es donde se arregla.
 *
 * Lo que **no** se exige es que sumen 100: la ficha de un fondo real deja unas
 * décimas en caja, y el reparto normaliza sobre el total que encuentre.
 * Pasarse de 100 sí es imposible, no incompleto.
 */
function classificationIsCoherent(
	value: { sector: string; sectorWeights: { weight: number }[] },
	ctx: z.RefinementCtx
): void {
	if (value.sector !== '' && value.sectorWeights.length > 0) {
		ctx.addIssue({
			code: 'custom',
			message: 'Elige una industria única o un desglose por industrias, no las dos'
		});
	}

	const total = value.sectorWeights.reduce((sum, w) => sum + w.weight, 0);
	if (total > 100) {
		ctx.addIssue({ code: 'custom', message: 'Los pesos del desglose suman más de 100 %' });
	}
}

/**
 * Los pesos que trae un envío del formulario, que los manda en un campo por
 * industria (`weight.technology`) en vez de en una lista.
 *
 * Un `<form>` no tiene forma de mandar un array de objetos, y la alternativa
 * —un campo oculto con JSON dentro— deja de funcionar en cuanto el navegador no
 * ejecuta el script, que es justo lo que estas pantallas evitan usando actions.
 * Las casillas vacías se descartan aquí: son las diez industrias en las que un
 * fondo sectorial no está.
 */
export function sectorWeightsFromForm(fd: FormData): { sector: string; weight: string }[] {
	const weights: { sector: string; weight: string }[] = [];

	for (const [key, value] of fd.entries()) {
		if (!key.startsWith('weight.') || typeof value !== 'string' || value.trim() === '') continue;

		weights.push({ sector: key.slice('weight.'.length), weight: value.trim() });
	}

	return weights;
}

export const assetCreateSchema = z
	.object({
		ticker: z.string().trim().toUpperCase().min(1, REQUIRED_ASSET_FIELDS),
		name: z.string().trim().min(1, REQUIRED_ASSET_FIELDS),
		assetType: z.string().trim().min(1, REQUIRED_ASSET_FIELDS),
		currency: z.string().trim().toUpperCase().min(1, REQUIRED_ASSET_FIELDS),
		exchange: z.string().trim().default(''),
		// Opcional y sin lista cerrada aquí: el backend acepta texto libre y lo
		// normaliza («Tecnología», «Technology», `technology` son el mismo sector),
		// así que repetir el vocabulario en este lado solo serviría para rechazar
		// grafías que el servidor sí entiende. El desplegable manda los valores
		// canónicos; una importación puede mandar cualquier otra cosa.
		sector: z.string().trim().default(''),
		// El desglose, para el activo que no cabe en una industria. Viaja
		// siempre, aunque esté vacío: en el backend un envío sin él lo borra.
		sectorWeights: z.array(sectorWeightFormSchema).default([])
	})
	.superRefine(classificationIsCoherent);

/**
 * Ajuste del precio manual de un activo.
 *
 * `price` se valida como número pero al backend viaja el texto tal cual lo
 * escribió el admin: convertirlo perdería los decimales de cola («190.00» →
 * «190»), que el backend guarda como los recibe.
 */
export const assetPriceSchema = z.object({
	id: z.string().trim().min(1, 'ID de activo requerido'),
	price: z.coerce.number().positive('Precio inválido'),
	currency: z.string().trim().min(1).default('USD')
});

/**
 * Edición completa de un activo del catálogo.
 *
 * Los cinco campos de identidad son los mismos del alta y comparten su mensaje,
 * porque el formulario los manda siempre enteros: editar el nombre reenvía
 * también el ticker que se conserva.
 *
 * Los dos últimos son de la edición y solo de ella. `isCurated` llega del
 * checkbox, que no envía nada cuando está desmarcado, así que se normaliza a
 * booleano aquí en vez de dejar que «ausente» signifique dos cosas. `price` es
 * opcional —quedarse en blanco deja el precio manual como estaba— y se valida
 * como número, pero al backend viaja el texto tal cual, igual que en el ajuste
 * rápido de la tabla.
 */
export const assetUpdateSchema = z
	.object({
		id: rowIdSchema,
		ticker: z.string().trim().toUpperCase().min(1, REQUIRED_ASSET_FIELDS),
		name: z.string().trim().min(1, REQUIRED_ASSET_FIELDS),
		assetType: z.string().trim().min(1, REQUIRED_ASSET_FIELDS),
		currency: z.string().trim().toUpperCase().min(1, REQUIRED_ASSET_FIELDS),
		exchange: z.string().trim().default(''),
		sector: z.string().trim().default(''),
		sectorWeights: z.array(sectorWeightFormSchema).default([]),
		isCurated: z
			.union([z.string(), z.boolean(), z.null(), z.undefined()])
			.transform((v) => v === true || v === 'on' || v === 'true'),
		price: z
			.string()
			.trim()
			.default('')
			.refine((v) => v === '' || (Number.isFinite(Number(v)) && Number(v) > 0), 'Precio inválido')
	})
	.superRefine(classificationIsCoherent);

/** Alta de una tasa de cambio. */
const REQUIRED_RATE_FIELDS = 'Moneda origen, destino y tasa son requeridos';

export const rateCreateSchema = z.object({
	fromCurrency: z.string().trim().toUpperCase().min(1, REQUIRED_RATE_FIELDS),
	toCurrency: z.string().trim().toUpperCase().min(1, REQUIRED_RATE_FIELDS),
	rate: z.string().trim().min(1, REQUIRED_RATE_FIELDS)
});

/** Ajuste de una tasa; el valor viaja como texto, igual que el precio. */
export const rateUpdateSchema = z.object({
	id: z.string().trim().min(1, 'ID de tasa requerido'),
	rate: z.coerce.number().positive('Tasa inválida')
});
