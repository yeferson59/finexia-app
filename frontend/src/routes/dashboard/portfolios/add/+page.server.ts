import type { Actions } from './$types';
import { redirect } from '@sveltejs/kit';
import * as portfolio from '$lib/api/portfolio';
import { portfolioCreateSchema } from '$lib/features/portfolio';

export const actions = {
	default: async ({ request, fetch, cookies }) => {
		const dataRequest = await request.formData();

		const { success, data, error } = await portfolioCreateSchema.safeParseAsync({
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

		const response = await portfolio.createPortfolio({ cookies, fetch }, data);

		if (!response.ok || !response.success) {
			return { success: false };
		}

		redirect(302, '/dashboard/portfolios');
	}
} satisfies Actions;
