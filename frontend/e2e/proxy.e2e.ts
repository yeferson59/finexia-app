import { expect, test } from '@playwright/test';

// The app is the backend's only door to the outside (docs/API.md §1.6). This
// runs against the production build, where the CSRF check is live: the unit
// specs cover the proxy and the check one at a time, and this is where their
// order in `hooks.server.ts` — and `svelte.config.js` switching SvelteKit's own
// check off — is seen to work.
test.describe('backend proxy', () => {
	test('forwards an OAuth token request posted from another origin', async ({ request }) => {
		// RFC 6749 §4.1.3: the client's server posts this form, never a page of
		// this app, so a foreign Origin is the normal case and not an attack.
		const res = await request.post('/oauth/token', {
			form: { grant_type: 'authorization_code', code: 'e2e' },
			headers: { origin: 'https://claude.ai', cookie: 'access_token_finexia=secret' }
		});

		expect(res.status()).toBe(200);
		const body = await res.json();
		expect(body.access_token).toBe('fnx_oat_e2e');
		expect(body.received).toEqual({ grantType: 'authorization_code', cookie: null });
	});

	test('still refuses a cross-site form aimed at the app itself', async ({ request }) => {
		const res = await request.post('/auth?/login', {
			form: { email: 'user@finexia.test', password: 'x' },
			headers: { origin: 'https://evil.example' }
		});

		expect(res.status()).toBe(403);
		expect(await res.text()).toBe('Cross-site POST form submissions are forbidden');
	});

	test('serves avatars from the backend under the app origin', async ({ request }) => {
		const res = await request.get('/users/0b7f9c7e-1111-4222-8333-444455556666/avatar');

		expect(res.status()).toBe(200);
		expect(res.headers()['content-type']).toBe('image/png');
	});

	test('keeps the rest of the API off the outside', async ({ request }) => {
		// The stub would answer this one with a 401 envelope, so a 404 from the
		// app means the request never left it.
		const res = await request.get('/auth/session');

		expect(res.status()).toBe(404);
	});
});
