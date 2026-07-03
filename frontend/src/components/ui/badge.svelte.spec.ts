import { page } from 'vitest/browser';
import { describe, it, expect } from 'vitest';
import { render } from 'vitest-browser-svelte';
import { createRawSnippet } from 'svelte';
import Badge from './badge.svelte';

function text(value: string) {
	return createRawSnippet(() => ({ render: () => `<span>${value}</span>` }));
}

describe('badge.svelte', () => {
	it('renders its children with the default neutral, sm, pill, uppercase styling', async () => {
		render(Badge, { children: text('Activo') });

		await expect.element(page.getByText('Activo')).toBeInTheDocument();

		const badge = document.querySelector('.badge') as HTMLElement;
		expect(badge.classList.contains('badge-neutral')).toBe(true);
		expect(badge.classList.contains('badge-sm')).toBe(true);
		expect(badge.classList.contains('badge-pill')).toBe(true);
		expect(badge.classList.contains('badge-uppercase')).toBe(true);
	});

	it('applies the requested tone and size', async () => {
		render(Badge, { tone: 'success', size: 'md', children: text('OK') });

		const badge = document.querySelector('.badge') as HTMLElement;
		expect(badge.classList.contains('badge-success')).toBe(true);
		expect(badge.classList.contains('badge-md')).toBe(true);
	});

	it('drops the pill and uppercase modifiers when disabled', async () => {
		render(Badge, { pill: false, uppercase: false, children: text('tag') });

		const badge = document.querySelector('.badge') as HTMLElement;
		expect(badge.classList.contains('badge-pill')).toBe(false);
		expect(badge.classList.contains('badge-uppercase')).toBe(false);
	});
});
