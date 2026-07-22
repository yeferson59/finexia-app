<script lang="ts">
	import { PlatformDetail, type Platform } from '$lib/features/platforms';
	import type { PageProps } from './$types';

	const { data }: PageProps = $props();

	const platform = $derived(data.platform as unknown as Platform);
</script>

<svelte:head>
	<title>{platform.name} - FINEXIA</title>
	<meta name="description" content={`Detalles de ${platform.name}`} />
</svelte:head>

<PlatformDetail {platform} />
