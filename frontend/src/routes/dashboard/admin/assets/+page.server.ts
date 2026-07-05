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
	createAsset: async ({ request, cookies, fetch }) => {
		const fd = await request.formData();
		const ticker = (fd.get('ticker') as string | null)?.trim().toUpperCase();
		const name = (fd.get('name') as string | null)?.trim();
		const assetType = fd.get('assetType') as string | null;
		const exchange = (fd.get('exchange') as string | null)?.trim() ?? '';
		const currency = (fd.get('currency') as string | null)?.trim().toUpperCase();

		if (!ticker || !name || !assetType || !currency) {
			return fail(400, { createError: 'Ticker, nombre, tipo y moneda son requeridos' });
		}

		const res = await authedFetch({ cookies, fetch }, '/assets', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ ticker, name, assetType, exchange, currency })
		});

		if (!res.ok) {
			const body = await res.json().catch(() => ({}));
			return fail(res.status, { createError: body.details ?? 'No se pudo crear el activo' });
		}

		return { createSuccess: true };
	},

	syncAsset: async ({ request, cookies, fetch }) => {
		const fd = await request.formData();
		const id = fd.get('id') as string;
		if (!id) return fail(400, { syncAssetError: 'ID requerido', syncAssetId: '' });

		const res = await authedFetch({ cookies, fetch }, `/assets/${id}/sync`, { method: 'POST' });
		if (!res.ok) {
			const body = await res.json().catch(() => ({}));
			return fail(res.status, {
				syncAssetError: body.details ?? body.message ?? 'Sincronización fallida',
				syncAssetId: id
			});
		}
		return { syncAssetSuccess: true, syncAssetId: id };
	},

	syncPrices: async ({ cookies, fetch }) => {
		const res = await authedFetch({ cookies, fetch }, '/assets/sync', { method: 'POST' });
		if (!res.ok) {
			const body = await res.json().catch(() => ({}));
			return fail(res.status, { syncError: body.details ?? 'La sincronización falló' });
		}
		const { data } = await res.json().catch(() => ({ data: null }));
		return { syncSuccess: true, synced: Array.isArray(data) ? data.length : 0 };
	},

	importAssets: async ({ request, cookies, fetch }) => {
		const fd = await request.formData();
		const file = fd.get('file');
		if (!(file instanceof File) || file.size === 0) {
			return fail(400, { importError: 'Selecciona un archivo CSV o Excel' });
		}

		const res = await authedFetch({ cookies, fetch }, '/assets/import', {
			method: 'POST',
			body: fd
		});

		const body = await res.json().catch(() => ({}));
		if (!res.ok) {
			return fail(res.status, { importError: body.details ?? 'No se pudo importar el archivo' });
		}

		return { importSuccess: true, importResult: body.data };
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

		const res = await authedFetch({ cookies, fetch }, `/portfolios/assets/${id}/price`, {
			method: 'PATCH',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ price: { value: priceStr, currency } })
		});

		if (!res.ok) {
			const body = await res.json().catch(() => ({}));
			return fail(res.status, {
				updateError: body.details ?? 'No se pudo actualizar el precio',
				errorId: id
			});
		}

		return { updateSuccess: true, updatedId: id };
	}
} satisfies Actions;
