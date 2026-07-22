import { env } from '$env/dynamic/private';
import { isRedirect, redirect } from '@sveltejs/kit';
import {
	ACCESS_COOKIE,
	REFRESH_COOKIE,
	clearSessionCookies,
	refreshAccessToken,
	type SessionEvent
} from '$lib/server/session';
import type { ApiEnvelope } from './types';

type AuthedEvent = SessionEvent;

/** Event shape the typed domain modules pass through to the client. */
export type ApiEvent = SessionEvent;

/**
 * Builds a fully-qualified backend URL. The single place that reads
 * `env.BASE_API`, so no loader/action constructs backend paths by hand (public
 * endpoints and logout, which can't use {@link authedFetch}, go through here).
 */
export function apiUrl(path: string): string {
	return `${env.BASE_API}${path}`;
}

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
	let accessToken = event.cookies.get(ACCESS_COOKIE);

	// Without an access token the backend answers 400 (missing JWT), not 401,
	// so resolve the session up front: refresh if possible, bail out otherwise.
	if (!accessToken) {
		const refreshToken = event.cookies.get(REFRESH_COOKIE);
		accessToken = (refreshToken && (await refreshAccessToken(event, refreshToken))) || undefined;
		if (!accessToken) {
			clearSessionCookies(event.cookies);
			redirect(302, '/auth');
		}
	}

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

/**
 * Flattened, typed view of a backend response used by the domain modules.
 *
 * Preserves everything the loaders/actions branch on today — the HTTP `ok`
 * and `status`, the envelope `success`/`message`/`details`/`action`, and the
 * typed `data` — so migrating a route to the API layer keeps its exact
 * behaviour without re-reading the raw `Response`.
 */
export interface ApiResult<T> {
	/** HTTP `res.ok`. */
	ok: boolean;
	/** HTTP status code (`0` when the backend was unreachable). */
	status: number;
	/** Envelope `success` flag. */
	success: boolean;
	/** Envelope `data`, or `null` when absent / on error. */
	data: T | null;
	message?: string;
	details?: string;
	action?: string;
}

/** Parses the standard envelope into an {@link ApiResult}. */
async function toResult<T>(res: Response | null): Promise<ApiResult<T>> {
	if (!res) return { ok: false, status: 0, success: false, data: null };
	const body = (await res.json().catch(() => ({}) as ApiEnvelope<T>)) as ApiEnvelope<T>;
	return {
		ok: res.ok,
		status: res.status,
		success: body.success ?? false,
		data: (body.data ?? null) as T | null,
		message: body.message,
		details: body.details,
		action: body.action
	};
}

/**
 * {@link authedFetch} + envelope parse. Redirects to `/auth` on an
 * unrecoverable 401 (throws the redirect, like `authedFetch`).
 */
export async function apiRequest<T>(
	event: AuthedEvent,
	path: string,
	init: RequestInit = {}
): Promise<ApiResult<T>> {
	return toResult<T>(await authedFetch(event, path, init));
}

/**
 * {@link authedFetchSafe} + envelope parse. Returns an {@link ApiResult} with
 * `ok: false` on network errors instead of throwing, for loaders that degrade
 * gracefully. A 401 still redirects to `/auth`.
 */
export async function apiRequestSafe<T>(
	event: AuthedEvent,
	path: string,
	init: RequestInit = {}
): Promise<ApiResult<T>> {
	return toResult<T>(await authedFetchSafe(event, path, init));
}
