import type { Actions, PageServerLoad } from './$types';
import * as market from '$lib/api/market';
import { fail } from '@sveltejs/kit';
import type { Asset } from '$lib/api/types';

export const load: PageServerLoad = async ({ cookies, fetch }) => {
	const res = await market.getAssets({ cookies, fetch }, { page: 1, limit: 100 });

	return {
		assets: res.success && Array.isArray(res.data) ? (res.data as Asset[]) : []
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

		const res = await market.createAsset(
			{ cookies, fetch },
			{ ticker, name, assetType, exchange, currency }
		);

		if (!res.ok) {
			return fail(res.status, { createError: res.details ?? 'No se pudo crear el activo' });
		}

		return { createSuccess: true };
	},

	syncAsset: async ({ request, cookies, fetch }) => {
		const fd = await request.formData();
		const id = fd.get('id') as string;
		if (!id) return fail(400, { syncAssetError: 'ID requerido', syncAssetId: '' });

		const res = await market.syncAsset({ cookies, fetch }, id);
		if (!res.ok) {
			return fail(res.status, {
				syncAssetError: res.details ?? res.message ?? 'Sincronización fallida',
				syncAssetId: id
			});
		}
		return { syncAssetSuccess: true, syncAssetId: id };
	},

	syncPrices: async ({ cookies, fetch }) => {
		const res = await market.syncAllAssets({ cookies, fetch });
		if (!res.ok) {
			return fail(res.status, { syncError: res.details ?? 'La sincronización falló' });
		}
		return { syncSuccess: true, synced: Array.isArray(res.data) ? res.data.length : 0 };
	},

	importAssets: async ({ request, cookies, fetch }) => {
		const fd = await request.formData();
		const file = fd.get('file');
		if (!(file instanceof File) || file.size === 0) {
			return fail(400, { importError: 'Selecciona un archivo CSV o Excel' });
		}

		const res = await market.importAssets({ cookies, fetch }, fd);

		if (!res.ok) {
			return fail(res.status, { importError: res.details ?? 'No se pudo importar el archivo' });
		}

		return { importSuccess: true, importResult: res.data };
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

		const res = await market.updateAssetPrice({ cookies, fetch }, id, {
			price: { value: priceStr, currency }
		});

		if (!res.ok) {
			return fail(res.status, {
				updateError: res.details ?? 'No se pudo actualizar el precio',
				errorId: id
			});
		}

		return { updateSuccess: true, updatedId: id };
	}
} satisfies Actions;
