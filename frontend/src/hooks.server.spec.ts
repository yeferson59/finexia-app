import { describe, it, expect, vi } from 'vitest';

vi.mock('$env/dynamic/private', () => ({ env: { BASE_API: 'https://api.finexia.example' } }));

const { handleFetch } = await import('./hooks.server');

type HandleFetchArgs = Parameters<typeof handleFetch>[0];

function buildArgs(overrides: {
	incomingUserAgent?: string | null;
	clientAddress?: string;
	outgoingUrl: string;
}): { args: HandleFetchArgs; fetch: ReturnType<typeof vi.fn> } {
	const fetch = vi.fn().mockResolvedValue(new Response(null));

	const incomingHeaders = new Headers();
	if (overrides.incomingUserAgent) {
		incomingHeaders.set('user-agent', overrides.incomingUserAgent);
	}

	const event = {
		request: new Request('https://finexia.example/auth/login', { headers: incomingHeaders }),
		getClientAddress: () =>
			overrides.clientAddress ??
			(() => {
				throw new Error('client address unavailable');
			})()
	};

	const args = {
		event,
		request: new Request(overrides.outgoingUrl),
		fetch
	} as unknown as HandleFetchArgs;

	return { args, fetch };
}

describe('handleFetch', () => {
	it('forwards the client IP and User-Agent to backend requests', async () => {
		const { args, fetch } = buildArgs({
			incomingUserAgent: 'Mozilla/5.0 (iPhone; CPU iPhone OS 18_0)',
			clientAddress: '203.0.113.7',
			outgoingUrl: 'https://api.finexia.example/auth/login'
		});

		await handleFetch(args);

		expect(fetch).toHaveBeenCalledTimes(1);
		const forwarded = fetch.mock.calls[0][0] as Request;
		expect(forwarded.headers.get('X-Forwarded-For')).toBe('203.0.113.7');
		expect(forwarded.headers.get('User-Agent')).toBe('Mozilla/5.0 (iPhone; CPU iPhone OS 18_0)');
	});

	it('leaves requests to other hosts untouched', async () => {
		const { args, fetch } = buildArgs({
			incomingUserAgent: 'Mozilla/5.0 (iPhone; CPU iPhone OS 18_0)',
			clientAddress: '203.0.113.7',
			outgoingUrl: 'https://other-service.example/ping'
		});

		await handleFetch(args);

		const forwarded = fetch.mock.calls[0][0] as Request;
		expect(forwarded.headers.get('X-Forwarded-For')).toBeNull();
		expect(forwarded.headers.get('User-Agent')).toBeNull();
	});

	it('still forwards the client IP when the incoming request has no User-Agent', async () => {
		const { args, fetch } = buildArgs({
			incomingUserAgent: null,
			clientAddress: '203.0.113.7',
			outgoingUrl: 'https://api.finexia.example/auth/login'
		});

		await handleFetch(args);

		const forwarded = fetch.mock.calls[0][0] as Request;
		expect(forwarded.headers.get('X-Forwarded-For')).toBe('203.0.113.7');
		expect(forwarded.headers.get('User-Agent')).toBeNull();
	});

	it('does not throw when getClientAddress() is unavailable (e.g. prerendering)', async () => {
		const { args, fetch } = buildArgs({
			incomingUserAgent: 'Mozilla/5.0',
			outgoingUrl: 'https://api.finexia.example/auth/login'
		});

		await expect(handleFetch(args)).resolves.toBeInstanceOf(Response);

		const forwarded = fetch.mock.calls[0][0] as Request;
		expect(forwarded.headers.get('X-Forwarded-For')).toBeNull();
		expect(forwarded.headers.get('User-Agent')).toBe('Mozilla/5.0');
	});
});
