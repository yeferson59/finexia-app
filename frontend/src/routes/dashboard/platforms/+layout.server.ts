import type { LayoutServerLoad } from './$types';
import * as platforms from '$lib/api/platforms';
import { resolveDisplayCurrency } from '$lib/shared/currency';

export const load: LayoutServerLoad = async ({ cookies, fetch, locals }) => {
	// El total de cada plataforma suma posiciones compradas en distintas
	// monedas, así que se pide en la de la cuenta: el backend convierte lo que
	// tenga tasa, deja el resto a valor nominal y lo marca.
	const currency = resolveDisplayCurrency(null, locals.user?.preferredCurrency);
	const res = await platforms.getSources({ cookies, fetch }, currency);

	if (!res.success) {
		return { platforms: [], currency };
	}

	return { platforms: res.data ?? [], currency };
};
