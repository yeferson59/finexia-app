import { page } from 'vitest/browser';
import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest';
import { render } from 'vitest-browser-svelte';
import CookieNotice from './cookie-notice.svelte';

const STORAGE_KEY = 'finexia:cookie-notice';

describe('cookie-notice.svelte', () => {
	beforeEach(() => localStorage.removeItem(STORAGE_KEY));
	afterEach(() => vi.restoreAllMocks());

	it('se muestra mientras no se haya cerrado', async () => {
		render(CookieNotice);

		await expect
			.element(page.getByRole('region', { name: 'Aviso de cookies' }))
			.toBeInTheDocument();
	});

	it('no vuelve a aparecer una vez aceptado', async () => {
		localStorage.setItem(STORAGE_KEY, 'dismissed');
		render(CookieNotice);

		await expect
			.element(page.getByRole('region', { name: 'Aviso de cookies' }))
			.not.toBeInTheDocument();
	});

	it('al aceptar se cierra y lo recuerda', async () => {
		render(CookieNotice);

		await page.getByRole('button', { name: 'Entendido' }).click();

		await expect
			.element(page.getByRole('region', { name: 'Aviso de cookies' }))
			.not.toBeInTheDocument();
		expect(localStorage.getItem(STORAGE_KEY)).toBe('dismissed');
	});

	it('se muestra igualmente si el almacenamiento no está disponible', async () => {
		// Modo privado y similares: leer localStorage lanza.
		vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
			throw new Error('acceso denegado');
		});
		render(CookieNotice);

		await expect
			.element(page.getByRole('region', { name: 'Aviso de cookies' }))
			.toBeInTheDocument();
	});
});
