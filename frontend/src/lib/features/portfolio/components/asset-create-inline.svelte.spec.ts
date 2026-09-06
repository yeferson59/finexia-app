import { page } from 'vitest/browser';
import { describe, it, expect, vi, afterEach } from 'vitest';
import { render } from 'vitest-browser-svelte';
import AssetCreateInline from './asset-create-inline.svelte';

const created = {
	id: 'a1',
	ticker: 'GEB',
	name: 'Grupo Energía Bogotá',
	assetType: 'stock',
	currency: 'COP',
	currentPrice: null,
	priceUpdatedAt: null,
	isCurated: false
};

function stubFetch(response: Partial<Response> & { json: () => Promise<unknown> }) {
	const fetchMock = vi.fn().mockResolvedValue({ ok: true, ...response });
	vi.stubGlobal('fetch', fetchMock);

	return fetchMock;
}

afterEach(() => {
	vi.unstubAllGlobals();
});

describe('asset-create-inline.svelte', () => {
	it('normalises the ticker and falls back to it when no name is given', async () => {
		const fetchMock = stubFetch({ json: async () => ({ success: true, data: created }) });
		const oncreated = vi.fn();

		render(AssetCreateInline, { ticker: ' geb ', oncreated, oncancel: () => {} });

		await page.getByRole('button', { name: 'Crear activo' }).click();

		expect(fetchMock).toHaveBeenCalledTimes(1);
		const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit];
		expect(url).toBe('/api/assets');
		expect(JSON.parse(init.body as string)).toMatchObject({
			ticker: 'GEB',
			name: 'GEB',
			assetType: 'stock',
			currency: 'USD'
		});
	});

	it('hands the created asset back so the combobox can select it', async () => {
		stubFetch({ json: async () => ({ success: true, data: created }) });
		const oncreated = vi.fn();

		render(AssetCreateInline, { ticker: 'GEB', oncreated, oncancel: () => {} });
		await page.getByRole('button', { name: 'Crear activo' }).click();

		await vi.waitFor(() => expect(oncreated).toHaveBeenCalledWith(created));
	});

	it('shows the backend message on a rejected creation and does not select anything', async () => {
		stubFetch({
			ok: false,
			status: 429,
			json: async () => ({ success: false, message: 'has añadido demasiados activos nuevos' })
		});
		const oncreated = vi.fn();

		render(AssetCreateInline, { ticker: 'GEB', oncreated, oncancel: () => {} });
		await page.getByRole('button', { name: 'Crear activo' }).click();

		await expect
			.element(page.getByText('has añadido demasiados activos nuevos'))
			.toBeInTheDocument();
		expect(oncreated).not.toHaveBeenCalled();
	});

	it('cancels without creating anything', async () => {
		const fetchMock = stubFetch({ json: async () => ({ success: true, data: created }) });
		const oncancel = vi.fn();

		render(AssetCreateInline, { ticker: 'GEB', oncreated: () => {}, oncancel });
		await page.getByRole('button', { name: 'Cancelar' }).click();

		expect(oncancel).toHaveBeenCalledTimes(1);
		expect(fetchMock).not.toHaveBeenCalled();
	});
});
