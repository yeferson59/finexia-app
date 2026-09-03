import type { Actions } from './$types';
import { redirect } from '@sveltejs/kit';
import * as portfolio from '$lib/api/portfolio';
import { portfolioEntrySchema } from '$lib/features/portfolio';

export const actions = {
	default: async ({ request, fetch, cookies, params }) => {
		const formData = await request.formData();

		const { success, error, data } = await portfolioEntrySchema.safeParseAsync({
			portfolioId: params.id,
			assetId: formData.get('assetId'),
			sourceId: formData.get('platformId'),
			quantity: formData.get('quantity'),
			price: formData.get('purchasePrice'),
			// Tres campos que solo significan algo juntos: el precio está en
			// `currency`, la cuenta pagó en `costCurrency` y `fxRate` es lo que
			// costaba una unidad de la primera en la segunda ese día.
			costCurrency: formData.get('costCurrency'),
			currency: formData.get('currency'),
			fxRate: formData.get('fxRate'),
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
