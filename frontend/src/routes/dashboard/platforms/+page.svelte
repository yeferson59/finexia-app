<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import PageHeader from '$lib/ui/page-header.svelte';
	import Button from '$lib/ui/button.svelte';
	import Card from '$lib/ui/card.svelte';
	import DataTable from '$lib/ui/data-table.svelte';
	import EmptyState from '$lib/ui/empty-state.svelte';
	import Pagination from '$lib/ui/pagination.svelte';
	import { PlatformAllocation, PlatformRow, rankByShare } from '$lib/features/platforms';
	import type { PageProps } from './$types';

	const { data }: PageProps = $props();

	// De mayor a menor: el orden es la primera cosa que cuenta esta pantalla.
	const ranked = $derived(rankByShare(data.platforms));

	const unconverted = $derived(
		data.platforms.reduce((sum, p) => sum + (p.positionsUnconverted ?? 0), 0)
	);

	// Contra un backend anterior a estas métricas no la informa ninguna, y una
	// columna de rayas parece un dato que falta en vez de uno que no existe.
	const showGain = $derived(data.platforms.some((p) => p.gainLoss !== undefined));

	// Una fila ocupa menos que una tarjeta, así que caben más antes de paginar.
	const PER_PAGE = 12;
	let page = $state(1);
	const pagedPlatforms = $derived(ranked.slice((page - 1) * PER_PAGE, page * PER_PAGE));

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

<PageHeader title="Plataformas" subtitle="Dónde está tu dinero y qué parte guarda cada sitio.">
	{#snippet actions()}
		<Button type="button" onclick={addNewPlatform}>Crear plataforma</Button>
	{/snippet}
</PageHeader>

{#if ranked.length === 0}
	<Card variant="elevated" padding="md">
		<EmptyState
			title="Todavía no hay plataformas"
			description="Registra el bróker, la casa de bolsa o la billetera donde tienes tu dinero y aparecerá aquí con lo que guarda."
		>
			{#snippet action()}
				<Button type="button" onclick={addNewPlatform}>Crear la primera</Button>
			{/snippet}
		</EmptyState>
	</Card>
{:else}
	<PlatformAllocation {ranked} currency={data.currency} {unconverted} />

	<Card variant="elevated" padding="none">
		<DataTable caption="Plataformas ordenadas por lo que guardan">
			<thead>
				<tr>
					<th scope="col">Plataforma</th>
					<th scope="col" class="num col-positions">Posiciones</th>
					<th scope="col" class="num">Invertido</th>
					{#if showGain}
						<th scope="col" class="num">Ganancia</th>
					{/if}
				</tr>
			</thead>
			<tbody>
				{#each pagedPlatforms as entry (entry.platform.id)}
					<PlatformRow {entry} count={ranked.length} {showGain} onView={viewDetails} />
				{/each}
			</tbody>
		</DataTable>

		<div class="pager">
			<Pagination bind:page total={ranked.length} perPage={PER_PAGE} label="plataformas" />
		</div>
	</Card>
{/if}

<style>
	.pager {
		padding: 0 1.25rem;
	}

	/* Su columna se esconde en estrecho —el importe es lo que no puede cortarse—
	   y el contador se lee en la línea bajo el nombre. */
	@media (max-width: 720px) {
		.col-positions {
			display: none;
		}

		.pager {
			padding: 0 0.9rem;
		}
	}
</style>
