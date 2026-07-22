/**
 * Contratos HTTP compartidos con el backend.
 *
 * Única fuente de verdad de los shapes que devuelve la API, mantenida a mano
 * contra `docs/API.md`. Antes de la Fase 2 estos tipos vivían duplicados en los
 * loaders/actions de `routes/`; aquí se centralizan para que un cambio de
 * contrato se refleje en un solo lugar.
 */

// ---------------------------------------------------------------------------
// Sobre de respuesta y paginación (§1.1 y §1.5 de docs/API.md)
// ---------------------------------------------------------------------------

/** Sobre estándar que envuelve toda respuesta JSON del backend. */
export interface ApiEnvelope<T = unknown> {
	success: boolean;
	message?: string;
	details?: string;
	/** Código estable de máquina en algunos flujos (p. ej. `auth:login:2fa`). */
	action?: string;
	data?: T;
	timestamp?: string;
}

/**
 * Bloque de metadatos de las rutas paginadas. Conserva nombres históricos por
 * área (`usersForPage`/`totalUsers`, …) accesibles vía índice.
 */
export interface PageMeta {
	currentPage: number;
	totalPages: number;
	previous: boolean;
	next: boolean;
	offset?: number;
	[key: string]: number | boolean | undefined;
}

/** `data` de una ruta paginada: lista de items + metadatos. */
export interface Paginated<T> {
	items: T[];
	metaData: PageMeta;
}

// ---------------------------------------------------------------------------
// Portfolios
// ---------------------------------------------------------------------------

/**
 * Resumen de un portfolio (`GET /portfolios/summary`). Superset de los subconjuntos
 * que antes tipaban por separado el dashboard, el layout de portfolios y el
 * selector de import; `displayCurrency` solo llega cuando se pide `?currency=`.
 */
export interface PortfolioSummary {
	id: string;
	name: string;
	description?: string;
	type: string;
	baseCurrency: string;
	displayCurrency?: string;
	isDefault?: boolean;
	riskId?: string;
	riskName: string;
	totalPositions: number;
	totalCostBase: string;
	totalMarketValue: string;
	totalGainLoss: string;
	totalGainLossPct: string;
	createdAt?: string;
}

/** Posición dentro de un portfolio (holdings de `GET /portfolios/:id`). */
export interface Holding {
	id: string;
	assetId: string;
	ticker: string;
	name: string;
	assetType: string;
	exchange: string;
	currency: string;
	quantity: string;
	price: string;
	marketPrice: string;
	costCurrency: string;
	category: string;
	entryDate: string;
	notes: string;
}

/** Detalle completo de un portfolio (`GET /portfolios/:id`). */
export interface PortfolioDetail {
	id: string;
	userId: string;
	name: string;
	description: string;
	type: string;
	baseCurrency: string;
	isDefault: boolean;
	riskId: string;
	riskName: string;
	createdAt: string;
	updatedAt: string;
	holdings: Holding[];
}

/** Nivel de riesgo del catálogo (`GET /portfolios/risks`). */
export interface Risk {
	id: string;
	name: string;
	description: string;
}

/** Asignación por categoría de activo (`GET /portfolios/allocation`). */
export interface AllocationItem {
	category: string;
	marketValue: string;
	percent: number;
}

/** Mayor transacción de un portfolio (`GET /portfolios/:id/top-transaction`). */
export interface TopTransaction {
	value: string;
	type: string;
	currency: string;
	assetTicker: string;
	assetName: string;
	transactionDate: string;
}

/** Punto de la serie de crecimiento. */
export interface GrowthDataPoint {
	date: string;
	totalValue: string;
	totalCostBase: string;
	gainLoss: string;
	gainLossPct: string;
}

/** Resumen agregado de la serie de crecimiento. */
export interface GrowthSummary {
	firstDate: string;
	initialValue: string;
	currentValue: string;
	totalGrowthPct: string;
}

/** Crecimiento (`GET /portfolios/growth` y `GET /portfolios/:id/growth`). */
export interface PortfolioGrowth {
	points: GrowthDataPoint[];
	summary: GrowthSummary;
}

// ---------------------------------------------------------------------------
// Transacciones
// ---------------------------------------------------------------------------

/** Transacción de una posición (`GET /portfolios/:id/assets/:symbol/transactions`). */
export interface Transaction {
	id: string;
	entryId: string;
	type: string;
	quantity: string;
	price: string;
	currency: string;
	fees: string;
	transactionDate: string;
	notes: string;
	createdAt: string;
}

/** Transacción del usuario con datos del activo (`GET /portfolios/transactions`). */
export interface UserTransaction extends Transaction {
	assetTicker: string;
	assetName: string;
}

/** `data` de las transacciones paginadas por asset. */
export interface PagedTransactions {
	data: Transaction[];
	total: number;
	page: number;
	limit: number;
	totalPages: number;
}

// ---------------------------------------------------------------------------
// Plataformas / fuentes
// ---------------------------------------------------------------------------

/** Plataforma / fuente (`GET /portfolios/sources`). */
export interface Platform {
	id: string;
	name: string;
	description: string;
	sourceType: string;
	/** Alias histórico de `sourceType` en algunas vistas/formularios. */
	type?: string;
	isActive: boolean;
	investments: number;
	totalValue: string;
	createdAt: string;
}

// ---------------------------------------------------------------------------
// Assets y tasas de cambio (mercado)
// ---------------------------------------------------------------------------

/** Precio de un asset. */
export interface AssetPrice {
	value: string;
	currency: string;
}

/** Asset del catálogo (`GET /portfolios/assets`). */
export interface Asset {
	id: string;
	ticker: string;
	name: string;
	assetType: string;
	currency: string;
	exchange?: string;
	currentPrice: AssetPrice | null;
	priceUpdatedAt: string | null;
}

/** Tasa de cambio (`GET /exchange-rates`). */
export interface ExchangeRate {
	id: string;
	fromCurrency: string;
	toCurrency: string;
	rate: string;
	rateDate: string;
	createdAt: string;
}

// ---------------------------------------------------------------------------
// Usuarios, preferencias, sesiones y 2FA
// ---------------------------------------------------------------------------

/** Usuario en el listado de administración (`GET /users`). */
export interface UserItem {
	id: string;
	name: string;
	email: string;
	emailVerified: boolean;
	createdAt: string;
	bannedAt: string | null;
	role: { name: string };
}

/** Invitación (`GET /users/invitations`). */
export interface InvitationItem {
	id: string;
	email: string;
	name: string;
	role: string;
	status: 'pending' | 'expired' | 'accepted' | 'revoked';
	expiresAt: string;
	createdAt: string;
}

/** Entrada de la waitlist (`GET /users/waitlist`). */
export interface WaitlistItem {
	id: string;
	email: string;
	status: 'pending' | 'invited' | 'registered';
	invitedAt: string | null;
	createdAt: string;
}

/** Preferencias del usuario (`GET /users/me/preferences`). */
export interface UserPreferences {
	userId: string;
	emailAlerts: boolean;
	weeklySummary: boolean;
}

/** Sesión activa del usuario (`GET /auth/sessions`). */
export interface ActiveSession {
	id: string;
	ipAddress: string | null;
	userAgent: string | null;
	location: string | null;
	createdAt: string;
	lastActiveAt: string;
	expiresAt: string;
	current: boolean;
}

/** Estado de la verificación en dos pasos (`GET /auth/2fa`). */
export interface TwoFactorStatus {
	enabled: boolean;
	pendingSetup: boolean;
	recoveryCodesLeft: number;
}
