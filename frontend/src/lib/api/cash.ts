/**
 * Efectivo (`/portfolios/cash`): los saldos de cada plataforma por moneda y
 * portafolio, y los depósitos, retiros e intereses que los cambian.
 */
import { apiRequest, apiRequestSafe, type ApiEvent, type ApiResult } from './client';
import type { CashBalance, CashMovement, PagedCashMovements } from './types';
import { cashBalanceSchema, pagedCashMovementsSchema } from './schemas';
import { z } from 'zod';

/**
 * `GET /portfolios/cash` — saldos del usuario, cada uno con su valor en
 * `currency` (o en la moneda de la cuenta si no se indica).
 */
export function getBalances(event: ApiEvent, currency?: string): Promise<ApiResult<CashBalance[]>> {
	const query = currency ? `?currency=${encodeURIComponent(currency)}` : '';
	return apiRequestSafe(event, `/portfolios/cash${query}`, {}, z.array(cashBalanceSchema));
}

/** `GET /portfolios/cash/movements` — movimientos, del más reciente al más antiguo. */
export function getMovements(
	event: ApiEvent,
	page: number,
	limit: number
): Promise<ApiResult<PagedCashMovements>> {
	return apiRequestSafe(
		event,
		`/portfolios/cash/movements?page=${page}&limit=${limit}`,
		{},
		pagedCashMovementsSchema
	);
}

/**
 * `POST /portfolios/cash/movements` — registra un movimiento. Si la plataforma
 * aún no guardaba esa moneda en ese portafolio, el backend abre el saldo.
 */
export function createMovement(
	event: ApiEvent,
	body: Record<string, unknown>
): Promise<ApiResult<CashMovement>> {
	return apiRequest<CashMovement>(event, '/portfolios/cash/movements', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
}

/**
 * `PUT /portfolios/cash/movements/:id` — reescribe un movimiento sobre el mismo
 * saldo. Portafolio, plataforma y moneda no se cambian aquí.
 */
export function updateMovement(
	event: ApiEvent,
	id: string,
	body: Record<string, unknown>
): Promise<ApiResult<CashMovement>> {
	return apiRequest<CashMovement>(event, `/portfolios/cash/movements/${id}`, {
		method: 'PUT',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
}

/** `DELETE /portfolios/cash/movements/:id` — borra un movimiento. */
export function deleteMovement(event: ApiEvent, id: string): Promise<ApiResult<unknown>> {
	return apiRequest<unknown>(event, `/portfolios/cash/movements/${id}`, { method: 'DELETE' });
}
