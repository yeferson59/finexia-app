import { json } from '@sveltejs/kit';
import { authedFetch } from '$lib/server/api';
import type { RequestHandler } from './$types';

// Proxies the multipart upload (file + mapping) to the backend parser so the
// browser never talks to the API directly.
export const POST: RequestHandler = async ({ request, cookies, fetch }) => {
	const form = await request.formData();

	const res = await authedFetch({ cookies, fetch }, '/portfolios/transactions/import/preview', {
		method: 'POST',
		body: form
	});

	return json(await res.json(), { status: res.status });
};
