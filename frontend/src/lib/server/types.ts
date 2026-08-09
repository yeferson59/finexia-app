import type { Cookies } from '@sveltejs/kit';

export type SessionEvent = {
	cookies: Cookies;
	fetch: typeof fetch;
};

export type RefreshResult = {
	accessToken: string;
	refreshToken: string | null;
	refreshMaxAge: number | null;
};

export type SessionResponse = {
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
