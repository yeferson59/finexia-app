import z from 'zod';
import type { Actions, PageServerLoad } from './$types';
import * as auth from '$lib/api/auth';
import { fail, redirect } from '@sveltejs/kit';

export const load: PageServerLoad = async ({ url, fetch }) => {
	const token = url.searchParams.get('token');
	if (!token) {
		return {
			valid: false as const,
			reason: 'Ingresa tu correo para recibir un enlace de verificación.'
		};
	}

	const res = await auth.validateEmailVerificationToken(fetch, token);
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
			.object({ token: z.string().min(1) })
			.safeParseAsync({ token: formData.get('token') });

		if (!parsed.success) {
			return fail(400, { errors: { server: 'Falta el token de verificación.' } });
		}

		const res = await auth.confirmEmailVerification(fetch, parsed.data.token);

		if (!res.ok) {
			const body = await res.json().catch(() => ({}));
			return fail(res.status, {
				errors: { server: body.details ?? body.message ?? 'No se pudo verificar el correo' }
			});
		}

		redirect(303, '/auth?verified=1');
	},
	resend: async ({ request, fetch }) => {
		const formData = await request.formData();

		const parsed = await z
			.object({ email: z.email().min(2) })
			.safeParseAsync({ email: formData.get('email') });

		if (!parsed.success) {
			return fail(400, { errors: { email: 'Ingresa un email válido' } });
		}

		// The backend response never reveals whether the email exists or is
		// already verified, so the UI reports the same generic success.
		await auth.requestEmailVerification(fetch, parsed.data.email);

		return { resent: true as const };
	}
} satisfies Actions;
