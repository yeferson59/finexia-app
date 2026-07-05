<script lang="ts">
	import { page } from '$app/state';
	import LoginRegister from '$components/auth/login-register.svelte';
	import type { ActionData, PageData } from './$types';

	let { form, data }: { form: ActionData; data: PageData } = $props();

	const notice = $derived(
		page.url.searchParams.has('registered')
			? 'Cuenta creada. Revisa tu correo para verificar tu cuenta antes de iniciar sesión.'
			: page.url.searchParams.has('verified')
				? 'Correo verificado. Ya puedes iniciar sesión.'
				: undefined
	);
</script>

<svelte:head>
	<title>Iniciar sesión - FINEXIA</title>
	<meta name="description" content="Inicia sesión o crea una cuenta en FINEXIA" />
</svelte:head>

{#if notice}
	<p class="auth-notice" role="status">{notice}</p>
{/if}

<LoginRegister {form} selfRegistrationEnabled={data.selfRegistrationEnabled} />

<style>
	.auth-notice {
		position: relative;
		z-index: 20;
		max-width: 32rem;
		margin: 1.5rem auto 0;
		padding: 0.85rem 1.25rem;
		background: rgba(34, 201, 126, 0.08);
		border: 1px solid rgba(34, 201, 126, 0.25);
		border-radius: 0.75rem;
		color: var(--text, #eceae5);
		font-size: 0.85rem;
		text-align: center;
	}
</style>
