/**
 * Transacciones — contratos HTTP como schemas Zod.
 *
 * Parte de `lib/api/schemas`: el porqué de que sean schemas y no interfaces
 * está en el `index.ts` de la carpeta.
 */

import { z } from 'zod';

// ---------------------------------------------------------------------------
// Transacciones
// ---------------------------------------------------------------------------

/** Transacción de una posición (`GET /portfolios/:id/assets/:symbol/transactions`). */
export const transactionSchema = z.object({
	id: z.string(),
	entryId: z.string(),
	type: z.string(),
	quantity: z.string(),
	price: z.string(),
	currency: z.string(),
	// `price` está en `currency` —la moneda en la que cotizó la operación— y
	// `fxRate` es lo que valía una unidad de esa moneda en `costCurrency` el día
	// de la operación. El coste real es `price * fxRate`: sin la tasa, una
	// compra de LVMH en euros liquidada en dólares se lee como si el broker
	// hubiera cobrado 606,60 USD.
	//
	// Opcionales para tolerar un backend anterior a la migración 000029: sin
	// ellos se asume tasa 1, que es lo que valía toda transacción de entonces.
	fxRate: z.string().optional(),
	costCurrency: z.string().optional(),
	fees: z.string(),
	// La comisión no siempre se cobra del mismo lado que la ejecución: el
	// bróker que cotizó en EUR pudo cargarla en USD. Es `currency` o
	// `costCurrency`, nunca una tercera; ausente significa `currency`.
	feesCurrency: z.string().optional(),
	transactionDate: z.string(),
	notes: z.string(),
	createdAt: z.string()
});

/** Transacción del usuario con datos del activo (`GET /portfolios/transactions`). */
export const userTransactionSchema = transactionSchema.extend({
	assetTicker: z.string(),
	assetName: z.string()
});

/** `data` de las transacciones paginadas por asset. */
export const pagedTransactionsSchema = z.object({
	data: z.array(transactionSchema),
	total: z.number(),
	page: z.number(),
	limit: z.number(),
	totalPages: z.number()
});
