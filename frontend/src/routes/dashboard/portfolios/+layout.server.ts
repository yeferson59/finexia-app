import type { LayoutServerLoad } from './$types';
import * as portfolio from '$lib/api/portfolio';
import { resolveDisplayCurrency } from '$lib/shared/currency';

export const load: LayoutServerLoad = async ({ cookies, fetch, locals }) => {
	// Se pide en la moneda de la cuenta: el listado suma los valores de todos
	// los portafolios y sin conversión esa suma mezclaba las monedas base de
	// cada uno. El backend deja sin convertir lo que no tenga tasa y lo marca,
	// así que la página sabe qué puede sumar.
	const currency = resolveDisplayCurrency(null, locals.user?.preferredCurrency);
	const res = await portfolio.getSummaries({ cookies, fetch }, currency);

	if (!res.success) {
		return { portfolios: [], currency, success: false };
	}

	return { portfolios: res.data ?? [], currency, success: true };
};
