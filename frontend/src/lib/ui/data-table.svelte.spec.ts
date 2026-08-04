import { page } from 'vitest/browser';
import { describe, it, expect } from 'vitest';
import { render } from 'vitest-browser-svelte';
import { createRawSnippet } from 'svelte';
import DataTable from './data-table.svelte';

const rows = createRawSnippet(() => ({
	render: () =>
		`<tbody><tr><td class="cell-ticker">AAPL</td><td class="cell-name">Apple Inc.</td></tr></tbody>`
}));

describe('data-table.svelte', () => {
	it('renders the given rows inside a table', async () => {
		render(DataTable, { children: rows });

		await expect.element(page.getByRole('table')).toBeInTheDocument();
		await expect.element(page.getByRole('cell', { name: 'Apple Inc.' })).toBeInTheDocument();
	});

	it('names the table for assistive tech without showing the caption', async () => {
		render(DataTable, { caption: 'Posiciones del portafolio', children: rows });

		await expect
			.element(page.getByRole('table', { name: 'Posiciones del portafolio' }))
			.toBeInTheDocument();
		expect(document.querySelector('caption')?.classList.contains('sr-only')).toBe(true);
	});

	it('shows the caption when asked to', async () => {
		render(DataTable, { caption: 'Movimientos', showCaption: true, children: rows });

		expect(document.querySelector('caption')?.classList.contains('sr-only')).toBe(false);
	});

	it('pins the header only when stickyHeader is set', async () => {
		const { rerender } = await render(DataTable, { children: rows });
		expect(document.querySelector('table')?.classList.contains('sticky-header')).toBe(false);

		await rerender({ stickyHeader: true, children: rows });
		expect(document.querySelector('table')?.classList.contains('sticky-header')).toBe(true);
	});
});
