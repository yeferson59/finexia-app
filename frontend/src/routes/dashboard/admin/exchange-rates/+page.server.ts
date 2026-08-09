import type { Actions, PageServerLoad } from './$types';
import * as market from '$lib/api/market';
import { fail } from '@sveltejs/kit';
import type { ExchangeRate } from '$lib/api/types';
import { rateCreateSchema, rateUpdateSchema } from '$lib/features/admin';

export const load: PageServerLoad = async ({ cookies, fetch }) => {
	const res = await market.getExchangeRates({ cookies, fetch }, { page: 1, limit: 100 });

	return {
		rates: res.success && Array.isArray(res.data) ? (res.data as ExchangeRate[]) : []
	};
};

export const actions = {
	createRate: async ({ request, cookies, fetch }) => {
		const fd = await request.formData();

		const parsed = rateCreateSchema.safeParse({
			fromCurrency: fd.get('fromCurrency') ?? '',
			toCurrency: fd.get('toCurrency') ?? '',
			rate: fd.get('rate') ?? ''
		});

		if (!parsed.success) {
			return fail(400, { createError: parsed.error.issues[0].message });
		}

		const res = await market.createRate({ cookies, fetch }, parsed.data);

		if (!res.ok) {
			return fail(res.status, {
				createError: res.details ?? 'No se pudo crear la tasa de cambio'
			});
		}

		return { createSuccess: true };
	},

	refreshRates: async ({ cookies, fetch }) => {
		const res = await market.refreshExchangeRates({ cookies, fetch });

		if (!res.ok) {
			return fail(res.status, {
				refreshError: res.details ?? 'No se pudo actualizar desde el feed público'
			});
		}

		// El número de pares que volvieron es lo único que hace falta para el
		// aviso; la tabla ya se recarga sola con el `load` que dispara la acción.
		return { refreshSuccess: true, refreshedCount: Array.isArray(res.data) ? res.data.length : 0 };
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
		const id = (fd.get('id') as string | null) ?? '';
		// Igual que el precio de un activo: al backend viaja el texto original.
		const rateStr = (fd.get('rate') as string | null) ?? '';

		const parsed = rateUpdateSchema.safeParse({ id, rate: rateStr });

		if (!parsed.success) {
			return fail(400, { updateError: parsed.error.issues[0].message, errorId: id });
		}

		const res = await market.updateRate({ cookies, fetch }, parsed.data.id, { rate: rateStr });

		if (!res.ok) {
			return fail(res.status, {
				updateError: res.details ?? 'No se pudo actualizar la tasa',
				errorId: parsed.data.id
			});
		}

		return { updateSuccess: true, updatedId: parsed.data.id };
	}
} satisfies Actions;
