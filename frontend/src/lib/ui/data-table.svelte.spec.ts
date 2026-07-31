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
});
