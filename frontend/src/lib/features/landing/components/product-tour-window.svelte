<script lang="ts">
	/*
	 * Marco de la maqueta: barra de ventana, menú lateral y barra superior,
	 * calcados del dashboard real. Todo el bloque es `aria-hidden`; el texto
	 * equivalente lo pone el pie de `product-tour.svelte`.
	 */
	import TourViewSummary from './tour-view-summary.svelte';
	import TourViewPortfolios from './tour-view-portfolios.svelte';
	import TourViewTransactions from './tour-view-transactions.svelte';
	import TourViewReports from './tour-view-reports.svelte';
	import { TOUR_NAV, type TourView } from '../product-tour';

	let { view }: { view: TourView } = $props();
</script>

<div class="win" aria-hidden="true">
	<div class="win-bar">
		<span class="dot"></span><span class="dot"></span><span class="dot"></span>
		<div class="win-url">finexia.me/dashboard</div>
	</div>

	<div class="win-body">
		<aside class="win-side">
			<div class="side-brand">
				<svg width="18" height="18" viewBox="0 0 30 30" fill="none">
					<rect width="30" height="30" rx="7" fill="var(--amber)" />
					<path
						d="M7 22L12.5 14.5L16.5 18.5L23 9"
						stroke="#0c0a06"
						stroke-width="3"
						stroke-linecap="round"
						stroke-linejoin="round"
					/>
				</svg>
				<span>FINEXIA</span>
			</div>
			<div class="side-label">Menú principal</div>
			<ul class="side-nav">
				{#each TOUR_NAV as item (item)}
					<li class:active={item === view.nav}>{item}</li>
				{/each}
			</ul>
		</aside>

		<div class="win-main">
			<div class="win-top">
				<span class="crumb">{view.crumb}</span>
				<span class="top-chips">
					<span class="chip">USD</span>
					<span class="chip">Modo oculto</span>
				</span>
			</div>

			<div class="win-view">
				{#if view.id === 'resumen'}
					<TourViewSummary />
				{:else if view.id === 'portafolios'}
					<TourViewPortfolios />
				{:else if view.id === 'transacciones'}
					<TourViewTransactions />
				{:else}
					<TourViewReports />
				{/if}
			</div>
		</div>
	</div>
</div>

<style>
	.win {
		border: 1px solid var(--border-strong);
		border-radius: 14px;
		overflow: hidden;
		background: rgba(255, 255, 255, 0.02);
		box-shadow: 0 30px 80px rgba(0, 0, 0, 0.45);
		backdrop-filter: blur(10px);
	}

	.win-bar {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 10px 14px;
		border-bottom: 1px solid var(--border);
		background: rgba(255, 255, 255, 0.022);
	}

	.dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: rgba(255, 255, 255, 0.13);
	}

	.win-url {
		flex: 1;
		text-align: center;
		font-family: var(--font-mono);
		font-size: 10px;
		letter-spacing: 0.04em;
		color: var(--text-dim);
	}

	.win-body {
		display: grid;
		grid-template-columns: 168px minmax(0, 1fr);
		min-height: 430px;
	}

	.win-side {
		padding: 16px 12px;
		border-right: 1px solid var(--border);
		background: rgba(0, 0, 0, 0.18);
	}

	.side-brand {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 0 6px 16px;
		font-family: var(--font-display);
		font-size: 12px;
		font-weight: 600;
		letter-spacing: 0.1em;
		color: var(--text);
	}

	.side-label {
		padding: 0 6px;
		font-family: var(--font-mono);
		font-size: 8.5px;
		letter-spacing: 0.18em;
		text-transform: uppercase;
		color: var(--text-dim);
		margin-bottom: 10px;
	}

	.side-nav {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.side-nav li {
		position: relative;
		padding: 7px 10px;
		border-radius: 6px;
		border: 1px solid transparent;
		font-size: 11.5px;
		color: var(--text-muted);
	}

	.side-nav li.active {
		background: rgba(212, 145, 42, 0.1);
		border-color: rgba(212, 145, 42, 0.22);
		color: var(--amber-light);
	}

	.side-nav li.active::before {
		content: '';
		position: absolute;
		left: -1px;
		top: 50%;
		transform: translateY(-50%);
		width: 2px;
		height: 12px;
		border-radius: 2px;
		background: var(--amber);
	}

	.win-main {
		display: flex;
		flex-direction: column;
		min-width: 0;
	}

	.win-top {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
		padding: 11px 18px;
		border-bottom: 1px solid var(--border);
	}

	.crumb {
		font-family: var(--font-mono);
		font-size: 9.5px;
		letter-spacing: 0.12em;
		text-transform: uppercase;
		color: var(--text-dim);
	}

	.top-chips {
		display: flex;
		gap: 6px;
	}

	.chip {
		padding: 3px 9px;
		border: 1px solid var(--border);
		border-radius: 5px;
		font-family: var(--font-mono);
		font-size: 9px;
		letter-spacing: 0.06em;
		color: var(--text-muted);
		white-space: nowrap;
	}

	.win-view {
		flex: 1;
		padding: 18px;
	}

	@media (max-width: 780px) {
		.win-body {
			grid-template-columns: minmax(0, 1fr);
			min-height: 0;
		}

		.win-side {
			display: none;
		}
	}

	@media (max-width: 560px) {
		.win-view {
			padding: 14px;
		}

		.win-top {
			padding: 10px 14px;
		}

		.chip:last-child {
			display: none;
		}
	}
</style>
