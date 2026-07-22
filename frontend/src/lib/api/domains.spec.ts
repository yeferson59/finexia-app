import { describe, it, expect, vi } from 'vitest';
import { ACCESS_COOKIE } from '$lib/server/session';
import { createMockCookies, jsonResponse } from '$lib/server/testing';
import * as portfolio from './portfolio';
import * as transactions from './transactions';
import * as platforms from './platforms';
import * as market from './market';
import * as user from './user';
import * as auth from './auth';

/** Authed event whose `fetch` is a spy returning `response`. */
function authedEvent(response = jsonResponse({ success: true, data: [] })) {
	const cookies = createMockCookies({ [ACCESS_COOKIE]: 'access-1' });
	const fetch = vi.fn().mockResolvedValue(response);
	return { event: { cookies, fetch }, fetch };
}

/** Reads the [url, init] of the backend call a domain function made. */
function lastCall(fetch: ReturnType<typeof vi.fn>): [string, RequestInit] {
	const [url, init] = fetch.mock.calls.at(-1) as [string, RequestInit];
	return [String(url), init ?? {}];
}

describe('portfolio module', () => {
	it('getSummaries hits /portfolios/summary with the currency query and a Bearer header', async () => {
		const { event, fetch } = authedEvent(jsonResponse({ success: true, data: [{ id: 'p1' }] }));

		const res = await portfolio.getSummaries(event, 'USD');

		const [url, init] = lastCall(fetch);
		expect(url).toContain('/portfolios/summary?currency=USD');
		expect((init.headers as Record<string, string>).Authorization).toBe('Bearer access-1');
		expect(res.success).toBe(true);
		expect(res.data).toEqual([{ id: 'p1' }]);
	});

	it('updatePortfolio sends a PATCH with a JSON body to /portfolios/:id', async () => {
		const { event, fetch } = authedEvent(jsonResponse({ success: true }));

		await portfolio.updatePortfolio(event, 'p1', { name: 'Nuevo' });

		const [url, init] = lastCall(fetch);
		expect(url).toContain('/portfolios/p1');
		expect(init.method).toBe('PATCH');
		expect(init.body).toBe(JSON.stringify({ name: 'Nuevo' }));
	});

	it('propagates an error envelope as ok:false with the status', async () => {
		const { event } = authedEvent(jsonResponse({ success: false, message: 'x' }, { status: 500 }));

		const res = await portfolio.getRisks(event);

		expect(res.ok).toBe(false);
		expect(res.status).toBe(500);
		expect(res.data).toBeNull();
	});
});

describe('transactions module', () => {
	it('getAssetTransactions encodes page and limit in the path', async () => {
		const { event, fetch } = authedEvent(jsonResponse({ success: true, data: { data: [] } }));

		await transactions.getAssetTransactions(event, 'p1', 'AAPL', 2, 50);

		const [url] = lastCall(fetch);
		expect(url).toContain('/portfolios/p1/assets/AAPL/transactions?page=2&limit=50');
	});
});

describe('platforms module', () => {
	it('deleteSource issues a DELETE to /portfolios/sources/:id', async () => {
		const { event, fetch } = authedEvent(jsonResponse({ success: true }));

		await platforms.deleteSource(event, 's1');

		const [url, init] = lastCall(fetch);
		expect(url).toContain('/portfolios/sources/s1');
		expect(init.method).toBe('DELETE');
	});
});

describe('market module', () => {
	it('searchAssets encodes the search term and returns the raw Response', async () => {
		const { event, fetch } = authedEvent(jsonResponse({ success: true, data: [] }));

		const res = await market.searchAssets(event, { search: 'ap pl', limit: '5' });

		const [url] = lastCall(fetch);
		expect(url).toContain('/portfolios/assets?search=ap%20pl&page=1&limit=5');
		expect(res).toBeInstanceOf(Response);
	});
});

describe('user module', () => {
	it('banUser sends a PATCH to /users/:id/ban with the ban flag', async () => {
		const { event, fetch } = authedEvent(jsonResponse({ success: true }));

		await user.banUser(event, 'u1', { ban: true });

		const [url, init] = lastCall(fetch);
		expect(url).toContain('/users/u1/ban');
		expect(init.method).toBe('PATCH');
		expect(init.body).toBe(JSON.stringify({ ban: true }));
	});
});

describe('auth module (public)', () => {
	it('login POSTs the credentials to /auth/login without a Bearer header', async () => {
		const fetch = vi.fn().mockResolvedValue(jsonResponse({ success: true }));

		const res = await auth.login(fetch, { email: 'a@b.co', password: 'supersecret' });

		const [url, init] = lastCall(fetch);
		expect(url).toContain('/auth/login');
		expect(init.method).toBe('POST');
		expect((init.headers as Record<string, string>).Authorization).toBeUndefined();
		expect(res).toBeInstanceOf(Response);
	});
});
