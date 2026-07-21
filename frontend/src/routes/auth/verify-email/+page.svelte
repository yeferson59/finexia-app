<script lang="ts">
	import { enhance } from '$app/forms';
	import { resolve } from '$app/paths';
	import Button from '$lib/ui/button.svelte';
	import Input from '$lib/ui/input.svelte';
	import type { PageProps } from './$types';

	const { data, form }: PageProps = $props();

	const authHref = resolve('/auth');
	let email = $state('');
	let submitting = $state(false);
	let confirmForm: HTMLFormElement | undefined = $state();

	function fieldError(field: string): string | undefined {
		if (!form || !('errors' in form) || !form.errors) return undefined;
		return (form.errors as Record<string, string>)[field];
	}

	$effect(() => {
		// The link requires no extra input from the user, so confirm it
		// automatically as soon as the page loads a valid token.
		if (data.valid && !form && confirmForm) {
			confirmForm.requestSubmit();
		}
	});
</script>

<svelte:head>
	<title>Verificar correo — FINEXIA</title>
	<meta name="robots" content="noindex" />
</svelte:head>

<main class="wrap">
	<div class="card">
		<div class="brand">
			<span class="brand-mark">◆</span>
			<span class="brand-name">FINEXIA</span>
		</div>

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

			{#if fieldError('server')}
				<p class="eyebrow">Verificación de correo</p>
				<h1 class="title">No se pudo verificar</h1>
				<p class="subtitle">{fieldError('server')}</p>
				<a class="back-link" href={authHref}>Volver a iniciar sesión</a>
			{:else}
				<p class="eyebrow">Verificación de correo</p>
				<h1 class="title">Verificando tu correo…</h1>
				<p class="subtitle">Un momento, estamos confirmando tu cuenta.</p>
			{/if}
		{:else if form?.resent}
			<p class="eyebrow">Enlace enviado</p>
			<h1 class="title">Revisa tu correo</h1>
			<p class="subtitle">
				Si <strong>{email}</strong> tiene una cuenta sin verificar, te enviamos un nuevo enlace de verificación.
				El enlace caduca pronto, así que úsalo cuanto antes.
			</p>
			<a class="back-link" href={authHref}>Volver a iniciar sesión</a>
		{:else}
			<p class="eyebrow">Enlace no válido</p>
			<h1 class="title">{data.reason}</h1>
			<p class="subtitle">Ingresa tu email y te enviaremos un nuevo enlace de verificación.</p>

			<form
				method="POST"
				action="?/resend"
				use:enhance={() => {
					submitting = true;
					return async ({ update }) => {
						submitting = false;
						await update({ reset: false });
					};
				}}
			>
				<div class="fields">
					<Input
						label="Email"
						name="email"
						type="email"
						placeholder="tu@email.com"
						bind:value={email}
						required
						error={fieldError('email')}
					/>
				</div>

				<Button type="submit" loading={submitting} class="submit">Reenviar enlace</Button>
			</form>

			<a class="back-link" href={authHref}>Volver a iniciar sesión</a>
		{/if}
	</div>
</main>

<style>
	.wrap {
		min-height: 100vh;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 1.5rem;
		background: var(--bg, #08090a);
	}

	.card {
		width: 100%;
		max-width: 26rem;
		background: var(--surface, #0e1011);
		border: 1px solid var(--border, #1e2023);
		border-radius: 1rem;
		padding: 2.25rem 2rem;
	}

	.brand {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		margin-bottom: 1.75rem;
	}

	.brand-mark {
		color: var(--amber, #d4912a);
		font-size: 1rem;
	}

	.brand-name {
		font-weight: 600;
		letter-spacing: 0.12em;
		font-size: 0.95rem;
		color: var(--text, #eceae5);
	}

	.eyebrow {
		font-family: var(--font-mono, monospace);
		font-size: 0.68rem;
		font-weight: 500;
		letter-spacing: 0.2em;
		text-transform: uppercase;
		color: var(--amber, #d4912a);
		margin: 0 0 0.5rem 0;
	}

	.title {
		font-size: 1.5rem;
		font-weight: 600;
		color: var(--text, #eceae5);
		margin: 0 0 0.5rem 0;
	}

	.subtitle {
		font-size: 0.9rem;
		color: var(--text-dim, #8a8780);
		line-height: 1.6;
		margin: 0 0 1.75rem 0;
	}

	.fields {
		display: flex;
		flex-direction: column;
		gap: 1rem;
		margin-bottom: 1.25rem;
	}

	:global(.submit) {
		width: 100%;
	}

	.back-link {
		display: inline-block;
		margin-top: 1.5rem;
		font-size: 0.85rem;
		color: var(--amber, #d4912a);
		text-decoration: none;
	}

	.back-link:hover {
		text-decoration: underline;
	}
</style>
