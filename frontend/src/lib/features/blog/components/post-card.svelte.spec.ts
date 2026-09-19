import { page } from 'vitest/browser';
import { describe, it, expect } from 'vitest';
import { render } from 'vitest-browser-svelte';
import PostCard from './post-card.svelte';
import type { PostMeta } from '../blog';

function post(over: Partial<PostMeta> = {}): PostMeta {
	return {
		slug: 'por-que-existe-finexia',
		title: 'Por qué existe Finexia',
		description: 'Tu dinero está repartido en cinco sitios y ninguno te enseña el total.',
		date: '2026-09-15',
		tags: ['producto'],
		author: 'Equipo Finexia',
		draft: false,
		readingMinutes: 4,
		...over
	};
}

describe('post-card.svelte', () => {
	it('enseña el titular, el resumen, la fecha y la lectura', async () => {
		render(PostCard, { post: post() });

		await expect.element(page.getByText('Por qué existe Finexia')).toBeInTheDocument();
		await expect.element(page.getByText('Tu dinero está repartido')).toBeInTheDocument();
		await expect.element(page.getByText('15 de septiembre de 2026')).toBeInTheDocument();
		await expect.element(page.getByText('4 min de lectura')).toBeInTheDocument();
	});

	it('lleva del titular al artículo', async () => {
		render(PostCard, { post: post({ slug: 'otro-articulo' }) });

		await expect
			.element(page.getByRole('link', { name: 'Por qué existe Finexia' }))
			.toHaveAttribute('href', '/blog/otro-articulo');
	});

	// Las etiquetas son enlaces propios, no adorno: desde una entrada se salta a
	// todo lo que hay escrito sobre lo mismo.
	it('enlaza cada etiqueta a su página, con la tilde ya fuera de la URL', async () => {
		render(PostCard, { post: post({ tags: ['guías'] }) });

		await expect
			.element(page.getByRole('link', { name: 'guías' }))
			.toHaveAttribute('href', '/blog/tag/guias');
	});

	it('marca la fecha para que una máquina también la lea', async () => {
		render(PostCard, { post: post() });

		const time = document.querySelector('time');
		expect(time?.getAttribute('datetime')).toBe('2026-09-15');
	});

	it('marca un borrador como tal', async () => {
		render(PostCard, { post: post({ draft: true }) });

		await expect.element(page.getByText('Borrador')).toBeInTheDocument();
	});

	it('no marca un artículo publicado', async () => {
		render(PostCard, { post: post() });

		await expect.element(page.getByText('Borrador')).not.toBeInTheDocument();
	});
});
