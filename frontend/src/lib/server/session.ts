import { env } from '$env/dynamic/private';
import { dev } from '$app/environment';
import type { Cookies } from '@sveltejs/kit';

export const ACCESS_COOKIE = 'access_token_finexia';
export const REFRESH_COOKIE = 'refresh_token';

const ACCESS_COOKIE_MAX_AGE = 60 * 60 * 24 * 7;
// Fallback when the backend response doesn't carry a Max-Age; mirrors the
// backend default JWT_REFRESH_DURATION (30 days).
const DEFAULT_REFRESH_MAX_AGE = 60 * 60 * 24 * 30;

export type SessionEvent = {
	cookies: Cookies;
	fetch: typeof fetch;
};

type RefreshResult = {
	accessToken: string;
	refreshToken: string | null;
	refreshMaxAge: number | null;
};

/**
 * Extracts the rotated refresh token (and its real Max-Age) from a backend
 * response, so the cookie the frontend re-issues expires in step with the
 * backend's configured JWT_REFRESH_DURATION instead of a hardcoded value.
 */
export function parseRefreshSetCookie(
	response: Response
): { value: string; maxAge: number | null } | null {
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

export function setAccessCookie(cookies: Cookies, token: string): void {
	cookies.set(ACCESS_COOKIE, token, {
		path: '/',
		httpOnly: true,
		secure: !dev,
		maxAge: ACCESS_COOKIE_MAX_AGE,
		sameSite: 'lax'
	});
}

export function setRefreshCookie(cookies: Cookies, token: string, maxAge: number | null): void {
	cookies.set(REFRESH_COOKIE, token, {
		path: '/',
		httpOnly: true,
		secure: !dev,
		maxAge: maxAge ?? DEFAULT_REFRESH_MAX_AGE,
		sameSite: 'lax'
	});
}

export function clearSessionCookies(cookies: Cookies): void {
	cookies.delete(ACCESS_COOKIE, { path: '/' });
	cookies.delete(REFRESH_COOKIE, { path: '/' });
}

// Single-flight: concurrent requests carrying the same refresh token (e.g. link
// preload racing with the click navigation) must not each POST /auth/refresh.
// The backend rotates the refresh token on every call, so two concurrent calls
// with the same token would trip reuse detection and revoke the whole family.
// We dedupe by sharing the in-flight promise keyed by the refresh token; each
// request then sets its own cookies from the shared result.
const inFlightRefreshes = new Map<string, Promise<RefreshResult | null>>();

async function performRefresh(
	event: SessionEvent,
	refreshToken: string
): Promise<RefreshResult | null> {
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
export async function refreshAccessToken(
	event: SessionEvent,
	refreshToken: string
): Promise<string | null> {
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
