/**
 * Portfolios: resumen, detalle, catálogo de riesgos, asignación, crecimiento y
 * creación/edición de portfolios y posiciones. Encapsula path + método + parseo
 * sobre `client.ts`; los tipos viven en `types.ts`.
 */
import { apiRequest, apiRequestSafe, type ApiEvent, type ApiResult } from './client';
import type {
	AllocationItem,
	PortfolioDetail,
	PortfolioGrowth,
	PortfolioSummary,
	Risk,
	TopTransaction
} from './types';
import {
	allocationItemSchema,
	portfolioDetailSchema,
	portfolioGrowthSchema,
	portfolioSummarySchema,
	riskSchema,
	topTransactionSchema
} from './schemas';
import { z } from 'zod';

/** `GET /portfolios/summary` — resumen de los portfolios (opcionalmente en `currency`). */
export function getSummaries(
	event: ApiEvent,
	currency?: string
): Promise<ApiResult<PortfolioSummary[]>> {
	const query = currency ? `?currency=${encodeURIComponent(currency)}` : '';
	return apiRequestSafe(event, `/portfolios/summary${query}`, {}, z.array(portfolioSummarySchema));
}

/** `GET /portfolios/:id` — detalle de un portfolio con sus holdings. */
export function getPortfolio(event: ApiEvent, id: string): Promise<ApiResult<PortfolioDetail>> {
	return apiRequestSafe(event, `/portfolios/${id}`, {}, portfolioDetailSchema);
}

/** `GET /portfolios/risks` — catálogo de niveles de riesgo. */
export function getRisks(event: ApiEvent): Promise<ApiResult<Risk[]>> {
	return apiRequestSafe(event, '/portfolios/risks', {}, z.array(riskSchema));
}

/** `GET /portfolios/allocation` — asignación por categoría de activo. */
export function getAllocation(event: ApiEvent): Promise<ApiResult<AllocationItem[]>> {
	return apiRequestSafe(event, '/portfolios/allocation', {}, z.array(allocationItemSchema));
}

/** `GET /portfolios/growth` — crecimiento agregado (soporta `since`/`period`). */
export function getAggregateGrowth(
	event: ApiEvent,
	opts: { since?: string; period?: string } = {}
): Promise<ApiResult<PortfolioGrowth>> {
	const params = new URLSearchParams();
	if (opts.since) params.set('since', opts.since);
	if (opts.period) params.set('period', opts.period);
	const query = params.toString() ? `?${params}` : '';
	return apiRequestSafe(event, `/portfolios/growth${query}`, {}, portfolioGrowthSchema);
}

/** `GET /portfolios/:id/growth` — crecimiento de un portfolio. */
export function getPortfolioGrowth(
	event: ApiEvent,
	id: string
): Promise<ApiResult<PortfolioGrowth>> {
	return apiRequestSafe(event, `/portfolios/${id}/growth`, {}, portfolioGrowthSchema);
}

/** `GET /portfolios/:id/top-transaction` — mayor transacción del portfolio. */
export function getTopTransaction(event: ApiEvent, id: string): Promise<ApiResult<TopTransaction>> {
	return apiRequestSafe(event, `/portfolios/${id}/top-transaction`, {}, topTransactionSchema);
}

/** `POST /portfolios` — crea un portfolio. */
export function createPortfolio(
	event: ApiEvent,
	body: Record<string, unknown>
): Promise<ApiResult<unknown>> {
	return apiRequest<unknown>(event, '/portfolios', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
}

/** `PATCH /portfolios/:id` — actualiza un portfolio. */
export function updatePortfolio(
	event: ApiEvent,
	id: string,
	body: Record<string, unknown>
): Promise<ApiResult<unknown>> {
	return apiRequest<unknown>(event, `/portfolios/${id}`, {
		method: 'PATCH',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
}

/** `POST /portfolios/entries` — crea una posición (entry) en un portfolio. */
export function createEntry(
	event: ApiEvent,
	body: Record<string, unknown>
): Promise<ApiResult<unknown>> {
	return apiRequest<unknown>(event, '/portfolios/entries', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
}
