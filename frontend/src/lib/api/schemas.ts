/**
 * Contratos HTTP del backend, como schemas Zod.
 *
 * **Fuente de verdad de los shapes de la API**: los tipos de `types.ts` se
 * derivan de aquí con `z.infer`, así que un contrato se escribe una sola vez.
 * Se mantienen a mano contra `docs/API.md`.
 *
 * Que sean schemas y no interfaces tiene un motivo: en desarrollo, la capa de
 * API valida con ellos lo que responde el backend (ver `client.ts`), de modo
 * que una divergencia deja un aviso en consola en vez de propagarse como un
 * `undefined` tres componentes más abajo. En producción no se ejecutan.
 */

import { z } from 'zod';

// ---------------------------------------------------------------------------
// Paginación (§1.5 de docs/API.md)
// ---------------------------------------------------------------------------

/**
 * Bloque de metadatos de las rutas paginadas. Conserva nombres históricos por
 * área (`usersForPage`/`totalUsers`, …), que llegan como claves extra.
 */
export const pageMetaSchema = z
	.object({
		currentPage: z.number(),
		totalPages: z.number(),
		previous: z.boolean(),
		next: z.boolean(),
		offset: z.number().optional()
	})
	.catchall(z.union([z.number(), z.boolean()]));

/** `data` de una ruta paginada: lista de items + metadatos. */
export function paginatedSchema<T extends z.ZodType>(item: T) {
	return z.object({ items: z.array(item), metaData: pageMetaSchema });
}

// ---------------------------------------------------------------------------
// Portfolios
// ---------------------------------------------------------------------------

/**
 * Resumen de un portfolio (`GET /portfolios/summary`). Superset de los
 * subconjuntos que antes tipaban por separado el dashboard, el layout de
 * portfolios y el selector de import; `displayCurrency` solo llega cuando se
 * pide `?currency=`.
 */
export const portfolioSummarySchema = z.object({
	id: z.string(),
	name: z.string(),
	description: z.string().optional(),
	type: z.string(),
	baseCurrency: z.string(),
	displayCurrency: z.string().optional(),
	// `false` cuando se pidió `?currency=` y no había tasa: los totales se
	// quedaron en `baseCurrency` en vez de fallar la petición entera, así que
	// esta fila no se puede sumar con las demás. Opcional por el backend previo.
	fxConverted: z.boolean().optional(),
	isDefault: z.boolean().optional(),
	riskId: z.string().optional(),
	riskName: z.string(),
	totalPositions: z.number(),
	totalCostBase: z.string(),
	totalMarketValue: z.string(),
	totalGainLoss: z.string(),
	totalGainLossPct: z.string(),
	createdAt: z.string().optional(),
	// Posiciones cuyos importes no están en la moneda base porque no había
	// tasa: siguen sumadas en los totales, así que un valor > 0 significa que
	// los totales mezclan monedas. Opcional para tolerar un backend anterior.
	positionsUnconverted: z.number().optional()
});

/** Posición dentro de un portfolio (holdings de `GET /portfolios/:id`). */
export const holdingSchema = z.object({
	id: z.string(),
	assetId: z.string(),
	ticker: z.string(),
	name: z.string(),
	assetType: z.string(),
	exchange: z.string(),
	currency: z.string(),
	quantity: z.string(),
	price: z.string(),
	marketPrice: z.string(),
	costCurrency: z.string(),
	category: z.string(),
	entryDate: z.string(),
	notes: z.string(),
	// Totales ya convertidos a la moneda base del portafolio: los únicos
	// importes del holding que se pueden sumar entre posiciones, porque `price`
	// y `marketPrice` vienen cada uno en su propia moneda. `fxConverted: false`
	// significa que faltó la tasa y los dos totales están sin convertir.
	//
	// Opcionales para tolerar un backend anterior a estos campos: quien los
	// consume vuelve al cálculo nativo cuando no llegan.
	costBasisBase: z.string().optional(),
	marketValueBase: z.string().optional(),
	fxConverted: z.boolean().optional()
});

/** Detalle completo de un portfolio (`GET /portfolios/:id`). */
export const portfolioDetailSchema = z.object({
	id: z.string(),
	userId: z.string(),
	name: z.string(),
	description: z.string(),
	type: z.string(),
	baseCurrency: z.string(),
	isDefault: z.boolean(),
	riskId: z.string(),
	riskName: z.string(),
	createdAt: z.string(),
	updatedAt: z.string(),
	holdings: z.array(holdingSchema)
});

