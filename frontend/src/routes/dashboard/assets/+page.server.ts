import type { PageServerLoad } from './$types';
import * as portfolio from '$lib/api/portfolio';
import { resolveDisplayCurrency } from '$lib/shared/currency';
import type { AssetHolding } from '$lib/api/types';

export const load: PageServerLoad = async ({ cookies, fetch, url, locals }) => {
	// Mismo contrato que el panel: sin `?currency=` manda la moneda de la
	// cuenta. Aquí no es opcional que haya una — las filas suman posiciones de
	// portafolios que pueden estar en monedas distintas, así que el peso de cada
	// activo solo significa algo si todo llega convertido a la misma.
	const currency = resolveDisplayCurrency(
		url.searchParams.get('currency'),
		locals.user?.preferredCurrency
	);

	const res = await portfolio.getAssetHoldings({ cookies, fetch }, currency);

	const holdings: AssetHolding[] = res.ok && res.success && Array.isArray(res.data) ? res.data : [];

	return { holdings, currency };
};
