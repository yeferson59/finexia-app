import type { Actions } from './$types';
import { fail, redirect } from '@sveltejs/kit';
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

		/*
		 * `fail`, no un objeto suelto: devolver `{ success: false }` responde 200,
		 * así que el navegador da el alta por buena. Y el mensaje sale del schema,
		 * no del `error.message` de Zod, que es el JSON de las incidencias crudo.
		 */
		if (!success) {
			return fail(400, { error: error.issues[0].message });
		}

		const response = await portfolio.createPortfolio({ cookies, fetch }, data);

		if (!response.ok || !response.success) {
			return fail(response.status >= 400 ? response.status : 400, {
				error: 'No pudimos crear el portafolio. Vuelve a intentarlo en un momento.'
			});
		}

		redirect(303, '/dashboard/portfolios');
	}
} satisfies Actions;
