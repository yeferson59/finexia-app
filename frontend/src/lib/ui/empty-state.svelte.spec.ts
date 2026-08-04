import { page } from 'vitest/browser';
import { describe, it, expect } from 'vitest';
import { render } from 'vitest-browser-svelte';
import { createRawSnippet } from 'svelte';
import EmptyState from './empty-state.svelte';

function markup(html: string) {
	return createRawSnippet(() => ({ render: () => html }));
}

describe('empty-state.svelte', () => {
	it('renders the title on its own', async () => {
		render(EmptyState, { title: 'Sin posiciones registradas' });

		await expect.element(page.getByText('Sin posiciones registradas')).toBeInTheDocument();
		expect(document.querySelector('.empty-description')).toBeNull();
		expect(document.querySelector('.empty-icon')).toBeNull();
		expect(document.querySelector('.empty-action')).toBeNull();
	});

	it('renders the description, icon and action when given', async () => {
		render(EmptyState, {
			title: 'Sin portafolios',
			description: 'Crea el primero para ver tu patrimonio.',
			icon: markup('<svg data-testid="icon"></svg>'),
			action: markup('<a href="/dashboard/portfolios">Crear portafolio</a>')
		});

		await expect
			.element(page.getByText('Crea el primero para ver tu patrimonio.'))
			.toBeInTheDocument();
		await expect.element(page.getByRole('link', { name: 'Crear portafolio' })).toBeInTheDocument();
		expect(document.querySelector('.empty-icon')?.getAttribute('aria-hidden')).toBe('true');
	});

	it('defaults to the md size and adds the dashed outline only when bordered', async () => {
		const { rerender } = await render(EmptyState, { title: 'Vacío' });

		let root = document.querySelector('.empty-state') as HTMLElement;
		expect(root.classList.contains('empty-md')).toBe(true);
		expect(root.classList.contains('empty-bordered')).toBe(false);

		await rerender({ title: 'Vacío', size: 'sm', bordered: true });

		root = document.querySelector('.empty-state') as HTMLElement;
		expect(root.classList.contains('empty-sm')).toBe(true);
		expect(root.classList.contains('empty-bordered')).toBe(true);
	});
});
