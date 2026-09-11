import { describe, it, expect, vi } from 'vitest';

vi.mock('$env/dynamic/private', () => ({ env: { BASE_API: 'http://backend.internal:8080' } }));

const { isPublicApiPath, proxyToBackend } = await import('./proxy');

type ProxyEvent = Parameters<typeof proxyToBackend>[0];

const AVATAR_PATH = '/users/0b7f9c7e-1111-4222-8333-444455556666/avatar';

function proxyEvent(
	path: string,
	init: RequestInit = {},
	backend: Response | Error = new Response(null, { status: 200 })
) {
	const fetch =
		backend instanceof Error
			? vi.fn().mockRejectedValue(backend)
			: vi.fn().mockResolvedValue(backend);
	const url = new URL(`https://finexia.example${path}`);
	const event = { request: new Request(url, init), url, fetch } as unknown as ProxyEvent;

	return { event, fetch };
}

describe('isPublicApiPath', () => {
	it.each([
		'/mcp',
		'/.well-known/oauth-protected-resource',
		'/.well-known/oauth-protected-resource/mcp',
		'/.well-known/oauth-authorization-server',
		'/.well-known/oauth-authorization-server/mcp',
		'/oauth/register',
		'/oauth/authorize',
		'/oauth/token',
		AVATAR_PATH
	])('exposes %s', (path) => {
		expect(isPublicApiPath(path)).toBe(true);
	});

	it.each([
		// A page of this app, not of the backend.
		'/oauth/consent',
		// The authenticated upload, and the rest of /users.
		'/users/me/avatar',
		'/users/me',
		'/auth/session',
		'/mcp/',
		'/mcp/tools',
		'/.well-known/security.txt',
		'/dashboard'
	])('keeps %s private', (path) => {
		expect(isPublicApiPath(path)).toBe(false);
	});
});

describe('proxyToBackend', () => {
	it('forwards the method, path, query and body to the same path on the backend', async () => {
		const { event, fetch } = proxyEvent('/oauth/token?trace=1', {
			method: 'POST',
			headers: { 'content-type': 'application/x-www-form-urlencoded' },
			body: 'grant_type=authorization_code&code=abc'
		});

		await proxyToBackend(event);

		const [target, init] = fetch.mock.calls[0];
		expect(target).toBe('http://backend.internal:8080/oauth/token?trace=1');
		expect(init.method).toBe('POST');
		expect(new TextDecoder().decode(init.body)).toBe('grant_type=authorization_code&code=abc');
		expect(init.headers.get('content-type')).toBe('application/x-www-form-urlencoded');
	});

	it("forwards the client's own credentials, never this app's session cookie", async () => {
		const { event, fetch } = proxyEvent('/mcp', {
			method: 'POST',
			headers: {
				authorization: 'Bearer fnx_oat_abc',
				'mcp-protocol-version': '2025-06-18',
				cookie: 'access_token_finexia=secret',
				'x-forwarded-for': '198.51.100.9'
			},
			body: '{}'
		});

		await proxyToBackend(event);

		const [, init] = fetch.mock.calls[0];
		expect(init.headers.get('authorization')).toBe('Bearer fnx_oat_abc');
		expect(init.headers.get('mcp-protocol-version')).toBe('2025-06-18');
		expect(init.headers.get('cookie')).toBeNull();
		// Set by handleFetch from the real client address, not trusted from here.
		expect(init.headers.get('x-forwarded-for')).toBeNull();
		expect(init.credentials).toBe('omit');
	});

	it('hands the redirect of /oauth/authorize back to the browser instead of following it', async () => {
		const consent = 'https://finexia.example/oauth/consent?request=42';
		const { event, fetch } = proxyEvent(
			'/oauth/authorize?client_id=abc',
			{},
			new Response(null, { status: 302, headers: { location: consent } })
		);

		const res = await proxyToBackend(event);

		expect(fetch.mock.calls[0][1].redirect).toBe('manual');
		expect(res.status).toBe(302);
		expect(res.headers.get('location')).toBe(consent);
	});

	it('returns the backend answer minus the headers that describe the hop or set cookies here', async () => {
		const { event } = proxyEvent(
			'/mcp',
			{ method: 'POST', body: '{}' },
			new Response('{"message":"invalid or missing token"}', {
				status: 401,
				headers: {
					'content-type': 'application/json',
					'www-authenticate': 'Bearer realm="finexia"',
					'cache-control': 'no-store',
					'set-cookie': 'backend=1; Path=/',
					'content-encoding': 'gzip',
					connection: 'keep-alive'
				}
			})
		);

		const res = await proxyToBackend(event);

		expect(res.status).toBe(401);
		expect(await res.text()).toBe('{"message":"invalid or missing token"}');
		expect(res.headers.get('www-authenticate')).toBe('Bearer realm="finexia"');
		expect(res.headers.get('content-type')).toBe('application/json');
		expect(res.headers.get('cache-control')).toBe('no-store');
		expect(res.headers.get('set-cookie')).toBeNull();
		expect(res.headers.get('content-encoding')).toBeNull();
		expect(res.headers.get('connection')).toBeNull();
	});

	it('passes a bodiless answer through, such as the 204 of a CORS preflight', async () => {
		const { event } = proxyEvent(
			'/oauth/token',
			{ method: 'OPTIONS' },
			new Response(null, { status: 204, headers: { 'access-control-allow-origin': '*' } })
		);

		const res = await proxyToBackend(event);

		expect(res.status).toBe(204);
		expect(res.body).toBeNull();
		expect(res.headers.get('access-control-allow-origin')).toBe('*');
	});

	it('streams the avatar image through with its caching headers', async () => {
		const { event } = proxyEvent(
			AVATAR_PATH,
			{},
			new Response(new Uint8Array([0x89, 0x50, 0x4e, 0x47]), {
				headers: { 'content-type': 'image/png', 'cache-control': 'public, max-age=86400' }
			})
		);

		const res = await proxyToBackend(event);

		expect(res.headers.get('content-type')).toBe('image/png');
		expect(res.headers.get('cache-control')).toBe('public, max-age=86400');
		expect(new Uint8Array(await res.arrayBuffer())).toEqual(
			new Uint8Array([0x89, 0x50, 0x4e, 0x47])
		);
	});

	it('answers 502 when the backend cannot be reached', async () => {
		const { event } = proxyEvent(
			'/mcp',
			{ method: 'POST', body: '{}' },
			new TypeError('fetch failed')
		);

		const res = await proxyToBackend(event);

		expect(res.status).toBe(502);
	});
});
