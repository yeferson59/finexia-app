import { page } from 'vitest/browser';
import { describe, it, expect, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import { createRawSnippet } from 'svelte';
import Button from './button.svelte';

/** A snippet that renders the given plain text, for the `children` prop. */
function text(value: string) {
	return createRawSnippet(() => ({ render: () => `<span>${value}</span>` }));
}

describe('button.svelte', () => {
	it('renders its children and defaults to the primary + md variant', async () => {
		render(Button, { children: text('Guardar') });

		const button = page.getByRole('button', { name: 'Guardar' });
		await expect.element(button).toBeInTheDocument();
		await expect.element(button).toHaveClass('btn-primary');
		await expect.element(button).toHaveClass('btn-md');
	});

	it('applies the requested variant and size classes', async () => {
		render(Button, { variant: 'secondary', size: 'lg', children: text('Cancelar') });

		const button = page.getByRole('button', { name: 'Cancelar' });
		await expect.element(button).toHaveClass('btn-secondary');
		await expect.element(button).toHaveClass('btn-lg');
	});

	it('is disabled when either disabled or loading is set', async () => {
		const { rerender } = await render(Button, { disabled: true, children: text('X') });
		await expect.element(page.getByRole('button')).toBeDisabled();

		await rerender({ disabled: false, loading: true, children: text('X') });
		await expect.element(page.getByRole('button')).toBeDisabled();
	});

	it('fires onclick when clicked', async () => {
		const onclick = vi.fn();
		render(Button, { onclick, type: 'button', children: text('Enviar') });

		await page.getByRole('button', { name: 'Enviar' }).click();
		expect(onclick).toHaveBeenCalledOnce();
	});

	it('does not fire onclick while loading (button is disabled)', async () => {
		const onclick = vi.fn();
		render(Button, { onclick, loading: true, type: 'button', children: text('Enviar') });

		await page
			.getByRole('button')
			.click({ force: true })
			.catch(() => {});
		expect(onclick).not.toHaveBeenCalled();
	});

	it('honors the type attribute', async () => {
		render(Button, { type: 'submit', children: text('OK') });
		await expect.element(page.getByRole('button')).toHaveAttribute('type', 'submit');
	});
});
