<script lang="ts">
	import { enhance } from '$app/forms';
	import Button from '$lib/ui/button.svelte';
	import Input from '$lib/ui/input.svelte';
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
	class="form-content"
	id="two-factor-form"
	use:enhance={() => {
		isSubmitting = true;
		return async ({ update }) => {
			await update({ reset: false });
			isSubmitting = false;
		};
	}}
>
	<p class="two-factor-title">Verificación en dos pasos</p>
	<p class="two-factor-copy">
		Ingresa el código de 6 dígitos de tu aplicación de autenticación o uno de tus códigos de
		recuperación.
	</p>

	<input type="hidden" name="token" value={token} />

	<Input
		label="Código de verificación"
		id="two-factor-code"
		name="code"
		type="text"
		placeholder="123456"
		autocomplete="one-time-code"
		bind:value={twoFactorCode}
		error={errors['code']}
		required
	/>

	{#if errors['server']}
		<p class="error-server" role="alert">{errors['server']}</p>
	{/if}

	<Button type="submit" variant="primary" size="lg" loading={isSubmitting} fullWidth>
		{isSubmitting ? 'Verificando...' : 'Verificar'}
	</Button>
</form>

<style>
	.two-factor-title {
		font-size: 1.1rem;
		font-weight: 700;
		color: var(--text-primary);
		margin: 0;
	}

	.two-factor-copy {
		font-size: 0.9rem;
		line-height: 1.6;
		color: var(--text-secondary);
		margin: 0;
	}
</style>
