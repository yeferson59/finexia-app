<script lang="ts">
	import { enhance } from '$app/forms';
	import { resolve } from '$app/paths';
	import Button from '$lib/ui/button.svelte';
	import Input from '$lib/ui/input.svelte';

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

	const authHref = resolve('/auth');
	let name = $derived(data.valid ? data.name : '');
	let password = $state('');
	let confirmPassword = $state('');
	let submitting = $state(false);

	function fieldError(field: string): string | undefined {
		if (!form || !('errors' in form) || !form.errors) return undefined;
		if (Array.isArray(form.errors)) {
			return form.errors.find((i) => i.path?.[0] === field)?.message;
		}
		return (form.errors as Record<string, string>)[field];
	}
</script>

<main class="wrap">
	<div class="card">
		<div class="brand">
			<span class="brand-mark">◆</span>
			<span class="brand-name">FINEXIA</span>
		</div>

		{#if !data.valid}
			<h1 class="title">Invitación no válida</h1>
			<p class="subtitle">{data.reason}</p>
			<a class="back-link" href={authHref}>Ir a iniciar sesión</a>
		{:else}
			<p class="eyebrow">Invitación</p>
			<h1 class="title">Activa tu cuenta</h1>
			<p class="subtitle">
				Estás activando el acceso para <strong>{data.email}</strong>. Crea una contraseña para
				entrar.
			</p>

			<form
				method="POST"
				action="?/accept"
				use:enhance={() => {
					submitting = true;
					return async ({ update }) => {
						submitting = false;
						await update();
					};
				}}
			>
				<input type="hidden" name="token" value={data.token} />

				<div class="fields">
					<Input label="Nombre" name="name" bind:value={name} required error={fieldError('name')} />
					<Input
						label="Contraseña"
						name="password"
						type="password"
						bind:value={password}
						required
						error={fieldError('password')}
					/>
					<Input
						label="Confirmar contraseña"
						name="confirmPassword"
						type="password"
						bind:value={confirmPassword}
						required
						error={fieldError('confirmPassword')}
					/>
				</div>

				<p class="hint">Entre 8 y 20 caracteres.</p>

				{#if fieldError('server')}
					<p class="server-error">{fieldError('server')}</p>
				{/if}

				<Button type="submit" loading={submitting} class="submit">Crear cuenta y entrar</Button>
			</form>

			<a class="back-link" href={authHref}>Ya tengo cuenta</a>
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
	}

	.hint {
		font-size: 0.78rem;
		color: var(--text-dim, #8a8780);
		margin: 0.6rem 0 1rem 0;
	}

	.server-error {
		font-size: 0.82rem;
		color: var(--red, #ef4444);
		margin: 0 0 1rem 0;
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
