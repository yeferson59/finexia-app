import z from 'zod';
import type { Actions } from './$types';
import { redirect } from '@sveltejs/kit';
import * as platforms from '$lib/api/platforms';

export const actions = {
	default: async ({ request, cookies, fetch }) => {
		const formData = await request.formData();

		const { success, error, data } = await z
			.object({
				name: z.coerce.string().min(2),
				description: z.coerce.string().optional(),
				type: z.coerce.string().min(2)
			})
			.safeParseAsync({
				name: formData.get('name'),
				description: formData.get('description'),
				type: formData.get('type')
			});

		if (!success) {
			return { error: error.message };
		}

		const res = await platforms.createSource({ cookies, fetch }, data);

		if (!res.ok || !res.success) {
			return { error: 'Failed to add platform' };
		}

		redirect(303, '/dashboard/platforms');
	}
} satisfies Actions;
