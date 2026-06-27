import type { Actions, PageServerLoad } from './$types';
import { authedFetch } from '$lib/server/api';
import { fail } from '@sveltejs/kit';

interface Asset {
	id: string;
	ticker: string;
	name: string;
	assetType: string;
	currency: string;
	currentPrice: { value: string; currency: string } | null;
	priceUpdatedAt: string | null;
}

export const load: PageServerLoad = async ({ cookies, fetch }) => {
	const event = { cookies, fetch };

	const res = await authedFetch(event, '/portfolios/assets?page=1&limit=100');
	const { data, success } = await res.json();

	return {
		assets: success && Array.isArray(data) ? (data as Asset[]) : []
	};
};

export const actions = {
	syncPrices: async ({ cookies, fetch }) => {
		const res = await authedFetch({ cookies, fetch }, '/assets/sync', { method: 'POST' });
		if (!res.ok) {
			const body = await res.json().catch(() => ({}));
			return fail(res.status, { syncError: body.details ?? 'La sincronización falló' });
		}
		const { data } = await res.json().catch(() => ({ data: null }));
		return { syncSuccess: true, synced: Array.isArray(data) ? data.length : 0 };
	},

	updatePrice: async ({ request, cookies, fetch }) => {
		const fd = await request.formData();
		const id = fd.get('id') as string;
		const priceStr = fd.get('price') as string;
		const currency = (fd.get('currency') as string) || 'USD';

		if (!id) return fail(400, { updateError: 'ID de activo requerido', errorId: id });

		const price = parseFloat(priceStr);
		if (isNaN(price) || price <= 0) {
			return fail(400, { updateError: 'Precio inválido', errorId: id });
		}

		const res = await authedFetch(
			{ cookies, fetch },
			`/portfolios/assets/${id}/price`,
			{
				method: 'PATCH',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ price: { value: priceStr, currency } })
			}
		);

		if (!res.ok) {
			const body = await res.json().catch(() => ({}));
			return fail(res.status, { updateError: body.details ?? 'No se pudo actualizar el precio', errorId: id });
		}

		return { updateSuccess: true, updatedId: id };
	}
} satisfies Actions;
