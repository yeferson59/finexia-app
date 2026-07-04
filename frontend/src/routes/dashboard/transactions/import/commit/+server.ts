import { json } from '@sveltejs/kit';
import { authedFetch } from '$lib/server/api';
import type { RequestHandler } from './$types';

// Proxies the confirmed import (file + mapping + portfolio/platform) to the
// backend, which persists every valid row in one database transaction.
export const POST: RequestHandler = async ({ request, cookies, fetch }) => {
	const form = await request.formData();

	const res = await authedFetch({ cookies, fetch }, '/portfolios/transactions/import', {
		method: 'POST',
		body: form
	});

	return json(await res.json(), { status: res.status });
};
