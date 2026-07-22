import { json } from '@sveltejs/kit';
import * as transactions from '$lib/api/transactions';
import type { RequestHandler } from './$types';

// Proxies the multipart upload (file + mapping) to the backend parser so the
// browser never talks to the API directly.
export const POST: RequestHandler = async ({ request, cookies, fetch }) => {
	const form = await request.formData();

	const res = await transactions.importPreview({ cookies, fetch }, form);

	return json(await res.json(), { status: res.status });
};
