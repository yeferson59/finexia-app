import type { Handle } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';

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

export const handle: Handle = async ({ event, resolve }) => {
	event.locals.user = null;
	event.locals.session = null;

	const accessToken = event.cookies.get('access_token_finexia');

	if (accessToken && needsSession(event.url.pathname)) {
		const res = await event.fetch(`${env.BASE_API}/auth/session`, {
			headers: {
				Authorization: `Bearer ${accessToken}`
			}
		});

		if (res.ok) {
			const { data, success }: SessionResponse = await res.json();

			if (success) {
				event.locals.user = data.user;
				event.locals.session = data.session;
			}
		}
	}

	const response = await resolve(event);

	return withRobots(event, response);
};
