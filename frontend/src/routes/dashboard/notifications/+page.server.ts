import type { Actions, PageServerLoad } from './$types';
import { z } from 'zod';
import { fail } from '@sveltejs/kit';
import * as user from '$lib/api/user';
import type { UserPreferences } from '$lib/api/types';

export const load: PageServerLoad = async ({ cookies, fetch }) => {
	const prefsRes = await user.getPreferences({ cookies, fetch });

	let preferences: UserPreferences = { userId: '', emailAlerts: true, weeklySummary: true };
	if (prefsRes.ok && prefsRes.success && prefsRes.data) preferences = prefsRes.data;

	return { preferences };
};

export const actions = {
	updatePreferences: async ({ request, fetch, cookies }) => {
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

		const res = await user.updatePreferences({ cookies, fetch }, parsed.data);

		if (!res.ok) {
			return fail(res.status, {
				action: 'updatePreferences',
				error: 'Error al guardar las preferencias'
			});
		}

		return { action: 'updatePreferences', success: true };
	}
} satisfies Actions;
