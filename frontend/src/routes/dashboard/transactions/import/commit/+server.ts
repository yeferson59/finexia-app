import { json } from '@sveltejs/kit';
import * as transactions from '$lib/api/transactions';
import type { RequestHandler } from './$types';

// Proxies the confirmed import (file + mapping + portfolio/platform) to the
// backend, which persists every valid row in one database transaction.
export const POST: RequestHandler = async ({ request, cookies, fetch }) => {
	const form = await request.formData();

	const res = await transactions.importCommit({ cookies, fetch }, form);

	return json(await res.json(), { status: res.status });
};
