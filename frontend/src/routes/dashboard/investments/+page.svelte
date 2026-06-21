<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import PageHeader from '$components/ui/page-header.svelte';
	import Card from '$components/ui/card.svelte';
	import { investmentStore } from '$lib/stores/investments.svelte';

	const opportunities = $derived(investmentStore.items);

	function viewDetails(id: string) {
		goto(resolve('/dashboard/investments/[id]', { id }));
	}

	function addNewProduct() {
		goto(resolve('/dashboard/investments/add'));
	}
</script>

<svelte:head>
	<title>Inversiones - FINEXIA</title>
	<meta name="description" content="Oportunidades y estrategias de inversión FINEXIA" />
</svelte:head>

<PageHeader
	title="Inversiones"
	subtitle="Descubre oportunidades alineadas con tu perfil y objetivos."
>
	{#snippet actions()}
		<button onclick={addNewProduct} class="btn-add-product">
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
			Agregar Producto
		</button>
	{/snippet}
</PageHeader>

<div class="panel-stack">
	<Card variant="elevated" padding="none">
		<div class="spotlight">
			<div>
				<p class="eyebrow">Recomendación destacada</p>
				<h2>Portafolio crecimiento equilibrado</h2>
				<p>
					Estrategia diseñada para optimizar rendimiento en un escenario mixto con sesgo
					tecnológico.
				</p>
			</div>
			<button class="action">Explorar estrategia</button>
		</div>
	</Card>

	<Card variant="elevated" padding="none">
		<div class="table-panel">
			<header class="table-head">
				<h2>Oportunidades activas</h2>
			</header>
			<div class="table">
				<div class="row heading">
					<span>Instrumento</span>
					<span>Riesgo</span>
					<span>ROI esperado</span>
					<span>Horizonte</span>
					<span></span>
				</div>
				{#each opportunities as item (item.id)}
					<button
						class="row row-interactive"
						onclick={() => viewDetails(item.id)}
						aria-label={`Ver detalles de ${item.name}`}
					>
						<span>{item.name}</span>
						<span>{item.riskLevel}</span>
						<span class="positive">{item.expectedROI}%</span>
						<span>{item.horizon} meses</span>
						<span class="row-icon">→</span>
					</button>
				{/each}
			</div>
		</div>
	</Card>
</div>

<style>
	.panel-stack {
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	.btn-add-product {
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

	.btn-add-product:hover {
		transform: translateY(-2px);
		box-shadow: 0 10px 25px rgba(212, 145, 42, 0.25);
	}

	.spotlight {
		display: flex;
		align-items: flex-end;
		justify-content: space-between;
		gap: 1rem;
		padding: 1.5rem;
	}

	.eyebrow {
		margin: 0 0 0.6rem;
		font-size: 0.72rem;
		letter-spacing: 0.7px;
		text-transform: uppercase;
		color: rgba(236, 234, 229, 0.46);
	}

	.spotlight h2 {
		margin: 0;
		color: var(--amber-light);
	}

	.spotlight p {
		margin: 0.6rem 0 0;
		color: rgba(236, 234, 229, 0.62);
		max-width: 62ch;
	}

	.action {
		border: none;
		border-radius: 8px;
		padding: 0.8rem 1.2rem;
		background: var(--amber);
		color: #0d0800;
		font-weight: 700;
		cursor: pointer;
	}

	.table-panel {
		padding: 1.3rem;
	}

	.table-head h2 {
		margin: 0 0 1rem;
		color: var(--text);
		font-size: 1.15rem;
	}

	.table {
		display: grid;
		gap: 0.55rem;
	}

	.row {
		display: grid;
		grid-template-columns: 1.4fr 0.8fr 0.8fr 0.8fr 0.2fr;
		gap: 0.7rem;
		padding: 0.85rem;
		border-radius: 10px;
		background: rgba(255, 255, 255, 0.022);
	}

	.row.heading {
		background: rgba(212, 145, 42, 0.12);
		font-size: 0.75rem;
		font-weight: 700;
		letter-spacing: 0.5px;
		text-transform: uppercase;
		color: rgba(236, 234, 229, 0.75);
	}

	.row.row-interactive {
		border: none;
		cursor: pointer;
		transition: all 0.3s ease;
		text-align: left;
		font-family: inherit;
		font-size: inherit;
		color: inherit;
	}

	.row.row-interactive:hover {
		background: var(--border-strong);
		transform: translateX(4px);
	}

	.row-icon {
		display: flex;
		align-items: center;
		justify-content: center;
		color: var(--amber);
		font-weight: 700;
		opacity: 0;
		transition: opacity 0.3s ease;
	}

	.row.row-interactive:hover .row-icon {
		opacity: 1;
	}

	.row span {
		font-size: 0.85rem;
		color: var(--text);
	}

	.positive {
		color: var(--green);
	}

	@media (max-width: 768px) {
		.btn-add-product {
			width: 100%;
		}

		.row {
			grid-template-columns: 1fr;
		}

		.row-icon {
			display: none;
		}
	}
</style>
