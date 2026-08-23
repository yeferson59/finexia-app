<script lang="ts">
	/*
	 * "Dentro de Finexia": recorrido por las cuatro vistas del dashboard.
	 *
	 * Las maquetas replican la interfaz real (misma tipografía, mismos tokens,
	 * misma jerarquía) con cifras de ejemplo, para que quien llega a la landing
	 * vea el producto que va a usar y no una ilustración genérica. La ventana es
	 * decorativa (`aria-hidden`); cada pestaña lleva debajo un pie de texto que
	 * describe la vista, así que la información también existe sin verla.
	 *
	 * La cabecera va a la izquierda y las pestañas a la derecha sobre el mismo
	 * filete: la sección abre con un eje horizontal en vez de con el tercer
	 * bloque centrado seguido.
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

<section class="band" id="producto">
	<div class="wrap block">
		<div class="tour-head reveal">
			<div class="tour-head-text">
				<div class="eyebrow">El producto</div>
				<h2 class="sec-title">Esto es lo que verás al entrar</h2>
			</div>

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
		</div>

		<div
			id="tour-panel-{view.id}"
			class="tour-panel reveal"
			role="tabpanel"
			aria-labelledby="tour-tab-{view.id}"
			tabindex="-1"
		>
			<ProductTourWindow {view} />

			<div class="tour-caption">
				<div>
					<h3>{view.title}</h3>
					<p>{view.description}</p>
				</div>
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
	.tour-head {
		display: flex;
		align-items: flex-end;
		justify-content: space-between;
		gap: 48px;
		padding-bottom: 26px;
		margin-bottom: 32px;
		border-bottom: 1px solid var(--border);
	}

	.tour-head-text {
		max-width: 620px;
	}

	.tour-tabs {
		display: flex;
		gap: 4px;
		flex-shrink: 0;
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

	/* El pie deja de ser una pila centrada: descripción a la izquierda,
	   capacidades alineadas a la derecha. */
	.tour-caption {
		display: grid;
		grid-template-columns: minmax(0, 1fr) 380px;
		gap: 56px;
		align-items: start;
		margin-top: 30px;
	}

	.tour-caption h3 {
		font-family: var(--font-display);
		font-weight: 500;
		font-size: 21px;
		letter-spacing: -0.01em;
	}

	.tour-caption p {
		margin-top: 10px;
		max-width: 62ch;
		font-size: 14.5px;
		font-weight: 300;
		line-height: 1.65;
		color: var(--text-muted);
		text-wrap: pretty;
	}

	.tour-points {
		display: flex;
		flex-wrap: wrap;
		justify-content: flex-end;
		gap: 8px;
		margin: 0;
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

	@media (max-width: 1040px) {
		.tour-head {
			flex-direction: column;
			align-items: flex-start;
			gap: 26px;
		}
		.tour-tabs {
			width: 100%;
		}
		.tour-caption {
			grid-template-columns: minmax(0, 1fr);
			gap: 22px;
		}
		.tour-points {
			justify-content: flex-start;
		}
	}

	@media (max-width: 640px) {
		.tour-tab {
			padding: 9px 14px;
			font-size: 13px;
		}
	}
</style>
