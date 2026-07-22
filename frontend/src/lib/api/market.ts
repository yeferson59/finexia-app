/**
 * Mercado: catálogo de assets y tasas de cambio, más las operaciones de
 * administración (alta, import, sincronización y ajuste de precio/tasa).
 */
import {
	apiRequest,
	apiRequestSafe,
	authedFetchSafe,
	type ApiEvent,
	type ApiResult
} from './client';
import type { Asset, ExchangeRate } from './types';

// --- Assets ---------------------------------------------------------------

/** `GET /portfolios/assets` — catálogo de assets (paginado). */
export function getAssets(
	event: ApiEvent,
	opts: { page?: number; limit?: number } = {}
): Promise<ApiResult<Asset[]>> {
	const page = opts.page ?? 1;
	const limit = opts.limit ?? 100;
	return apiRequestSafe<Asset[]>(event, `/portfolios/assets?page=${page}&limit=${limit}`);
}

/**
 * `GET /portfolios/assets` con búsqueda — para el combobox de activos. Devuelve
 * la `Response` cruda porque el endpoint `/api/assets` la proxya tal cual.
 */
export function searchAssets(
	event: ApiEvent,
	opts: { search?: string; limit?: string } = {}
): Promise<Response | null> {
	const limit = opts.limit ?? '10';
	const search = opts.search?.trim();
	const path = search
		? `/portfolios/assets?search=${encodeURIComponent(search)}&page=1&limit=${limit}`
		: `/portfolios/assets?page=1&limit=${limit}`;
	return authedFetchSafe(event, path);
}

/** `POST /assets` — crea un asset. */
export function createAsset(
	event: ApiEvent,
	body: Record<string, unknown>
): Promise<ApiResult<unknown>> {
	return apiRequest<unknown>(event, '/assets', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
}

/** `POST /assets/import` — import masivo de assets (multipart). */
export function importAssets(event: ApiEvent, form: FormData): Promise<ApiResult<unknown>> {
	return apiRequest<unknown>(event, '/assets/import', { method: 'POST', body: form });
}

/** `POST /assets/:id/sync` — sincroniza el precio de un asset. */
export function syncAsset(event: ApiEvent, id: string): Promise<ApiResult<unknown>> {
	return apiRequest<unknown>(event, `/assets/${id}/sync`, { method: 'POST' });
}

/** `POST /assets/sync` — sincroniza los precios de todos los assets. */
export function syncAllAssets(event: ApiEvent): Promise<ApiResult<unknown[]>> {
	return apiRequest<unknown[]>(event, '/assets/sync', { method: 'POST' });
}

/** `PATCH /portfolios/assets/:id/price` — fija el precio manual de un asset. */
export function updateAssetPrice(
	event: ApiEvent,
	id: string,
	body: Record<string, unknown>
): Promise<ApiResult<unknown>> {
	return apiRequest<unknown>(event, `/portfolios/assets/${id}/price`, {
		method: 'PATCH',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
}

// --- Exchange rates -------------------------------------------------------

/** `GET /exchange-rates` — tasas de cambio (paginado). */
export function getExchangeRates(
	event: ApiEvent,
	opts: { page?: number; limit?: number } = {}
): Promise<ApiResult<ExchangeRate[]>> {
	const page = opts.page ?? 1;
	const limit = opts.limit ?? 100;
	return apiRequestSafe<ExchangeRate[]>(event, `/exchange-rates?page=${page}&limit=${limit}`);
}

/** `POST /exchange-rates` — crea una tasa. */
export function createRate(
	event: ApiEvent,
	body: Record<string, unknown>
): Promise<ApiResult<unknown>> {
	return apiRequest<unknown>(event, '/exchange-rates', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
}

/** `POST /exchange-rates/import` — import masivo de tasas (multipart). */
export function importRates(event: ApiEvent, form: FormData): Promise<ApiResult<unknown>> {
	return apiRequest<unknown>(event, '/exchange-rates/import', { method: 'POST', body: form });
}

/** `POST /exchange-rates/:id/sync` — sincroniza una tasa. */
export function syncRate(event: ApiEvent, id: string): Promise<ApiResult<unknown>> {
	return apiRequest<unknown>(event, `/exchange-rates/${id}/sync`, { method: 'POST' });
}

/** `POST /exchange-rates/sync` — sincroniza todas las tasas. */
export function syncAllRates(event: ApiEvent): Promise<ApiResult<unknown[]>> {
	return apiRequest<unknown[]>(event, '/exchange-rates/sync', { method: 'POST' });
}

/** `PATCH /exchange-rates/:id` — actualiza una tasa. */
export function updateRate(
	event: ApiEvent,
	id: string,
	body: Record<string, unknown>
): Promise<ApiResult<unknown>> {
	return apiRequest<unknown>(event, `/exchange-rates/${id}`, {
		method: 'PATCH',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
}
