import { fail } from '@sveltejs/kit';
import { createSupportCheckout, getSupportConfig, getSupportContribution } from '$lib/api/support';
import { ACCESS_COOKIE, REFRESH_COOKIE } from '$lib/server/session';
import { supportAmountSchema, supportResult } from '$lib/features/support';
import type { Actions, PageServerLoad } from './$types';

const UNAVAILABLE = 'Los aportes con Bold no están disponibles ahora mismo.';

// Bold devuelve a quien paga aquí con `?bold-order-id=…&bold-tx-status=…`. El
// estado lo da el backend, que ya lo sabe por el webhook o se lo pregunta a
// Bold; el de la URL solo sirve si el backend no contesta.
export const load: PageServerLoad = async ({ url, fetch, cookies }) => {
	const orderId = url.searchParams.get('bold-order-id');

	const [config, contribution] = await Promise.all([
		getSupportConfig(fetch),
		// Bold no admite ids de más de 60 caracteres: uno más largo no es nuestro.
		orderId && orderId.length <= 60 ? getSupportContribution(fetch, orderId) : null
	]);

	let result = null;
	// Un 404 es una orden que no existe: no se enseña nada, diga lo que diga la
	// URL. Solo si el backend no contesta se cae al estado de la URL.
	if (contribution && contribution.status !== 404) {
		result = supportResult(
			contribution.data?.status ?? null,
			contribution.data?.totalCharged ?? null,
			url.searchParams.get('bold-tx-status')
		);
	}

	// Quien tiene sesión vuelve al panel y no a la portada. La página es pública
	// y el hook no valida la sesión aquí: basta con ver la cookie, y si ya no
	// vale el panel lo manda a entrar.
	const signedIn = Boolean(cookies.get(ACCESS_COOKIE) ?? cookies.get(REFRESH_COOKIE));

	return { enabled: config.data?.enabled ?? false, result, signedIn };
};

export const actions: Actions = {
	// El backend guarda la orden y la firma; el navegador abre con ella la
	// pasarela de Bold.
	pagar: async ({ request, fetch }) => {
		const parsed = supportAmountSchema.safeParse(Object.fromEntries(await request.formData()));
		if (!parsed.success) return fail(400, { error: parsed.error.issues[0].message });

		const res = await createSupportCheckout(fetch, parsed.data);
		if (res.status === 429) {
			return fail(429, {
				error: 'Demasiados intentos seguidos. Espera unos minutos y vuelve a probar.'
			});
		}
		// El formulario ya valida el mismo rango; un 400 aquí es que los topes
		// del backend cambiaron sin los de `support.ts`.
		if (res.status === 400) return fail(400, { error: 'Ese monto no se puede aportar.' });
		if (!res.ok || !res.data) return fail(503, { error: UNAVAILABLE });

		return { checkout: res.data };
	}
};