/** Nivel de riesgo del catálogo (`GET /portfolios/risks`). */
export const riskSchema = z.object({
	id: z.string(),
	name: z.string(),
	description: z.string()
});

/** Asignación por categoría de activo (`GET /portfolios/allocation`). */
export const allocationItemSchema = z.object({
	category: z.string(),
	marketValue: z.string(),
	percent: z.number(),
	// Moneda en la que están todos los `marketValue` de la respuesta, y
	// posiciones de esta categoría que no pudieron convertirse a ella.
	// Opcionales para tolerar un backend anterior.
	currency: z.string().optional(),
	positionsUnconverted: z.number().optional()
});

/**
 * Un activo con todo lo que el usuario tiene de él, sumado a través de sus
 * portafolios (`GET /portfolios/holdings`).
 *
 * `quantity` son unidades y solo significa algo dentro de la fila: sumar
 * acciones con bitcoins no da nada. `marketValue` es lo que sí compara, y viene
 * en `displayCurrency` para todas las filas —igual que la asignación—, que es
 * lo que hace que `percent` quiera decir algo.
 *
 * `marketPrice` vacío no es precio cero: es una posición valorada a coste, en la
 * que cada entrada pagó el suyo y ningún número representa al activo. Eso lo
 * dice `priceSource`.
 */
export const assetHoldingSchema = z.object({
	assetId: z.string(),
	ticker: z.string(),
	name: z.string(),
	assetType: z.string(),
	exchange: z.string(),
	/** Moneda en la que cotiza el activo, que es la de `marketPrice`. */
	currency: z.string(),
	quantity: z.string(),
	marketPrice: z.string(),
	marketValue: z.string(),
	percent: z.number(),
	/** Moneda de `marketValue`, la misma en todas las filas. */
	displayCurrency: z.string(),
	/** En cuántos portafolios del usuario aparece el activo. */
	portfolios: z.number(),
	priceSource: z.string(),
	positionsUnconverted: z.number()
});

/** Mayor transacción de un portfolio (`GET /portfolios/:id/top-transaction`). */
export const topTransactionSchema = z.object({
	value: z.string(),
	type: z.string(),
	currency: z.string(),
	assetTicker: z.string(),
	assetName: z.string(),
	transactionDate: z.string()
});

/** Punto de la serie de crecimiento. */
export const growthDataPointSchema = z.object({
	date: z.string(),
	totalValue: z.string(),
	totalCostBase: z.string(),
	gainLoss: z.string(),
	gainLossPct: z.string(),
	// Portafolios sumados a esta fecha sin tasa con la que convertirlos, y por
	// tanto contados a valor nominal. Opcional por si el backend va por detrás.
	portfoliosUnconverted: z.number().optional(),
	// Dinero que el dueño metió (positivo) o sacó (negativo) entre el punto
	// anterior y este, reconstruido de las transacciones. Es lo que hay que
	// descontar de la variación del valor para que quede rentabilidad.
	// Opcional por si el backend va por detrás.
	netFlow: z.string().optional()
});

/**
 * Resumen agregado de la serie de crecimiento.
 *
 * Lleva dos lecturas distintas que no hay que confundir: `totalGrowthPct` mide
 * cuánto se movió el **valor** entre el primer snapshot y el último —abrir un
 * portafolio o añadir una posición cuenta como crecimiento— mientras que
 * `gainLoss`/`gainLossPct` son el beneficio del último punto, mercado menos
 * capital invertido, que es el rendimiento de verdad.
 *
 * Las dos últimas son opcionales para tolerar un backend anterior.
 */
export const growthSummarySchema = z.object({
	firstDate: z.string(),
	initialValue: z.string(),
	currentValue: z.string(),
	totalGrowthPct: z.string(),
	gainLoss: z.string().optional(),
	gainLossPct: z.string().optional(),
	/** Moneda en la que están todos los importes de la serie. */
	currency: z.string().optional()
});

/** Crecimiento (`GET /portfolios/growth` y `GET /portfolios/:id/growth`). */
export const portfolioGrowthSchema = z.object({
	points: z.array(growthDataPointSchema),
	summary: growthSummarySchema
});

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
	fees: z.string(),
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

// ---------------------------------------------------------------------------
// Plataformas / fuentes
// ---------------------------------------------------------------------------

