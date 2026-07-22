<script lang="ts">
	import { enhance } from '$app/forms';
	import { resolve } from '$app/paths';
	import Button from '$lib/ui/button.svelte';
	import Input from '$lib/ui/input.svelte';
	import PasswordInput from './password-input.svelte';
	import type { AuthActionResult } from '../types';
	import { parseErrors } from '../utils';

	let {
		form,
		slideDirection = 'right',
		onSwitchToLogin
	}: {
		form: AuthActionResult;
		slideDirection?: 'left' | 'right';
		onSwitchToLogin: () => void;
	} = $props();

	let registerName = $state('');
	let registerEmail = $state('');
	let registerPassword = $state('');
	let registerConfirmPassword = $state('');
	let agreeTerms = $state(false);
	let isSubmitting = $state(false);

	const errors = $derived(form?.type === 'register' ? parseErrors(form.errors) : {});
</script>

<form
	method="POST"
	action="?/register"
	class="form-content"
	class:slide-left={slideDirection === 'left'}
	id="register-form"
	use:enhance={() => {
		isSubmitting = true;
		return async ({ update }) => {
			await update({ reset: false });
			isSubmitting = false;
		};
	}}
>
	<Input
		label="Nombre completo"
		id="register-name"
		name="name"
		type="text"
		placeholder="Juan Pérez"
		bind:value={registerName}
		error={errors['name']}
		required
	/>

	<Input
		label="Email"
		id="register-email"
		name="email"
		type="email"
		placeholder="tu@email.com"
		bind:value={registerEmail}
		error={errors['email']}
		required
	/>

	<PasswordInput
		label="Contraseña"
		id="register-password"
		name="password"
		placeholder="Crea una contraseña segura"
		bind:value={registerPassword}
		error={errors['password']}
	/>

	<PasswordInput
		label="Confirmar contraseña"
		id="register-confirm"
		name="confirmPassword"
		placeholder="Repite tu contraseña"
		bind:value={registerConfirmPassword}
		error={errors['confirmPassword']}
	/>

	<div class="consent">
		<input
			type="checkbox"
			id="terms"
			name="terms"
			class="consent-input"
			bind:checked={agreeTerms}
		/>
		<label for="terms" class="consent-label">
			Autorizo el tratamiento de mis datos personales según la
			<a href={resolve('/privacidad')} target="_blank" rel="noopener">Política de Privacidad</a>
			y acepto los
			<a href={resolve('/terminos')} target="_blank" rel="noopener">Términos y Condiciones</a>.
		</label>
	</div>
	{#if errors['terms']}
		<span class="error-message">{errors['terms']}</span>
	{/if}

	{#if errors['server']}
		<p class="error-server" role="alert">{errors['server']}</p>
	{/if}

	{#if form?.type === 'register' && form.duplicateEmail}
		<button type="button" onclick={onSwitchToLogin} class="resend-link">
			Iniciar sesión con este correo
		</button>
		<a href={resolve('/auth/forgot-password')} class="resend-link"> ¿Olvidaste tu contraseña? </a>
	{/if}

	<Button type="submit" variant="primary" size="lg" loading={isSubmitting} fullWidth>
		{isSubmitting ? 'Creando cuenta...' : 'Crear cuenta'}
	</Button>

	<div class="form-switch">
		¿Ya tienes cuenta?
		<button type="button" onclick={onSwitchToLogin} class="switch-link"> Inicia sesión </button>
	</div>
</form>

<style>
	.consent {
		display: flex;
		align-items: flex-start;
		gap: 0.75rem;
	}

	.consent-input {
		appearance: none;
		width: 20px;
		height: 20px;
		margin-top: 1px;
		border: 1.5px solid rgba(212, 145, 42, 0.3);
		border-radius: 6px;
		background: rgba(255, 255, 255, 0.03);
		cursor: pointer;
		transition: all 0.25s ease;
		position: relative;
		flex-shrink: 0;
	}

	.consent-input:hover {
		border-color: rgba(212, 145, 42, 0.5);
	}

	.consent-input:checked {
		background: var(--amber);
		border-color: var(--amber);
	}

	.consent-input:checked::after {
		content: '✓';
		position: absolute;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
		color: #0d0800;
		font-size: 0.875rem;
		font-weight: 700;
	}

	.consent-input:focus-visible {
		outline: 2px solid var(--amber);
		outline-offset: 2px;
	}

	.consent-label {
		font-size: 0.8rem;
		line-height: 1.5;
		color: var(--text-secondary);
		cursor: pointer;
		letter-spacing: 0.2px;
		font-weight: 500;
	}

	.consent-label a {
		color: var(--gold-primary);
		text-decoration: underline;
		text-underline-offset: 2px;
		font-weight: 600;
	}

	.consent-label a:hover {
		color: var(--gold-light);
	}

	.error-message {
		font-size: 0.75rem;
		color: var(--error-color);
		margin-top: -1rem;
		letter-spacing: 0.2px;
		font-weight: 500;
	}
</style>
