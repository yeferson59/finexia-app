<script lang="ts">
	/*
	 * El segundo paso del login. El titular —«Verificación en dos pasos»— lo
	 * pone el contenedor, en la columna de la izquierda; aquí va solo el campo.
	 */
	import { enhance } from '$app/forms';
	import { resolve } from '$app/paths';
	import AuthField from './auth-field.svelte';
	import type { AuthActionResult } from '../types';
	import { parseErrors } from '../utils';

	let { form, token }: { form: AuthActionResult; token: string } = $props();

	let twoFactorCode = $state('');
	let isSubmitting = $state(false);

	const errors = $derived(form?.type === 'login' ? parseErrors(form.errors) : {});
</script>

<form
	method="POST"
	action="?/twoFactor"
	class="lp-fields"
	id="two-factor-form"
	use:enhance={() => {
		isSubmitting = true;
		return async ({ update }) => {
			await update({ reset: false });
			isSubmitting = false;
		};
	}}
>
	<input type="hidden" name="token" value={token} />

	<AuthField
		label="Código de verificación"
		id="two-factor-code"
		name="code"
		autocomplete="one-time-code"
		placeholder="123456"
		hint="Los 6 dígitos de la app, o uno de tus códigos de recuperación si no tienes el teléfono."
		bind:value={twoFactorCode}
		error={errors['code']}
	/>

	{#if errors['server']}
		<p class="lp-alert error" role="alert">{errors['server']}</p>
	{/if}

	<button type="submit" class="lp-btn block" disabled={isSubmitting}>
		{isSubmitting ? 'Verificando…' : 'Verificar'}
	</button>

	<p class="restart">
		<a href={resolve('/auth')} class="lp-link" data-sveltekit-reload>Entrar con otra cuenta</a>
	</p>
</form>

<style>
	.restart {
		margin: 0;
		font-size: 15px;
	}
</style>
