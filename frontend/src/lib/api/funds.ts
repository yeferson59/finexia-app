/**
 * Fondos de inversión (`/portfolios/funds`): posiciones de unidades cuyo valor
 * escribe el dueño desde el extracto, como marcas por fecha.
 */
import { apiRequest, apiRequestSafe, type ApiEvent, type ApiResult } from './client';
import type { Fund, FundMark, FundMovement, FundPerformance, PublicFund } from './types';
import {
	fundMarkSchema,
	fundMovementSchema,
	fundPerformanceSchema,
	fundSchema,
	publicFundSchema
} from './schemas';
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

/** `GET /portfolios/funds/:id/movements` — aportes y retiros, del más reciente. */
export function getMovements(event: ApiEvent, assetId: string): Promise<ApiResult<FundMovement[]>> {
	return apiRequestSafe(
		event,
		`/portfolios/funds/${assetId}/movements`,
		{},
		z.array(fundMovementSchema)
	);
}

/**
 * `POST /portfolios/funds/:id/contributions` — dinero que entra a un fondo que
 * se sigue por saldo. Las unidades las calcula el backend.
 */
export function contribute(
	event: ApiEvent,
	assetId: string,
	body: Record<string, unknown>
): Promise<ApiResult<FundMovement>> {
	return apiRequest<FundMovement>(event, `/portfolios/funds/${assetId}/contributions`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
}

/** `POST /portfolios/funds/:id/withdrawals` — dinero que sale de una posición del fondo. */
export function withdraw(
	event: ApiEvent,
	assetId: string,
	body: Record<string, unknown>
): Promise<ApiResult<FundMovement>> {
	return apiRequest<FundMovement>(event, `/portfolios/funds/${assetId}/withdrawals`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
}

/** `DELETE /portfolios/funds/movements/:txnId` — borra un aporte o un retiro. */
export function deleteMovement(event: ApiEvent, txnId: string): Promise<ApiResult<unknown>> {
	return apiRequest<unknown>(event, `/portfolios/funds/movements/${txnId}`, { method: 'DELETE' });
}

/** `GET /portfolios/funds/:id/performance` — rentabilidad por periodo, dinero y serie. */
export function getPerformance(
	event: ApiEvent,
	assetId: string
): Promise<ApiResult<FundPerformance>> {
	return apiRequestSafe(
		event,
		`/portfolios/funds/${assetId}/performance`,
		{},
		fundPerformanceSchema
	);
}

/**
 * `POST /portfolios/funds/:id/marks/bulk` — varias marcas de una vez: la tabla
 * que trae un extracto. Cada una reemplaza la de su día.
 */
export function saveMarks(
	event: ApiEvent,
	assetId: string,
	body: Record<string, unknown>
): Promise<ApiResult<{ saved: number }>> {
	return apiRequest<{ saved: number }>(event, `/portfolios/funds/${assetId}/marks/bulk`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
}

/**
 * `GET /portfolios/funds/catalog?q=` — los fondos cuyo valor de unidad publica
 * la Superintendencia Financiera, por nombre o código; los más grandes primero.
 */
export function searchPublicFunds(event: ApiEvent, q: string): Promise<ApiResult<PublicFund[]>> {
	return apiRequestSafe(
		event,
		`/portfolios/funds/catalog?q=${encodeURIComponent(q.trim())}`,
		{},
		z.array(publicFundSchema)
	);
}

/**
 * `PUT /portfolios/funds/:id/link` — enlaza un fondo por unidades a uno de la
 * Superfinanciera: sus valores publicados, desde la primera compra, pasan a ser
 * sus marcas. Las que escribió el dueño no se tocan.
 */
export function linkFund(
	event: ApiEvent,
	assetId: string,
	publicFundId: string
): Promise<ApiResult<Fund>> {
	return apiRequest<Fund>(event, `/portfolios/funds/${assetId}/link`, {
		method: 'PUT',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ publicFundId })
	});
}

/** `DELETE /portfolios/funds/:id/link` — deshace el enlace y borra los valores que trajo. */
export function unlinkFund(event: ApiEvent, assetId: string): Promise<ApiResult<Fund>> {
	return apiRequest<Fund>(event, `/portfolios/funds/${assetId}/link`, { method: 'DELETE' });
}
