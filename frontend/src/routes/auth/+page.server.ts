import z from 'zod';
import type { Actions } from './$types';
import { env } from '$env/dynamic/private';
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

		const response = await fetch(`${env.BASE_API}/auth/login`, {
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
	register: async ({ request }) => {
		const formData = await request.formData();

		const { success, data, error } = await z
			.object({
				name: z.string().min(2),
				email: z.email().min(2),
				password: z.string().min(8),
				confirmPassword: z.string().min(8).max(18),
				terms: z.coerce.boolean()
			})
			.safeParseAsync({
				name: formData.get('name'),
				email: formData.get('email'),
				password: formData.get('password'),
				confirmPassword: formData.get('confirmPassword'),
				terms: formData.get('terms')
			});

		if (!success) {
			return { success: false, errors: error.issues };
		}

		if (data.password !== data.confirmPassword) {
			return { success: false, errors: { confirmPassword: 'Passwords do not match' } };
		}

		if (!data.terms) {
			return { success: false, errors: { terms: 'You must accept the terms and conditions' } };
		}

		const response = await fetch(`${env.BASE_API}/auth/register`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(data)
		});

		if (!response.ok) {
			const data = await response.json();
			return { success: false, errors: { server: data.message || 'Registration failed' } };
		}

		const { success: registeredSuccess, message: registeredMessage } = await response.json();

		if (!registeredSuccess) {
			return { success: false, errors: { server: registeredMessage || 'Registration failed' } };
		}

		redirect(302, '/login');
	}
} satisfies Actions;
