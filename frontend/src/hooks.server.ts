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

export const handle: Handle = async ({ event, resolve }) => {
	const accessToken = event.cookies.get('access_token_finexia');

	if (!accessToken) {
		event.locals.user = null;
		event.locals.session = null;

		const response = await resolve(event);

		return response;
	}

	const res = await event.fetch(`${env.BASE_API}/auth/session`, {
		headers: {
			Authorization: `Bearer ${accessToken}`
		}
	});

	if (!res.ok) {
		event.locals.user = null;
		event.locals.session = null;

		const response = await resolve(event);

		return response;
	}

	const { data, success }: SessionResponse = await res.json();

	if (!success) {
		event.locals.user = null;
		event.locals.session = null;

		const response = await resolve(event);

		return response;
	}

	event.locals.user = data.user;
	event.locals.session = data.session;

	const response = await resolve(event);

	return response;
};
