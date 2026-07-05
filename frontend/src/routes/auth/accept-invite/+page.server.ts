import z from 'zod';
import type { Actions, PageServerLoad } from './$types';
import { env } from '$env/dynamic/private';
import { fail, redirect } from '@sveltejs/kit';

export const load: PageServerLoad = async ({ url, fetch }) => {
	const token = url.searchParams.get('token');
	if (!token) {
		return { valid: false as const, reason: 'Falta el token de invitación.' };
	}

	const res = await fetch(`${env.BASE_API}/auth/invitations?token=${encodeURIComponent(token)}`);
	const body = await res.json().catch(() => ({}));

	if (!res.ok || !body.success) {
		const reason =
			res.status === 410
				? 'Esta invitación ha expirado. Pídele a un administrador que te la reenvíe.'
				: 'Esta invitación no es válida o ya fue utilizada.';
		return { valid: false as const, reason };
	}

	return {
		valid: true as const,
		token,
		email: (body.data?.email ?? '') as string,
		name: (body.data?.name ?? '') as string
	};
};

export const actions = {
	accept: async ({ request, fetch }) => {
		const formData = await request.formData();

		const parsed = await z
			.object({
				token: z.string().min(1),
				name: z.string().min(2).max(254),
				// Mirror the backend bounds (min=8,max=20) so login never rejects it.
				password: z.string().min(8).max(20),
				confirmPassword: z.string().min(8).max(20)
			})
			.safeParseAsync({
				token: formData.get('token'),
				name: formData.get('name'),
				password: formData.get('password'),
				confirmPassword: formData.get('confirmPassword')
			});

		if (!parsed.success) {
			return fail(400, { errors: parsed.error.issues });
		}

		if (parsed.data.password !== parsed.data.confirmPassword) {
			return fail(400, { errors: { confirmPassword: 'Las contraseñas no coinciden' } });
		}

		const res = await fetch(`${env.BASE_API}/auth/invitations/accept`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({
				token: parsed.data.token,
				name: parsed.data.name,
				password: parsed.data.password
			})
		});

		if (!res.ok) {
			const body = await res.json().catch(() => ({}));
			return fail(res.status, {
				errors: { server: body.details ?? body.message ?? 'No se pudo activar la cuenta' }
			});
		}

		// Account provisioned — send them to the login page to sign in.
		redirect(303, '/auth?invited=1');
	}
} satisfies Actions;
