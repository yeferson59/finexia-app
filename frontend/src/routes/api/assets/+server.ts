import { json } from '@sveltejs/kit';
import { authedFetchSafe } from '$lib/server/api';
import type { RequestHandler } from './$types';

export const GET: RequestHandler = async ({ url, cookies, fetch }) => {
	const search = url.searchParams.get('search') ?? '';
	const limit = url.searchParams.get('limit') ?? '10';

	const endpoint = search.trim()
		? `/portfolios/assets?search=${encodeURIComponent(search.trim())}&page=1&limit=${limit}`
		: `/portfolios/assets?page=1&limit=${limit}`;

	const res = await authedFetchSafe({ cookies, fetch }, endpoint);
	if (!res?.ok) return json({ success: false, data: [] });

	return json(await res.json());
};
