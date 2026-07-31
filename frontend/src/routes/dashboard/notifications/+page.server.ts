import type { Actions, PageServerLoad } from './$types';
import { fail } from '@sveltejs/kit';
import * as user from '$lib/api/user';
import type { UserPreferences } from '$lib/api/types';
import { notificationPreferencesSchema } from '$lib/features/settings';

export const load: PageServerLoad = async ({ cookies, fetch }) => {
	const prefsRes = await user.getPreferences({ cookies, fetch });

	let preferences: UserPreferences = { userId: '', emailAlerts: true, weeklySummary: true };
	if (prefsRes.ok && prefsRes.success && prefsRes.data) preferences = prefsRes.data;

	return { preferences };
};

export const actions = {
	updatePreferences: async ({ request, fetch, cookies }) => {
		const formData = await request.formData();

		const parsed = notificationPreferencesSchema.safeParse({
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
