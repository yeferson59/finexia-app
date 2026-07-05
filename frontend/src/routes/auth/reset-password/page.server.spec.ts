import { describe, it, expect, vi } from 'vitest';
import { isRedirect } from '@sveltejs/kit';
import { load, actions } from './+page.server';
import { jsonResponse } from '$lib/server/testing';

type LoadEvent = Parameters<typeof load>[0];
type ConfirmEvent = Parameters<typeof actions.confirm>[0];

function buildLoadEvent(url: string, fetch: ReturnType<typeof vi.fn>) {
	return { url: new URL(url), fetch } as unknown;
}

function buildConfirmEvent(fields: Record<string, string>, fetch: ReturnType<typeof vi.fn>) {
	const formData = new FormData();
	for (const [key, value] of Object.entries(fields)) formData.append(key, value);

	return {
		request: { formData: async () => formData },
		fetch
	} as unknown;
}

describe('reset-password load', () => {
	it('reports invalid when no token is present', async () => {
		const fetch = vi.fn();
		const result = await load(
			buildLoadEvent('https://app.test/auth/reset-password', fetch) as LoadEvent
		);

		expect(result).toEqual({ valid: false, reason: 'Falta el token de recuperación.' });
		expect(fetch).not.toHaveBeenCalled();
	});

	it('reports expired when the backend returns 410', async () => {
		const fetch = vi.fn().mockResolvedValue(jsonResponse({ success: false }, { status: 410 }));
		const result = await load(
			buildLoadEvent('https://app.test/auth/reset-password?token=abc', fetch) as LoadEvent
		);

		expect(result).toEqual({
			valid: false,
			reason: 'Este enlace ha expirado. Solicita uno nuevo.'
		});
	});

	it('reports invalid for any other backend failure', async () => {
		const fetch = vi.fn().mockResolvedValue(jsonResponse({ success: false }, { status: 400 }));
		const result = await load(
			buildLoadEvent('https://app.test/auth/reset-password?token=abc', fetch) as LoadEvent
		);

		expect(result).toEqual({
			valid: false,
			reason: 'Este enlace no es válido o ya fue utilizado.'
		});
	});

	it('returns the token when the backend confirms it is valid', async () => {
		const fetch = vi.fn().mockResolvedValue(jsonResponse({ success: true }));
		const result = await load(
			buildLoadEvent('https://app.test/auth/reset-password?token=abc', fetch) as LoadEvent
		);

		expect(result).toEqual({ valid: true, token: 'abc' });
	});
});

describe('reset-password confirm action', () => {
	const validFields = { token: 'abc', password: 'supersecret', confirmPassword: 'supersecret' };

	it('fails validation for a short password without hitting the backend', async () => {
		const fetch = vi.fn();
		const result = await actions.confirm(
			buildConfirmEvent(
				{ ...validFields, password: 'short', confirmPassword: 'short' },
				fetch
			) as ConfirmEvent
		);

		expect(result?.status).toBe(400);
		expect(fetch).not.toHaveBeenCalled();
	});

	it('fails when the passwords do not match', async () => {
		const fetch = vi.fn();
		const result = await actions.confirm(
			buildConfirmEvent({ ...validFields, confirmPassword: 'different1' }, fetch) as ConfirmEvent
		);

		expect(result?.status).toBe(400);
		expect(result?.data.errors).toEqual({ confirmPassword: 'Las contraseñas no coinciden' });
		expect(fetch).not.toHaveBeenCalled();
	});

	it('surfaces the backend error when the token is rejected', async () => {
		const fetch = vi
			.fn()
			.mockResolvedValue(jsonResponse({ details: 'invalid password reset link' }, { status: 400 }));

		const result = await actions.confirm(buildConfirmEvent(validFields, fetch) as ConfirmEvent);

		expect(result?.status).toBe(400);
		expect(result?.data.errors).toEqual({ server: 'invalid password reset link' });
	});

	it('redirects to /auth on success', async () => {
		const fetch = vi.fn().mockResolvedValue(jsonResponse({ success: true }));

		let thrown: unknown;
		try {
			await actions.confirm(buildConfirmEvent(validFields, fetch) as ConfirmEvent);
		} catch (e) {
			thrown = e;
		}

		expect(isRedirect(thrown)).toBe(true);
		expect((thrown as { status: number }).status).toBe(303);
		expect((thrown as { location: string }).location).toBe('/auth?reset=1');
	});
});
