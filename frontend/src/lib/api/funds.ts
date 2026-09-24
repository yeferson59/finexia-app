/**
 * Fondos de inversión (`/portfolios/funds`): posiciones de unidades cuyo valor
 * escribe el dueño desde el extracto, como marcas por fecha.
 */
import { apiRequest, apiRequestSafe, type ApiEvent, type ApiResult } from './client';
import type { Fund, FundMark } from './types';
import { fundMarkSchema, fundSchema } from './schemas';
import { z } from 'zod';

/** `GET /portfolios/funds` — los fondos del usuario, por nombre. */
export function getFunds(event: ApiEvent): Promise<ApiResult<Fund[]>> {
	return apiRequestSafe(event, '/portfolios/funds', {}, z.array(fundSchema));
}

/**
 * `POST /portfolios/funds` — crea el fondo con su primera compra y, si se
 * conoce, lo que vale hoy una unidad.
 */
export function createFund(
	event: ApiEvent,
	body: Record<string, unknown>
): Promise<ApiResult<Fund>> {
	return apiRequest<Fund>(event, '/portfolios/funds', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
}

/** `DELETE /portfolios/funds/:id` — deja de seguir un fondo que ya nadie guarda. */
export function deleteFund(event: ApiEvent, assetId: string): Promise<ApiResult<unknown>> {
	return apiRequest<unknown>(event, `/portfolios/funds/${assetId}`, { method: 'DELETE' });
}

/** `GET /portfolios/funds/:id/marks` — las marcas de un fondo, de la más reciente. */
export function getMarks(event: ApiEvent, assetId: string): Promise<ApiResult<FundMark[]>> {
	return apiRequestSafe(event, `/portfolios/funds/${assetId}/marks`, {}, z.array(fundMarkSchema));
}

/**
 * `POST /portfolios/funds/:id/marks` — lo que valía una unidad un día. Si ya
 * había una marca ese día, la reemplaza; con fecha pasada, corrige la gráfica
 * desde ese día.
 */
export function saveMark(
	event: ApiEvent,
	assetId: string,
	body: Record<string, unknown>
): Promise<ApiResult<FundMark>> {
	return apiRequest<FundMark>(event, `/portfolios/funds/${assetId}/marks`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
}

/** `DELETE /portfolios/funds/:id/marks/:date` — borra la marca de un día (`YYYY-MM-DD`). */
export function deleteMark(
	event: ApiEvent,
	assetId: string,
	date: string
): Promise<ApiResult<unknown>> {
	return apiRequest<unknown>(event, `/portfolios/funds/${assetId}/marks/${date}`, {
		method: 'DELETE'
	});
}
