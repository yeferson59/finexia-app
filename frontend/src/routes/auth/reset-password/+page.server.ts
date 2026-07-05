import z from 'zod';
import type { Actions, PageServerLoad } from './$types';
import { env } from '$env/dynamic/private';
import { fail, redirect } from '@sveltejs/kit';

export const load: PageServerLoad = async ({ url, fetch }) => {
	const token = url.searchParams.get('token');
	if (!token) {
		return { valid: false as const, reason: 'Falta el token de recuperación.' };
	}

	const res = await fetch(`${env.BASE_API}/auth/password-reset?token=${encodeURIComponent(token)}`);
	const body = await res.json().catch(() => ({}));

	if (!res.ok || !body.success) {
		const reason =
			res.status === 410
				? 'Este enlace ha expirado. Solicita uno nuevo.'
				: 'Este enlace no es válido o ya fue utilizado.';
		return { valid: false as const, reason };
	}

	return { valid: true as const, token };
};

export const actions = {
	confirm: async ({ request, fetch }) => {
		const formData = await request.formData();

		const parsed = await z
			.object({
				token: z.string().min(1),
				// Mirror the backend bounds (min=8,max=20) so login never rejects it.
				password: z.string().min(8).max(20),
				confirmPassword: z.string().min(8).max(20)
			})
			.safeParseAsync({
				token: formData.get('token'),
				password: formData.get('password'),
				confirmPassword: formData.get('confirmPassword')
			});

		if (!parsed.success) {
			return fail(400, { errors: parsed.error.issues });
		}

		if (parsed.data.password !== parsed.data.confirmPassword) {
			return fail(400, { errors: { confirmPassword: 'Las contraseñas no coinciden' } });
		}

		const res = await fetch(`${env.BASE_API}/auth/password-reset/confirm`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ token: parsed.data.token, password: parsed.data.password })
		});

		if (!res.ok) {
			const body = await res.json().catch(() => ({}));
			return fail(res.status, {
				errors: { server: body.details ?? body.message ?? 'No se pudo restablecer la contraseña' }
			});
		}

		// Password rotated and every session revoked — send them to the login
		// page to sign in with the new password.
		redirect(303, '/auth?reset=1');
	}
} satisfies Actions;
