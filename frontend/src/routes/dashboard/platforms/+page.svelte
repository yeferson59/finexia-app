<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import PageHeader from '$lib/ui/page-header.svelte';
	import Card from '$lib/ui/card.svelte';
	import Pagination from '$lib/ui/pagination.svelte';
	import { PlatformCard } from '$lib/features/platforms';
	import type { PageProps } from './$types';

	const { data }: PageProps = $props();

	const PER_PAGE = 9;
	let page = $state(1);
	const pagedPlatforms = $derived(data.platforms.slice((page - 1) * PER_PAGE, page * PER_PAGE));

	function viewDetails(id: string) {
		goto(resolve('/dashboard/platforms/[id]', { id }));
	}

	function addNewPlatform() {
		goto(resolve('/dashboard/platforms/add'));
	}
</script>

<svelte:head>
	<title>Plataformas de Inversión - FINEXIA</title>
	<meta name="description" content="Gestiona tus plataformas de inversión" />
</svelte:head>

<PageHeader
	title="Plataformas de Inversión"
	subtitle="Administra todas tus plataformas y corredurías en un solo lugar."
>
	{#snippet actions()}
		<button onclick={addNewPlatform} class="btn-add-platform">
			<svg
				width="18"
				height="18"
				viewBox="0 0 24 24"
				fill="none"
				stroke="currentColor"
				stroke-width="2"
			>
				<path d="M12 5v14M5 12h14" />
			</svg>
			Agregar Plataforma
		</button>
	{/snippet}
</PageHeader>

<Card variant="elevated" padding="none">
	<div class="table-panel">
		<header class="table-head">
			<h2>Tus Plataformas</h2>
			<p class="platform-count">
				{data.platforms.length} plataforma{data.platforms.length !== 1 ? 's' : ''}
			</p>
		</header>

		{#if data.platforms.length === 0}
			<div class="empty-state">
				<svg
					width="64"
					height="64"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="1.5"
				>
					<rect x="3" y="3" width="18" height="18" rx="2" />
					<path d="M3 9h18" />
					<path d="M9 3v18" />
				</svg>
				<h3>No hay plataformas registradas</h3>
				<p>Comienza agregando tu primera plataforma de inversión</p>
				<button onclick={addNewPlatform} class="btn-empty-action">Agregar Plataforma</button>
			</div>
		{:else}
			<div class="platforms-grid">
				{#each pagedPlatforms as platform (platform.id)}
					<PlatformCard {platform} onView={viewDetails} />
				{/each}
			</div>
			<Pagination bind:page total={data.platforms.length} perPage={PER_PAGE} label="plataformas" />
		{/if}
	</div>
</Card>

<style>
	.btn-add-platform {
		display: flex;
		align-items: center;
		gap: 0.6rem;
		padding: 0.85rem 1.5rem;
		border: none;
		border-radius: 10px;
		background: var(--amber);
		color: #0d0800;
		font-weight: 700;
		font-family: var(--font-body);
		font-size: 0.95rem;
		cursor: pointer;
		transition: all 0.3s ease;
		letter-spacing: 0.3px;
		white-space: nowrap;
	}

	.btn-add-platform:hover {
		transform: translateY(-2px);
		box-shadow: 0 10px 25px rgba(212, 145, 42, 0.25);
	}

	.table-panel {
		padding: 1.5rem;
	}

	.table-head {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 1.5rem;
		border-bottom: 1px solid var(--border);
		padding-bottom: 1rem;
	}

	.table-head h2 {
		margin: 0;
		color: var(--text);
		font-size: 1.15rem;
	}

	.platform-count {
		margin: 0;
		color: rgba(236, 234, 229, 0.6);
		font-size: 0.9rem;
	}

	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 1rem;
		padding: 3rem 2rem;
		text-align: center;
		color: rgba(236, 234, 229, 0.6);
	}

	.empty-state svg {
		color: rgba(212, 145, 42, 0.4);
	}

	.empty-state h3 {
		margin: 0;
		color: var(--text);
		font-size: 1.2rem;
	}

	.empty-state p {
		margin: 0 0 0.5rem;
		color: rgba(236, 234, 229, 0.6);
	}

	.btn-empty-action {
		margin-top: 0.5rem;
		padding: 0.75rem 1.5rem;
		border: 1.5px solid var(--amber);
		border-radius: 8px;
		background: transparent;
		color: var(--amber);
		font-weight: 600;
		cursor: pointer;
		transition: all 0.3s ease;
	}

	.btn-empty-action:hover {
		background: var(--border);
		transform: translateY(-2px);
	}

	.platforms-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
		gap: 1.5rem;
		animation: fade-in 0.4s ease-out;
	}

	@keyframes fade-in {
		from {
			opacity: 0;
			transform: translateY(10px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	@media (max-width: 768px) {
		.btn-add-platform {
			width: 100%;
		}

		.platforms-grid {
			grid-template-columns: 1fr;
		}

		.table-head {
			flex-direction: column;
			align-items: flex-start;
			gap: 0.5rem;
		}
	}
</style>
