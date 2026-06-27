import z from 'zod';
import type { Actions } from './$types';
import { env } from '$env/dynamic/private';
import { dev } from '$app/environment';
import { redirect, fail } from '@sveltejs/kit';

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
			return fail(400, { type: 'login' as const, errors: loginDto.error.issues });
		}

		const response = await fetch(`${env.BASE_API}/auth/login`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ email: loginDto.data.email, password: loginDto.data.password })
		});

		if (!response.ok) {
			const data = await response.json().catch(() => ({}));
			return fail(response.status, {
				type: 'login' as const,
				errors: { server: data.message || 'Credenciales incorrectas' }
			});
		}

		const { data, success, message } = await response.json();

		if (!success) {
			return fail(400, { type: 'login' as const, errors: { server: message || 'Error al iniciar sesión' } });
		}

		cookies.set('access_token_finexia', data.accessToken, {
			path: '/',
			httpOnly: true,
			secure: !dev,
			maxAge: 60 * 60 * 24 * 7,
			expires: new Date(Date.now() + 60 * 60 * 24 * 7 * 1000),
			sameSite: 'lax'
		});

		const rawRefreshToken = response.headers.get('set-cookie')?.match(/refresh_token=([^;]+)/)?.[1];
		if (rawRefreshToken) {
			cookies.set('refresh_token', rawRefreshToken, {
				path: '/',
				httpOnly: true,
				secure: !dev,
				maxAge: 60 * 60 * 24 * 30,
				sameSite: 'lax'
			});
		}

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
			return fail(400, { type: 'register' as const, errors: error.issues });
		}

		if (data.password !== data.confirmPassword) {
			return fail(400, {
				type: 'register' as const,
				errors: { confirmPassword: 'Las contraseñas no coinciden' }
			});
		}

		if (!data.terms) {
			return fail(400, {
				type: 'register' as const,
				errors: { terms: 'Debes aceptar los términos y condiciones' }
			});
		}

		const response = await fetch(`${env.BASE_API}/auth/register`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(data)
		});

		if (!response.ok) {
			const body = await response.json().catch(() => ({}));
			return fail(response.status, {
				type: 'register' as const,
				errors: { server: body.message || 'Error al registrarse' }
			});
		}

		const { success: registeredSuccess, message: registeredMessage } = await response.json();

		if (!registeredSuccess) {
			return fail(400, {
				type: 'register' as const,
				errors: { server: registeredMessage || 'Error al registrarse' }
			});
		}

		redirect(302, '/auth');
	}
} satisfies Actions;
