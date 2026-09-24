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
 * Un fondo del catálogo de la Superintendencia Financiera
 * (`GET /portfolios/funds/catalog`): un tipo de participación de un FIC, con el
 * último valor de unidad que publicó. Cada tipo de participación tiene el suyo.
 */
export const publicFundSchema = z.object({
	/** Los cinco códigos que lo identifican: «5-31-3644-1-501». */
	id: z.string(),
	entityName: z.string(),
	fundName: z.string(),
	/** «FIC DE MERCADO MONETARIO», «FIC DE TIPO GENERAL»… */
	fundKind: z.string(),
	fundCode: z.number(),
	participation: z.number(),
	unitValue: z.string(),
	valueDate: z.string(),
	investors: z.number()
});

/** El fondo publicado cuyos valores de unidad valoran un fondo del usuario. */
export const publicFundLinkSchema = z.object({
	id: z.string(),
	entityName: z.string(),
	fundName: z.string(),
	participation: z.number()
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
	/**
	 * El fondo de la Superfinanciera del que salen sus valores de unidad; `null`
	 * cuando los escribe el dueño.
	 */
	publicFund: publicFundLinkSchema.nullable().default(null),
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
	/** `user` la escribió el dueño; `public`, la publicó la Superfinanciera. */
	source: z.enum(['user', 'public']).default('user'),
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

/**
 * La rentabilidad de un fondo en un periodo. `from`, `days`, `pct` y `eaPct`
 * van `null` cuando el historial no llega tan atrás; `eaPct` también en un
 * periodo de menos de 28 días, donde anualizar exagera.
 */
export const fundPeriodSchema = z.object({
	key: z.enum(['30d', '90d', '180d', '365d', 'ytd', 'inception']),
	from: z.string().nullable(),
	to: z.string(),
	days: z.number().nullable(),
	pct: z.string().nullable(),
	eaPct: z.string().nullable()
});

/** El valor de unidad de un día: la serie que dibuja la gráfica del fondo. */
export const fundPointSchema = z.object({
	date: z.string(),
	unitValue: z.string()
});

/**
 * Cómo le fue a un fondo (`GET /portfolios/funds/:id/performance`): la
 * rentabilidad por periodo, como la publica la entidad, y el dinero que entró y
 * salió. En un fondo por saldo, `unitValue` y la serie son un índice base 100.
 */
export const fundPerformanceSchema = z.object({
	assetId: z.string(),
	tracking: z.enum(['units', 'balance']),
	currency: z.string(),
	valuedOn: z.string().nullable(),
	unitValue: z.string().nullable(),
	units: z.string(),
	value: z.string(),
	cost: z.string(),
	pricedAtCost: z.boolean(),
	invested: z.string(),
	withdrawn: z.string(),
	fees: z.string(),
	realizedGain: z.string(),
	/** `null` mientras el fondo vale lo que costó. */
	unrealizedGain: z.string().nullable(),
	periods: z.array(fundPeriodSchema),
	series: z.array(fundPointSchema)
});
