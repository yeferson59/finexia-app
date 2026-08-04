<script lang="ts">
	/*
	 * "Dentro de Finexia": recorrido por las cuatro vistas del dashboard.
	 *
	 * Las maquetas replican la interfaz real (misma tipografía, mismos tokens,
	 * misma jerarquía) con cifras de ejemplo, para que quien llega a la landing
	 * vea el producto que va a usar y no una ilustración genérica. La ventana es
	 * decorativa (`aria-hidden`); cada pestaña lleva debajo un pie de texto que
	 * describe la vista, así que la información también existe sin verla.
	 */
	import ProductTourWindow from './product-tour-window.svelte';
	import { TOUR_VIEWS, type TourView } from '../product-tour';

	let active = $state<TourView['id']>('resumen');

	const view = $derived(TOUR_VIEWS.find((v) => v.id === active) ?? TOUR_VIEWS[0]);

	/** Flechas ← → entre pestañas, como pide el patrón `tablist` de WAI-ARIA. */
	function onTabKey(event: KeyboardEvent, index: number) {
		const delta = event.key === 'ArrowRight' ? 1 : event.key === 'ArrowLeft' ? -1 : 0;
		if (delta === 0) return;
		event.preventDefault();
		const next = TOUR_VIEWS[(index + delta + TOUR_VIEWS.length) % TOUR_VIEWS.length];
		active = next.id;
		document.getElementById(`tour-tab-${next.id}`)?.focus();
	}
</script>

<section class="wrap block" id="producto">
	<div class="sec-head reveal">
		<div class="eyebrow">El producto</div>
		<h2 class="sec-title">Esto es lo que verás<br />al entrar</h2>
		<p class="sec-desc">
			Cuatro vistas que trabajan sobre los mismos datos: el resumen de tu patrimonio, tus
			portafolios, cada movimiento registrado y los reportes que puedes descargar.
		</p>
	</div>

	<div class="tour reveal">
		<div class="tour-tabs" role="tablist" aria-label="Vistas del panel de Finexia">
			{#each TOUR_VIEWS as v, i (v.id)}
				<button
					id="tour-tab-{v.id}"
					class="tour-tab"
					class:active={active === v.id}
					role="tab"
					type="button"
					aria-selected={active === v.id}
					aria-controls="tour-panel-{v.id}"
					tabindex={active === v.id ? 0 : -1}
					onclick={() => (active = v.id)}
					onkeydown={(e) => onTabKey(e, i)}
				>
					{v.tab}
				</button>
			{/each}
		</div>

		<div
			id="tour-panel-{view.id}"
			class="tour-panel"
			role="tabpanel"
			aria-labelledby="tour-tab-{view.id}"
			tabindex="-1"
		>
			<ProductTourWindow {view} />

			<div class="tour-caption">
				<h3>{view.title}</h3>
				<p>{view.description}</p>
				<ul class="tour-points">
					{#each view.points as point (point)}
						<li>{point}</li>
					{/each}
				</ul>
			</div>
		</div>
	</div>
</section>

<style>
	.tour {
		max-width: 1080px;
		margin: 0 auto;
	}

	.tour-tabs {
		display: flex;
		gap: 4px;
		margin: 0 auto 22px;
		width: fit-content;
		max-width: 100%;
		padding: 4px;
		border: 1px solid var(--border);
		border-radius: 10px;
		background: var(--surface);
		overflow-x: auto;
		scrollbar-width: none;
	}

	.tour-tabs::-webkit-scrollbar {
		display: none;
	}

	.tour-tab {
		flex-shrink: 0;
		padding: 9px 18px;
		border: none;
		border-radius: 7px;
		background: transparent;
		color: var(--text-muted);
		font-family: var(--font-body);
		font-size: 13.5px;
		font-weight: 500;
		cursor: pointer;
		transition:
			background 0.2s,
			color 0.2s;
	}

	.tour-tab:hover {
		color: var(--text);
		background: rgba(255, 255, 255, 0.04);
	}

	.tour-tab.active {
		background: rgba(212, 145, 42, 0.12);
		color: var(--amber-light);
	}

	.tour-panel:focus {
		outline: none;
	}

	.tour-caption {
		max-width: 720px;
		margin: 26px auto 0;
		text-align: center;
	}

	.tour-caption h3 {
		font-family: var(--font-display);
		font-weight: 500;
		font-size: 20px;
		letter-spacing: -0.01em;
	}

	.tour-caption p {
		margin-top: 10px;
		font-size: 14.5px;
		font-weight: 300;
		line-height: 1.65;
		color: var(--text-muted);
	}

	.tour-points {
		display: flex;
		flex-wrap: wrap;
		justify-content: center;
		gap: 8px;
		margin-top: 18px;
		list-style: none;
		padding: 0;
	}

	.tour-points li {
		padding: 5px 12px;
		border: 1px solid var(--border);
		border-radius: 999px;
		background: var(--surface);
		font-family: var(--font-mono);
		font-size: 11px;
		letter-spacing: 0.02em;
		color: var(--text-muted);
	}

	@media (max-width: 640px) {
		.tour-tabs {
			width: 100%;
			justify-content: flex-start;
		}

		.tour-tab {
			padding: 9px 14px;
			font-size: 13px;
		}
	}
</style>
