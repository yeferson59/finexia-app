/**
 * Assets y tasas — contratos HTTP como schemas Zod.
 *
 * Parte de `lib/api/schemas`: el porqué de que sean schemas y no interfaces
 * está en el `index.ts` de la carpeta.
 */

import { z } from 'zod';

// ---------------------------------------------------------------------------
// Assets y tasas de cambio (mercado)
// ---------------------------------------------------------------------------

/**
 * Una industria y su peso dentro de un activo, en porcentaje (33.1, no 0.331).
 *
 * `weight` llega como número del backend —es un porcentaje con cuatro
 * decimales, no un importe— y se acepta también como texto para no depender de
 * cómo lo serialice quien lo mande.
 */
export const sectorWeightSchema = z.object({
	sector: z.string(),
	weight: z.coerce.number()
});

/** Precio de un asset. */
export const assetPriceSchema = z.object({
	value: z.string(),
	currency: z.string()
});

/** Asset del catálogo (`GET /portfolios/assets`). */
export const assetSchema = z.object({
	id: z.string(),
	ticker: z.string(),
	name: z.string(),
	assetType: z.string(),
	currency: z.string(),
	exchange: z.string().optional(),
	/**
	 * Industria del activo, vacía cuando nadie lo ha clasificado. Opcional para
	 * tolerar un backend anterior.
	 */
	sector: z.string().optional(),
	/**
	 * El desglose del activo que no cabe en una sola industria: un ETF de
	 * mercado ancho. Es **excluyente** con `sector` —el backend no guarda los
	 * dos—, así que una lista con filas significa que `sector` está vacío por
	 * ese motivo y no porque falte clasificar.
	 *
	 * Opcional, como `sector`, para tolerar un backend anterior: el schema solo
	 * avisa del contrato en desarrollo, no transforma la respuesta, así que un
	 * `default` aquí sería una promesa que en ejecución no se cumple.
	 */
	sectorWeights: z.array(sectorWeightSchema).optional(),
	currentPrice: assetPriceSchema.nullable(),
	priceUpdatedAt: z.string().nullable(),
	/**
	 * `true` si lo curó el operador y lo ve todo el mundo; `false` si lo aportó
	 * un usuario, en cuyo caso solo lo ven quienes lo aportaron (API §2.8).
	 */
	isCurated: z.boolean().optional()
});

/**
 * Resultado de un import masivo (`POST /assets/import`,
 * `POST /exchange-rates/import`).
 *
 * Un import es parcial por diseño: las filas correctas entran aunque otras
 * fallen, y por eso el resultado lleva las tres cuentas más el detalle por
 * fila de lo que se quedó fuera.
 */
export const importResultSchema = z.object({
	totalRows: z.number(),
	imported: z.number(),
	skipped: z.number(),
	errors: z.array(z.object({ row: z.number(), message: z.string() }))
});

/** Tasa de cambio (`GET /exchange-rates`). */
export const exchangeRateSchema = z.object({
	id: z.string(),
	fromCurrency: z.string(),
	toCurrency: z.string(),
	rate: z.string(),
	rateDate: z.string(),
	// Quién puso la tasa: `manual` si la escribió un administrador (o vino de
	// una hoja importada) y el nombre de la fuente si la publicó un feed
	// público. Con `.default` para que una respuesta anterior a la columna, sin
	// el campo, siga validando en vez de tumbar la pantalla de administración.
	source: z.string().default('manual'),
	createdAt: z.string()
});
