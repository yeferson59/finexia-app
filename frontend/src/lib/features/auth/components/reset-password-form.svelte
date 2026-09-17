<script lang="ts">
	import { enhance } from '$app/forms';
	import { resolve } from '$app/paths';
	import PublicShell from '$lib/ui/public-shell.svelte';
	import PasswordInput from './password-input.svelte';
	import { parseErrors } from '../utils';

	type ResetData =
		{ valid: false; reason: string } | { valid: true; token: string; reason?: undefined };

	let {
		data,
		form
	}: {
		data: ResetData;
		form: {
			errors?: Record<string, string> | Array<{ path: PropertyKey[]; message: string }>;
		} | null;
	} = $props();

	let password = $state('');
	let confirmPassword = $state('');
	let submitting = $state(false);

	const errors = $derived(parseErrors(form?.errors));
</script>

<PublicShell>
	{#snippet heading()}
		{#if data.valid}
			<h1 class="shell-title">Elige una contraseña nueva</h1>
			<p class="shell-note">Al guardarla se cierran todas tus sesiones abiertas.</p>
		{:else}
			<h1 class="shell-title">Este enlace ya no sirve</h1>
		{/if}
	{/snippet}

	{#if !data.valid}
		<p class="shell-copy">{data.reason}</p>
		<p class="shell-after">
			<a class="lp-link" href={resolve('/auth/forgot-password')}>Pedir un enlace nuevo</a>
		</p>
	{:else}
		<form
			method="POST"
			action="?/confirm"
			class="lp-fields"
			use:enhance={() => {
				submitting = true;
				return async ({ update }) => {
					submitting = false;
					await update();
				};
			}}
		>
			<input type="hidden" name="token" value={data.token} />

			<PasswordInput
				label="Contraseña nueva"
				id="reset-password"
				name="password"
				autocomplete="new-password"
				hint="Entre 8 y 20 caracteres."
				bind:value={password}
				error={errors['password']}
			/>
			<PasswordInput
				label="Repite la contraseña"
				id="reset-confirm"
				name="confirmPassword"
				autocomplete="new-password"
				bind:value={confirmPassword}
				error={errors['confirmPassword']}
			/>

			{#if errors['server']}
				<p class="lp-alert error" role="alert">{errors['server']}</p>
			{/if}

			<button type="submit" class="lp-btn block" disabled={submitting}>
				{submitting ? 'Guardando…' : 'Guardar contraseña'}
			</button>
		</form>

		<p class="shell-after">
			<a class="lp-link" href={resolve('/auth')}>Volver a iniciar sesión</a>
		</p>
	{/if}
</PublicShell>
