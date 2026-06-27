import type { Handle } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';
import { dev } from '$app/environment';

type SessionResponse = {
	data: {
		user: {
			name: string;
			email: string;
			emailVerified: boolean;
			image: string;
			role: string;
			preferredCurrency: string;
			createdAt: string;
			updatedAt: string;
		};
		session: {
			id: string;
			userId: string;
			expiresAt: string;
			ipAddress: string | null;
			userAgent: string | null;
			createdAt: string;
		};
	};
	success: boolean;
	message: string;
	details: string;
};

// Private areas must never be indexed even if a URL leaks past robots.txt.
const PRIVATE_PREFIXES = ['/dashboard', '/auth', '/demo'];

function withRobots(event: Parameters<Handle>[0]['event'], response: Response): Response {
	if (PRIVATE_PREFIXES.some((prefix) => event.url.pathname.startsWith(prefix))) {
		response.headers.set('X-Robots-Tag', 'noindex, nofollow');
	}

	return response;
}

// Only private areas need the session; validating it on public pages (landing,
// sitemap, etc.) would add a backend round-trip to the most-visited routes.
function needsSession(pathname: string): boolean {
	return PRIVATE_PREFIXES.some((prefix) => pathname.startsWith(prefix));
}

type RefreshResult = { accessToken: string; refreshToken: string | null };

// Single-flight: concurrent requests carrying the same refresh token (e.g. link
// preload racing with the click navigation) must not each POST /auth/refresh.
// The backend rotates the refresh token on every call, so two concurrent calls
// with the same token would trip reuse detection and revoke the whole family.
// We dedupe by sharing the in-flight promise keyed by the refresh token; each
// request then sets its own cookies from the shared result.
const inFlightRefreshes = new Map<string, Promise<RefreshResult | null>>();

async function performRefresh(
	event: Parameters<Handle>[0]['event'],
	refreshToken: string
): Promise<RefreshResult | null> {
	const res = await event.fetch(`${env.BASE_API}/auth/refresh`, {
		method: 'POST',
		headers: { Cookie: `refresh_token=${refreshToken}` }
	});

	if (!res.ok) return null;

	const { data, success } = await res.json();
	if (!success || !data?.accessToken) return null;

	const newRefreshToken =
		res.headers.get('set-cookie')?.match(/refresh_token=([^;]+)/)?.[1] ?? null;

	return { accessToken: data.accessToken as string, refreshToken: newRefreshToken };
}

async function refreshAccessToken(
	event: Parameters<Handle>[0]['event'],
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

	if (result.refreshToken) {
		event.cookies.set('refresh_token', result.refreshToken, {
			path: '/',
			httpOnly: true,
			secure: !dev,
			maxAge: 60 * 60 * 24 * 30,
			sameSite: 'lax'
		});
	}

	return result.accessToken;
}

async function resolveSession(
	event: Parameters<Handle>[0]['event'],
	accessToken: string
): Promise<boolean> {
	const res = await event.fetch(`${env.BASE_API}/auth/session`, {
		headers: { Authorization: `Bearer ${accessToken}` }
	});

	if (!res.ok) return false;

	const { data, success }: SessionResponse = await res.json();
	if (success) {
		event.locals.user = data.user;
		event.locals.session = data.session;
		return true;
	}
	return false;
}

export const handle: Handle = async ({ event, resolve }) => {
	event.locals.user = null;
	event.locals.session = null;

	if (needsSession(event.url.pathname)) {
		const accessToken = event.cookies.get('access_token_finexia');
		const refreshToken = event.cookies.get('refresh_token');

		if (accessToken) {
			const valid = await resolveSession(event, accessToken);

			if (!valid && refreshToken) {
				const newAccessToken = await refreshAccessToken(event, refreshToken);

				if (newAccessToken) {
					event.cookies.set('access_token_finexia', newAccessToken, {
						path: '/',
						httpOnly: true,
						secure: !dev,
						maxAge: 60 * 60 * 24 * 7,
						sameSite: 'lax'
					});
					await resolveSession(event, newAccessToken);
				} else {
					event.cookies.delete('access_token_finexia', { path: '/' });
					event.cookies.delete('refresh_token', { path: '/' });
				}
			} else if (!valid) {
				// Token inválido sin refresh token disponible — eliminar para evitar
				// que acciones de formulario envíen un token expirado al backend.
				event.cookies.delete('access_token_finexia', { path: '/' });
			}
		} else if (refreshToken) {
			const newAccessToken = await refreshAccessToken(event, refreshToken);

			if (newAccessToken) {
				event.cookies.set('access_token_finexia', newAccessToken, {
					path: '/',
					httpOnly: true,
					secure: !dev,
					maxAge: 60 * 60 * 24 * 7,
					sameSite: 'lax'
				});
				await resolveSession(event, newAccessToken);
			} else {
				event.cookies.delete('refresh_token', { path: '/' });
			}
		}
	}

	const response = await resolve(event);

	return withRobots(event, response);
};
