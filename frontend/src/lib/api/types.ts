/**
 * Contratos HTTP compartidos con el backend.
 *
 * Única fuente de verdad de los shapes que devuelve la API: los tipos se
 * **derivan** de los schemas de `schemas.ts` con `z.infer`, así que un cambio de
 * contrato se escribe una sola vez y no puede desalinearse del schema con el
 * que la capa de API valida en desarrollo.
 *
 * Este módulo sigue siendo el sitio del que todo el mundo importa los tipos
 * (`import type { Asset } from '$lib/api/types'`), que en compilación se borran.
 * Quien necesite el schema en tiempo de ejecución lo importa de `schemas.ts`.
 */

import type { z } from 'zod';
import type {
	activeSessionSchema,
	allocationItemSchema,
	assetPriceSchema,
	assetSchema,
	exchangeRateSchema,
	growthDataPointSchema,
	growthSummarySchema,
	holdingSchema,
	importResultSchema,
	invitationItemSchema,
	marketCredentialSchema,
	marketPriceSchema,
	marketProviderSchema,
	marketRateSchema,
	marketSyncResultSchema,
	pageMetaSchema,
	pagedTransactionsSchema,
	platformSchema,
	portfolioDetailSchema,
	portfolioGrowthSchema,
	portfolioSummarySchema,
	riskSchema,
	topTransactionSchema,
	transactionSchema,
	twoFactorStatusSchema,
	userItemSchema,
	userPreferencesSchema,
	userTransactionSchema,
	waitlistItemSchema
} from './schemas';

// ---------------------------------------------------------------------------
// Sobre de respuesta y paginación (§1.1 y §1.5 de docs/API.md)
// ---------------------------------------------------------------------------

/**
 * Sobre estándar que envuelve toda respuesta JSON del backend.
 *
 * No se deriva de un schema porque es genérico sobre `data` y solo se usa para
 * tipar el parseo del sobre en `client.ts`; lo que se valida es el `data`.
 */
export interface ApiEnvelope<T = unknown> {
	success: boolean;
	message?: string;
	details?: string;
	/** Código estable de máquina en algunos flujos (p. ej. `auth:login:2fa`). */
	action?: string;
	data?: T;
	timestamp?: string;
}

/** Bloque de metadatos de las rutas paginadas. */
export type PageMeta = z.infer<typeof pageMetaSchema>;

/** `data` de una ruta paginada: lista de items + metadatos. */
export interface Paginated<T> {
	items: T[];
	metaData: PageMeta;
}

// ---------------------------------------------------------------------------
// Portfolios
// ---------------------------------------------------------------------------

/** Resumen de un portfolio (`GET /portfolios/summary`). */
export type PortfolioSummary = z.infer<typeof portfolioSummarySchema>;

/** Posición dentro de un portfolio (holdings de `GET /portfolios/:id`). */
export type Holding = z.infer<typeof holdingSchema>;

/** Detalle completo de un portfolio (`GET /portfolios/:id`). */
export type PortfolioDetail = z.infer<typeof portfolioDetailSchema>;

/** Nivel de riesgo del catálogo (`GET /portfolios/risks`). */
export type Risk = z.infer<typeof riskSchema>;

/** Asignación por categoría de activo (`GET /portfolios/allocation`). */
export type AllocationItem = z.infer<typeof allocationItemSchema>;

/** Mayor transacción de un portfolio (`GET /portfolios/:id/top-transaction`). */
export type TopTransaction = z.infer<typeof topTransactionSchema>;

/** Punto de la serie de crecimiento. */
export type GrowthDataPoint = z.infer<typeof growthDataPointSchema>;

/** Resumen agregado de la serie de crecimiento. */
export type GrowthSummary = z.infer<typeof growthSummarySchema>;

/** Crecimiento (`GET /portfolios/growth` y `GET /portfolios/:id/growth`). */
export type PortfolioGrowth = z.infer<typeof portfolioGrowthSchema>;

// ---------------------------------------------------------------------------
// Transacciones
// ---------------------------------------------------------------------------

/** Transacción de una posición (`GET /portfolios/:id/assets/:symbol/transactions`). */
export type Transaction = z.infer<typeof transactionSchema>;

/** Transacción del usuario con datos del activo (`GET /portfolios/transactions`). */
export type UserTransaction = z.infer<typeof userTransactionSchema>;

/** `data` de las transacciones paginadas por asset. */
export type PagedTransactions = z.infer<typeof pagedTransactionsSchema>;

// ---------------------------------------------------------------------------
// Plataformas / fuentes
// ---------------------------------------------------------------------------

/** Plataforma / fuente (`GET /portfolios/sources`). */
export type Platform = z.infer<typeof platformSchema>;

// ---------------------------------------------------------------------------
// Assets y tasas de cambio (mercado)
// ---------------------------------------------------------------------------

/** Precio de un asset. */
export type AssetPrice = z.infer<typeof assetPriceSchema>;

/** Asset del catálogo (`GET /portfolios/assets`). */
export type Asset = z.infer<typeof assetSchema>;

/** Resultado de un import masivo (`POST /assets/import`, `POST /exchange-rates/import`). */
export type ImportResult = z.infer<typeof importResultSchema>;

/** Tasa de cambio (`GET /exchange-rates`). */
export type ExchangeRate = z.infer<typeof exchangeRateSchema>;

// ---------------------------------------------------------------------------
// Usuarios, preferencias, sesiones y 2FA
// ---------------------------------------------------------------------------

/** Usuario en el listado de administración (`GET /users`). */
export type UserItem = z.infer<typeof userItemSchema>;

/** Invitación (`GET /users/invitations`). */
export type InvitationItem = z.infer<typeof invitationItemSchema>;

/** Entrada de la waitlist (`GET /users/waitlist`). */
export type WaitlistItem = z.infer<typeof waitlistItemSchema>;

/** Preferencias del usuario (`GET /users/me/preferences`). */
export type UserPreferences = z.infer<typeof userPreferencesSchema>;

/** Sesión activa del usuario (`GET /auth/sessions`). */
export type ActiveSession = z.infer<typeof activeSessionSchema>;

/** Estado de la verificación en dos pasos (`GET /auth/2fa`). */
export type TwoFactorStatus = z.infer<typeof twoFactorStatusSchema>;

/** Proveedor de datos de mercado para el que se puede aportar una clave. */
export type MarketProvider = z.infer<typeof marketProviderSchema>;

/** Estado de una clave de proveedor del usuario (`GET /market/credentials`). */
export type MarketCredential = z.infer<typeof marketCredentialSchema>;

/** Un precio que trajo la clave del propio usuario (`POST /market/sync`). */
export type MarketPrice = z.infer<typeof marketPriceSchema>;

/** Una tasa de cambio que trajo la clave del propio usuario. */
export type MarketRate = z.infer<typeof marketRateSchema>;

/** Resultado de `POST /market/sync`. */
export type MarketSyncResult = z.infer<typeof marketSyncResultSchema>;
