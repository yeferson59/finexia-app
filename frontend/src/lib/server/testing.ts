import type { Cookies } from '@sveltejs/kit';

export type MockCookies = Cookies & {
	/** Records every cookies.set() call, in order. */
	setCalls: { name: string; value: string; opts: Record<string, unknown> }[];
	/** Names passed to cookies.delete(), in order. */
	deleteCalls: string[];
};

/**
 * Minimal in-memory {@link Cookies} implementation for server-side auth tests.
 * Only the members our code touches (get/set/delete) are backed by real state;
 * the rest are present to satisfy the type and throw if unexpectedly used.
 */
export function createMockCookies(initial: Record<string, string> = {}): MockCookies {
	const store = new Map<string, string>(Object.entries(initial));
	const setCalls: MockCookies['setCalls'] = [];
	const deleteCalls: string[] = [];

	const cookies = {
		get: (name: string) => store.get(name),
		getAll: () => [...store.entries()].map(([name, value]) => ({ name, value })),
		set: (name: string, value: string, opts: Record<string, unknown>) => {
			store.set(name, value);
			setCalls.push({ name, value, opts });
		},
		delete: (name: string) => {
			store.delete(name);
			deleteCalls.push(name);
		},
		serialize: () => {
			throw new Error('serialize() not implemented in mock');
		},
		setCalls,
		deleteCalls
	};

	return cookies as unknown as MockCookies;
}

/** Builds a JSON Response with an optional set-cookie header for refresh rotation. */
export function jsonResponse(
	body: unknown,
	init: { status?: number; setCookie?: string } = {}
): Response {
	const headers = new Headers({ 'content-type': 'application/json' });
	if (init.setCookie) headers.append('set-cookie', init.setCookie);
	return new Response(JSON.stringify(body), { status: init.status ?? 200, headers });
}
