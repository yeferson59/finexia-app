import type { Actions, PageServerLoad } from './$types';
import * as market from '$lib/api/market';
import { fail } from '@sveltejs/kit';
import type { ExchangeRate } from '$lib/api/types';

export const load: PageServerLoad = async ({ cookies, fetch }) => {
	const res = await market.getExchangeRates({ cookies, fetch }, { page: 1, limit: 100 });

	return {
		rates: res.success && Array.isArray(res.data) ? (res.data as ExchangeRate[]) : []
	};
};

export const actions = {
	createRate: async ({ request, cookies, fetch }) => {
		const fd = await request.formData();
		const fromCurrency = (fd.get('fromCurrency') as string | null)?.trim().toUpperCase();
		const toCurrency = (fd.get('toCurrency') as string | null)?.trim().toUpperCase();
		const rate = fd.get('rate') as string | null;

		if (!fromCurrency || !toCurrency || !rate) {
			return fail(400, { createError: 'Moneda origen, destino y tasa son requeridos' });
		}

		const res = await market.createRate({ cookies, fetch }, { fromCurrency, toCurrency, rate });

		if (!res.ok) {
			return fail(res.status, {
				createError: res.details ?? 'No se pudo crear la tasa de cambio'
			});
		}

		return { createSuccess: true };
	},

	importRates: async ({ request, cookies, fetch }) => {
		const fd = await request.formData();
		const file = fd.get('file');
		if (!(file instanceof File) || file.size === 0) {
			return fail(400, { importError: 'Selecciona un archivo CSV o Excel' });
		}

		const res = await market.importRates({ cookies, fetch }, fd);

		if (!res.ok) {
			return fail(res.status, { importError: res.details ?? 'No se pudo importar el archivo' });
		}

		return { importSuccess: true, importResult: res.data };
	},

	updateRate: async ({ request, cookies, fetch }) => {
		const fd = await request.formData();
		const id = fd.get('id') as string;
		const rate = fd.get('rate') as string;

		if (!id) return fail(400, { updateError: 'ID de tasa requerido', errorId: id });

		const value = parseFloat(rate);
		if (isNaN(value) || value <= 0) {
			return fail(400, { updateError: 'Tasa inválida', errorId: id });
		}

		const res = await market.updateRate({ cookies, fetch }, id, { rate });

		if (!res.ok) {
			return fail(res.status, {
				updateError: res.details ?? 'No se pudo actualizar la tasa',
				errorId: id
			});
		}

		return { updateSuccess: true, updatedId: id };
	}
} satisfies Actions;
