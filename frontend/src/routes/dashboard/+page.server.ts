import z from 'zod';
import type { Actions } from './$types';
import { env } from '$env/dynamic/private';
import { redirect } from '@sveltejs/kit';

export const actions = {
	logout: async ({ cookies, fetch }) => {
		const token = cookies.get('access_token_finexia');

		if (!token) return { success: false };

		const response = await fetch(`${env.API_URL}/auth/logout`, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json',
				Authorization: `Bearer ${token}`
			}
		});

		if (!response.ok) {
			return { success: false };
		}

		const { success } = await response.json();

		if (!success) {
			return { success: false };
		}

		cookies.delete('access_token_finexia', { path: '/' });

		return redirect(302, '/auth');
	}
} satisfies Actions;
