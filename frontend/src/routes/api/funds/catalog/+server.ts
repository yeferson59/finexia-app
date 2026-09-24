import { json } from '@sveltejs/kit';
import * as funds from '$lib/api/funds';
import type { RequestHandler } from './$types';

/**
 * El buscador de fondos de la Superfinanciera. Va por aquí y no directo al
 * backend por lo mismo que `/api/assets`: el token vive en una cookie httpOnly
 * que el componente no puede leer.
 */
export const GET: RequestHandler = async ({ url, cookies, fetch }) => {
	const q = url.searchParams.get('q') ?? '';

	const res = await funds.searchPublicFunds({ cookies, fetch }, q);
	if (!res.ok || !res.success) {
		return json({ success: false, status: res.status, data: [] }, { status: res.status || 500 });
	}

	return json({ success: true, data: res.data ?? [] });
};
