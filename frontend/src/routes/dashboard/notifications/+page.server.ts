import type { Actions, PageServerLoad } from './$types';
import { z } from 'zod';
import { env } from '$env/dynamic/private';
import { fail } from '@sveltejs/kit';
import { authedFetchSafe } from '$lib/server/api';

interface UserPreferences {
	userId: string;
	emailAlerts: boolean;
	weeklySummary: boolean;
}

export const load: PageServerLoad = async ({ cookies, fetch }) => {
	const prefsRes = await authedFetchSafe({ cookies, fetch }, '/users/me/preferences');

	let preferences: UserPreferences = { userId: '', emailAlerts: true, weeklySummary: true };
	if (prefsRes?.ok) {
		const { data, success } = await prefsRes.json();
		if (success && data) preferences = data;
	}

	return { preferences };
};

export const actions = {
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
	}
} satisfies Actions;
