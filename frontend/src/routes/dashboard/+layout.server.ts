import { env } from '$env/dynamic/private';
import type { LayoutServerLoad } from './$types';
import { redirect } from '@sveltejs/kit';

type SessionResponse = {
	data: {
		user: {
			id: string;
			name: string;
			email: string;
			verified: boolean;
			image: string;
			preferredCurrency: string;
			createdAt: string;
			updatedAt: string;
		};
		session: {
			id: string;
			userId: string;
			expiresAt: string;
			ipAddress: string;
			userAgent: string;
			createdAt: string;
			updatedAt: string;
		};
	};
	success: boolean;
	message: string;
	details: string;
};

export const load: LayoutServerLoad = async ({ cookies, fetch }) => {
	const accessToken = cookies.get('access_token_finexia');

	if (!accessToken) {
		return redirect(303, '/auth');
	}

	const response = await fetch(`${env.BASE_API}/auth/session`, {
		headers: {
			Authorization: `Bearer ${accessToken}`
		}
	});

	if (!response.ok) {
		return redirect(303, '/auth');
	}

	const { data, success }: SessionResponse = await response.json();

	if (!success) {
		return redirect(303, '/auth');
	}

	return data;
};
