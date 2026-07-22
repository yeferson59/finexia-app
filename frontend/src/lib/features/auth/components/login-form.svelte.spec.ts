import { page } from 'vitest/browser';
import { describe, it, expect } from 'vitest';
import { render } from 'vitest-browser-svelte';
import LoginForm from './login-form.svelte';

const noop = () => {};

describe('login-form.svelte', () => {
	it('renders the email and password fields', async () => {
		render(LoginForm, { form: null, onSwitchToRegister: noop });

		await expect.element(page.getByPlaceholder('tu@email.com')).toBeInTheDocument();
		await expect.element(page.getByPlaceholder('Ingresa tu contraseña')).toBeInTheDocument();
	});

	it('renders a server error coming from the login action result', async () => {
		render(LoginForm, {
			form: { type: 'login', errors: { server: 'Credenciales incorrectas' } },
			onSwitchToRegister: noop
		});

		const alert = page.getByRole('alert');
		await expect.element(alert).toBeInTheDocument();
		await expect.element(alert).toHaveTextContent('Credenciales incorrectas');
	});

	it('maps a zod issue array into a field-level error message', async () => {
		render(LoginForm, {
			form: { type: 'login', errors: [{ path: ['email'], message: 'Correo inválido' }] },
			onSwitchToRegister: noop
		});

		await expect.element(page.getByText('Correo inválido')).toBeInTheDocument();
	});

	it('offers to resend the verification link when the account is unverified', async () => {
		render(LoginForm, {
			form: { type: 'login', errors: { server: 'Verifica tu correo' }, unverified: true },
			onSwitchToRegister: noop
		});

		await expect
			.element(page.getByRole('link', { name: 'Reenviar enlace de verificación' }))
			.toBeInTheDocument();
	});
});
