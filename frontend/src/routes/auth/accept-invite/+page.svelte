<script lang="ts">
	import { AcceptInviteForm } from '$lib/features/auth';
	import type { PageProps } from './$types';

	const { data, form }: PageProps = $props();
</script>

<svelte:head>
	<title>Activar cuenta — FINEXIA</title>
	<meta name="robots" content="noindex" />
</svelte:head>

<AcceptInviteForm {data} {form} />
