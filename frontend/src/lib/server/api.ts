import { env } from '$env/dynamic/private';
import { isRedirect, redirect } from '@sveltejs/kit';
import {
	ACCESS_COOKIE,
	REFRESH_COOKIE,
	clearSessionCookies,
	refreshAccessToken,
	type SessionEvent
} from '$lib/server/session';

type AuthedEvent = SessionEvent;

function doFetch(
	event: AuthedEvent,
	path: string,
	init: RequestInit,
	accessToken: string | undefined
): Promise<Response> {
	return event.fetch(`${env.BASE_API}${path}`, {
		...init,
		headers: {
			...(accessToken ? { Authorization: `Bearer ${accessToken}` } : {}),
			...init.headers
		}
	});
}

/**
 * Server-side fetch to the backend API with the access token attached.
 *
 * On a 401 the access token is refreshed (single-flight, shared with
 * `hooks.server.ts`) and the request retried once, covering the window where
 * the token expires between the hook and the loader. Only when the refresh
 * token itself is rejected are the session cookies cleared and the request
 * redirected to `/auth`, so loaders never silently render with empty data when
 * the session has expired — and a still-valid refresh token is never discarded
 * over a transient 401.
 */
export async function authedFetch(
	event: AuthedEvent,
	path: string,
	init: RequestInit = {}
): Promise<Response> {
	const accessToken = event.cookies.get(ACCESS_COOKIE);

	let res = await doFetch(event, path, init, accessToken);

	if (res.status === 401) {
		const refreshToken = event.cookies.get(REFRESH_COOKIE);
		const newAccessToken = refreshToken ? await refreshAccessToken(event, refreshToken) : null;

		if (newAccessToken) {
			res = await doFetch(event, path, init, newAccessToken);
			if (res.status !== 401) return res;
		}

		clearSessionCookies(event.cookies);
		redirect(302, '/auth');
	}

	return res;
}

/**
 * Like {@link authedFetch} but returns `null` on network errors instead of
 * throwing, for loaders that want to degrade gracefully (e.g. dashboard widgets)
 * when the backend is unreachable. A 401 still redirects to `/auth`.
 */
export async function authedFetchSafe(
	event: AuthedEvent,
	path: string,
	init: RequestInit = {}
): Promise<Response | null> {
	try {
		return await authedFetch(event, path, init);
	} catch (err) {
		if (isRedirect(err)) throw err;
		return null;
	}
}
