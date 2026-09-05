import { describe, it, expect, vi, beforeEach } from 'vitest';
import { isRedirect, isHttpError } from '@sveltejs/kit';
import { actions, load } from './+page.server';
import { createMockCookies } from '$lib/server/testing';
import * as user from '$lib/api/user';

vi.mock('$lib/api/user', () => ({
	getOAuthConsent: vi.fn(),
	decideOAuthConsent: vi.fn()
}));

const REQUEST_ID = '0b7f9c7e-1111-4222-8333-444455556666';

const consent = {
	requestId: REQUEST_ID,
	clientName: 'Claude',
	redirectUri: 'https://claude.ai/api/mcp/auth_callback',
	scopes: ['mcp:read'],
	expiresAt: '2026-09-05T00:00:00Z'
};

const session = { id: 'sess-1' } as App.Locals['session'];
const account = { name: 'Jane' } as App.Locals['user'];

beforeEach(() => vi.clearAllMocks());

function loadEvent(url: string, locals: Partial<App.Locals> = { session, user: account }) {
	return {
		url: new URL(url),
		locals: { user: null, session: null, ...locals },
		cookies: createMockCookies(),
		fetch: vi.fn()
	} as unknown as Parameters<typeof load>[0];
}

/**
 * La acción se invoca desde `action="?/decide"`, que **reemplaza la query
 * entera**: la URL del POST no lleva el `?request=…` con el que se cargó la
 * página. Por eso el evento se construye aquí con esa URL pelada — es la que
 * llega de verdad, y montarla con el parámetro haría pasar un test que la
 * realidad suspende.
 */
function decideEvent(fields: Record<string, string>) {
	const formData = new FormData();
	for (const [key, value] of Object.entries(fields)) formData.append(key, value);

	return {
		url: new URL('https://finexia.test/oauth/consent?/decide'),
		request: { formData: async () => formData },
		cookies: createMockCookies(),
		fetch: vi.fn()
	} as unknown as Parameters<typeof actions.decide>[0];
}

async function thrownBy(promise: unknown) {
	try {
		await promise;
	} catch (e) {
		return e;
	}

	return undefined;
}

describe('oauth consent load', () => {
	it('devuelve la petición para la pantalla', async () => {
		vi.mocked(user.getOAuthConsent).mockResolvedValue({
			ok: true,
			status: 200,
			success: true,
			data: consent
		});

		const result = await load(
			loadEvent(`https://finexia.test/oauth/consent?request=${REQUEST_ID}`)
		);

		expect(result).toEqual({ consent, requestId: REQUEST_ID });
	});

	it('manda a iniciar sesión y anota a dónde volver', async () => {
		const event = loadEvent(`https://finexia.test/oauth/consent?request=${REQUEST_ID}`, {
			session: null,
			user: null
		});

		const thrown = await thrownBy(load(event));

		expect(isRedirect(thrown)).toBe(true);
		expect((thrown as { location: string }).location).toBe('/auth');
		// Sin esto el login aterriza en /dashboard y la petición se pierde: el
		// cliente se queda esperando un código que ya nunca llega.
		expect(event.cookies.get('finexia_return_to')).toBe(`/oauth/consent?request=${REQUEST_ID}`);
	});

	it('sin id no hay nada que aprobar', async () => {
		const thrown = await thrownBy(load(loadEvent('https://finexia.test/oauth/consent')));

		expect(isHttpError(thrown)).toBe(true);
		expect((thrown as { status: number }).status).toBe(400);
	});
});

describe('oauth consent decide', () => {
	it('aprueba y redirige a donde diga el backend', async () => {
		const redirectTo = 'https://claude.ai/api/mcp/auth_callback?code=abc&state=xyz';
		vi.mocked(user.decideOAuthConsent).mockResolvedValue({
			ok: true,
			status: 200,
			success: true,
			data: { redirectTo }
		});

		const thrown = await thrownBy(
			actions.decide(decideEvent({ request: REQUEST_ID, decision: 'approve' }))
		);

		expect(user.decideOAuthConsent).toHaveBeenCalledWith(expect.anything(), REQUEST_ID, true);
		expect(isRedirect(thrown)).toBe(true);
		expect((thrown as { location: string }).location).toBe(redirectTo);
	});

	it('cualquier decisión que no sea "approve" es denegar', async () => {
		vi.mocked(user.decideOAuthConsent).mockResolvedValue({
			ok: true,
			status: 200,
			success: true,
			data: { redirectTo: 'https://claude.ai/cb?error=access_denied' }
		});

		await thrownBy(actions.decide(decideEvent({ request: REQUEST_ID, decision: 'deny' })));

		expect(user.decideOAuthConsent).toHaveBeenCalledWith(expect.anything(), REQUEST_ID, false);
	});

	it('sin id en el cuerpo responde 400 y no llama al backend', async () => {
		const thrown = await thrownBy(actions.decide(decideEvent({ decision: 'approve' })));

		expect(isHttpError(thrown)).toBe(true);
		expect((thrown as { status: number }).status).toBe(400);
		expect(user.decideOAuthConsent).not.toHaveBeenCalled();
	});

	it('una petición caducada es un 404, no un 500', async () => {
		vi.mocked(user.decideOAuthConsent).mockResolvedValue({
			ok: false,
			status: 404,
			success: false,
			data: null
		});

		const thrown = await thrownBy(
			actions.decide(decideEvent({ request: REQUEST_ID, decision: 'approve' }))
		);

		expect((thrown as { status: number }).status).toBe(404);
	});
});
