import type { Actions } from './$types';
import { z } from 'zod';
import { redirect } from '@sveltejs/kit';
import { authedFetch } from '$lib/server/api';

export const actions = {
	default: async ({ request, fetch, cookies }) => {
		const dataRequest = await request.formData();

		const { success, data, error } = await z
			.object({
				name: z.string().min(1),
				description: z.string().nullable(),
				type: z.string().min(1),
				riskId: z.uuid(),
				currency: z.string().min(1),
				priceValue: z.coerce.number().nonnegative().default(0),
				isDefault: z.coerce.boolean()
			})
			.safeParseAsync({
				name: dataRequest.get('name'),
				description: dataRequest.get('description'),
				type: dataRequest.get('type'),
				riskId: dataRequest.get('riskId'),
				currency: dataRequest.get('currency'),
				priceValue: dataRequest.get('priceValue'),
				isDefault: dataRequest.get('isDefault')
			});

		if (!success) {
			return { success, error: error.message };
		}

		const response = await authedFetch({ cookies, fetch }, '/portfolios', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(data)
		});

		if (!response.ok) {
			return { success: false };
		}

		const { success: responseSuccess } = await response.json();

		if (!responseSuccess) {
			return { success: false };
		}

		redirect(302, '/dashboard/portfolios');
	}
} satisfies Actions;
