/**
 * Schemas Zod de los formularios de portfolios.
 *
 * Estaban declarados dentro de las actions de `routes/dashboard/portfolios/**`,
 * uno por archivo, con las mismas reglas escritas de formas distintas. Aquí
 * viven juntos para que el contrato de un mismo campo (la cantidad, el precio,
 * la fecha) no dependa de qué formulario lo envía.
 */

import { z } from 'zod';

/** Alta de un portfolio (`routes/dashboard/portfolios/add`). */
export const portfolioCreateSchema = z.object({
	name: z.string().min(1),
	description: z.string().nullable(),
	type: z.string().min(1),
	riskId: z.uuid(),
	currency: z.string().min(1),
	priceValue: z.coerce.number().nonnegative().default(0),
	isDefault: z.coerce.boolean()
});

/** Edición de un portfolio (`routes/dashboard/portfolios/[id]`). */
export const portfolioUpdateSchema = z.object({
	name: z.string().min(2, 'El nombre debe tener al menos 2 caracteres'),
	description: z.string().optional().default(''),
	type: z.string().min(1),
	riskId: z.uuid(),
	isDefault: z.coerce.boolean()
});

/** Alta de una posición (`routes/dashboard/portfolios/[id]/add`). */
export const portfolioEntrySchema = z.object({
	portfolioId: z.uuid(),
	assetId: z.uuid(),
	sourceId: z.uuid(),
	quantity: z.coerce.number().positive(),
	price: z.coerce.number().positive(),
	// El precio y su moneda viajan separados, así que una moneda vacía o basura
	// no falla: se guarda y el coste queda etiquetado con algo que no es lo que
	// se pagó. Se exige el código ISO de tres letras que espera el backend.
	costCurrency: z.coerce
		.string()
		.trim()
		.toUpperCase()
		.regex(/^[A-Z]{3}$/, 'Moneda inválida: usa un código ISO de tres letras'),
	entryDate: z.coerce.date(),
	notes: z.coerce.string().optional()
});

/**
 * Campos comunes de una transacción. El precio admite 0 (una entrega o un
 * split no cuestan nada) aunque la cantidad nunca lo sea.
 */
const transactionSchema = z.object({
	type: z.string().min(1),
	quantity: z.coerce.number().positive(),
	price: z.coerce.number().min(0),
	currency: z.string().default('USD'),
	fees: z.coerce.number().min(0).default(0),
	transactionDate: z.coerce.date(),
	notes: z.string().optional()
});

/** Alta de una transacción sobre una posición existente. */
export const transactionCreateSchema = transactionSchema.extend({ entryId: z.uuid() });

/** Edición de una transacción ya registrada. */
export const transactionUpdateSchema = transactionSchema.extend({ txnId: z.uuid() });

/** Borrado de una transacción: solo hace falta identificarla. */
export const transactionDeleteSchema = z.object({ txnId: z.uuid() });
