import { page, userEvent } from 'vitest/browser';
import { describe, it, expect } from 'vitest';
import { render } from 'vitest-browser-svelte';
import LoginRegister from './login-register.svelte';

describe('login-register.svelte (container)', () => {
	it('shows the login form by default', async () => {
		render(LoginRegister, { form: null });

		await expect.element(page.getByLabelText('Correo electrónico')).toBeInTheDocument();
		await expect.element(page.getByLabelText('Contraseña', { exact: true })).toBeInTheDocument();
		await expect
			.element(page.getByRole('tab', { name: 'Iniciar sesión' }))
			.toHaveAttribute('aria-selected', 'true');
	});

	it('switches to the register form and reveals the extra fields when self-registration is enabled', async () => {
		render(LoginRegister, { form: null, selfRegistrationEnabled: true });

		await page.getByRole('tab', { name: 'Crear cuenta' }).click();

		await expect.element(page.getByLabelText('Nombre')).toBeInTheDocument();
		await expect.element(page.getByLabelText('Repite la contraseña')).toBeInTheDocument();
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
		await expect.element(page.getByLabelText('Nombre')).not.toBeInTheDocument();
	});

	it('defaults to the invite-only notice when the prop is omitted', async () => {
		render(LoginRegister, { form: null });

		await page.getByRole('tab', { name: 'Crear cuenta' }).click();

		await expect.element(page.getByText('Registro por invitación')).toBeInTheDocument();
	});

	it('moves between the two modes with the arrow keys', async () => {
		render(LoginRegister, { form: null });

		const login = page.getByRole('tab', { name: 'Iniciar sesión' });
		await login.click();
		await userEvent.keyboard('{ArrowDown}');

		await expect
			.element(page.getByRole('tab', { name: 'Crear cuenta' }))
			.toHaveAttribute('aria-selected', 'true');
	});

	it('opens on the register tab when the returning action was a registration', async () => {
		render(LoginRegister, {
			form: { type: 'register', errors: { server: 'Ya existe una cuenta con este correo.' } },
			selfRegistrationEnabled: true
		});

		await expect
			.element(page.getByRole('tab', { name: 'Crear cuenta' }))
			.toHaveAttribute('aria-selected', 'true');
		await expect
			.element(page.getByRole('alert'))
			.toHaveTextContent('Ya existe una cuenta con este correo.');
	});

	it('shows the notice that a previous screen sent back', async () => {
		render(LoginRegister, { form: null, notice: 'Correo verificado. Ya puedes iniciar sesión.' });

		await expect
			.element(page.getByRole('status'))
			.toHaveTextContent('Correo verificado. Ya puedes iniciar sesión.');
	});
});
