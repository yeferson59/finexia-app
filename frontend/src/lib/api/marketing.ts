/**
 * Marketing (público): alta en la waitlist. Sin sesión, como {@link auth}.
 */
import { apiUrl } from './client';

type Fetch = typeof fetch;

/** `POST /marketing/waitlists` — alta en la waitlist. Devuelve la `Response` cruda. */
export function joinWaitlist(fetchFn: Fetch, email: string): Promise<Response> {
	return fetchFn(apiUrl('/marketing/waitlists'), {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ email })
	});
}
