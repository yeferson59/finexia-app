import { page } from 'vitest/browser';
import { describe, it, expect } from 'vitest';
import { render } from 'vitest-browser-svelte';
import RegisterForm from './register-form.svelte';

const noop = () => {};

describe('register-form.svelte', () => {
	it('renders the name, email and password fields', async () => {
		render(RegisterForm, { form: null, onSwitchToLogin: noop });

		await expect.element(page.getByPlaceholder('Juan Pérez')).toBeInTheDocument();
		await expect.element(page.getByPlaceholder('Crea una contraseña segura')).toBeInTheDocument();
		await expect.element(page.getByPlaceholder('Repite tu contraseña')).toBeInTheDocument();
	});

	it('shows a terms error from the register action result', async () => {
		render(RegisterForm, {
			form: {
				type: 'register',
				errors: { terms: 'Debes aceptar los términos y condiciones' }
			},
			onSwitchToLogin: noop
		});

		await expect
			.element(page.getByText('Debes aceptar los términos y condiciones'))
			.toBeInTheDocument();
	});

	it('offers login shortcuts when the email is already registered', async () => {
		render(RegisterForm, {
			form: {
				type: 'register',
				errors: { server: 'Ya existe una cuenta con este correo.' },
				duplicateEmail: true
			},
			onSwitchToLogin: noop
		});

		await expect
			.element(page.getByRole('button', { name: 'Iniciar sesión con este correo' }))
			.toBeInTheDocument();
	});
});
