import { env } from '$env/dynamic/private';
import { dev } from '$app/environment';
import type { Cookies } from '@sveltejs/kit';
import type { RefreshResult, SessionEvent } from './types';

export const ACCESS_COOKIE = 'access_token_finexia';
export const REFRESH_COOKIE = 'refresh_token';

const ACCESS_COOKIE_MAX_AGE = 60 * 60 * 24 * 7;
// Fallback when the backend response doesn't carry a Max-Age; mirrors the
// backend default JWT_REFRESH_DURATION (30 days).
const DEFAULT_REFRESH_MAX_AGE = 60 * 60 * 24 * 30;

/**
 * Extracts the rotated refresh token (and its real Max-Age) from a backend
 * response, so the cookie the frontend re-issues expires in step with the
 * backend's configured JWT_REFRESH_DURATION instead of a hardcoded value.
 */
export function parseRefreshSetCookie(response: Response) {
	const setCookies =
		typeof response.headers.getSetCookie === 'function'
			? response.headers.getSetCookie()
			: (response.headers.get('set-cookie')?.split(/,(?=\s*\w+=)/) ?? []);

	for (const cookie of setCookies) {
		const match = cookie.match(new RegExp(`(?:^|[,\\s])${REFRESH_COOKIE}=([^;,\\s]+)`));
		if (!match) continue;

		const maxAge = cookie.match(/;\s*max-age=(\d+)/i)?.[1];
		return { value: match[1], maxAge: maxAge ? Number(maxAge) : null };
	}

	return null;
}

export function setAccessCookie(cookies: Cookies, token: string) {
	cookies.set(ACCESS_COOKIE, token, {
		path: '/',
		httpOnly: true,
		secure: !dev,
		maxAge: ACCESS_COOKIE_MAX_AGE,
		sameSite: 'lax'
	});
}

export function setRefreshCookie(cookies: Cookies, token: string, maxAge: number | null) {
	cookies.set(REFRESH_COOKIE, token, {
		path: '/',
		httpOnly: true,
		secure: !dev,
		maxAge: maxAge ?? DEFAULT_REFRESH_MAX_AGE,
		sameSite: 'lax'
	});
}

export function clearSessionCookies(cookies: Cookies) {
	cookies.delete(ACCESS_COOKIE, { path: '/' });
	cookies.delete(REFRESH_COOKIE, { path: '/' });
}

/**
 * Cookie de "vuelve aquí después de iniciar sesión".
 *
 * Existe por una página concreta: el consentimiento OAuth. A esa se llega desde
 * un enlace que abre el cliente MCP, así que llegar sin sesión es lo normal, y
 * mandar al usuario a `/dashboard` tras el login pierde la petición que estaba
 * autorizando — el cliente se queda esperando un código que ya nunca llega.
 *
 * Va en cookie y no en query porque las acciones de formulario de SvelteKit
 * postean a `?/login`, lo que se lleva por delante cualquier parámetro que
 * hubiera en la URL.
 */
const RETURN_COOKIE = 'finexia_return_to';

// Un minuto de margen sobre lo que tarda un login. Más allá de eso, quien
// vuelva es alguien que abandonó el flujo, y llevarlo a una pantalla de
// consentimiento caducada es peor que llevarlo al dashboard.
const RETURN_COOKIE_MAX_AGE = 15 * 60;

/**
 * Guarda a dónde volver. Solo acepta rutas locales: cualquier cosa que empiece
 * por `//` o traiga esquema es un redirect abierto esperando a que alguien lo
 * encadene con un login legítimo, así que se descarta en silencio.
 */
export function setReturnTo(cookies: Cookies, path: string) {
	if (!isLocalPath(path)) return;

	cookies.set(RETURN_COOKIE, path, {
		path: '/',
		httpOnly: true,
		secure: !dev,
		maxAge: RETURN_COOKIE_MAX_AGE,
		sameSite: 'lax'
	});
}

/**
 * Lee el destino y lo consume: es de un solo uso, para que un login posterior
 * no reenvíe a una pantalla que ya se resolvió.
 */
export function takeReturnTo(cookies: Cookies): string | null {
	const target = cookies.get(RETURN_COOKIE);
	if (!target) return null;

	cookies.delete(RETURN_COOKIE, { path: '/' });

	return isLocalPath(target) ? target : null;
}

/**
 * Se valida al escribir *y* al leer. Escribir vale para lo que este código
 * genera; leer vale para lo que llegue en una cookie que el navegador ya tenía.
 */
function isLocalPath(path: string): boolean {
	return path.startsWith('/') && !path.startsWith('//') && !path.includes('\\');
}

// Single-flight: concurrent requests carrying the same refresh token (e.g. link
// preload racing with the click navigation) must not each POST /auth/refresh.
// The backend rotates the refresh token on every call, so two concurrent calls
// with the same token would trip reuse detection and revoke the whole family.
// We dedupe by sharing the in-flight promise keyed by the refresh token; each
// request then sets its own cookies from the shared result.
const inFlightRefreshes = new Map<string, Promise<RefreshResult | null>>();

async function performRefresh(event: SessionEvent, refreshToken: string) {
	const res = await event.fetch(`${env.BASE_API}/auth/refresh`, {
		method: 'POST',
		headers: { Cookie: `${REFRESH_COOKIE}=${refreshToken}` }
	});

	// A 5xx is a backend problem, not a verdict on the token: throw so callers
	// keep the cookies instead of logging the user out over a transient outage.
	if (res.status >= 500) {
		throw new Error(`refresh failed with status ${res.status}`);
	}

	if (!res.ok) return null;

	const { data, success } = await res.json();
	if (!success || !data?.accessToken) return null;

	const rotated = parseRefreshSetCookie(res);

	return {
		accessToken: data.accessToken as string,
		refreshToken: rotated?.value ?? null,
		refreshMaxAge: rotated?.maxAge ?? null
	};
}

/**
 * Exchanges the refresh token for a new access token and updates both session
 * cookies. Returns the new access token, or `null` when the backend rejected
 * the refresh token (in which case the caller decides whether to clear
 * cookies). Network errors and backend 5xx responses are thrown, NOT returned
 * as `null`: a transient outage must never be treated as an invalid session.
 */
export async function refreshAccessToken(event: SessionEvent, refreshToken: string) {
	let pending = inFlightRefreshes.get(refreshToken);

	if (!pending) {
		pending = performRefresh(event, refreshToken).finally(() => {
			inFlightRefreshes.delete(refreshToken);
		});
		inFlightRefreshes.set(refreshToken, pending);
	}

	const result = await pending;
	if (!result) return null;

	setAccessCookie(event.cookies, result.accessToken);
	if (result.refreshToken) {
		setRefreshCookie(event.cookies, result.refreshToken, result.refreshMaxAge);
	}

	return result.accessToken;
}
