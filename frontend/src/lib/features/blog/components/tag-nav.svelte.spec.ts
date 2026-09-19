import { page } from 'vitest/browser';
import { describe, it, expect } from 'vitest';
import { render } from 'vitest-browser-svelte';
import TagNav from './tag-nav.svelte';
import type { TagSummary } from '../blog';

const tags: TagSummary[] = [
	{ slug: 'producto', label: 'producto', count: 2 },
	{ slug: 'guias', label: 'guías', count: 1 }
];

describe('tag-nav.svelte', () => {
	it('lista las etiquetas con cuántos artículos lleva cada una', async () => {
		render(TagNav, { tags });

		await expect.element(page.getByRole('link', { name: 'producto' })).toBeInTheDocument();
		await expect.element(page.getByRole('link', { name: 'guías' })).toBeInTheDocument();
		expect(document.querySelectorAll('.count')).toHaveLength(2);
	});

	// Desde una etiqueta hay que poder volver a verlo todo, y la salida tiene
	// que estar donde estaba la entrada.
	it('deja «Todos» delante, marcado cuando no hay etiqueta activa', async () => {
		render(TagNav, { tags });

		await expect
			.element(page.getByRole('link', { name: 'Todos' }))
			.toHaveAttribute('aria-current', 'page');
	});

	it('marca la etiqueta que se está viendo y desmarca «Todos»', async () => {
		render(TagNav, { tags, active: 'guias' });

		await expect
			.element(page.getByRole('link', { name: 'guías' }))
			.toHaveAttribute('aria-current', 'page');
		expect(document.querySelector('a[href="/blog"]')?.hasAttribute('aria-current')).toBe(false);
	});

	it('no pinta nada cuando no hay etiquetas', async () => {
		render(TagNav, { tags: [] });

		expect(document.querySelector('.tag-nav')).toBeNull();
	});
});
