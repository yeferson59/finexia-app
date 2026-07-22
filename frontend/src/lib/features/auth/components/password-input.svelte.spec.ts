import { page } from 'vitest/browser';
import { describe, it, expect } from 'vitest';
import { render } from 'vitest-browser-svelte';
import PasswordInput from './password-input.svelte';

describe('password-input.svelte', () => {
	it('toggles password visibility when the show/hide button is clicked', async () => {
		render(PasswordInput, {
			label: 'Contraseña',
			id: 'login-password',
			name: 'password',
			placeholder: 'Ingresa tu contraseña'
		});

		const password = page.getByPlaceholder('Ingresa tu contraseña');
		await expect.element(password).toHaveAttribute('type', 'password');

		await page.getByRole('button', { name: 'Mostrar contraseña' }).click();
		await expect.element(password).toHaveAttribute('type', 'text');
	});

	it('surfaces a field error passed down from the parent form', async () => {
		render(PasswordInput, {
			label: 'Contraseña',
			name: 'password',
			placeholder: 'Ingresa tu contraseña',
			error: 'La contraseña es obligatoria'
		});

		await expect.element(page.getByText('La contraseña es obligatoria')).toBeInTheDocument();
	});
});
