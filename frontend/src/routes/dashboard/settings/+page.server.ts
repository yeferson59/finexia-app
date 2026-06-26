import type { Actions, PageServerLoad } from './$types';
import { z } from 'zod';
import { env } from '$env/dynamic/private';
import { fail } from '@sveltejs/kit';

interface UserPreferences {
	userId: string;
	emailAlerts: boolean;
	weeklySummary: boolean;
}

export const load: PageServerLoad = async ({ locals, cookies, fetch }) => {
	const accessToken = cookies.get('access_token_finexia');
	const headers = { Authorization: `Bearer ${accessToken}` };

	const prefsRes = await fetch(`${env.BASE_API}/users/me/preferences`, { headers }).catch(
		() => null
	);

	let preferences: UserPreferences = { userId: '', emailAlerts: true, weeklySummary: true };
	if (prefsRes?.ok) {
		const { data, success } = await prefsRes.json();
		if (success && data) preferences = data;
	}

	return { user: locals.user, preferences };
};

export const actions = {
	updateProfile: async ({ request, fetch, cookies }) => {
		const accessToken = cookies.get('access_token_finexia');
		const formData = await request.formData();

		const schema = z.object({
			name: z.string().min(2, 'El nombre debe tener al menos 2 caracteres').max(254),
			preferredCurrency: z
				.string()
				.length(3, 'La moneda debe ser un código de 3 caracteres')
				.toUpperCase(),
			image: z.string().optional()
		});

		const parsed = schema.safeParse({
			name: formData.get('name'),
			preferredCurrency: formData.get('preferredCurrency'),
			image: formData.get('image') || undefined
		});

		if (!parsed.success) {
			return fail(400, { action: 'updateProfile', error: parsed.error.issues[0].message });
		}

		const res = await fetch(`${env.BASE_API}/users/me`, {
			method: 'PATCH',
			headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${accessToken}` },
			body: JSON.stringify(parsed.data)
		});

		if (!res.ok) {
			const body = await res.json().catch(() => ({}));
			return fail(res.status, {
				action: 'updateProfile',
				error: body.details ?? 'Error al actualizar el perfil'
			});
		}

		return { action: 'updateProfile', success: true };
	},

	updatePreferences: async ({ request, fetch, cookies }) => {
		const accessToken = cookies.get('access_token_finexia');
		const formData = await request.formData();

		const schema = z.object({
			emailAlerts: z.coerce.boolean(),
			weeklySummary: z.coerce.boolean()
		});

		const parsed = schema.safeParse({
			emailAlerts: formData.get('emailAlerts'),
			weeklySummary: formData.get('weeklySummary')
		});

		if (!parsed.success) {
			return fail(400, {
				action: 'updatePreferences',
				error: parsed.error.issues[0].message
			});
		}

		const res = await fetch(`${env.BASE_API}/users/me/preferences`, {
			method: 'PATCH',
			headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${accessToken}` },
			body: JSON.stringify(parsed.data)
		});

		if (!res.ok) {
			return fail(res.status, {
				action: 'updatePreferences',
				error: 'Error al guardar las preferencias'
			});
		}

		return { action: 'updatePreferences', success: true };
	},

	changePassword: async ({ request, fetch, cookies }) => {
		const accessToken = cookies.get('access_token_finexia');
		const formData = await request.formData();

		const schema = z
			.object({
				currentPassword: z.string().min(8, 'La contraseña debe tener al menos 8 caracteres'),
				newPassword: z.string().min(8, 'La nueva contraseña debe tener al menos 8 caracteres'),
				confirmPassword: z.string()
			})
			.refine((d) => d.newPassword === d.confirmPassword, {
				message: 'Las contraseñas no coinciden',
				path: ['confirmPassword']
			});

		const parsed = schema.safeParse({
			currentPassword: formData.get('currentPassword'),
			newPassword: formData.get('newPassword'),
			confirmPassword: formData.get('confirmPassword')
		});

		if (!parsed.success) {
			return fail(400, { action: 'changePassword', error: parsed.error.issues[0].message });
		}

		const res = await fetch(`${env.BASE_API}/users/me/password`, {
			method: 'PATCH',
			headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${accessToken}` },
			body: JSON.stringify({
				currentPassword: parsed.data.currentPassword,
				newPassword: parsed.data.newPassword
			})
		});

		if (!res.ok) {
			const errorMsg =
				res.status === 400 ? 'Contraseña actual incorrecta' : 'Error al cambiar la contraseña';
			return fail(res.status, { action: 'changePassword', error: errorMsg });
		}

		return { action: 'changePassword', success: true };
	}
} satisfies Actions;
