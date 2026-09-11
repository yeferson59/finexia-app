import type { RequestEvent } from '@sveltejs/kit';
import { apiUrl } from './client';

/**
 * The part of the backend that is reachable from outside this app.
 *
 * The backend lives on a private network: the browser never talks to it, and
 * everything the app needs is fetched by its own server. But a slice of the API
 * is used by others — an MCP client, a connector running OAuth discovery, an
 * `<img>` tag — and so needs a public URL. `hooks.server.ts` forwards exactly
 * these paths, unchanged, to the same path on the backend, and the backend
 * advertises this app's origin as its own (`PUBLIC_URL`, docs/API.md §1.6).
 *
 * The list is closed on purpose: a route that is not here does not exist from
 * the outside, which is the whole point of taking the backend off the internet.
 */
const PUBLIC_API_PATHS: readonly RegExp[] = [
	// The MCP server (docs/API.md §2.11).
	/^\/mcp$/,
	// OAuth discovery, with and without the suffix the spec derives from `/mcp`.
	/^\/\.well-known\/oauth-(?:protected-resource|authorization-server)(?:\/mcp)?$/,
	// The three OAuth endpoints the client drives itself. The fourth step of the
	// flow, `/oauth/consent`, is a page of this app and never reaches here.
	/^\/oauth\/(?:register|authorize|token)$/,
	// Public avatars (§2.3). Ids only: `/users/me/avatar` is the upload, which is
	// authenticated and goes through the settings action instead.
	/^\/users\/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\/avatar$/i
];

/**
 * Request headers that cross over to the backend. An allowlist, and what
 * matters is what it leaves out: the `Cookie`. None of these routes uses this
 * app's session, and never forwarding it is what lets them skip the CSRF check
 * (`$lib/server/csrf`): with no cookie there is no ambient authority for a
 * cross-site form to borrow.
 *
 * The client IP and User-Agent are absent because `handleFetch` sets them, as
 * it does on every other backend call.
 */
const FORWARDED_REQUEST_HEADERS = [
	'accept',
	'authorization',
	'content-type',
	'if-modified-since',
	'if-none-match',
	'last-event-id',
	'mcp-protocol-version',
	'mcp-session-id'
];

/**
 * Response headers that do not travel back. The hop-by-hop ones describe the
 * connection to the backend, not the client's; `content-encoding` and
 * `content-length` would describe bytes `fetch` has already decoded; and
 * `set-cookie` because the backend has nothing to store on this origin.
 */
const DROPPED_RESPONSE_HEADERS = [
	'connection',
	'keep-alive',
	'proxy-connection',
	'te',
	'trailer',
	'transfer-encoding',
	'upgrade',
	'content-encoding',
	'content-length',
	'set-cookie'
];

/** Statuses that carry no body; `new Response` throws if handed one. */
const NULL_BODY_STATUSES = new Set([101, 204, 205, 304]);

export function isPublicApiPath(pathname: string) {
	return PUBLIC_API_PATHS.some((path) => path.test(pathname));
}

/**
 * Forwards the request to the same path on the backend and returns its answer
 * untouched, bar the headers above.
 *
 * `redirect: 'manual'` because the redirect `/oauth/authorize` answers with is
 * meant for the browser, not for this server. `credentials: 'omit'` because
 * `event.fetch` would otherwise attach this app's cookies whenever the backend
 * sits on a subdomain of it. The response body is streamed, never buffered.
 */
export async function proxyToBackend(event: Pick<RequestEvent, 'request' | 'url' | 'fetch'>) {
	const { request, url } = event;

	const headers = new Headers();
	for (const name of FORWARDED_REQUEST_HEADERS) {
		const value = request.headers.get(name);
		if (value !== null) headers.set(name, value);
	}

	let res: Response;
	try {
		res = await event.fetch(apiUrl(url.pathname + url.search), {
			method: request.method,
			headers,
			// Read whole: these bodies are small (a JSON-RPC call, a token form),
			// and streaming one in would need `duplex: 'half'` all the way down.
			body:
				request.method === 'GET' || request.method === 'HEAD'
					? undefined
					: await request.arrayBuffer(),
			redirect: 'manual',
			credentials: 'omit'
		});
	} catch {
		return new Response('Bad Gateway', { status: 502 });
	}

	const responseHeaders = new Headers(res.headers);
	for (const name of DROPPED_RESPONSE_HEADERS) responseHeaders.delete(name);

	const hasBody = request.method !== 'HEAD' && !NULL_BODY_STATUSES.has(res.status);

	return new Response(hasBody ? res.body : null, {
		status: res.status,
		statusText: res.statusText,
		headers: responseHeaders
	});
}
