/**
 * Flujos de autenticación públicos (login, registro, 2FA-login, reset y
 * verificación de email, invitaciones) y logout.
 *
 * A diferencia del resto de módulos, estas funciones no llevan sesión: reciben
 * el `fetch` del evento y construyen la URL con {@link apiUrl}. Devuelven la
 * `Response` cruda porque las actions necesitan leer cabeceras `Set-Cookie`,
 * códigos `action` y branch por status, sin perder ningún matiz.
 */
import { apiUrl } from './client';

type Fetch = typeof fetch;

/** POST con cuerpo JSON al backend, sin autenticación. */
function postJson(fetchFn: Fetch, path: string, body: unknown): Promise<Response> {
	return fetchFn(apiUrl(path), {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
}

/** `POST /auth/login`. */
export function login(fetchFn: Fetch, body: { email: string; password: string }): Promise<Response> {
	return postJson(fetchFn, '/auth/login', body);
}

/** `POST /auth/2fa/login` — segundo paso del login 2FA. */
export function twoFactorLogin(
	fetchFn: Fetch,
	body: { token: string; code: string }
): Promise<Response> {
	return postJson(fetchFn, '/auth/2fa/login', body);
}

/** `POST /auth/register`. */
export function register(fetchFn: Fetch, body: Record<string, unknown>): Promise<Response> {
	return postJson(fetchFn, '/auth/register', body);
}

/** `POST /auth/password-reset` — solicita el enlace de recuperación. */
export function requestPasswordReset(fetchFn: Fetch, email: string): Promise<Response> {
	return postJson(fetchFn, '/auth/password-reset', { email });
}

/** `GET /auth/password-reset?token=` — valida un token de reset. */
export function validatePasswordResetToken(fetchFn: Fetch, token: string): Promise<Response> {
	return fetchFn(apiUrl(`/auth/password-reset?token=${encodeURIComponent(token)}`));
}

/** `POST /auth/password-reset/confirm` — confirma el reset con nueva contraseña. */
export function confirmPasswordReset(
	fetchFn: Fetch,
	body: { token: string; password: string }
): Promise<Response> {
	return postJson(fetchFn, '/auth/password-reset/confirm', body);
}

/** `GET /auth/invitations?token=` — valida un token de invitación. */
export function validateInvitation(fetchFn: Fetch, token: string): Promise<Response> {
	return fetchFn(apiUrl(`/auth/invitations?token=${encodeURIComponent(token)}`));
}

/** `POST /auth/invitations/accept` — acepta la invitación fijando contraseña. */
export function acceptInvitation(
	fetchFn: Fetch,
	body: { token: string; name: string; password: string }
): Promise<Response> {
	return postJson(fetchFn, '/auth/invitations/accept', body);
}

/** `POST /auth/verify-email` — (re)envía el enlace de verificación. */
export function requestEmailVerification(fetchFn: Fetch, email: string): Promise<Response> {
	return postJson(fetchFn, '/auth/verify-email', { email });
}

/** `GET /auth/verify-email?token=` — valida un token de verificación. */
export function validateEmailVerificationToken(fetchFn: Fetch, token: string): Promise<Response> {
	return fetchFn(apiUrl(`/auth/verify-email?token=${encodeURIComponent(token)}`));
}

/** `POST /auth/verify-email/confirm` — marca el email como verificado. */
export function confirmEmailVerification(fetchFn: Fetch, token: string): Promise<Response> {
	return postJson(fetchFn, '/auth/verify-email/confirm', { token });
}

/**
 * `POST /auth/logout` — cierra la sesión actual. Se pasa el access token (Bearer)
 * y el refresh token (cookie) explícitamente porque el logout no usa el flujo
 * de {@link authedFetch}.
 */
export function logout(
	fetchFn: Fetch,
	tokens: { accessToken: string; refreshToken?: string; refreshCookieName: string }
): Promise<Response> {
	return fetchFn(apiUrl('/auth/logout'), {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json',
			Authorization: `Bearer ${tokens.accessToken}`,
			...(tokens.refreshToken
				? { Cookie: `${tokens.refreshCookieName}=${tokens.refreshToken}` }
				: {})
		}
	});
}
