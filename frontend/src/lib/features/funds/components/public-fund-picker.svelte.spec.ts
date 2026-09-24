import { page, userEvent } from 'vitest/browser';
import { describe, it, expect, vi, afterEach } from 'vitest';
import { render } from 'vitest-browser-svelte';
import PublicFundPicker from './public-fund-picker.svelte';

const fiducuenta = {
	id: '5-31-2852-1-800',
	entityName: 'Fiduciaria Bancolombia S.A. Sociedad Fiduciaria',
	fundName: 'FONDO DE INVERSIÓN COLECTIVA ABIERTO FIDUCUENTA',
	fundKind: 'FIC DE MERCADO MONETARIO',
	fundCode: 2852,
	participation: 800,
	unitValue: '48354.833953',
	valueDate: '2026-09-22T00:00:00Z',
	investors: 852578
};

function stubCatalog(status = 200, data: unknown[] = [fiducuenta]) {
	const fetchMock = vi.fn(async () =>
		Response.json(status === 200 ? { success: true, data } : { success: false, data: [] }, {
			status
		})
	);
	vi.stubGlobal('fetch', fetchMock);

	return fetchMock;
}

afterEach(() => {
	vi.unstubAllGlobals();
});

describe('public-fund-picker.svelte', () => {
	it('searches the catalog and carries the chosen fund in a hidden field', async () => {
		const fetchMock = stubCatalog();
		const onSelect = vi.fn();
		render(PublicFundPicker, { onSelect });
		const hidden = () => document.querySelector<HTMLInputElement>('input[name="publicFundId"]');

		await userEvent.type(page.getByRole('combobox'), 'fiducuenta');

		const option = page.getByRole('option', { name: /Fiducuenta/ });
		await expect.element(option).toBeVisible();
		await expect.element(option).toMatchTextContent(/Participación 800/);
		expect(fetchMock).toHaveBeenCalledWith('/api/funds/catalog?q=fiducuenta');

		// Chosen with the keyboard, as a list of options is.
		await userEvent.keyboard('{ArrowDown}{Enter}');

		await expect.element(page.getByText('Fiducuenta', { exact: true })).toBeVisible();
		expect(hidden()?.value).toBe(fiducuenta.id);
		expect(onSelect).toHaveBeenCalledWith(fiducuenta);

		await page.getByRole('button', { name: 'Cambiar' }).click();

		expect(hidden()?.value).toBe('');
	});

	it('does not search for a single letter', async () => {
		const fetchMock = stubCatalog();
		render(PublicFundPicker, {});

		await userEvent.type(page.getByRole('combobox'), 'f');
		await new Promise((r) => setTimeout(r, 400));

		expect(fetchMock).not.toHaveBeenCalled();
	});

	it('says so when the Superfinanciera does not answer', async () => {
		stubCatalog(503);
		render(PublicFundPicker, {});

		await userEvent.type(page.getByRole('combobox'), 'renta');

		await expect.element(page.getByRole('alert')).toMatchTextContent(/no respondió/);
	});
});
