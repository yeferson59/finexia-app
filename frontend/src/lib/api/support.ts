/**
 * Aportes de `/apoyar` (`docs/API.md` §2.13). Las tres primeras son públicas,
 * sin sesión, como {@link marketing}: la página la ve cualquiera. Las dos de
 * administración van con la sesión, como el resto del panel.
 *
 * El backend guarda cada orden, la firma con la llave secreta de Bold y sigue
 * su estado por el webhook. Aquí no hay llaves: solo se piden la orden firmada y
 * su estado.
 */
import { apiRequestSafe, publicRequest, type ApiEvent, type ApiResult } from './client';
import {
	adminContributionSchema,
	paginatedSchema,
	supportCheckoutSchema,
	supportConfigSchema,
	supportContributionSchema,
	supportSummarySchema
} from './schemas';
import type {
	AdminContribution,
	Paginated,
	SupportCheckout,
	SupportConfig,
	SupportContribution,
	SupportStatus,
	SupportSummary
} from './types';

type Fetch = typeof fetch;

/** `GET /support` — si se pueden recibir aportes ahora mismo. */
export function getSupportConfig(fetchFn: Fetch) {
	return publicRequest<SupportConfig>(fetchFn, '/support', {}, supportConfigSchema);
}

/** `POST /support/checkouts` — crea y firma una orden por `amount` pesos. */
export function createSupportCheckout(fetchFn: Fetch, amount: number) {
	return publicRequest<SupportCheckout>(
		fetchFn,
		'/support/checkouts',
		{
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ amount })
		},
		supportCheckoutSchema
	);
}

/**
 * `GET /support/checkouts/:orderId` — el estado de una orden. El backend le
 * pregunta a Bold si aún no está resuelta, así que es el dato bueno aunque el
 * webhook no haya llegado.
 */
export function getSupportContribution(fetchFn: Fetch, orderId: string) {
	return publicRequest<SupportContribution>(
		fetchFn,
		`/support/checkouts/${encodeURIComponent(orderId)}`,
		{},
		supportContributionSchema
	);
}

/**
 * `GET /support/contributions` — los aportes, del más nuevo al más viejo
 * (admin). Sin `status`, todos.
 */
export function getAdminContributions(
	event: ApiEvent,
	opts: { page?: number; limit?: number; status?: SupportStatus } = {}
): Promise<ApiResult<Paginated<AdminContribution>>> {
	const params = new URLSearchParams({
		page: String(opts.page ?? 1),
		limit: String(opts.limit ?? 25)
	});
	if (opts.status) params.set('status', opts.status);

	return apiRequestSafe(
		event,
		`/support/contributions?${params}`,
		{},
		paginatedSchema(adminContributionSchema)
	);
}

/** `GET /support/contributions/summary` — cuánto se ha aportado (admin). */
export function getSupportSummary(event: ApiEvent): Promise<ApiResult<SupportSummary>> {
	return apiRequestSafe(event, '/support/contributions/summary', {}, supportSummarySchema);
}
