/**
 * Usuarios y cuenta: perfil, preferencias, avatar, contraseña, administración
 * de usuarios/invitaciones/waitlist y gestión de sesiones + 2FA (`/auth/*`
 * autenticado, disparado desde ajustes).
 */
import { apiRequest, apiRequestSafe, type ApiEvent, type ApiResult } from './client';
import type {
	ActiveSession,
	InvitationItem,
	Paginated,
	TwoFactorStatus,
	UserItem,
	UserPreferences,
	WaitlistItem
} from './types';

// --- Perfil / preferencias del usuario ------------------------------------

/** `GET /users/me/preferences` — preferencias del usuario. */
export function getPreferences(event: ApiEvent): Promise<ApiResult<UserPreferences>> {
	return apiRequestSafe<UserPreferences>(event, '/users/me/preferences');
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
	return apiRequestSafe<Paginated<UserItem>>(event, `/users?page=${page}&limit=${limit}`);
}

/** `GET /users/invitations` — listado paginado de invitaciones (admin). */
export function getInvitations(
	event: ApiEvent,
	opts: { page?: number; limit?: number } = {}
): Promise<ApiResult<Paginated<InvitationItem>>> {
	const page = opts.page ?? 1;
	const limit = opts.limit ?? 50;
	return apiRequestSafe<Paginated<InvitationItem>>(
		event,
		`/users/invitations?page=${page}&limit=${limit}`
	);
}

/** `GET /users/waitlist` — listado paginado de la waitlist (admin). */
export function getWaitlist(
	event: ApiEvent,
	opts: { page?: number; limit?: number } = {}
): Promise<ApiResult<Paginated<WaitlistItem>>> {
	const page = opts.page ?? 1;
	const limit = opts.limit ?? 50;
	return apiRequestSafe<Paginated<WaitlistItem>>(
		event,
		`/users/waitlist?page=${page}&limit=${limit}`
	);
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
	return apiRequestSafe<ActiveSession[]>(event, '/auth/sessions');
}

/** `GET /auth/2fa` — estado de la verificación en dos pasos. */
export function getTwoFactorStatus(event: ApiEvent): Promise<ApiResult<TwoFactorStatus>> {
	return apiRequestSafe<TwoFactorStatus>(event, '/auth/2fa');
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
