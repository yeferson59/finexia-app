<script lang="ts">
	import { VerifyEmailPanel } from '$lib/features/auth';
	import type { PageProps } from './$types';

	const { data, form }: PageProps = $props();
</script>

<svelte:head>
	<title>Verificar correo — Finexia</title>
	<meta name="robots" content="noindex" />
</svelte:head>

<VerifyEmailPanel {data} {form} />