/** Plataforma / fuente (`GET /portfolios/sources`). */
export const platformSchema = z.object({
	id: z.string(),
	name: z.string(),
	description: z.string(),
	sourceType: z.string(),
	/** Alias histórico de `sourceType` en algunas vistas/formularios. */
	type: z.string().optional(),
	isActive: z.boolean(),
	investments: z.number(),
	totalValue: z.string(),
	/**
	 * Moneda en la que viene `totalValue`. El total suma posiciones compradas
	 * en monedas distintas, así que sin esto la cifra no tiene unidad — y la
	 * vista le ponía un "$" fijo. Opcional mientras convivan backends viejos.
	 */
	displayCurrency: z.string().optional(),
	/**
	 * Posiciones que siguen contadas en `totalValue` a valor nominal porque no
	 * había tasa para su moneda: si es > 0 el total mezcla monedas.
	 */
	positionsUnconverted: z.number().optional(),
	createdAt: z.string()
});

// ---------------------------------------------------------------------------
// Assets y tasas de cambio (mercado)
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// Usuarios, preferencias, sesiones y 2FA
// ---------------------------------------------------------------------------

/** Usuario en el listado de administración (`GET /users`). */
export const userItemSchema = z.object({
	id: z.string(),
	name: z.string(),
	email: z.string(),
	emailVerified: z.boolean(),
	createdAt: z.string(),
	bannedAt: z.string().optional().nullable(),
	role: z.object({ name: z.string() })
});

/** Invitación (`GET /users/invitations`). */
export const invitationItemSchema = z.object({
	id: z.string(),
	email: z.string(),
	name: z.string(),
	role: z.string(),
	status: z.enum(['pending', 'expired', 'accepted', 'revoked']),
	expiresAt: z.string(),
	createdAt: z.string()
});

/** Entrada de la waitlist (`GET /users/waitlist`). */
export const waitlistItemSchema = z.object({
	id: z.string(),
	email: z.string(),
	status: z.enum(['pending', 'invited', 'registered']),
	invitedAt: z.string().optional().nullable(),
	createdAt: z.string()
});

/** Preferencias del usuario (`GET /users/me/preferences`). */
export const userPreferencesSchema = z.object({
	userId: z.string(),
	emailAlerts: z.boolean(),
	weeklySummary: z.boolean()
});

/** Sesión activa del usuario (`GET /auth/sessions`). */
export const activeSessionSchema = z.object({
	id: z.string(),
	ipAddress: z.string().nullable(),
	userAgent: z.string().nullable(),
	location: z.string().nullable(),
	createdAt: z.string(),
	lastActiveAt: z.string(),
	expiresAt: z.string(),
	current: z.boolean()
});

/** Estado de la verificación en dos pasos (`GET /auth/2fa`). */
export const twoFactorStatusSchema = z.object({
	enabled: z.boolean(),
	pendingSetup: z.boolean(),
	recoveryCodesLeft: z.number()
});

/** Proveedor de datos de mercado para el que se puede aportar una clave. */
export const marketProviderSchema = z.enum(['finnhub', 'alphavantage']);

/**
 * Estado de una clave de proveedor del usuario (`GET /market/credentials`).
 *
 * No hay campo para la clave, y es deliberado: una vez guardada solo se puede
 * reemplazar o borrar, nunca leer. `last4` existe para que la UI pueda
 * identificarla sin tenerla.
 */
export const marketCredentialSchema = z.object({
	provider: marketProviderSchema,
	last4: z.string(),
	status: z.enum(['active', 'invalid', 'rate_limited']),
	lastVerifiedAt: z.string().nullable(),
	lastError: z.string().optional(),
	createdAt: z.string(),
	updatedAt: z.string()
});

/** Un precio que trajo la clave del propio usuario (`POST /market/sync`). */
export const marketPriceSchema = z.object({
	assetId: z.string(),
	ticker: z.string(),
	price: z.string(),
	source: marketProviderSchema,
	fetchedAt: z.string()
});

/** Una tasa de cambio que trajo la clave del propio usuario. */
export const marketRateSchema = z.object({
	fromCurrency: z.string(),
	toCurrency: z.string(),
	rate: z.string(),
	source: marketProviderSchema,
	fetchedAt: z.string()
});

/**
 * Resultado de `POST /market/sync`.
 *
 * Las dos mitades van nombradas porque sincronizar no es solo precios: sin tasa
 * de cambio una posición en otra moneda no se puede valorar, y una respuesta
 * con precios pero sin tasas ha hecho medio trabajo.
 */
export const marketSyncResultSchema = z.object({
	prices: z.array(marketPriceSchema),
	rates: z.array(marketRateSchema)
});
