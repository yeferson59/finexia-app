import { page } from 'vitest/browser';
import { describe, it, expect } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Input from './input.svelte';

describe('input.svelte', () => {
	it('associates the label with the input via id', async () => {
		render(Input, { label: 'Correo', id: 'email', name: 'email' });

		const input = page.getByLabelText('Correo');
		await expect.element(input).toBeInTheDocument();
	});

	it('marks a required field with an asterisk', async () => {
		render(Input, { label: 'Correo', id: 'email', required: true });
		await expect.element(page.getByText('*')).toBeInTheDocument();
	});

	it('shows the error message and applies the error class', async () => {
		render(Input, { label: 'Correo', id: 'email', error: 'Correo inválido' });

		await expect.element(page.getByText('Correo inválido')).toBeInTheDocument();
		await expect.element(page.getByLabelText('Correo')).toHaveClass('input-error');
	});

	it('reflects typed text into the input value', async () => {
		render(Input, { label: 'Nombre', id: 'name' });

		const input = page.getByLabelText('Nombre');
		await input.fill('Yeferson');
		await expect.element(input).toHaveValue('Yeferson');
	});

	it('forwards the type and placeholder attributes', async () => {
		render(Input, { label: 'Clave', id: 'pw', type: 'password', placeholder: '••••' });

		const input = page.getByLabelText('Clave');
		await expect.element(input).toHaveAttribute('type', 'password');
		await expect.element(input).toHaveAttribute('placeholder', '••••');
	});
});
