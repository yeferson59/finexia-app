/**
 * Efectivo (`/portfolios/cash`): los saldos de cada plataforma por moneda y
 * portafolio, los depósitos, retiros e intereses que los cambian, y la tasa que
 * rinde cada cuenta.
 */
import { apiRequest, apiRequestSafe, type ApiEvent, type ApiResult } from './client';
import type {
	CashBalance,
	CashMove,
	CashMovement,
	CashPocket,
	CashRate,
	PagedCashMovements
} from './types';
import {
	cashBalanceSchema,
	cashPocketSchema,
	cashRateSchema,
	pagedCashMovementsSchema
} from './schemas';
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

/**
 * `GET /portfolios/cash/rates` — todas las versiones de las tasas, por
 * plataforma y moneda, de la más nueva a la más vieja.
 */
export function getRates(event: ApiEvent): Promise<ApiResult<CashRate[]>> {
	return apiRequestSafe(event, '/portfolios/cash/rates', {}, z.array(cashRateSchema));
}

/**
 * `POST /portfolios/cash/rates` — anota una tasa, o una versión nueva que cierra
 * la anterior la víspera.
 */
export function createRate(
	event: ApiEvent,
	body: Record<string, unknown>
): Promise<ApiResult<CashRate>> {
	return apiRequest<CashRate>(event, '/portfolios/cash/rates', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
}

/** `PUT /portfolios/cash/rates/:id` — corrige los valores de la versión más reciente. */
export function updateRate(
	event: ApiEvent,
	id: string,
	body: Record<string, unknown>
): Promise<ApiResult<CashRate>> {
	return apiRequest<CashRate>(event, `/portfolios/cash/rates/${id}`, {
		method: 'PUT',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
}

/** `POST /portfolios/cash/rates/:id/end` — la cuenta deja de rendir desde `endsOn`. */
export function endRate(
	event: ApiEvent,
	id: string,
	body: Record<string, unknown>
): Promise<ApiResult<CashRate>> {
	return apiRequest<CashRate>(event, `/portfolios/cash/rates/${id}/end`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
}

/**
 * `POST /portfolios/cash/interest/recalculate` — tira los días que la cuenta ya
 * tenía calculados desde una fecha y los vuelve a calcular sobre lo que guarda
 * hoy. Es lo que arregla un movimiento anotado con fecha pasada.
 */
export function recalculateInterest(
	event: ApiEvent,
	body: Record<string, unknown>
): Promise<ApiResult<unknown>> {
	return apiRequest<unknown>(event, '/portfolios/cash/interest/recalculate', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
}

/** `DELETE /portfolios/cash/rates/:id` — borra la versión más reciente. */
export function deleteRate(event: ApiEvent, id: string): Promise<ApiResult<unknown>> {
	return apiRequest<unknown>(event, `/portfolios/cash/rates/${id}`, { method: 'DELETE' });
}

/**
 * `GET /portfolios/cash/pockets` — los bolsillos del usuario, abiertos y
 * cerrados, por plataforma y moneda.
 */
export function getPockets(event: ApiEvent): Promise<ApiResult<CashPocket[]>> {
	return apiRequestSafe(event, '/portfolios/cash/pockets', {}, z.array(cashPocketSchema));
}

/** `POST /portfolios/cash/pockets` — abre un bolsillo flexible en una cuenta. */
export function createPocket(
	event: ApiEvent,
	body: Record<string, unknown>
): Promise<ApiResult<CashPocket>> {
	return apiRequest<CashPocket>(event, '/portfolios/cash/pockets', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
}

/** `PUT /portfolios/cash/pockets/:id` — lo único que se cambia es el nombre. */
export function renamePocket(
	event: ApiEvent,
	id: string,
	body: Record<string, unknown>
): Promise<ApiResult<CashPocket>> {
	return apiRequest<CashPocket>(event, `/portfolios/cash/pockets/${id}`, {
		method: 'PUT',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
}

/** `DELETE /portfolios/cash/pockets/:id` — borra un bolsillo sin movimientos. */
export function deletePocket(event: ApiEvent, id: string): Promise<ApiResult<unknown>> {
	return apiRequest<unknown>(event, `/portfolios/cash/pockets/${id}`, { method: 'DELETE' });
}

/**
 * `POST /portfolios/cash/movements/move` — mueve dinero entre dos saldos de una
 * misma cuenta dentro de un portafolio. Son dos patas en una sola transacción y
 * se compensan, así que la rentabilidad del portafolio no se mueve.
 */
export function moveCash(
	event: ApiEvent,
	body: Record<string, unknown>
): Promise<ApiResult<CashMove>> {
	return apiRequest<CashMove>(event, '/portfolios/cash/movements/move', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
}
