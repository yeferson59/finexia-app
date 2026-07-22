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
		onSwitchToRegister
	}: {
		form: AuthActionResult;
		slideDirection?: 'left' | 'right';
		onSwitchToRegister: () => void;
	} = $props();

	let loginEmail = $state('');
	let loginPassword = $state('');
	let isSubmitting = $state(false);

	const errors = $derived(form?.type === 'login' ? parseErrors(form.errors) : {});
</script>

<form
	method="POST"
	action="?/login"
	class="form-content"
	class:slide-left={slideDirection === 'left'}
	id="login-form"
	use:enhance={() => {
		isSubmitting = true;
		return async ({ update }) => {
			await update({ reset: false });
			isSubmitting = false;
		};
	}}
>
	<Input
		label="Email"
		id="login-email"
		name="email"
		type="email"
		placeholder="tu@email.com"
		bind:value={loginEmail}
		error={errors['email']}
		required
	/>

	<PasswordInput
		label="Contraseña"
		id="login-password"
		name="password"
		placeholder="Ingresa tu contraseña"
		bind:value={loginPassword}
		error={errors['password']}
	/>

	<div class="form-footer">
		<a href={resolve('/auth/forgot-password')} class="forgot-link">¿Olvidaste tu contraseña?</a>
	</div>

	{#if errors['server']}
		<p class="error-server" role="alert">{errors['server']}</p>
	{/if}

	{#if form?.type === 'login' && form.unverified}
		<a href={resolve('/auth/verify-email')} class="resend-link">
			Reenviar enlace de verificación
		</a>
	{/if}

	<Button type="submit" variant="primary" size="lg" loading={isSubmitting} fullWidth>
		{isSubmitting ? 'Iniciando sesión...' : 'Iniciar sesión'}
	</Button>

	<div class="form-switch">
		¿No tienes cuenta?
		<button type="button" onclick={onSwitchToRegister} class="switch-link"> Crear una </button>
	</div>
</form>
