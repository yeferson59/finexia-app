<script lang="ts">
	/*
	 * "El producto": recorrido por cuatro vistas del panel.
	 *
	 * Las maquetas replican la interfaz real (misma tipografía, mismos tokens,
	 * misma jerarquía) con las cifras del ejemplo, para que quien llega a la
	 * portada vea el producto que va a usar y no una ilustración. La ventana es
	 * decorativa (`aria-hidden`); cada pestaña lleva debajo un texto que describe
	 * la vista, así que la información también existe sin verla.
	 *
	 * Sobre el papel de la portada, la ventana es lo único oscuro de la sección:
	 * se lee como lo que es, una pantalla del panel.
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

<section class="lp-section" id="producto" aria-labelledby="producto-title">
	<div class="lp-wrap">
		<div class="head">
			<h2 class="lp-h2" id="producto-title">Esto es lo que verás al entrar</h2>

			<div class="tabs" role="tablist" aria-label="Vistas del panel de Finexia">
				{#each TOUR_VIEWS as v, i (v.id)}
					<button
						id="tour-tab-{v.id}"
						class="tab"
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
			class="panel"
			role="tabpanel"
			aria-labelledby="tour-tab-{view.id}"
			tabindex="-1"
		>
			<ProductTourWindow {view} />

			<div class="caption">
				<div>
					<h3>{view.title}</h3>
					<p>{view.description}</p>
				</div>
				<ul class="points">
					{#each view.points as point (point)}
						<li>{point}</li>
					{/each}
				</ul>
			</div>
		</div>
	</div>
</section>

<style>
	.head {
		display: flex;
		align-items: flex-end;
		justify-content: space-between;
		gap: 24px 48px;
		flex-wrap: wrap;
		margin-bottom: 40px;
	}

	.head .lp-h2 {
		max-width: 16ch;
	}

	.tabs {
		display: flex;
		gap: 4px;
		max-width: 100%;
		overflow-x: auto;
		scrollbar-width: none;
	}

	.tabs::-webkit-scrollbar {
		display: none;
	}

	.tab {
		flex-shrink: 0;
		min-height: 44px;
		padding: 0 16px;
		border: 1px solid var(--lp-rule);
		border-radius: 999px;
		background: transparent;
		color: var(--lp-ink);
		font-family: var(--lp-font);
		font-size: 15px;
		font-weight: 500;
		cursor: pointer;
		transition: border-color 0.15s ease;
	}

	.tab:hover {
		border-color: var(--lp-ink);
	}

	.tab[aria-selected='true'] {
		border-color: var(--lp-ink);
		background: var(--lp-ink);
		color: var(--lp-paper);
	}

	.panel:focus {
		outline: none;
	}

	.caption {
		display: grid;
		grid-template-columns: minmax(0, 1fr) minmax(0, 380px);
		gap: 24px 64px;
		align-items: start;
		margin-top: 36px;
	}

	.caption h3 {
		margin: 0;
		font-size: 24px;
		font-stretch: 110%;
		font-weight: 620;
		letter-spacing: -0.015em;
		line-height: 1.15;
	}

	.caption p {
		margin: 12px 0 0;
		max-width: 60ch;
		color: var(--lp-ink-2);
		text-wrap: pretty;
	}

	.points {
		margin: 0;
		padding: 0;
		list-style: none;
		border-top: 1px solid var(--lp-rule);
	}

	.points li {
		padding: 10px 0;
		border-bottom: 1px solid var(--lp-rule);
		font-size: 15px;
	}

	@media (max-width: 900px) {
		.caption {
			grid-template-columns: minmax(0, 1fr);
		}
	}

	@media (max-width: 640px) {
		.head {
			margin-bottom: 28px;
		}
		/* Cuatro pestañas no caben en una fila: pasan a dos en vez de esconder la
		   última tras un desplazamiento que no se ve. */
		.tabs {
			flex-wrap: wrap;
		}
	}
</style>
