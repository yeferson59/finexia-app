/**
 * Usuarios y mercado propio — contratos HTTP como schemas Zod.
 *
 * Parte de `lib/api/schemas`: el porqué de que sean schemas y no interfaces
 * está en el `index.ts` de la carpeta.
 */

import { z } from 'zod';

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
