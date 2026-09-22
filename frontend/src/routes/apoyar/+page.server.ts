import { fail } from '@sveltejs/kit';
import { boldKeys, createCheckout, fetchPayment, isOrderId } from '$lib/server/bold';
import { supportAmountSchema, supportResult } from '$lib/features/support';
import type { Actions, PageServerLoad } from './$types';

// Bold devuelve a quien paga aquí con `?bold-order-id=…&bold-tx-status=…`. El
// estado se pregunta a Bold: el de la URL solo sirve si Bold no contesta.
export const load: PageServerLoad = async ({ url, fetch }) => {
	const keys = boldKeys();
	const orderId = url.searchParams.get('bold-order-id');

	let result = null;
	if (keys && isOrderId(orderId)) {
		const payment = await fetchPayment(fetch, keys, orderId);
		result = supportResult(
			payment?.status ?? null,
			payment?.total ?? null,
			url.searchParams.get('bold-tx-status')
		);
	}

	return { enabled: keys !== null, result };
};

export const actions: Actions = {
	// Crea y firma la orden; el navegador abre con ella la pasarela de Bold.
	pagar: async ({ request, url }) => {
		const keys = boldKeys();
		if (!keys)
			return fail(503, { error: 'Los aportes con Bold no están disponibles ahora mismo.' });

		const parsed = supportAmountSchema.safeParse(Object.fromEntries(await request.formData()));
		if (!parsed.success) return fail(400, { error: parsed.error.issues[0].message });

		return { checkout: createCheckout(keys, parsed.data, new URL('/apoyar', url.origin).href) };
	}
};
