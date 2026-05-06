<script lang="ts">
	const opportunities = [
		{ name: 'Tecnología IA', risk: 'Medio', roi: '12.4%', horizon: '18 meses' },
		{ name: 'Energía Renovable', risk: 'Bajo', roi: '8.1%', horizon: '24 meses' },
		{ name: 'Mercados Emergentes', risk: 'Alto', roi: '16.8%', horizon: '36 meses' }
	];
</script>

<svelte:head>
	<title>Inversiones - FINEXIA</title>
	<meta name="description" content="Oportunidades y estrategias de inversión FINEXIA" />
</svelte:head>

<header class="page-header">
	<h1 class="page-title">Inversiones</h1>
	<p class="page-subtitle">Descubre oportunidades alineadas con tu perfil y objetivos.</p>
</header>

<section class="spotlight panel">
	<div>
		<p class="eyebrow">Recomendación destacada</p>
		<h2>Portafolio crecimiento equilibrado</h2>
		<p>
			Estrategia diseñada para optimizar rendimiento en un escenario mixto con sesgo tecnológico.
		</p>
	</div>
	<button class="action">Explorar estrategia</button>
</section>

<section class="panel table-panel">
	<header class="table-head">
		<h2>Oportunidades activas</h2>
	</header>
	<div class="table">
		<div class="row heading">
			<span>Instrumento</span>
			<span>Riesgo</span>
			<span>ROI esperado</span>
			<span>Horizonte</span>
		</div>
		{#each opportunities as item (item.name)}
			<div class="row">
				<span>{item.name}</span>
				<span>{item.risk}</span>
				<span class="positive">{item.roi}</span>
				<span>{item.horizon}</span>
			</div>
		{/each}
	</div>
</section>

<style>
	.page-header {
		margin-bottom: 2rem;
		padding-bottom: 1.5rem;
		border-bottom: 1px solid rgba(212, 175, 55, 0.1);
	}

	.page-title {
		margin: 0 0 0.5rem;
		font-size: 2.35rem;
		font-weight: 700;
		letter-spacing: 0.5px;
		color: #d4af37;
		font-family: 'Poppins', system-ui, sans-serif;
	}

	.page-subtitle {
		margin: 0;
		color: rgba(224, 224, 224, 0.62);
	}

	.panel {
		border: 1px solid rgba(212, 175, 55, 0.15);
		border-radius: 16px;
		background: linear-gradient(135deg, rgba(26, 31, 46, 0.9) 0%, rgba(32, 39, 56, 0.9) 100%);
		box-shadow:
			0 20px 60px rgba(0, 0, 0, 0.3),
			inset 0 1px 0 rgba(255, 255, 255, 0.05);
		backdrop-filter: blur(16px);
	}

	.spotlight {
		display: flex;
		align-items: flex-end;
		justify-content: space-between;
		gap: 1rem;
		padding: 1.5rem;
		margin-bottom: 1.5rem;
	}

	.eyebrow {
		margin: 0 0 0.6rem;
		font-size: 0.72rem;
		letter-spacing: 0.7px;
		text-transform: uppercase;
		color: rgba(224, 224, 224, 0.46);
	}

	.spotlight h2 {
		margin: 0;
		color: #e8c547;
	}

	.spotlight p {
		margin: 0.6rem 0 0;
		color: rgba(224, 224, 224, 0.62);
		max-width: 62ch;
	}

	.action {
		border: none;
		border-radius: 8px;
		padding: 0.8rem 1.2rem;
		background: linear-gradient(135deg, #d4af37, #e8c547);
		color: #0f1419;
		font-weight: 700;
		cursor: pointer;
	}

	.table-panel {
		padding: 1.3rem;
	}

	.table-head h2 {
		margin: 0 0 1rem;
		color: #e0e0e0;
		font-size: 1.15rem;
	}

	.table {
		display: grid;
		gap: 0.55rem;
	}

	.row {
		display: grid;
		grid-template-columns: 1.4fr 0.8fr 0.8fr 0.8fr;
		gap: 0.7rem;
		padding: 0.85rem;
		border-radius: 10px;
		background: rgba(15, 20, 25, 0.42);
	}

	.row.heading {
		background: rgba(212, 175, 55, 0.12);
		font-size: 0.75rem;
		font-weight: 700;
		letter-spacing: 0.5px;
		text-transform: uppercase;
		color: rgba(224, 224, 224, 0.75);
	}

	.row span {
		font-size: 0.85rem;
		color: #e0e0e0;
	}

	.positive {
		color: #2ecc71;
	}

	@media (max-width: 768px) {
		.page-title {
			font-size: 1.85rem;
		}

		.spotlight {
			flex-direction: column;
			align-items: flex-start;
		}

		.row {
			grid-template-columns: 1fr;
		}
	}
</style>
