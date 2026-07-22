import { page } from 'vitest/browser';
import { describe, it, expect } from 'vitest';
import { render } from 'vitest-browser-svelte';
import LoginRegister from './login-register.svelte';

describe('login-register.svelte (container)', () => {
	it('shows the login form by default', async () => {
		render(LoginRegister, { form: null });

		await expect.element(page.getByPlaceholder('tu@email.com')).toBeInTheDocument();
		await expect.element(page.getByPlaceholder('Ingresa tu contraseña')).toBeInTheDocument();
		await expect
			.element(page.getByRole('tab', { name: 'Iniciar sesión' }))
			.toHaveAttribute('aria-selected', 'true');
	});

	it('switches to the register form and reveals the extra fields when self-registration is enabled', async () => {
		render(LoginRegister, { form: null, selfRegistrationEnabled: true });

		await page.getByRole('tab', { name: 'Crear cuenta' }).click();

		await expect.element(page.getByPlaceholder('Juan Pérez')).toBeInTheDocument();
		await expect.element(page.getByPlaceholder('Repite tu contraseña')).toBeInTheDocument();
		await expect
			.element(page.getByRole('tab', { name: 'Crear cuenta' }))
			.toHaveAttribute('aria-selected', 'true');
	});

	it('shows an invite-only notice instead of the form while self-registration is disabled', async () => {
		render(LoginRegister, { form: null, selfRegistrationEnabled: false });

		await page.getByRole('tab', { name: 'Crear cuenta' }).click();

		await expect.element(page.getByText('Registro por invitación')).toBeInTheDocument();
		await expect
			.element(page.getByRole('link', { name: 'Unirme a la lista de espera' }))
			.toBeInTheDocument();
		await expect.element(page.getByPlaceholder('Juan Pérez')).not.toBeInTheDocument();
	});

	it('defaults to the invite-only notice when the prop is omitted', async () => {
		render(LoginRegister, { form: null });

		await page.getByRole('tab', { name: 'Crear cuenta' }).click();

		await expect.element(page.getByText('Registro por invitación')).toBeInTheDocument();
	});
});
