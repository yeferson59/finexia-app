import { page } from 'vitest/browser';
import { describe, it, expect } from 'vitest';
import { render } from 'vitest-browser-svelte';
import { createRawSnippet } from 'svelte';
import RailSection from './rail-section.svelte';

function text(value: string) {
	return createRawSnippet(() => ({ render: () => `<span>${value}</span>` }));
}

describe('rail-section.svelte', () => {
	it('titles the block with an h2 and a top rule by default', async () => {
		render(RailSection, {
			title: 'Destino',
			description: 'Dónde quedan los movimientos.',
			children: text('Contenido')
		});

		await expect.element(page.getByRole('heading', { level: 2, name: 'Destino' })).toBeVisible();
		await expect.element(page.getByText('Dónde quedan los movimientos.')).toBeVisible();
		await expect.element(page.getByText('Contenido')).toBeVisible();
		expect(document.querySelector('.rail-section.divider-top')).not.toBeNull();
	});

	/* Dentro de un grupo el `<h2>` ya lo puso el grupo: repetirlo rompería el
	   esquema de encabezados con el que se navega la página. */
	it('drops to an h3 when the block lives inside a group', async () => {
		render(RailSection, { title: 'Contraseña', level: 3, children: text('X') });

		await expect.element(page.getByRole('heading', { level: 3, name: 'Contraseña' })).toBeVisible();
		expect(document.querySelector('.rail-section.nested')).not.toBeNull();
	});

	it('puts the rule below when asked, and omits it entirely when told to', async () => {
		render(RailSection, { title: 'A', divider: 'bottom', children: text('X') });
		expect(document.querySelector('.rail-section.divider-bottom')).not.toBeNull();

		render(RailSection, { title: 'B', divider: 'none', children: text('Y') });
		const plain = document.querySelectorAll('.rail-section')[1];
		expect(plain.classList.contains('divider-top')).toBe(false);
		expect(plain.classList.contains('divider-bottom')).toBe(false);
	});

	/* El contenido se corta a 34rem porque un nombre no se escribe en una caja de
	   mil píxeles; una tabla sí necesita el ancho entero. */
	it('caps the content width, and lets a table have the whole rail', async () => {
		render(RailSection, { title: 'Campos', children: text('X') });
		const capped = document.querySelector('.rail-section') as HTMLElement;
		expect(capped.style.getPropertyValue('--rail-content-max')).toBe('34rem');

		render(RailSection, { title: 'Tabla', contentMax: 'none', children: text('Y') });
		const wide = document.querySelectorAll('.rail-section')[1] as HTMLElement;
		expect(wide.style.getPropertyValue('--rail-content-max')).toBe('none');
	});

	/* `rail-fields` no es una clase con ámbito: la define `routes/layout.css`,
	   que es donde vive el aspecto de un campo, dentro y fuera de un carril. */
	it('marks the content column as a field column only when asked', async () => {
		render(RailSection, { title: 'Sin campos', children: text('X') });
		expect(document.querySelector('.content')?.classList.contains('rail-fields')).toBe(false);

		render(RailSection, { title: 'Con campos', fields: true, children: text('Y') });
		expect(document.querySelectorAll('.content')[1].classList.contains('rail-fields')).toBe(true);
	});

	it('renders the rail aside under the description', async () => {
		render(RailSection, {
			title: 'Asistentes',
			description: 'Conecta un cliente MCP.',
			aside: text('Solo lectura'),
			children: text('X')
		});

		await expect.element(page.getByText('Solo lectura')).toBeVisible();
	});
});
