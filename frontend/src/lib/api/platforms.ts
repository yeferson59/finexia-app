/**
 * Plataformas / fuentes (`/portfolios/sources`): listado, alta, edición y borrado.
 */
import { apiRequest, apiRequestSafe, type ApiEvent, type ApiResult } from './client';
import type { Platform } from './types';

/** `GET /portfolios/sources` — plataformas del usuario. */
export function getSources(event: ApiEvent): Promise<ApiResult<Platform[]>> {
	return apiRequestSafe<Platform[]>(event, '/portfolios/sources');
}

/** `POST /portfolios/sources` — crea una plataforma. */
export function createSource(
	event: ApiEvent,
	body: Record<string, unknown>
): Promise<ApiResult<unknown>> {
	return apiRequest<unknown>(event, '/portfolios/sources', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
}

/** `PATCH /portfolios/sources/:id` — actualiza una plataforma. */
export function updateSource(
	event: ApiEvent,
	id: string,
	body: Record<string, unknown>
): Promise<ApiResult<unknown>> {
	return apiRequest<unknown>(event, `/portfolios/sources/${id}`, {
		method: 'PATCH',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
}

/** `DELETE /portfolios/sources/:id` — elimina una plataforma. */
export function deleteSource(event: ApiEvent, id: string): Promise<ApiResult<unknown>> {
	return apiRequest<unknown>(event, `/portfolios/sources/${id}`, { method: 'DELETE' });
}
