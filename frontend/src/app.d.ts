// See https://svelte.dev/docs/kit/types#app.d.ts
// for information about these interfaces
declare global {
	namespace App {
		// interface Error {}
		interface Locals {
			user: {
				name: string;
				email: string;
				emailVerified: boolean;
				image: string;
				role: string;
				preferredCurrency: string;
				createdAt: string;
				updatedAt: string;
			} | null;
			session: {
				id: string;
				userId: string;
				expiresAt: string;
				ipAddress: string | null;
				userAgent: string | null;
				createdAt: string;
			} | null;
		}
		// interface PageData {}
		// interface PageState {}
		// interface Platform {}
	}
}

export {};
