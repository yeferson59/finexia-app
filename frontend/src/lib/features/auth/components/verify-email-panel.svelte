<script lang="ts">
	import { enhance } from '$app/forms';
	import { resolve } from '$app/paths';
	import PublicShell from '$lib/ui/public-shell.svelte';
	import AuthField from './auth-field.svelte';

	type VerifyData =
		{ valid: false; reason: string } | { valid: true; token: string; reason?: string };

	let {
		data,
		form
	}: {
		data: VerifyData;
		form: { resent?: boolean; errors?: Record<string, string> } | null;
	} = $props();

	const authHref = resolve('/auth');
	let email = $state('');
	let submitting = $state(false);
	let confirmForm: HTMLFormElement | undefined = $state();

	const serverError = $derived(form?.errors?.server);

	$effect(() => {
		// The link requires no extra input from the user, so confirm it
		// automatically as soon as the page loads a valid token.
		if (data.valid && !form && confirmForm) {
			confirmForm.requestSubmit();
		}
	});
</script>

<PublicShell>
	{#snippet heading()}
		{#if data.valid && serverError}
			<h1 class="shell-title">No pudimos verificar tu correo</h1>
		{:else if data.valid}
			<h1 class="shell-title">Verificando tu correo</h1>
		{:else if form?.resent}
			<h1 class="shell-title">Revisa tu correo</h1>
		{:else}
			<h1 class="shell-title">Verifica tu correo</h1>
		{/if}
	{/snippet}

	{#if data.valid}
		<form
			method="POST"
			action="?/confirm"
			bind:this={confirmForm}
			use:enhance={() => {
				submitting = true;
				return async ({ update }) => {
					submitting = false;
					await update({ reset: false });
				};
			}}
		>
			<input type="hidden" name="token" value={data.token} />
		</form>

		{#if serverError}
			<p class="lp-alert error" role="alert">{serverError}</p>
			<p class="shell-after">
				<a class="lp-link" href={resolve('/auth/verify-email')}>Pedir un enlace nuevo</a>
			</p>
		{:else}
			<p class="shell-copy" role="status">
				Un momento: estamos confirmando tu cuenta. Al terminar te llevamos a iniciar sesión.
			</p>
		{/if}
	{:else if form?.resent}
		<p class="shell-copy" role="status">
			Si <strong>{email}</strong> tiene una cuenta sin verificar, te enviamos un enlace nuevo. Caduca
			pronto, así que úsalo cuanto antes.
		</p>
		<p class="shell-after"><a class="lp-link" href={authHref}>Volver a iniciar sesión</a></p>
	{:else}
		<p class="shell-copy">{data.reason}</p>

		<form
			method="POST"
			action="?/resend"
			class="lp-fields resend"
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
				id="verify-email"
				name="email"
				type="email"
				autocomplete="email"
				placeholder="tu@correo.com"
				bind:value={email}
				error={form?.errors?.email}
			/>

			<button type="submit" class="lp-btn block" disabled={submitting}>
				{submitting ? 'Enviando…' : 'Enviar enlace'}
			</button>
		</form>

		<p class="shell-after"><a class="lp-link" href={authHref}>Volver a iniciar sesión</a></p>
	{/if}
</PublicShell>

<style>
	.resend {
		margin-top: 28px;
	}
</style>
