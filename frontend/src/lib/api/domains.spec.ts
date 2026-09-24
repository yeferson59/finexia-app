import { describe, it, expect, vi } from 'vitest';
import { ACCESS_COOKIE } from '$lib/server/session';
import { createMockCookies, jsonResponse } from '$lib/server/testing';
import * as portfolio from './portfolio';
import * as transactions from './transactions';
import * as platforms from './platforms';
import * as market from './market';
import * as user from './user';
import * as auth from './auth';
import * as support from './support';
import * as funds from './funds';

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

describe('funds module', () => {
	it('saveMark POSTs the mark to /portfolios/funds/:id/marks', async () => {
		const { event, fetch } = authedEvent(jsonResponse({ success: true }));

		await funds.saveMark(event, 'f1', { date: '2026-09-30T00:00:00Z', unitValue: 12431.22 });

		const [url, init] = lastCall(fetch);
		expect(url).toContain('/portfolios/funds/f1/marks');
		expect(init.method).toBe('POST');
		expect(init.body).toBe(JSON.stringify({ date: '2026-09-30T00:00:00Z', unitValue: 12431.22 }));
	});

	it('deleteMark names the day in the path', async () => {
		const { event, fetch } = authedEvent(jsonResponse({ success: true }));

		await funds.deleteMark(event, 'f1', '2026-09-30');

		const [url, init] = lastCall(fetch);
		expect(url).toContain('/portfolios/funds/f1/marks/2026-09-30');
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

	it('createMCPToken POSTs the name and the lifetime to /auth/mcp-tokens', async () => {
		const { event, fetch } = authedEvent(
			jsonResponse({ success: true, data: { id: 't1', token: 'fnx_mcp_x' } })
		);

		const res = await user.createMCPToken(event, { name: 'Claude Desktop', expiresInDays: 90 });

		const [url, init] = lastCall(fetch);
		expect(url).toContain('/auth/mcp-tokens');
		expect(init.method).toBe('POST');
		expect(init.body).toBe(JSON.stringify({ name: 'Claude Desktop', expiresInDays: 90 }));
		// El secreto solo llega aquí: es la respuesta de crear, y nada lo guarda.
		expect((res.data as { token: string } | null)?.token).toBe('fnx_mcp_x');
	});

	it('rotateMCPToken POSTs to /auth/mcp-tokens/:id/rotate', async () => {
		const { event, fetch } = authedEvent(jsonResponse({ success: true, data: {} }));

		await user.rotateMCPToken(event, 't1', { expiresInDays: 0 });

		const [url, init] = lastCall(fetch);
		expect(url).toContain('/auth/mcp-tokens/t1/rotate');
		expect(init.method).toBe('POST');
		expect(init.body).toBe(JSON.stringify({ expiresInDays: 0 }));
	});

	it('deleteMCPToken sends a DELETE to /auth/mcp-tokens/:id', async () => {
		const { event, fetch } = authedEvent(jsonResponse({ success: true }));

		await user.deleteMCPToken(event, 't1');

		const [url, init] = lastCall(fetch);
		expect(url).toContain('/auth/mcp-tokens/t1');
		expect(init.method).toBe('DELETE');
	});

	it('getMCPTokens reads /auth/mcp-tokens with the Bearer header', async () => {
		const { event, fetch } = authedEvent(jsonResponse({ success: true, data: [] }));

		await user.getMCPTokens(event);

		const [url, init] = lastCall(fetch);
		expect(url).toContain('/auth/mcp-tokens');
		expect((init.headers as Record<string, string>).Authorization).toBe('Bearer access-1');
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

describe('support module (public)', () => {
	it('createSupportCheckout POSTs the amount without any session header', async () => {
		const fetch = vi
			.fn()
			.mockResolvedValue(
				jsonResponse({ success: true, data: { orderId: 'FNX-APOYO-1' } }, { status: 201 })
			);

		const res = await support.createSupportCheckout(fetch, 20000);

		const [url, init] = lastCall(fetch);
		expect(url).toMatch(/\/support\/checkouts$/);
		expect(init.method).toBe('POST');
		expect(init.body).toBe(JSON.stringify({ amount: 20000 }));
		expect((init.headers as Record<string, string>).Authorization).toBeUndefined();
		expect(res.ok).toBe(true);
		expect(res.data).toEqual({ orderId: 'FNX-APOYO-1' });
	});

	it('getSupportContribution encodes the order id into the path', async () => {
		const fetch = vi.fn().mockResolvedValue(jsonResponse({ success: true, data: {} }));

		await support.getSupportContribution(fetch, 'FNX-APOYO-1/../x');

		expect(lastCall(fetch)[0]).toMatch(/\/support\/checkouts\/FNX-APOYO-1%2F\.\.%2Fx$/);
	});

	it('degrades to ok:false with status 0 when the backend is unreachable', async () => {
		const fetch = vi.fn().mockRejectedValue(new TypeError('fetch failed'));

		const res = await support.getSupportConfig(fetch);

		expect(res).toMatchObject({ ok: false, status: 0, data: null });
	});
});
