import { describe, it, expect, vi } from 'vitest';
import { isRedirect } from '@sveltejs/kit';
import { authedFetch, authedFetchSafe } from './api';
import { ACCESS_COOKIE, REFRESH_COOKIE } from './session';
import { createMockCookies, jsonResponse } from './testing';

/** A refresh response the backend would send: new access token + rotated cookie. */
function refreshOk(accessToken = 'refreshed-access') {
	return jsonResponse(
		{ success: true, data: { accessToken } },
		{ setCookie: `refresh_token=rotated; Max-Age=2592000; Path=/; HttpOnly` }
	);
}

describe('authedFetch', () => {
	it('attaches the access token as a Bearer header and returns the response', async () => {
		const cookies = createMockCookies({ [ACCESS_COOKIE]: 'access-1' });
		const fetch = vi.fn().mockResolvedValue(jsonResponse({ ok: true }));

		const res = await authedFetch({ cookies, fetch }, '/portfolios');

		expect(res.status).toBe(200);
		const [, init] = fetch.mock.calls[0];
		expect((init.headers as Record<string, string>).Authorization).toBe('Bearer access-1');
	});

	it('refreshes up front when only a refresh token is present, then calls the backend', async () => {
		const cookies = createMockCookies({ [REFRESH_COOKIE]: 'refresh-1' });
		const fetch = vi
			.fn()
			.mockResolvedValueOnce(refreshOk('minted-access')) // /auth/refresh
			.mockResolvedValueOnce(jsonResponse({ ok: true })); // actual request

		const res = await authedFetch({ cookies, fetch }, '/portfolios');

		expect(res.status).toBe(200);
		expect(cookies.get(ACCESS_COOKIE)).toBe('minted-access');
		const [, init] = fetch.mock.calls[1];
		expect((init.headers as Record<string, string>).Authorization).toBe('Bearer minted-access');
	});

	it('clears cookies and redirects to /auth when there is no session at all', async () => {
		const cookies = createMockCookies();
		const fetch = vi.fn();

		let thrown: unknown;
		try {
			await authedFetch({ cookies, fetch }, '/portfolios');
		} catch (e) {
			thrown = e;
		}

		expect(isRedirect(thrown)).toBe(true);
		expect((thrown as { location: string }).location).toBe('/auth');
		expect(cookies.deleteCalls).toEqual([ACCESS_COOKIE, REFRESH_COOKIE]);
		expect(fetch).not.toHaveBeenCalled();
	});

	it('retries once with a fresh token after a 401 and returns the retried response', async () => {
		const cookies = createMockCookies({
			[ACCESS_COOKIE]: 'stale',
			[REFRESH_COOKIE]: 'refresh-1'
		});
		const fetch = vi
			.fn()
			.mockResolvedValueOnce(jsonResponse({ error: 'expired' }, { status: 401 })) // first call
			.mockResolvedValueOnce(refreshOk('fresh-access')) // /auth/refresh
			.mockResolvedValueOnce(jsonResponse({ ok: true })); // retry

		const res = await authedFetch({ cookies, fetch }, '/portfolios');

		expect(res.status).toBe(200);
		expect(fetch).toHaveBeenCalledTimes(3);
		const [, retryInit] = fetch.mock.calls[2];
		expect((retryInit.headers as Record<string, string>).Authorization).toBe('Bearer fresh-access');
	});

	it('redirects to /auth when the refresh token is also rejected on a 401', async () => {
		const cookies = createMockCookies({
			[ACCESS_COOKIE]: 'stale',
			[REFRESH_COOKIE]: 'bad-refresh'
		});
		const fetch = vi
			.fn()
			.mockResolvedValueOnce(jsonResponse({ error: 'expired' }, { status: 401 }))
			.mockResolvedValueOnce(jsonResponse({ success: false }, { status: 401 })); // refresh rejected

		let thrown: unknown;
		try {
			await authedFetch({ cookies, fetch }, '/portfolios');
		} catch (e) {
			thrown = e;
		}

		expect(isRedirect(thrown)).toBe(true);
		expect((thrown as { location: string }).location).toBe('/auth');
		expect(cookies.deleteCalls).toEqual([ACCESS_COOKIE, REFRESH_COOKIE]);
	});

	it('forwards custom headers and the request method to the backend', async () => {
		const cookies = createMockCookies({ [ACCESS_COOKIE]: 'access-1' });
		const fetch = vi.fn().mockResolvedValue(jsonResponse({ ok: true }));

		await authedFetch({ cookies, fetch }, '/portfolios', {
			method: 'POST',
			headers: { 'X-Test': 'yes' }
		});

		const [, init] = fetch.mock.calls[0];
		expect(init.method).toBe('POST');
		expect((init.headers as Record<string, string>)['X-Test']).toBe('yes');
		expect((init.headers as Record<string, string>).Authorization).toBe('Bearer access-1');
	});
});

describe('authedFetchSafe', () => {
	it('returns null instead of throwing when the backend fetch errors', async () => {
		const cookies = createMockCookies({ [ACCESS_COOKIE]: 'access-1' });
		const fetch = vi.fn().mockRejectedValue(new Error('ECONNREFUSED'));

		const res = await authedFetchSafe({ cookies, fetch }, '/dashboard');

		expect(res).toBeNull();
	});

	it('re-throws redirects so an expired session still bounces to /auth', async () => {
		const cookies = createMockCookies();
		const fetch = vi.fn();

		let thrown: unknown;
		try {
			await authedFetchSafe({ cookies, fetch }, '/dashboard');
		} catch (e) {
			thrown = e;
		}

		expect(isRedirect(thrown)).toBe(true);
		expect((thrown as { location: string }).location).toBe('/auth');
	});

	it('passes through a successful response unchanged', async () => {
		const cookies = createMockCookies({ [ACCESS_COOKIE]: 'access-1' });
		const fetch = vi.fn().mockResolvedValue(jsonResponse({ ok: true }));

		const res = await authedFetchSafe({ cookies, fetch }, '/dashboard');

		expect(res?.status).toBe(200);
	});
});
