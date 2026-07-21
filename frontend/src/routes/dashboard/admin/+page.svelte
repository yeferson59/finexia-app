<script lang="ts">
	import PageHeader from '$lib/ui/page-header.svelte';
	import Card from '$lib/ui/card.svelte';
	import Stat from '$lib/ui/stat.svelte';
	import { resolve } from '$app/paths';
	import { goto } from '$app/navigation';

	import type { PageProps } from './$types';

	const { data }: PageProps = $props();

	function formatDate(iso: string | null): string {
		if (!iso) return 'Sin datos';
		return new Intl.DateTimeFormat('es', {
			dateStyle: 'medium',
			timeStyle: 'short'
		}).format(new Date(iso));
	}
</script>

<svelte:head>
	<title>Admin — FINEXIA</title>
</svelte:head>

<PageHeader
	eyebrow="Administración"
	title="Panel de Control"
	subtitle="Gestión y estado del sistema."
/>

<section class="stats-grid">
	<Card padding="md">
		<Stat label="Usuarios registrados" value={data.totalUsers} tone="highlight" />
	</Card>
	<Card padding="md">
		<Stat label="Activos en sistema" value={data.totalAssets} />
	</Card>
	<Card padding="md">
		<Stat label="Tasas de cambio" value={data.totalRates} />
	</Card>
	<Card padding="md">
		<Stat label="Última sincronización" value={formatDate(data.lastSync)} />
	</Card>
</section>

<section class="shortcuts">
	<h2 class="section-title">Accesos rápidos</h2>
	<div class="shortcut-grid">
		<Card padding="md" hover onclick={() => goto(resolve('/dashboard/admin/users'))}>
			<div class="shortcut-card">
				<svg
					width="24"
					height="24"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="1.5"
				>
					<path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
					<circle cx="9" cy="7" r="4"></circle>
					<path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
					<path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
				</svg>
				<div>
					<p class="shortcut-title">Gestionar Usuarios</p>
					<p class="shortcut-desc">Ver, crear y eliminar usuarios del sistema</p>
				</div>
			</div>
		</Card>
		<Card padding="md" hover onclick={() => goto(resolve('/dashboard/admin/assets'))}>
			<div class="shortcut-card">
				<svg
					width="24"
					height="24"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="1.5"
				>
					<ellipse cx="12" cy="5" rx="9" ry="3"></ellipse>
					<path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3"></path>
					<path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"></path>
				</svg>
				<div>
					<p class="shortcut-title">Gestionar Activos</p>
					<p class="shortcut-desc">Sincronizar y actualizar precios de activos</p>
				</div>
			</div>
		</Card>
		<Card padding="md" hover onclick={() => goto(resolve('/dashboard/admin/exchange-rates'))}>
			<div class="shortcut-card">
				<svg
					width="24"
					height="24"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="1.5"
				>
					<polyline points="17 1 21 5 17 9"></polyline>
					<path d="M3 11V9a4 4 0 0 1 4-4h14"></path>
					<polyline points="7 23 3 19 7 15"></polyline>
					<path d="M21 13v2a4 4 0 0 1-4 4H3"></path>
				</svg>
				<div>
					<p class="shortcut-title">Gestionar Tasas de Cambio</p>
					<p class="shortcut-desc">Sincronizar y actualizar tasas de cambio de divisas</p>
				</div>
			</div>
		</Card>
	</div>
</section>

<style>
	.stats-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 1.25rem;
		margin-bottom: 2.5rem;
	}

	.shortcuts {
		animation: fade-in 0.4s ease-out both;
		animation-delay: 0.1s;
	}

	.section-title {
		font-family: var(--font-mono);
		font-size: 0.625rem;
		font-weight: 500;
		text-transform: uppercase;
		letter-spacing: 0.2em;
		color: var(--text-dim);
		margin: 0 0 1rem 0;
	}

	.shortcut-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
		gap: 1rem;
	}

	.shortcut-card {
		display: flex;
		align-items: flex-start;
		gap: 1rem;
		color: var(--text-muted);
	}

	.shortcut-card svg {
		flex-shrink: 0;
		color: var(--amber);
		margin-top: 2px;
	}

	.shortcut-title {
		font-size: 0.95rem;
		font-weight: 600;
		color: var(--text);
		margin: 0 0 0.25rem 0;
	}

	.shortcut-desc {
		font-size: 0.82rem;
		color: var(--text-muted);
		margin: 0;
	}

	@keyframes fade-in {
		from {
			opacity: 0;
			transform: translateY(8px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}
</style>
