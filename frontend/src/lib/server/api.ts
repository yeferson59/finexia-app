import { env } from '$env/dynamic/private';
import { isRedirect, redirect } from '@sveltejs/kit';
import type { Cookies } from '@sveltejs/kit';

type AuthedEvent = {
	cookies: Cookies;
	fetch: typeof fetch;
};

/**
 * Server-side fetch to the backend API with the access token attached.
 *
 * On a 401 the session cookies are cleared and the request is redirected to
 * `/auth`, so loaders never silently render with empty data when the session has
 * expired (which previously left users staring at a blank dashboard). The token
 * refresh itself still happens in `hooks.server.ts` before loaders run; this is
 * the safety net for the rare case where the token expires in between.
 */
export async function authedFetch(
	event: AuthedEvent,
	path: string,
	init: RequestInit = {}
): Promise<Response> {
	const accessToken = event.cookies.get('access_token_finexia');

	const res = await event.fetch(`${env.BASE_API}${path}`, {
		...init,
		headers: {
			...(accessToken ? { Authorization: `Bearer ${accessToken}` } : {}),
			...init.headers
		}
	});

	if (res.status === 401) {
		event.cookies.delete('access_token_finexia', { path: '/' });
		event.cookies.delete('refresh_token', { path: '/' });
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
