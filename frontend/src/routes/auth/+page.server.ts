import z from 'zod';
import type { Actions } from './$types';
import { env } from '$/config/env';
import { NODE_ENV } from '$env/static/private';
import { redirect } from '@sveltejs/kit';

export const actions = {
	login: async ({ request, cookies, fetch }) => {
		const formData = await request.formData();
		const loginDto = await z
			.object({
				email: z.email().min(2),
				password: z.string().min(8)
			})
			.safeParseAsync({
				email: formData.get('email'),
				password: formData.get('password')
			});

		if (!loginDto.success) {
			return { success: false, errors: loginDto.error.issues };
		}

		const response = await fetch(`${env.baseApi}/auth/login`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ email: loginDto.data.email, password: loginDto.data.password })
		});

		if (!response.ok) {
			const data = await response.json();
			return { success: false, errors: { server: data.message || 'Login failed' } };
		}

		const { data, success, message } = await response.json();

		if (!success) {
			return { success: false, errors: { server: message || 'Login failed' } };
		}

		cookies.set('access_token_finexia', data.accessToken, {
			path: '/',
			httpOnly: true,
			secure: NODE_ENV === 'production',
			maxAge: 60 * 60 * 24 * 7,
			expires: new Date(Date.now() + 60 * 60 * 24 * 7 * 1000),
			sameSite: 'lax'
		});

		return redirect(302, '/dashboard');
	},
	register: async (event) => {
		// TODO register the user
	}
} satisfies Actions;
