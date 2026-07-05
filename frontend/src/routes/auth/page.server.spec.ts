import { describe, it, expect, vi, afterEach } from 'vitest';
import { isRedirect } from '@sveltejs/kit';
import { actions, load } from './+page.server';
import { ACCESS_COOKIE, REFRESH_COOKIE } from '$lib/server/session';
import { createMockCookies, jsonResponse, type MockCookies } from '$lib/server/testing';
import { features } from '$/config/features';

vi.mock('$/config/features', () => ({ features: { selfRegistration: false } }));

type LoginEvent = Parameters<typeof actions.login>[0];
type RegisterEvent = Parameters<typeof actions.register>[0];

afterEach(() => vi.unstubAllGlobals());

/**
 * Builds the slice of a RequestEvent the auth actions read. `login` uses the
 * event's `fetch`, while `register` uses the global `fetch`, so the same mock is
 * wired into both to keep the tests uniform.
 */
function buildEvent(
	fields: Record<string, string>,
	fetch: ReturnType<typeof vi.fn>,
	cookies: MockCookies = createMockCookies()
) {
	const formData = new FormData();
	for (const [key, value] of Object.entries(fields)) formData.append(key, value);

	vi.stubGlobal('fetch', fetch);

	return {
		request: { formData: async () => formData },
		cookies,
		fetch
	} as unknown;
}

const validLogin = { email: 'user@finexia.me', password: 'supersecret' };
const validRegister = {
	name: 'Jane Doe',
	email: 'jane@finexia.me',
	password: 'supersecret',
	confirmPassword: 'supersecret',
	terms: 'on'
};

describe('load', () => {
	it('exposes the selfRegistration feature flag to the page', () => {
		expect(load({} as Parameters<typeof load>[0])).toEqual({
			selfRegistrationEnabled: features.selfRegistration
		});
	});
});

describe('login action', () => {
	it('fails validation for a bad email without hitting the backend', async () => {
		const fetch = vi.fn();
		const result = await actions.login(
			buildEvent({ email: 'nope', password: 'supersecret' }, fetch) as LoginEvent
		);

		expect(result?.status).toBe(400);
		expect(result?.data.type).toBe('login');
		expect(Array.isArray(result?.data.errors)).toBe(true);
		expect(fetch).not.toHaveBeenCalled();
	});

	it('fails validation for a short password', async () => {
		const fetch = vi.fn();
		const result = await actions.login(
			buildEvent({ email: 'user@finexia.me', password: 'short' }, fetch) as LoginEvent
		);

		expect(result?.status).toBe(400);
		expect(fetch).not.toHaveBeenCalled();
	});

	it('surfaces the backend error message when credentials are rejected', async () => {
		const fetch = vi
			.fn()
			.mockResolvedValue(jsonResponse({ message: 'Credenciales incorrectas' }, { status: 401 }));

		const result = await actions.login(buildEvent(validLogin, fetch) as LoginEvent);

		expect(result?.status).toBe(401);
		expect(result?.data.type).toBe('login');
		expect(result?.data.errors).toEqual({ server: 'Credenciales incorrectas' });
	});

	it('fails when the backend responds ok but success is false', async () => {
		const fetch = vi
			.fn()
			.mockResolvedValue(jsonResponse({ success: false, message: 'Cuenta bloqueada' }));

		const result = await actions.login(buildEvent(validLogin, fetch) as LoginEvent);

		expect(result?.status).toBe(400);
		expect(result?.data.errors).toEqual({ server: 'Cuenta bloqueada' });
	});

	it('flags an unverified account so the UI can offer to resend the link', async () => {
		const fetch = vi.fn().mockResolvedValue(
			jsonResponse(
				{ message: 'email not verified', action: 'auth:login:unverified' },
				{ status: 403 }
			)
		);

		const result = await actions.login(buildEvent(validLogin, fetch) as LoginEvent);
		const data = result?.data as { unverified?: boolean; errors?: Record<string, string> };

		expect(result?.status).toBe(403);
		expect(data.unverified).toBe(true);
		expect(data.errors).toEqual({
			server: 'Debes verificar tu correo antes de iniciar sesión.'
		});
	});

	it('sets session cookies and redirects to /dashboard on success', async () => {
		const cookies = createMockCookies();
		const fetch = vi
			.fn()
			.mockResolvedValue(
				jsonResponse(
					{ success: true, data: { accessToken: 'access-token' } },
					{ setCookie: 'refresh_token=refresh-token; Max-Age=2592000; Path=/; HttpOnly' }
				)
			);

		let thrown: unknown;
		try {
			await actions.login(buildEvent(validLogin, fetch, cookies) as LoginEvent);
		} catch (e) {
			thrown = e;
		}

		expect(isRedirect(thrown)).toBe(true);
		expect((thrown as { status: number }).status).toBe(302);
		expect((thrown as { location: string }).location).toBe('/dashboard');
		expect(cookies.get(ACCESS_COOKIE)).toBe('access-token');
		expect(cookies.get(REFRESH_COOKIE)).toBe('refresh-token');
	});

	it('still redirects on success when the backend sends no refresh cookie', async () => {
		const cookies = createMockCookies();
		const fetch = vi
			.fn()
			.mockResolvedValue(jsonResponse({ success: true, data: { accessToken: 'access-token' } }));

		let thrown: unknown;
		try {
			await actions.login(buildEvent(validLogin, fetch, cookies) as LoginEvent);
		} catch (e) {
			thrown = e;
		}

		expect(isRedirect(thrown)).toBe(true);
		expect(cookies.get(ACCESS_COOKIE)).toBe('access-token');
		expect(cookies.get(REFRESH_COOKIE)).toBeUndefined();
	});
});

