import type { Actions, PageServerLoad } from './$types';
import { authedFetch } from '$lib/server/api';
import { fail } from '@sveltejs/kit';

interface ExchangeRate {
	id: string;
	fromCurrency: string;
	toCurrency: string;
	rate: string;
	rateDate: string;
	createdAt: string;
}

export const load: PageServerLoad = async ({ cookies, fetch }) => {
	const event = { cookies, fetch };

	const res = await authedFetch(event, '/exchange-rates?page=1&limit=100');
	const { data, success } = await res.json();

	return {
		rates: success && Array.isArray(data) ? (data as ExchangeRate[]) : []
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

		const res = await authedFetch({ cookies, fetch }, '/exchange-rates', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ fromCurrency, toCurrency, rate })
		});

		if (!res.ok) {
			const body = await res.json().catch(() => ({}));
			return fail(res.status, {
				createError: body.details ?? 'No se pudo crear la tasa de cambio'
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

		const res = await authedFetch({ cookies, fetch }, '/exchange-rates/import', {
			method: 'POST',
			body: fd
		});

		const body = await res.json().catch(() => ({}));
		if (!res.ok) {
			return fail(res.status, { importError: body.details ?? 'No se pudo importar el archivo' });
		}

		return { importSuccess: true, importResult: body.data };
	},

	syncRate: async ({ request, cookies, fetch }) => {
		const fd = await request.formData();
		const id = fd.get('id') as string;
		if (!id) return fail(400, { syncRateError: 'ID requerido', syncRateId: '' });

		const res = await authedFetch({ cookies, fetch }, `/exchange-rates/${id}/sync`, {
			method: 'POST'
		});
		if (!res.ok) {
			const body = await res.json().catch(() => ({}));
			return fail(res.status, {
				syncRateError: body.details ?? body.message ?? 'Sincronización fallida',
				syncRateId: id
			});
		}
		return { syncRateSuccess: true, syncRateId: id };
	},

	syncRates: async ({ cookies, fetch }) => {
		const res = await authedFetch({ cookies, fetch }, '/exchange-rates/sync', { method: 'POST' });
		if (!res.ok) {
			const body = await res.json().catch(() => ({}));
			return fail(res.status, { syncError: body.details ?? 'La sincronización falló' });
		}
		const { data } = await res.json().catch(() => ({ data: null }));
		return { syncSuccess: true, synced: Array.isArray(data) ? data.length : 0 };
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

		const res = await authedFetch({ cookies, fetch }, `/exchange-rates/${id}`, {
			method: 'PATCH',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ rate })
		});

		if (!res.ok) {
			const body = await res.json().catch(() => ({}));
			return fail(res.status, {
				updateError: body.details ?? 'No se pudo actualizar la tasa',
				errorId: id
			});
		}

		return { updateSuccess: true, updatedId: id };
	}
} satisfies Actions;
