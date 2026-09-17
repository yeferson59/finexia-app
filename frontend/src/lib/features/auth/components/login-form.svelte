<script lang="ts">
	import { enhance } from '$app/forms';
	import { resolve } from '$app/paths';
	import AuthField from './auth-field.svelte';
	import PasswordInput from './password-input.svelte';
	import type { AuthActionResult } from '../types';
	import { parseErrors } from '../utils';

	let { form }: { form: AuthActionResult } = $props();

	let loginEmail = $state('');
	let loginPassword = $state('');
	let isSubmitting = $state(false);

	const errors = $derived(form?.type === 'login' ? parseErrors(form.errors) : {});
</script>

<form
	method="POST"
	action="?/login"
	class="lp-fields"
	id="login-form"
	use:enhance={() => {
		isSubmitting = true;
		return async ({ update }) => {
			await update({ reset: false });
			isSubmitting = false;
		};
	}}
>
	<AuthField
		label="Correo electrónico"
		id="login-email"
		name="email"
		type="email"
		autocomplete="email"
		placeholder="tu@correo.com"
		bind:value={loginEmail}
		error={errors['email']}
	/>

	<PasswordInput
		label="Contraseña"
		id="login-password"
		name="password"
		autocomplete="current-password"
		bind:value={loginPassword}
		error={errors['password']}
	>
		{#snippet action()}
			<a href={resolve('/auth/forgot-password')} class="lp-link">¿Olvidaste tu contraseña?</a>
		{/snippet}
	</PasswordInput>

	{#if errors['server']}
		<p class="lp-alert error" role="alert">
			{errors['server']}
			{#if form?.type === 'login' && form.unverified}
				<a href={resolve('/auth/verify-email')} class="lp-link">Reenviar enlace de verificación</a>
			{/if}
		</p>
	{/if}

	<button type="submit" class="lp-btn block" disabled={isSubmitting}>
		{isSubmitting ? 'Entrando…' : 'Iniciar sesión'}
	</button>
</form>

<style>
	.lp-alert .lp-link {
		display: block;
		margin-top: 6px;
		color: var(--lp-ink);
	}
</style>
