import { describe, it, expect } from 'vitest';
import { isRedirect } from '@sveltejs/kit';
import { load } from './+layout.server';

type LoadEvent = Parameters<typeof load>[0];

const user = { name: 'Jane' } as App.Locals['user'];
const session = { id: 'sess-1' } as App.Locals['session'];

function loadWith(locals: Partial<App.Locals>) {
	return load({ locals: { user: null, session: null, ...locals } } as LoadEvent);
}

describe('auth layout load', () => {
	it('redirects an already-authenticated user to /dashboard', async () => {
		let thrown: unknown;
		try {
			await loadWith({ user, session });
		} catch (e) {
			thrown = e;
		}

		expect(isRedirect(thrown)).toBe(true);
		expect((thrown as { status: number }).status).toBe(303);
		expect((thrown as { location: string }).location).toBe('/dashboard');
	});

	it('does nothing when there is no session', async () => {
		await expect(loadWith({ user: null, session: null })).resolves.toBeUndefined();
	});

	it('does not redirect when only a user (but no session) is present', async () => {
		await expect(loadWith({ user, session: null })).resolves.toBeUndefined();
	});
});
