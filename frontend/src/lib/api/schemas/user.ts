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

/**
 * Token personal del endpoint MCP (`GET /auth/mcp-tokens`).
 *
 * No hay campo para el secreto, igual que en `marketCredentialSchema` y por el
 * mismo motivo: el backend solo guarda su hash, así que un listado no puede
 * devolverlo. `last4` es lo único que queda para reconocer cuál es cuál.
 */
export const mcpTokenSchema = z.object({
	id: z.string(),
	name: z.string(),
	last4: z.string(),
	expiresAt: z.string().nullable(),
	lastUsedAt: z.string().nullable(),
	rotatedAt: z.string().nullable(),
	createdAt: z.string(),
	expired: z.boolean()
});

/**
 * Lo mismo, con el secreto: es la respuesta de crear y de rotar, las dos únicas
 * veces que el token viaja en claro.
 */
export const mcpTokenSecretSchema = mcpTokenSchema.extend({
	token: z.string()
});

/**
 * Una petición de autorización OAuth esperando aprobación
 * (`GET /auth/oauth/consent/:id`).
 *
 * Todo lo que hay aquí lo escribió el backend leyendo la fila que aparcó en
 * `/oauth/authorize`; el navegador solo llevó el id. `clientName` y `logoUri`
 * los eligió quien registró el cliente, así que se pintan como texto y como
 * imagen, nunca como HTML ni como enlace de confianza.
 */
export const oauthConsentSchema = z.object({
	requestId: z.string(),
	clientName: z.string(),
	clientUri: z.string().optional(),
	logoUri: z.string().optional(),
	redirectUri: z.string(),
	scopes: z.array(z.string()),
	expiresAt: z.string()
});

/**
 * Una aplicación conectada (`GET /auth/oauth-grants`).
 *
 * No hay campo para los tokens, por el mismo motivo que en `mcpTokenSchema`:
 * el backend solo guarda sus hashes. Lo que identifica la conexión es el nombre
 * del cliente y cuándo se usó por última vez.
 */
export const oauthGrantSchema = z.object({
	id: z.string(),
	clientName: z.string(),
	clientUri: z.string().optional(),
	scopes: z.array(z.string()),
	lastUsedAt: z.string().nullable(),
	createdAt: z.string()
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
