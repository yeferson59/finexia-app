<script lang="ts">
	import { resolve } from '$app/paths';
	import PageHeader from '$lib/ui/page-header.svelte';
	import Button from '$lib/ui/button.svelte';
	import EmptyState from '$lib/ui/empty-state.svelte';
	import { todayLocalDateString } from '$lib/shared/format/date';
	import { FundCreateForm, FundList, FundMarkForm, type Fund } from '$lib/features/funds';
	import type { PageProps } from './$types';

	const { data }: PageProps = $props();

	const today = todayLocalDateString();

	/* Sin portafolio o sin plataforma no hay dónde poner el fondo. */
	const canCreate = $derived(data.portfolios.length > 0 && data.platforms.length > 0);

	let creating = $state(false);

	/* El fondo cuyo valor se está actualizando. Se guarda el id, no el fondo: tras
	   guardar, la página se recarga y el diálogo lee la versión nueva. */
	let markingId = $state<string | null>(null);
	const marking = $derived<Fund | null>(data.funds.find((f) => f.assetId === markingId) ?? null);
</script>

<svelte:head>
	<title>Fondos - FINEXIA</title>
	<meta
		name="description"
		content="Tus fondos de inversión: lo que valen según tu extracto y lo que ganan"
	/>
</svelte:head>

<PageHeader
	title="Fondos"
	subtitle="Fondos de inversión colectiva, de pensiones voluntarias y money market: rinden lo que resulte, no una tasa fija."
>
	{#snippet actions()}
		<Button type="button" onclick={() => (creating = true)} disabled={!canCreate}>
			Registrar fondo
		</Button>
	{/snippet}
</PageHeader>

{#if !canCreate}
	<p class="feedback prereq">
		Para registrar un fondo necesitas
		{#if data.portfolios.length === 0}
			un <a href={resolve('/dashboard/portfolios/add')}>portafolio</a>{data.platforms.length === 0
				? ' y '
				: ''}
		{/if}
		{#if data.platforms.length === 0}
			una <a href={resolve('/dashboard/platforms/add')}>plataforma activa</a>
		{/if}
		donde guardarlo.
	</p>
{/if}

{#if data.loadFailed}
	<p class="feedback error">No pudimos cargar tus fondos. Vuelve a intentarlo en un momento.</p>
{:else if data.funds.length === 0}
	<EmptyState
		bordered
		title="Todavía no hay fondos"
		description="Registra un fondo con las unidades y el valor de unidad de tu extracto. Cuando lo actualices, lo que gane o pierda contará en tu rentabilidad."
	>
		{#snippet action()}
			{#if canCreate}
				<Button type="button" onclick={() => (creating = true)}>Registrar el primer fondo</Button>
			{/if}
		{/snippet}
	</EmptyState>
{:else}
	<p class="rule">
		Un fondo vale sus unidades por el último valor de unidad que escribiste. Finexia no estima lo
		que pasa entre un extracto y otro: actualízalo con la fecha del extracto y la gráfica de
		crecimiento se corrige desde ese día.
	</p>
	<FundList
		funds={data.funds}
		{today}
		showPortfolio={data.portfolios.length > 1}
		onMark={(fund) => (markingId = fund.assetId)}
	/>
{/if}

<FundCreateForm
	open={creating}
	portfolios={data.portfolios}
	platforms={data.platforms}
	currency={data.currency}
	onClose={() => (creating = false)}
/>

<FundMarkForm
	fund={marking}
	marks={marking ? (data.marks[marking.assetId] ?? []) : []}
	onClose={() => (markingId = null)}
/>

<style>
	.prereq {
		margin: 0 0 1.5rem;
	}

	.prereq a {
		color: var(--amber);
	}

	.rule {
		max-width: 70ch;
		margin: 0 0 1.25rem;
		font-size: 0.84rem;
		line-height: 1.5;
		color: var(--text-muted);
	}
</style>
