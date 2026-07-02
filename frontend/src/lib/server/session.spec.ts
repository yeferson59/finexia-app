import { describe, it, expect } from 'vitest';
import { parseRefreshSetCookie } from './session';

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
