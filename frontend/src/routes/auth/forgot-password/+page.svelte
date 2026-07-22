<script lang="ts">
	import { ForgotPasswordForm } from '$lib/features/auth';
	import type { PageProps } from './$types';

	const { form }: PageProps = $props();
</script>

<svelte:head>
	<title>Recuperar contraseña — FINEXIA</title>
	<meta name="robots" content="noindex" />
</svelte:head>

<ForgotPasswordForm {form} />
