<script lang="ts">
	import { enhance } from '$app/forms';
	import { resolve } from '$app/paths';
	import Button from '$components/ui/button.svelte';
	import Input from '$components/ui/input.svelte';
	import type { PageProps } from './$types';

	const { form }: PageProps = $props();

	const authHref = resolve('/auth');
	let email = $state('');
	let submitting = $state(false);

	function fieldError(field: string): string | undefined {
		if (!form || !('errors' in form) || !form.errors) return undefined;
		return (form.errors as Record<string, string>)[field];
	}
</script>

<svelte:head>
	<title>Recuperar contraseña — FINEXIA</title>
	<meta name="robots" content="noindex" />
</svelte:head>

<main class="wrap">
	<div class="card">
		<div class="brand">
			<span class="brand-mark">◆</span>
			<span class="brand-name">FINEXIA</span>
		</div>

		{#if form?.sent}
			<p class="eyebrow">Enlace enviado</p>
			<h1 class="title">Revisa tu correo</h1>
			<p class="subtitle">
				Si <strong>{email}</strong> tiene una cuenta con nosotros, te enviamos un enlace para restablecer
				tu contraseña. El enlace caduca pronto, así que úsalo cuanto antes.
			</p>
			<a class="back-link" href={authHref}>Volver a iniciar sesión</a>
		{:else}
			<p class="eyebrow">Recuperar contraseña</p>
			<h1 class="title">¿Olvidaste tu contraseña?</h1>
			<p class="subtitle">
				Ingresa el email de tu cuenta y te enviaremos un enlace para crear una nueva contraseña.
			</p>

			<form
				method="POST"
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

				<Button type="submit" loading={submitting} class="submit">Enviar enlace</Button>
			</form>

			<a class="back-link" href={authHref}>Ya recuerdo mi contraseña</a>
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
