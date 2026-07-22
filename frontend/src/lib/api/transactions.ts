/**
 * Transacciones: listado del usuario, transacciones por posición, alta/edición,
 * wizard de import (preview/commit) y exports XLSX.
 */
import {
	apiRequest,
	apiRequestSafe,
	authedFetch,
	authedFetchSafe,
	type ApiEvent,
	type ApiResult
} from './client';
import type { PagedTransactions, UserTransaction } from './types';

/** `GET /portfolios/transactions` — transacciones recientes del usuario. */
export function getRecent(event: ApiEvent): Promise<ApiResult<UserTransaction[]>> {
	return apiRequestSafe<UserTransaction[]>(event, '/portfolios/transactions');
}

/** `GET /portfolios/:id/assets/:symbol/transactions` — transacciones paginadas de una posición. */
export function getAssetTransactions(
	event: ApiEvent,
	id: string,
	symbol: string,
	page: number,
	limit: number
): Promise<ApiResult<PagedTransactions>> {
	return apiRequestSafe<PagedTransactions>(
		event,
		`/portfolios/${id}/assets/${symbol}/transactions?page=${page}&limit=${limit}`
	);
}

/** `POST /portfolios/entries/:entryId/transactions` — crea una transacción. */
export function createTransaction(
	event: ApiEvent,
	entryId: string,
	body: Record<string, unknown>
): Promise<ApiResult<unknown>> {
	return apiRequest<unknown>(event, `/portfolios/entries/${entryId}/transactions`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
}

/** `PUT /portfolios/transactions/:txnId` — actualiza una transacción. */
export function updateTransaction(
	event: ApiEvent,
	txnId: string,
	body: Record<string, unknown>
): Promise<ApiResult<unknown>> {
	return apiRequest<unknown>(event, `/portfolios/transactions/${txnId}`, {
		method: 'PUT',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
}

/**
 * `POST /portfolios/transactions/import/preview` — preview del import.
 * Reenvía el multipart tal cual y devuelve la `Response` cruda (el endpoint la
 * proxya con su status hacia el navegador).
 */
export function importPreview(event: ApiEvent, form: FormData): Promise<Response> {
	return authedFetch(event, '/portfolios/transactions/import/preview', {
		method: 'POST',
		body: form
	});
}

/** `POST /portfolios/transactions/import` — commit del import masivo. */
export function importCommit(event: ApiEvent, form: FormData): Promise<Response> {
	return authedFetch(event, '/portfolios/transactions/import', {
		method: 'POST',
		body: form
	});
}

/**
 * `GET /portfolios/export/*` — descarga XLSX. Devuelve la `Response` cruda para
 * poder hacer streaming del cuerpo binario sin parsear.
 */
export function exportFile(event: ApiEvent, path: string): Promise<Response | null> {
	return authedFetchSafe(event, path);
}
