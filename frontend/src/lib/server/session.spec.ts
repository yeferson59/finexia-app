import { describe, it, expect, vi } from 'vitest';
import {
	ACCESS_COOKIE,
	REFRESH_COOKIE,
	clearSessionCookies,
	parseRefreshSetCookie,
	refreshAccessToken,
	setAccessCookie,
	setRefreshCookie
} from './session';
import { createMockCookies, jsonResponse } from './testing';

function responseWithSetCookie(...cookies: string[]): Response {
	const headers = new Headers();
	for (const cookie of cookies) {
		headers.append('set-cookie', cookie);
	}
	return new Response(null, { headers });
}

describe('parseRefreshSetCookie', () => {
	it('extracts the token and Max-Age from a backend cookie', () => {
		const res = responseWithSetCookie(
			'refresh_token=abc-123_XYZ=; max-age=2592000; path=/; HttpOnly; SameSite=Strict'
		);

		expect(parseRefreshSetCookie(res)).toEqual({ value: 'abc-123_XYZ=', maxAge: 2592000 });
	});

	it('returns null Max-Age when the attribute is missing', () => {
		const res = responseWithSetCookie('refresh_token=abc123; path=/; HttpOnly');

		expect(parseRefreshSetCookie(res)).toEqual({ value: 'abc123', maxAge: null });
	});

	it('finds the refresh token among multiple Set-Cookie headers', () => {
		const res = responseWithSetCookie(
			'other_cookie=zzz; path=/',
			'refresh_token=abc123; Max-Age=86400; path=/'
		);

		expect(parseRefreshSetCookie(res)).toEqual({ value: 'abc123', maxAge: 86400 });
	});

	it('returns null when no refresh token cookie is present', () => {
		const res = responseWithSetCookie('other_cookie=zzz; path=/');

		expect(parseRefreshSetCookie(res)).toBeNull();
	});

	it('does not match cookies whose name merely ends in refresh_token', () => {
		const res = responseWithSetCookie('not_refresh_token=zzz; path=/');

		expect(parseRefreshSetCookie(res)).toBeNull();
	});
});

describe('setAccessCookie', () => {
	it('writes the access token as an httpOnly, lax, rooted cookie', () => {
		const cookies = createMockCookies();

		setAccessCookie(cookies, 'access-abc');

		expect(cookies.get(ACCESS_COOKIE)).toBe('access-abc');
		const [call] = cookies.setCalls;
		expect(call.name).toBe(ACCESS_COOKIE);
		expect(call.opts).toMatchObject({
			path: '/',
			httpOnly: true,
			sameSite: 'lax',
			maxAge: 60 * 60 * 24 * 7
		});
	});
});

describe('setRefreshCookie', () => {
	it('uses the provided Max-Age when given', () => {
		const cookies = createMockCookies();

		setRefreshCookie(cookies, 'refresh-xyz', 12345);

		expect(cookies.get(REFRESH_COOKIE)).toBe('refresh-xyz');
		expect(cookies.setCalls[0].opts).toMatchObject({ maxAge: 12345, httpOnly: true, path: '/' });
	});

	it('falls back to the 30-day default when Max-Age is null', () => {
		const cookies = createMockCookies();

		setRefreshCookie(cookies, 'refresh-xyz', null);

		expect(cookies.setCalls[0].opts.maxAge).toBe(60 * 60 * 24 * 30);
	});
});

describe('clearSessionCookies', () => {
	it('deletes both the access and refresh cookies', () => {
		const cookies = createMockCookies({
			[ACCESS_COOKIE]: 'a',
			[REFRESH_COOKIE]: 'r'
		});

		clearSessionCookies(cookies);

		expect(cookies.get(ACCESS_COOKIE)).toBeUndefined();
		expect(cookies.get(REFRESH_COOKIE)).toBeUndefined();
		expect(cookies.deleteCalls).toEqual([ACCESS_COOKIE, REFRESH_COOKIE]);
	});
});

describe('refreshAccessToken', () => {
	it('exchanges a refresh token and updates both cookies from the response', async () => {
		const cookies = createMockCookies();
		const fetch = vi
			.fn()
			.mockResolvedValue(
				jsonResponse(
					{ success: true, data: { accessToken: 'new-access' } },
					{ setCookie: 'refresh_token=rotated-token; Max-Age=2592000; Path=/; HttpOnly' }
				)
			);

		const token = await refreshAccessToken({ cookies, fetch }, 'old-refresh');

		expect(token).toBe('new-access');
		expect(cookies.get(ACCESS_COOKIE)).toBe('new-access');
		expect(cookies.get(REFRESH_COOKIE)).toBe('rotated-token');
		expect(fetch).toHaveBeenCalledOnce();
	});

	it('sets the access cookie but leaves refresh untouched when no rotation is returned', async () => {
		const cookies = createMockCookies({ [REFRESH_COOKIE]: 'keep-me' });
		const fetch = vi
			.fn()
			.mockResolvedValue(jsonResponse({ success: true, data: { accessToken: 'new-access' } }));

		const token = await refreshAccessToken({ cookies, fetch }, 'keep-me');

		expect(token).toBe('new-access');
		expect(cookies.get(ACCESS_COOKIE)).toBe('new-access');
		expect(cookies.get(REFRESH_COOKIE)).toBe('keep-me');
	});

	it('returns null when the backend rejects the refresh token (non-5xx)', async () => {
		const cookies = createMockCookies();
		const fetch = vi.fn().mockResolvedValue(jsonResponse({ success: false }, { status: 401 }));

		const token = await refreshAccessToken({ cookies, fetch }, 'bad-refresh');

		expect(token).toBeNull();
		expect(cookies.get(ACCESS_COOKIE)).toBeUndefined();
	});

	it('returns null when the response is ok but carries no access token', async () => {
		const cookies = createMockCookies();
		const fetch = vi.fn().mockResolvedValue(jsonResponse({ success: true, data: {} }));

		expect(await refreshAccessToken({ cookies, fetch }, 'refresh')).toBeNull();
	});

	it('throws on a 5xx so a transient outage is not treated as an invalid session', async () => {
		const cookies = createMockCookies();
		const fetch = vi.fn().mockResolvedValue(jsonResponse({}, { status: 503 }));

		await expect(refreshAccessToken({ cookies, fetch }, 'refresh')).rejects.toThrow(/status 503/);
	});

	it('de-dupes concurrent refreshes that share the same token (single-flight)', async () => {
		const cookies = createMockCookies();
		const fetch = vi
			.fn()
			.mockResolvedValue(jsonResponse({ success: true, data: { accessToken: 'shared' } }));

		const [a, b] = await Promise.all([
			refreshAccessToken({ cookies, fetch }, 'same-token'),
			refreshAccessToken({ cookies, fetch }, 'same-token')
		]);

		expect(a).toBe('shared');
		expect(b).toBe('shared');
		expect(fetch).toHaveBeenCalledOnce();
	});
});