describe('register action', () => {
	it('fails validation for a short name', async () => {
		const fetch = vi.fn();
		const result = await actions.register(
			buildEvent({ ...validRegister, name: 'A' }, fetch) as RegisterEvent
		);

		expect(result?.status).toBe(400);
		expect(result?.data.type).toBe('register');
		expect(fetch).not.toHaveBeenCalled();
	});

	it('fails when the passwords do not match', async () => {
		const fetch = vi.fn();
		const result = await actions.register(
			buildEvent({ ...validRegister, confirmPassword: 'different1' }, fetch) as RegisterEvent
		);

		expect(result?.status).toBe(400);
		expect(result?.data.errors).toEqual({ confirmPassword: 'Las contraseñas no coinciden' });
		expect(fetch).not.toHaveBeenCalled();
	});

	it('fails when the terms are not accepted', async () => {
		const fetch = vi.fn();
		const { terms: _omit, ...withoutTerms } = validRegister;
		const result = await actions.register(buildEvent(withoutTerms, fetch) as RegisterEvent);

		expect(result?.status).toBe(400);
		expect(result?.data.errors).toEqual({
			terms: 'Debes aceptar los términos y condiciones'
		});
		expect(fetch).not.toHaveBeenCalled();
	});

	it('surfaces the backend error when registration is rejected', async () => {
		const fetch = vi
			.fn()
			.mockResolvedValue(jsonResponse({ message: 'Email ya registrado' }, { status: 409 }));

		const result = await actions.register(buildEvent(validRegister, fetch) as RegisterEvent);

		expect(result?.status).toBe(409);
		expect(result?.data.errors).toEqual({ server: 'Email ya registrado' });
	});

	it('fails when the backend responds ok but success is false', async () => {
		const fetch = vi.fn().mockResolvedValue(jsonResponse({ success: false, message: 'No válido' }));

		const result = await actions.register(buildEvent(validRegister, fetch) as RegisterEvent);

		expect(result?.status).toBe(400);
		expect(result?.data.errors).toEqual({ server: 'No válido' });
	});

	it('flags a duplicate email with a friendly message instead of the raw backend one', async () => {
		const fetch = vi.fn().mockResolvedValue(
			jsonResponse(
				{ message: 'email already registered', action: 'auth:register:duplicate' },
				{ status: 409 }
			)
		);

		const result = await actions.register(buildEvent(validRegister, fetch) as RegisterEvent);
		const data = result?.data as { duplicateEmail?: boolean; errors?: Record<string, string> };

		expect(result?.status).toBe(409);
		expect(data.duplicateEmail).toBe(true);
		expect(data.errors).toEqual({
			server: 'Ya existe una cuenta con este correo. Inicia sesión o recupera tu contraseña.'
		});
	});

	it('flags a disabled registration with a friendly message instead of the raw backend one', async () => {
		const fetch = vi.fn().mockResolvedValue(
			jsonResponse(
				{ message: 'self-registration is disabled', action: 'auth:register:disabled' },
				{ status: 403 }
			)
		);

		const result = await actions.register(buildEvent(validRegister, fetch) as RegisterEvent);
		const data = result?.data as { disabled?: boolean; errors?: Record<string, string> };

		expect(result?.status).toBe(403);
		expect(data.disabled).toBe(true);
		expect(data.errors).toEqual({
			server: 'El registro está cerrado durante la beta. Únete a la lista de espera y te invitaremos.'
		});
	});

	it('redirects to /auth?registered=1 after a successful registration', async () => {
		const fetch = vi
			.fn()
			.mockResolvedValue(jsonResponse({ success: true, message: 'Cuenta creada' }));

		let thrown: unknown;
		try {
			await actions.register(buildEvent(validRegister, fetch) as RegisterEvent);
		} catch (e) {
			thrown = e;
		}

		expect(isRedirect(thrown)).toBe(true);
		expect((thrown as { status: number }).status).toBe(302);
		expect((thrown as { location: string }).location).toBe('/auth?registered=1');
	});
});
