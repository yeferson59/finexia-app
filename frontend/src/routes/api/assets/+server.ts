import { json } from '@sveltejs/kit';
import * as market from '$lib/api/market';
import type { RequestHandler } from './$types';

export const GET: RequestHandler = async ({ url, cookies, fetch }) => {
	const search = url.searchParams.get('search') ?? '';
	const limit = url.searchParams.get('limit') ?? '10';

	const res = await market.searchAssets({ cookies, fetch }, { search, limit });
	if (!res?.ok) return json({ success: false, data: [] });

	return json(await res.json());
};

/**
 * Alta de un activo desde el buscador, para cuando el catálogo no tiene lo que
 * el usuario está registrando. Va por aquí y no directo al backend porque el
 * token de acceso vive en una cookie httpOnly que el componente no puede leer.
 *
 * El backend cura o aporta según el rol; esta ruta no distingue.
 */
export const POST: RequestHandler = async ({ request, cookies, fetch }) => {
	const body = await request.json().catch(() => null);
	if (!body || typeof body !== 'object') {
		return json({ success: false, message: 'Solicitud inválida' }, { status: 400 });
	}

	const { ticker, name, assetType, exchange, currency } = body as Record<string, unknown>;

	const res = await market.createAsset(
		{ cookies, fetch },
		{
			ticker: String(ticker ?? '')
				.trim()
				.toUpperCase(),
			name: String(name ?? '').trim(),
			assetType: String(assetType ?? ''),
			exchange: String(exchange ?? '').trim(),
			currency: String(currency ?? '')
				.trim()
				.toUpperCase()
		}
	);

	if (!res.ok) {
		// `details` lo pone el binder del backend, `action` los errores de
		// dominio; el primero que haya es el que explica qué pasó.
		return json(
			{ success: false, message: res.details ?? res.action ?? 'No se pudo crear el activo' },
			{ status: res.status || 500 }
		);
	}

	return json({ success: true, data: res.data });
};
