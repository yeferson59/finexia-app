<script lang="ts">
	import { enhance } from '$app/forms';
	import { resolve } from '$app/paths';
	import PublicShell from '$lib/ui/public-shell.svelte';
	import AuthField from './auth-field.svelte';
	import PasswordInput from './password-input.svelte';
	import { parseErrors } from '../utils';

	type InviteData =
		| { valid: false; reason: string }
		| { valid: true; token: string; email: string; name: string; reason?: undefined };

	let {
		data,
		form
	}: {
		data: InviteData;
		form: {
			errors?: Record<string, string> | Array<{ path: PropertyKey[]; message: string }>;
		} | null;
	} = $props();

	let name = $derived(data.valid ? data.name : '');
	let password = $state('');
	let confirmPassword = $state('');
	let submitting = $state(false);

	const errors = $derived(parseErrors(form?.errors));
</script>

<PublicShell>
	{#snippet heading()}
		{#if data.valid}
			<h1 class="shell-title">Activa tu cuenta</h1>
			<p class="shell-note">
				Te invitaron con <strong>{data.email}</strong>. Solo falta tu nombre y una contraseña.
			</p>
		{:else}
			<h1 class="shell-title">Esta invitación ya no sirve</h1>
		{/if}
	{/snippet}

	{#if !data.valid}
		<p class="shell-copy">{data.reason}</p>
		<p class="shell-after">
			¿Ya activaste tu cuenta? <a class="lp-link" href={resolve('/auth')}>Iniciar sesión</a>
		</p>
	{:else}
		<form
			method="POST"
			action="?/accept"
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

			<AuthField
				label="Nombre"
				id="invite-name"
				name="name"
				autocomplete="name"
				bind:value={name}
				error={errors['name']}
			/>
			<PasswordInput
				label="Contraseña"
				id="invite-password"
				name="password"
				autocomplete="new-password"
				hint="Entre 8 y 20 caracteres."
				bind:value={password}
				error={errors['password']}
			/>
			<PasswordInput
				label="Repite la contraseña"
				id="invite-confirm"
				name="confirmPassword"
				autocomplete="new-password"
				bind:value={confirmPassword}
				error={errors['confirmPassword']}
			/>

			{#if errors['server']}
				<p class="lp-alert error" role="alert">{errors['server']}</p>
			{/if}

			<button type="submit" class="lp-btn block" disabled={submitting}>
				{submitting ? 'Activando…' : 'Activar cuenta'}
			</button>
		</form>

		<p class="shell-after">
			¿Ya la activaste? <a class="lp-link" href={resolve('/auth')}>Iniciar sesión</a>
		</p>
	{/if}
</PublicShell>
