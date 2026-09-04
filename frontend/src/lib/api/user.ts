/**
 * Usuarios y cuenta: perfil, preferencias, avatar, contraseña, administración
 * de usuarios/invitaciones/waitlist y gestión de sesiones + 2FA (`/auth/*`
 * autenticado, disparado desde ajustes).
 */
import { apiRequest, apiRequestSafe, type ApiEvent, type ApiResult } from './client';
import type {
	ActiveSession,
	InvitationItem,
	MCPToken,
	MCPTokenSecret,
	OAuthConsent,
	OAuthGrant,
	Paginated,
	TwoFactorStatus,
	UserItem,
	UserPreferences,
	WaitlistItem
} from './types';
import {
	activeSessionSchema,
	invitationItemSchema,
	mcpTokenSchema,
	mcpTokenSecretSchema,
	oauthConsentSchema,
	oauthGrantSchema,
	paginatedSchema,
	twoFactorStatusSchema,
	userItemSchema,
	userPreferencesSchema,
	waitlistItemSchema
} from './schemas';
import { z } from 'zod';

// --- Perfil / preferencias del usuario ------------------------------------

/** `GET /users/me/preferences` — preferencias del usuario. */
export function getPreferences(event: ApiEvent): Promise<ApiResult<UserPreferences>> {
	return apiRequestSafe(event, '/users/me/preferences', {}, userPreferencesSchema);
}

