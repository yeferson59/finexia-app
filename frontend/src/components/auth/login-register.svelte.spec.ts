import { page } from 'vitest/browser';
import { describe, it, expect } from 'vitest';
import { render } from 'vitest-browser-svelte';
import LoginRegister from './login-register.svelte';

describe('login-register.svelte', () => {
	it('shows the login form by default', async () => {
		render(LoginRegister, { form: null });

		await expect.element(page.getByPlaceholder('tu@email.com')).toBeInTheDocument();
		await expect.element(page.getByPlaceholder('Ingresa tu contraseña')).toBeInTheDocument();
		await expect
			.element(page.getByRole('tab', { name: 'Iniciar sesión' }))
			.toHaveAttribute('aria-selected', 'true');
	});

	it('switches to the register form and reveals the extra fields', async () => {
		render(LoginRegister, { form: null });

		await page.getByRole('tab', { name: 'Crear cuenta' }).click();

		await expect.element(page.getByPlaceholder('Juan Pérez')).toBeInTheDocument();
		await expect.element(page.getByPlaceholder('Repite tu contraseña')).toBeInTheDocument();
		await expect
			.element(page.getByRole('tab', { name: 'Crear cuenta' }))
			.toHaveAttribute('aria-selected', 'true');
	});

	it('renders a server error coming from the login action result', async () => {
		render(LoginRegister, {
			form: { type: 'login', errors: { server: 'Credenciales incorrectas' } }
		});

		const alert = page.getByRole('alert');
		await expect.element(alert).toBeInTheDocument();
		await expect.element(alert).toHaveTextContent('Credenciales incorrectas');
	});

	it('maps a zod issue array into a field-level error message', async () => {
		render(LoginRegister, {
			form: { type: 'login', errors: [{ path: ['email'], message: 'Correo inválido' }] }
		});

		await expect.element(page.getByText('Correo inválido')).toBeInTheDocument();
	});

	it('toggles password visibility when the show/hide button is clicked', async () => {
		render(LoginRegister, { form: null });

		const password = page.getByPlaceholder('Ingresa tu contraseña');
		await expect.element(password).toHaveAttribute('type', 'password');

		await page.getByRole('button', { name: 'Mostrar contraseña' }).click();
		await expect.element(password).toHaveAttribute('type', 'text');
	});
});
