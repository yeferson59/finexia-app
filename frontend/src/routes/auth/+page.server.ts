import type { Actions, PageServerLoad } from './$types';
import * as auth from '$lib/api/auth';
import { redirect, fail } from '@sveltejs/kit';
import { parseRefreshSetCookie, setAccessCookie, setRefreshCookie } from '$lib/server/session';
import { features } from '$lib/shared/config/features';
import { loginSchema, twoFactorSchema, registerSchema } from '$lib/features/auth';

export const load: PageServerLoad = () => {
	return { selfRegistrationEnabled: features.selfRegistration };
};

export const actions = {
	login: async ({ request, cookies, fetch }) => {
		const formData = await request.formData();
		const loginDto = await loginSchema.safeParseAsync({
			email: formData.get('email'),
			password: formData.get('password')
		});

		if (!loginDto.success) {
			return fail(400, { type: 'login' as const, errors: loginDto.error.issues });
		}

		const response = await auth.login(fetch, {
			email: loginDto.data.email,
			password: loginDto.data.password
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

		const body = await response.json();
		const { data, success, message } = body;

		if (!success) {
			return fail(400, {
				type: 'login' as const,
				errors: { server: message || 'Error al iniciar sesión' }
			});
		}

		// Password accepted but the account has 2FA enabled: no session yet.
		// Hand the short-lived pending token to the client so it can submit
		// the authenticator code as a second step.
		if (body.action === 'auth:login:2fa') {
			return {
				type: 'login' as const,
				twoFactorRequired: true as const,
				twoFactorToken: (data?.twoFactorToken as string) ?? '',
				errors: {}
			};
		}

		setAccessCookie(cookies, data.accessToken);

		const rotatedRefresh = parseRefreshSetCookie(response);
		if (rotatedRefresh) {
			setRefreshCookie(cookies, rotatedRefresh.value, rotatedRefresh.maxAge);
		}

		return redirect(302, '/dashboard');
	},
	twoFactor: async ({ request, cookies, fetch }) => {
		const formData = await request.formData();
		const parsed = await twoFactorSchema.safeParseAsync({
			token: formData.get('token'),
			code: formData.get('code')
		});

		if (!parsed.success) {
			return fail(400, {
				type: 'login' as const,
				twoFactorRequired: true as const,
				twoFactorToken: String(formData.get('token') ?? ''),
				errors: { code: 'Ingresa el código de 6 dígitos o un código de recuperación.' }
			});
		}

		const response = await auth.twoFactorLogin(fetch, {
			token: parsed.data.token,
			code: parsed.data.code
		});

		if (!response.ok) {
			const body = await response.json().catch(() => ({}));
			// Pending token expired or burned its attempts: back to square one.
			if (body.action === 'auth:2fa:expired') {
				return fail(response.status, {
					type: 'login' as const,
					errors: {
						server: 'La verificación expiró. Vuelve a iniciar sesión con tu contraseña.'
					}
				});
			}
			return fail(response.status, {
				type: 'login' as const,
				twoFactorRequired: true as const,
				twoFactorToken: parsed.data.token,
				errors: { code: 'Código incorrecto o ya utilizado. Inténtalo de nuevo.' }
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

		const { success, data, error } = await registerSchema.safeParseAsync({
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

		const response = await auth.register(fetch, data);

		if (!response.ok) {
			const body = await response.json().catch(() => ({}));
			if (body.action === 'auth:register:disabled') {
				return fail(response.status, {
					type: 'register' as const,
					errors: {
						server:
							'El registro está cerrado durante la beta. Únete a la lista de espera y te invitaremos.'
					},
					disabled: true as const
				});
			}
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