/** `PATCH /users/me/preferences` — actualiza las preferencias. */
export function updatePreferences(
	event: ApiEvent,
	body: Record<string, unknown>
): Promise<ApiResult<unknown>> {
	return apiRequest<unknown>(event, '/users/me/preferences', {
		method: 'PATCH',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
}

/** `PATCH /users/me` — actualiza el perfil propio. */
export function updateProfile(
	event: ApiEvent,
	body: Record<string, unknown>
): Promise<ApiResult<unknown>> {
	return apiRequest<unknown>(event, '/users/me', {
		method: 'PATCH',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
}

/** `POST /users/me/avatar` — sube el avatar (multipart). */
export function uploadAvatar(
	event: ApiEvent,
	form: FormData
): Promise<ApiResult<{ image?: string }>> {
	return apiRequest<{ image?: string }>(event, '/users/me/avatar', { method: 'POST', body: form });
}

/** `PATCH /users/me/password` — cambia la contraseña. */
export function changePassword(
	event: ApiEvent,
	body: Record<string, unknown>
): Promise<ApiResult<unknown>> {
	return apiRequest<unknown>(event, '/users/me/password', {
		method: 'PATCH',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
}

// --- Administración de usuarios -------------------------------------------

/** `GET /users` — listado paginado de usuarios (admin). */
export function getUsers(
	event: ApiEvent,
	opts: { page?: number; limit?: number } = {}
): Promise<ApiResult<Paginated<UserItem>>> {
	const page = opts.page ?? 1;
	const limit = opts.limit ?? 20;
	return apiRequestSafe(
		event,
		`/users?page=${page}&limit=${limit}`,
		{},
		paginatedSchema(userItemSchema)
	);
}

/** `GET /users/invitations` — listado paginado de invitaciones (admin). */
export function getInvitations(
	event: ApiEvent,
	opts: { page?: number; limit?: number } = {}
): Promise<ApiResult<Paginated<InvitationItem>>> {
	const page = opts.page ?? 1;
	const limit = opts.limit ?? 50;
	return apiRequestSafe(
		event,
		`/users/invitations?page=${page}&limit=${limit}`,
		{},
		paginatedSchema(invitationItemSchema)
	);
}

/** `GET /users/waitlist` — listado paginado de la waitlist (admin). */
export function getWaitlist(
	event: ApiEvent,
	opts: { page?: number; limit?: number } = {}
): Promise<ApiResult<Paginated<WaitlistItem>>> {
	const page = opts.page ?? 1;
	const limit = opts.limit ?? 50;
	return apiRequestSafe(
		event,
		`/users/waitlist?page=${page}&limit=${limit}`,
		{},
		paginatedSchema(waitlistItemSchema)
	);
}

/** `DELETE /users/waitlist/:id` — elimina una entrada de la waitlist (admin). */
export function deleteWaitlistEntry(event: ApiEvent, id: string): Promise<ApiResult<unknown>> {
	return apiRequest<unknown>(event, `/users/waitlist/${id}`, { method: 'DELETE' });
}

/** `POST /users/invitations` — crea una invitación (admin). */
export function inviteUser(
	event: ApiEvent,
	body: Record<string, unknown>
): Promise<ApiResult<unknown>> {
	return apiRequest<unknown>(event, '/users/invitations', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
}

/** `POST /users/invitations/:id/resend` — reenvía una invitación (admin). */
export function resendInvitation(event: ApiEvent, id: string): Promise<ApiResult<unknown>> {
	return apiRequest<unknown>(event, `/users/invitations/${id}/resend`, { method: 'POST' });
}

/** `DELETE /users/invitations/:id` — revoca una invitación (admin). */
export function revokeInvitation(event: ApiEvent, id: string): Promise<ApiResult<unknown>> {
	return apiRequest<unknown>(event, `/users/invitations/${id}`, { method: 'DELETE' });
}

/** `DELETE /users/:id` — elimina un usuario (admin). */
export function deleteUser(event: ApiEvent, id: string): Promise<ApiResult<unknown>> {
	return apiRequest<unknown>(event, `/users/${id}`, { method: 'DELETE' });
}

/** `PATCH /users/:id/ban` — banea/desbanea un usuario (admin). */
export function banUser(
	event: ApiEvent,
	id: string,
	body: Record<string, unknown>
): Promise<ApiResult<unknown>> {
	return apiRequest<unknown>(event, `/users/${id}/ban`, {
		method: 'PATCH',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
}

// --- Sesiones y 2FA -------------------------------------------------------

/** `GET /auth/sessions` — sesiones activas del usuario. */
export function getSessions(event: ApiEvent): Promise<ApiResult<ActiveSession[]>> {
	return apiRequestSafe(event, '/auth/sessions', {}, z.array(activeSessionSchema));
}

/** `GET /auth/2fa` — estado de la verificación en dos pasos. */
export function getTwoFactorStatus(event: ApiEvent): Promise<ApiResult<TwoFactorStatus>> {
	return apiRequestSafe(event, '/auth/2fa', {}, twoFactorStatusSchema);
}

/** `DELETE /auth/sessions/:id` — revoca una sesión. */
export function revokeSession(event: ApiEvent, id: string): Promise<ApiResult<unknown>> {
	return apiRequest<unknown>(event, `/auth/sessions/${id}`, { method: 'DELETE' });
}

/** `POST /auth/sessions/revoke-others` — revoca las demás sesiones. */
export function revokeOtherSessions(event: ApiEvent): Promise<ApiResult<{ revoked: number }>> {
	return apiRequest<{ revoked: number }>(event, '/auth/sessions/revoke-others', { method: 'POST' });
}

/** `POST /auth/2fa/setup` — inicia el enrolamiento 2FA. */
export function setupTwoFactor(
	event: ApiEvent,
	body: Record<string, unknown>
): Promise<ApiResult<{ secret?: string; otpauthUrl?: string }>> {
	return apiRequest<{ secret?: string; otpauthUrl?: string }>(event, '/auth/2fa/setup', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
}

/** `POST /auth/2fa/enable` — confirma y activa 2FA. */
export function enableTwoFactor(
	event: ApiEvent,
	body: Record<string, unknown>
): Promise<ApiResult<{ recoveryCodes?: string[] }>> {
	return apiRequest<{ recoveryCodes?: string[] }>(event, '/auth/2fa/enable', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
}

/** `POST /auth/2fa/disable` — desactiva 2FA. */
export function disableTwoFactor(
	event: ApiEvent,
	body: Record<string, unknown>
): Promise<ApiResult<unknown>> {
	return apiRequest<unknown>(event, '/auth/2fa/disable', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
}

/** `POST /auth/2fa/recovery-codes` — regenera los códigos de recuperación. */
export function regenerateRecoveryCodes(
	event: ApiEvent,
	body: Record<string, unknown>
): Promise<ApiResult<{ recoveryCodes?: string[] }>> {
	return apiRequest<{ recoveryCodes?: string[] }>(event, '/auth/2fa/recovery-codes', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
}

// --- Tokens del endpoint MCP ---------------------------------------------
//
// El secreto solo viaja en la respuesta de crear y de rotar. Ninguna de estas
// funciones lo guarda en ningún sitio: se lo devuelve a la action, que lo pone
// una vez en el `form` de la página y ahí acaba su vida.

/** `GET /auth/mcp-tokens` — tokens del usuario, sin secreto. */
export function getMCPTokens(event: ApiEvent): Promise<ApiResult<MCPToken[]>> {
	return apiRequestSafe(event, '/auth/mcp-tokens', {}, z.array(mcpTokenSchema));
}

/** `POST /auth/mcp-tokens` — crea uno y devuelve su secreto, una sola vez. */
export function createMCPToken(
	event: ApiEvent,
	body: { name: string; expiresInDays: number }
): Promise<ApiResult<MCPTokenSecret>> {
	return apiRequest<MCPTokenSecret>(
		event,
		'/auth/mcp-tokens',
		{
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(body)
		},
		mcpTokenSecretSchema
	);
}

/** `POST /auth/mcp-tokens/:id/rotate` — reemplaza el secreto del token. */
export function rotateMCPToken(
	event: ApiEvent,
	id: string,
	body: { expiresInDays: number }
): Promise<ApiResult<MCPTokenSecret>> {
	return apiRequest<MCPTokenSecret>(
		event,
		`/auth/mcp-tokens/${id}/rotate`,
		{
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(body)
		},
		mcpTokenSecretSchema
	);
}

/** `DELETE /auth/mcp-tokens/:id` — revoca el token. */
export function deleteMCPToken(event: ApiEvent, id: string): Promise<ApiResult<unknown>> {
	return apiRequest<unknown>(event, `/auth/mcp-tokens/${id}`, { method: 'DELETE' });
}

// --- OAuth: consentimiento y aplicaciones conectadas -----------------------

/**
 * `GET /auth/oauth/consent/:id` — la petición de autorización que la pantalla
 * de consentimiento va a mostrar.
 *
 * El id es lo único que llegó por el navegador: cliente, ámbitos y URI de
 * retorno salen de la fila que el backend aparcó en `/oauth/authorize`, así que
 * un id manipulado no puede cambiar *sobre qué* se pregunta, solo apuntar a
 * otra petición (que no será suya y devolverá 404).
 */
export function getOAuthConsent(
	event: ApiEvent,
	requestId: string
): Promise<ApiResult<OAuthConsent>> {
	return apiRequestSafe(event, `/auth/oauth/consent/${requestId}`, {}, oauthConsentSchema);
}

/**
 * `POST /auth/oauth/consent/:id` — aprueba o deniega, y devuelve a dónde hay
 * que enviar el navegador.
 *
 * El redirect lo calcula el backend a partir de la URI registrada del cliente,
 * nunca de nada que venga del formulario: es la diferencia entre un flujo OAuth
 * y un redirect abierto.
 */
export function decideOAuthConsent(
	event: ApiEvent,
	requestId: string,
	approved: boolean
): Promise<ApiResult<{ redirectTo: string }>> {
	return apiRequest<{ redirectTo: string }>(event, `/auth/oauth/consent/${requestId}`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ approved })
	});
}

/** `GET /auth/oauth-grants` — aplicaciones externas con acceso a `/mcp`. */
export function listOAuthGrants(event: ApiEvent): Promise<ApiResult<OAuthGrant[]>> {
	return apiRequestSafe(event, '/auth/oauth-grants', {}, z.array(oauthGrantSchema));
}

/** `DELETE /auth/oauth-grants/:id` — desconecta una aplicación. */
export function revokeOAuthGrant(event: ApiEvent, grantId: string): Promise<ApiResult<unknown>> {
	return apiRequest<unknown>(event, `/auth/oauth-grants/${grantId}`, { method: 'DELETE' });
}
