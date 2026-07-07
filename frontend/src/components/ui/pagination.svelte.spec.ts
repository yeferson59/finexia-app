import { page } from 'vitest/browser';
import { describe, it, expect } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Pagination from './pagination.svelte';

describe('pagination.svelte', () => {
	it('does not render when everything fits on one page', async () => {
		render(Pagination, { total: 8, perPage: 10 });
		await expect
			.element(page.getByRole('navigation', { name: 'Paginación' }))
			.not.toBeInTheDocument();
	});

	it('renders a page button per page and a range summary', async () => {
		render(Pagination, { total: 25, perPage: 10, label: 'activos' });

		await expect.element(page.getByRole('navigation', { name: 'Paginación' })).toBeInTheDocument();
		await expect.element(page.getByText('1–10 de 25 activos')).toBeInTheDocument();
		await expect.element(page.getByRole('button', { name: 'Página 1' })).toBeInTheDocument();
		await expect.element(page.getByRole('button', { name: 'Página 3' })).toBeInTheDocument();
	});

	it('disables the previous control on the first page', async () => {
		render(Pagination, { total: 25, perPage: 10 });
		await expect.element(page.getByRole('button', { name: 'Página anterior' })).toBeDisabled();
	});

	it('advances the range summary when a later page is selected', async () => {
		render(Pagination, { total: 25, perPage: 10 });

		await page.getByRole('button', { name: 'Página 3' }).click();

		await expect.element(page.getByText('21–25 de 25 elementos')).toBeInTheDocument();
		await expect.element(page.getByRole('button', { name: 'Página siguiente' })).toBeDisabled();
	});
});
