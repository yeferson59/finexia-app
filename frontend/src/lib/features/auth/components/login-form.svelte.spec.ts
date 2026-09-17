import { page } from 'vitest/browser';
import { describe, it, expect } from 'vitest';
import { render } from 'vitest-browser-svelte';
import LoginForm from './login-form.svelte';

describe('login-form.svelte', () => {
	it('renders the email and password fields', async () => {
		render(LoginForm, { form: null });

		await expect.element(page.getByLabelText('Correo electrónico')).toBeInTheDocument();
		await expect.element(page.getByLabelText('Contraseña', { exact: true })).toBeInTheDocument();
	});

	it('renders a server error coming from the login action result', async () => {
		render(LoginForm, {
			form: { type: 'login', errors: { server: 'Credenciales incorrectas' } }
		});

		const alert = page.getByRole('alert');
		await expect.element(alert).toBeInTheDocument();
		await expect.element(alert).toHaveTextContent('Credenciales incorrectas');
	});

	it('links the forgotten-password flow from the password field', async () => {
		render(LoginForm, { form: null });

		await expect
			.element(page.getByRole('link', { name: '¿Olvidaste tu contraseña?' }))
			.toHaveAttribute('href', '/auth/forgot-password');
	});

	it('maps a zod issue array into a field-level error message', async () => {
		render(LoginForm, {
			form: { type: 'login', errors: [{ path: ['email'], message: 'Correo inválido' }] }
		});

		await expect.element(page.getByText('Correo inválido')).toBeInTheDocument();
		await expect
			.element(page.getByLabelText('Correo electrónico'))
			.toHaveAttribute('aria-invalid', 'true');
	});

	it('offers to resend the verification link when the account is unverified', async () => {
		render(LoginForm, {
			form: { type: 'login', errors: { server: 'Verifica tu correo' }, unverified: true }
		});

		await expect
			.element(page.getByRole('link', { name: 'Reenviar enlace de verificación' }))
			.toBeInTheDocument();
	});
});
