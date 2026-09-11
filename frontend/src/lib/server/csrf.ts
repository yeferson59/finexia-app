import { json, text } from '@sveltejs/kit';

/**
 * SvelteKit's origin check for form submissions, done here instead.
 *
 * SvelteKit rejects a POST/PUT/PATCH/DELETE with a form body whose `Origin` is
 * not the app's own, but it decides before `handle` runs and only for every
 * route or none. That rules out `/oauth/token`, which this app forwards to the
 * backend (`$lib/api/proxy`) and which by definition receives forms from other
 * servers (RFC 6749 §4.1.3), usually with no `Origin` at all. So the built-in
 * check is off in `svelte.config.js`, and `hooks.server.ts` runs this one for
 * everything the proxy does not take.
 *
 * It mirrors SvelteKit 2's point by point — same methods, same content types,
 * same answer — including rejecting a form that carries no `Origin`.
 */
const FORM_CONTENT_TYPES = new Set([
	'application/x-www-form-urlencoded',
	'multipart/form-data',
	'text/plain',
	// SvelteKit's binary encoding for remote-function forms.
	'application/x-sveltekit-formdata'
]);

const UNSAFE_METHODS = new Set(['POST', 'PUT', 'PATCH', 'DELETE']);

export function isCrossSiteFormSubmission(request: Request, url: URL) {
	const contentType = request.headers.get('content-type')?.split(';', 1)[0].trim() ?? '';

	return (
		UNSAFE_METHODS.has(request.method) &&
		FORM_CONTENT_TYPES.has(contentType.toLowerCase()) &&
		request.headers.get('origin') !== url.origin
	);
}

/** The same 403 SvelteKit answers with, so no client can tell the check moved. */
export function crossSiteFormResponse(request: Request) {
	const message = `Cross-site ${request.method} form submissions are forbidden`;

	if (request.headers.get('accept') === 'application/json') {
		return json({ message }, { status: 403 });
	}

	return text(message, { status: 403 });
}
