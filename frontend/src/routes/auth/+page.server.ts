import z from 'zod';
import type { Actions } from './$types';
import { env } from '$env/dynamic/private';
import { redirect, fail } from '@sveltejs/kit';
import { parseRefreshSetCookie, setAccessCookie, setRefreshCookie } from '$lib/server/session';

export const actions = {
	login: async ({ request, cookies, fetch }) => {
		const formData = await request.formData();
		const loginDto = await z
			.object({
				email: z.email().min(2),
				// El backend (LoginRequestDTO) valida min=8,max=20.
				password: z.string().min(8).max(20)
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
			if (data.action === 'auth:login:unverified') {
				return fail(response.status, {
					type: 'login' as const,
					errors: { server: 'Debes verificar tu correo antes de iniciar sesión.' },
					unverified: true as const
				});
			}
			return fail(response.status, {
				type: 'login' as const,
				errors: { server: data.message || 'Credenciales incorrectas' }
			});
		}

		const { data, success, message } = await response.json();

		if (!success) {
			return fail(400, {
				type: 'login' as const,
				errors: { server: message || 'Error al iniciar sesión' }
			});
		}

		setAccessCookie(cookies, data.accessToken);

		const rotatedRefresh = parseRefreshSetCookie(response);
		if (rotatedRefresh) {
			setRefreshCookie(cookies, rotatedRefresh.value, rotatedRefresh.maxAge);
		}

		return redirect(302, '/dashboard');
	},
	register: async ({ request }) => {
		const formData = await request.formData();

		const { success, data, error } = await z
			.object({
				name: z.string().min(2),
				email: z.email().min(2),
				// El backend (RegisterRequestDTO) valida min=8,max=20.
				password: z.string().min(8).max(20),
				confirmPassword: z.string().min(8).max(20),
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
			if (body.action === 'auth:register:duplicate') {
				return fail(response.status, {
					type: 'register' as const,
					errors: {
						server: 'Ya existe una cuenta con este correo. Inicia sesión o recupera tu contraseña.'
					},
					duplicateEmail: true as const
				});
			}
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

		redirect(302, '/auth?registered=1');
	}
} satisfies Actions;
