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
