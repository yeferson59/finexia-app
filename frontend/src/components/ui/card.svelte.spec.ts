import { page } from 'vitest/browser';
import { describe, it, expect, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import { createRawSnippet } from 'svelte';
import Card from './card.svelte';

function text(value: string) {
	return createRawSnippet(() => ({ render: () => `<span>${value}</span>` }));
}

describe('card.svelte', () => {
	it('renders a non-interactive div by default', async () => {
		render(Card, { children: text('Contenido') });

		await expect.element(page.getByText('Contenido')).toBeInTheDocument();
		expect(document.querySelector('div.card')).not.toBeNull();
		expect(page.getByRole('button').query()).toBeNull();
	});

	it('renders a labelled button and fires onclick when interactive', async () => {
		const onclick = vi.fn();
		render(Card, { onclick, ariaLabel: 'Abrir portafolio', children: text('Portafolio') });

		const button = page.getByRole('button', { name: 'Abrir portafolio' });
		await expect.element(button).toHaveClass('card-interactive');

		await button.click();
		expect(onclick).toHaveBeenCalledOnce();
	});

	it('applies the requested variant and padding classes', async () => {
		render(Card, { variant: 'elevated', padding: 'lg', children: text('X') });

		const card = document.querySelector('.card') as HTMLElement;
		expect(card.classList.contains('card-elevated')).toBe(true);
		expect(card.classList.contains('card-p-lg')).toBe(true);
	});
});
