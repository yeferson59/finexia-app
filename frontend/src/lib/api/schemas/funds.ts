/**
 * Fondos de inversión — contratos HTTP como schemas Zod.
 *
 * Parte de `lib/api/schemas`: el porqué de que sean schemas y no interfaces
 * está en el `index.ts` de la carpeta.
 */

import { z } from 'zod';

/** Lo que un portafolio guarda de un fondo, en una plataforma. */
export const fundPositionSchema = z.object({
	entryId: z.string(),
	portfolioId: z.string(),
	portfolioName: z.string(),
	sourceId: z.string(),
	sourceName: z.string(),
	units: z.string(),
	/** Lo que costaron las unidades, en la moneda del fondo. */
	cost: z.string()
});

/**
 * Un fondo que el usuario sigue (`GET /portfolios/funds`): unidades cuyo valor
 * escribe el dueño desde el extracto, porque ningún proveedor de mercado lo
 * cotiza.
 */
export const fundSchema = z.object({
	assetId: z.string(),
	/** Generado (`FND-…`): un fondo no tiene símbolo. */
	ticker: z.string(),
	name: z.string(),
	currency: z.string(),
	/** `units` sigue unidades y valor de unidad; `balance`, solo el saldo. */
	tracking: z.enum(['units', 'balance']),
	/** Todo lo que guardan sus posiciones, en `currency`. */
	units: z.string(),
	cost: z.string(),
	/** Las unidades a la última marca, o a su costo si no hay ninguna. */
	value: z.string(),
	/** Sin marcas vale lo que costó: su ganancia de cero no dice nada. */
	pricedAtCost: z.boolean(),
	/** La última marca; `null` hasta que haya una. */
	unitValue: z.string().nullable(),
	valuedOn: z.string().nullable(),
	marks: z.number(),
	positions: z.array(fundPositionSchema),
	createdAt: z.string()
});

/** Lo que valía un fondo un día (`GET /portfolios/funds/:id/marks`). */
export const fundMarkSchema = z.object({
	date: z.string(),
	unitValue: z.string(),
	/** El saldo con que se marcó un fondo que se sigue por saldo. */
	balance: z.string().nullable(),
	notes: z.string(),
	createdAt: z.string(),
	updatedAt: z.string()
});

/**
 * Un aporte o un retiro de un fondo (`GET /portfolios/funds/:id/movements`).
 * En un fondo que se sigue por saldo, `amount` es el dinero que se dijo y las
 * unidades salen de él; en uno por unidades, `amount` es unidades × precio.
 */
export const fundMovementSchema = z.object({
	txnId: z.string(),
	entryId: z.string(),
	portfolioId: z.string(),
	portfolioName: z.string(),
	sourceName: z.string(),
	kind: z.enum(['contribution', 'withdrawal']),
	date: z.string(),
	amount: z.string(),
	fees: z.string(),
	/** Un retiro de todo lo que guardaba la posición. */
	all: z.boolean(),
	units: z.string(),
	unitValue: z.string(),
	notes: z.string(),
	createdAt: z.string()
});
