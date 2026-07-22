<script lang="ts">
	import { ResetPasswordForm } from '$lib/features/auth';
	import type { PageProps } from './$types';

	const { data, form }: PageProps = $props();
</script>

<svelte:head>
	<title>Restablecer contraseña — FINEXIA</title>
	<meta name="robots" content="noindex" />
</svelte:head>

<ResetPasswordForm {data} {form} />
