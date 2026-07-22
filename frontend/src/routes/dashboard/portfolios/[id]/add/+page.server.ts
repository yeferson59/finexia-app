import type { Actions } from './$types';
import { z } from 'zod';
import { redirect } from '@sveltejs/kit';
import * as portfolio from '$lib/api/portfolio';

export const actions = {
	default: async ({ request, fetch, cookies, params }) => {
		const formData = await request.formData();

		const { success, error, data } = await z
			.object({
				portfolioId: z.uuid(),
				assetId: z.uuid(),
				sourceId: z.uuid(),
				quantity: z.coerce.number().positive(),
				price: z.coerce.number().positive(),
				costCurrency: z.coerce.string(),
				category: z.coerce.string().min(1),
				entryDate: z.coerce.date(),
				notes: z.coerce.string().optional()
			})
			.safeParseAsync({
				portfolioId: params.id,
				assetId: formData.get('assetId'),
				sourceId: formData.get('platformId'),
				quantity: formData.get('quantity'),
				price: formData.get('purchasePrice'),
				costCurrency: 'USD',
				category: formData.get('category'),
				entryDate: formData.get('purchaseDate'),
				notes: formData.get('notes')
			});

		if (!success) {
			return { success, error: error.message };
		}

		const response = await portfolio.createEntry({ cookies, fetch }, data);

		if (!response.ok || !response.success) {
			return { success: false };
		}

		redirect(303, `/dashboard/portfolios/${params.id}`);
	}
} satisfies Actions;
