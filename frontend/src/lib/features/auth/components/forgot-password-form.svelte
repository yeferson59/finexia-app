<script lang="ts">
	import { enhance } from '$app/forms';
	import { resolve } from '$app/paths';
	import PublicShell from '$lib/ui/public-shell.svelte';
	import AuthField from './auth-field.svelte';

	let { form }: { form: { sent?: boolean; errors?: Record<string, string> } | null } = $props();

	const authHref = resolve('/auth');
	let email = $state('');
	let submitting = $state(false);

	function fieldError(field: string): string | undefined {
		if (!form || !('errors' in form) || !form.errors) return undefined;
		return (form.errors as Record<string, string>)[field];
	}
</script>

<PublicShell>
	{#snippet heading()}
		{#if form?.sent}
			<h1 class="shell-title">Revisa tu correo</h1>
		{:else}
			<h1 class="shell-title">Recupera tu contraseña</h1>
			<p class="shell-note">Te enviamos un enlace para que elijas una nueva.</p>
		{/if}
	{/snippet}

	{#if form?.sent}
		<p class="shell-copy" role="status">
			Si <strong>{email}</strong> tiene una cuenta en Finexia, te acabamos de enviar el enlace. Caduca
			pronto, así que úsalo cuanto antes.
		</p>
		<p class="shell-copy">¿No llega? Mira en la carpeta de spam antes de pedir otro.</p>
		<p class="shell-after"><a class="lp-link" href={authHref}>Volver a iniciar sesión</a></p>
	{:else}
		<form
			method="POST"
			class="lp-fields"
			use:enhance={() => {
				submitting = true;
				return async ({ update }) => {
					submitting = false;
					await update({ reset: false });
				};
			}}
		>
			<AuthField
				label="Correo de tu cuenta"
				id="forgot-email"
				name="email"
				type="email"
				autocomplete="email"
				placeholder="tu@correo.com"
				bind:value={email}
				error={fieldError('email')}
			/>

			<button type="submit" class="lp-btn block" disabled={submitting}>
				{submitting ? 'Enviando…' : 'Enviar enlace'}
			</button>
		</form>

		<p class="shell-after">
			¿Ya la recuerdas? <a class="lp-link" href={authHref}>Iniciar sesión</a>
		</p>
	{/if}
</PublicShell>
