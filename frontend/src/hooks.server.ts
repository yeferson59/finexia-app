import type { Handle, HandleFetch } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';
import {
	ACCESS_COOKIE,
	REFRESH_COOKIE,
	clearSessionCookies,
	refreshAccessToken
} from '$lib/server/session';

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
const PRIVATE_PREFIXES = ['/dashboard', '/auth'];

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
		const accessToken = event.cookies.get(ACCESS_COOKIE);
		const refreshToken = event.cookies.get(REFRESH_COOKIE);

		if (accessToken) {
			const valid = await resolveSession(event, accessToken);

			if (!valid && refreshToken) {
				const newAccessToken = await refreshAccessToken(event, refreshToken);

				if (newAccessToken) {
					await resolveSession(event, newAccessToken);
				} else {
					clearSessionCookies(event.cookies);
				}
			} else if (!valid) {
				// Token inválido sin refresh token disponible — eliminar para evitar
				// que acciones de formulario envíen un token expirado al backend.
				event.cookies.delete(ACCESS_COOKIE, { path: '/' });
			}
		} else if (refreshToken) {
			const newAccessToken = await refreshAccessToken(event, refreshToken);

			if (newAccessToken) {
				await resolveSession(event, newAccessToken);
			} else {
				event.cookies.delete(REFRESH_COOKIE, { path: '/' });
			}
		}
	}

	const response = await resolve(event);

	return withRobots(event, response);
};

// Forward the real client IP and User-Agent on every server-side call to the
// backend. The backend keys its rate limiters by IP (without this header
// every user would share the SSR server's single IP and collide on the same
// limit buckets) and stamps sessions with both fields so users can recognize
// their devices under /dashboard/settings; without forwarding, every session
// would record the SSR server's own request instead of the browser's.
export const handleFetch: HandleFetch = async ({ event, request, fetch }) => {
	if (env.BASE_API && request.url.startsWith(env.BASE_API)) {
		try {
			request.headers.set('X-Forwarded-For', event.getClientAddress());
		} catch {
			// getClientAddress() throws when the address is unavailable (e.g.
			// prerendering); the backend then falls back to the direct peer IP.
		}

		const userAgent = event.request.headers.get('user-agent');
		if (userAgent) {
			request.headers.set('User-Agent', userAgent);
		}
	}

	return fetch(request);
};
