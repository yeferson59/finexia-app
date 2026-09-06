import type { Actions } from './$types';
import { fail, redirect } from '@sveltejs/kit';
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

		/*
		 * `fail`, no un objeto suelto: devolver `{ success: false }` responde 200,
		 * así que el navegador da el alta por buena. Y el mensaje es el del
		 * schema; `error.message` es el JSON de las incidencias de Zod en crudo, y
		 * el formulario lo pintaba tal cual detrás de «No se pudo registrar el
		 * activo:».
		 */
		if (!success) {
			return fail(400, { success: false, error: error.issues[0].message });
		}

		const response = await portfolio.createEntry({ cookies, fetch }, data);

		if (!response.ok || !response.success) {
			// El backend rechaza combinaciones de moneda y tasa que no pueden ser
			// ciertas, y el mensaje dice cuál —«USD no se convierte en sí misma a
			// 1.0638»—. Devolverlo es la diferencia entre corregir un campo y
			// reintentar a ciegas contra el mismo error.
			return fail(response.status >= 400 ? response.status : 400, {
				success: false,
				error: response.details || response.message || response.action
			});
		}

		redirect(303, `/dashboard/portfolios/${params.id}`);
	}
} satisfies Actions;
