import { describe, it, expect } from 'vitest';
import { isRedirect } from '@sveltejs/kit';
import { load } from './+layout.server';

type LoadEvent = Parameters<typeof load>[0];

const user = { name: 'Jane' } as App.Locals['user'];
const session = { id: 'sess-1' } as App.Locals['session'];

/**
 * Cookie jar mínimo: solo `get` y `delete`, que es todo lo que toca este load a
 * través de `takeReturnTo`. `deleted` se expone porque el destino de vuelta es
 * de un solo uso, y eso es lo que hay que poder comprobar.
 */
function cookieJar(initial: Record<string, string> = {}) {
	const store = { ...initial };
	const deleted: string[] = [];

	return {
		deleted,
		cookies: {
			get: (name: string) => store[name],
			delete: (name: string) => {
				delete store[name];
				deleted.push(name);
			}
		}
	};
}

function loadWith(locals: Partial<App.Locals>, initialCookies: Record<string, string> = {}) {
	const jar = cookieJar(initialCookies);
	const result = load({
		locals: { user: null, session: null, ...locals },
		cookies: jar.cookies
	} as unknown as LoadEvent);

	return { result, jar };
}

async function redirectFrom(promise: unknown) {
	try {
		await promise;
	} catch (e) {
		return e;
	}

	return undefined;
}

describe('auth layout load', () => {
	it('redirects an already-authenticated user to /dashboard', async () => {
		const { result } = loadWith({ user, session });
		const thrown = await redirectFrom(result);

		expect(isRedirect(thrown)).toBe(true);
		expect((thrown as { status: number }).status).toBe(303);
		expect((thrown as { location: string }).location).toBe('/dashboard');
	});

	// El caso que existe por la pantalla de consentimiento OAuth: se llega a
	// ella sin sesión, se pasa por el login, y el usuario tiene que volver a la
	// petición que estaba autorizando en vez de aterrizar en el dashboard.
	it('honours a pending return target and consumes it', async () => {
		const target = '/oauth/consent?request=abc';
		const { result, jar } = loadWith({ user, session }, { finexia_return_to: target });
		const thrown = await redirectFrom(result);

		expect((thrown as { location: string }).location).toBe(target);
		expect(jar.deleted).toContain('finexia_return_to');
	});

	// Un destino que salga del sitio es un redirect abierto encadenado a un
	// login legítimo, que es exactamente la forma en que se explotan.
	it('ignores a non-local return target', async () => {
		const { result } = loadWith({ user, session }, { finexia_return_to: '//evil.test/steal' });
		const thrown = await redirectFrom(result);

		expect((thrown as { location: string }).location).toBe('/dashboard');
	});

	it('does nothing when there is no session', async () => {
		await expect(loadWith({ user: null, session: null }).result).resolves.toBeUndefined();
	});

	it('does not redirect when only a user (but no session) is present', async () => {
		await expect(loadWith({ user, session: null }).result).resolves.toBeUndefined();
	});
});
