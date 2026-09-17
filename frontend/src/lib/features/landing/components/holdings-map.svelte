<script lang="ts">
	/*
	 * El mapa del hero: las dieciocho posiciones del ejemplo, cada una un bloque
	 * del ancho de su valor, agrupadas por plataforma, por portafolio o por tipo.
	 *
	 * Es lo que hace Finexia dicho sin palabras. Por plataforma, que es como está
	 * el dinero de verdad, cada fila mezcla colores: en Degiro hay jubilación y
	 * reserva a la vez. Por portafolio, que es como tú lo piensas, cada fila es
	 * de un color. Los bloques no cambian de tamaño entre vistas —la escala es
	 * común—, solo de sitio.
	 *
	 * Entra por plataforma y, la primera vez que se ve, se reordena solo por
	 * portafolio. Es el único movimiento que la página hace sin que se lo pidan;
	 * si el visitante ya ha tocado el control o prefiere menos movimiento, no pasa.
	 */
	import { onMount } from 'svelte';
	import {
		HOLDINGS,
		HOLDING_LENSES,
		HOLDINGS_TOTAL,
		PORTFOLIO_ORDER,
		formatUsd,
		layoutHoldings,
		type HoldingLens
	} from '../holdings';

	let lens = $state<HoldingLens>('platform');
	let touched = false;
	let hovered = $state<number | null>(null);
	let root: HTMLElement;

	const layout = $derived(layoutHoldings(lens));

	const titles: Record<HoldingLens, string> = {
		platform: 'Así está tu dinero: repartido en 5 plataformas.',
		portfolio: 'Así lo piensas tú: 3 portafolios.',
		type: 'Y por tipo de activo, si lo prefieres.'
	};

	const colors: Record<string, string> = {
		Jubilación: 'var(--lp-jubilacion)',
		Cripto: 'var(--lp-cripto)',
		Reserva: 'var(--lp-reserva)'
	};

	const detail = $derived.by(() => {
		if (hovered === null) return null;
		const h = HOLDINGS[hovered];
		return `${h.name} (${h.short}): ${formatUsd(h.value)} en ${h.platform}, dentro de ${h.portfolio}.`;
	});

	/** Un solo manejador para los dieciocho bloques: lee cuál es del `data-id`. */
	function onPointerOver(event: PointerEvent) {
		const el = (event.target as HTMLElement).closest<HTMLElement>('[data-id]');
		if (el) hovered = Number(el.dataset.id);
	}

	function choose(next: HoldingLens) {
		touched = true;
		lens = next;
	}

	onMount(() => {
		if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) return;

		let timer: ReturnType<typeof setTimeout> | undefined;
		const io = new IntersectionObserver(
			([entry]) => {
				if (!entry.isIntersecting) return;
				io.disconnect();
				timer = setTimeout(() => {
					if (!touched) lens = 'portfolio';
				}, 1400);
			},
			{ threshold: 0.6 }
		);
		io.observe(root);

		return () => {
			io.disconnect();
			clearTimeout(timer);
		};
	});
</script>

<figure class="map" bind:this={root} aria-labelledby="map-title">
	<div class="map-top">
		<p class="map-title" id="map-title" aria-live="polite">{titles[lens]}</p>

		<div class="lens" role="group" aria-label="Agrupar las posiciones por">
			<span class="lens-label" aria-hidden="true">Agrupar por</span>
			{#each HOLDING_LENSES as option (option.id)}
				<button
					type="button"
					class:on={lens === option.id}
					aria-pressed={lens === option.id}
					onclick={() => choose(option.id)}
				>
					{option.label}
				</button>
			{/each}
		</div>
	</div>

	<div class="plot" style="--rows:{layout.groups.length}">
		<ol class="rows">
			{#each layout.groups as group (lens + group.label)}
				<li class="row">
					<span class="row-name">{group.label}</span>
					<span class="row-count">
						{group.count}
						{group.count === 1 ? 'posición' : 'posiciones'}
					</span>
					<span class="row-value">{formatUsd(group.value)}</span>
				</li>
			{/each}
		</ol>

		<!--
			Los bloques son decorativos para un lector de pantalla: lo mismo, y
			completo, está en la tabla de debajo y en las filas de la izquierda.
			El puntero sí los lee, para decir qué posición es cada uno.
		-->
		<div
			class="blocks"
			aria-hidden="true"
			role="presentation"
			onpointerover={onPointerOver}
			onpointerleave={() => (hovered = null)}
		>
			{#each layout.blocks as block (block.id)}
				<span
					class="block"
					class:dim={hovered !== null && hovered !== block.id}
					class:dark-text={block.portfolio === 'Jubilación'}
					data-id={block.id}
					style="--row:{block.row}; --x:{block.x}; --w:{block.w}; background:{colors[
						block.portfolio
					]}"
				>
					<span class="block-label">{block.short}</span>
				</span>
			{/each}
		</div>
	</div>

	<table class="lp-sr-only">
		<caption>Las 18 posiciones del ejemplo</caption>
		<thead>
			<tr>
				<th scope="col">Posición</th>
				<th scope="col">Plataforma</th>
				<th scope="col">Portafolio</th>
				<th scope="col">Tipo</th>
				<th scope="col">Valor</th>
			</tr>
		</thead>
		<tbody>
			{#each HOLDINGS as h, i (i)}
				<tr>
					<th scope="row">{h.name}</th>
					<td>{h.platform}</td>
					<td>{h.portfolio}</td>
					<td>{h.type}</td>
					<td>{formatUsd(h.value)}</td>
				</tr>
			{/each}
		</tbody>
	</table>

	<figcaption class="map-foot">
		<ul class="legend" aria-label="Color de cada portafolio">
			{#each PORTFOLIO_ORDER as name (name)}
				<li><i style="background:{colors[name]}"></i>{name}</li>
			{/each}
		</ul>
		<p class="detail" aria-hidden="true">
			{detail ?? `Cada bloque es una posición. En total, ${formatUsd(HOLDINGS_TOTAL)}.`}
		</p>
	</figcaption>
</figure>

<style>
	.map {
		--label-w: 300px;
		--row-h: 58px;
		--bar-h: 34px;
		--bar-top: calc((var(--row-h) - var(--bar-h)) / 2);
		position: relative;
		margin: 0;
		padding: 28px 32px 22px;
		border-radius: 14px;
		background: var(--lp-paper-2);
	}

	.map-top {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 16px 32px;
		flex-wrap: wrap;
		padding-bottom: 18px;
		border-bottom: 1px solid var(--lp-rule);
	}

	.map-title {
		margin: 0;
		font-size: var(--lp-fs-h3);
		font-stretch: 108%;
		font-weight: 600;
		letter-spacing: -0.01em;
	}

	.lens {
		display: flex;
		align-items: center;
		gap: 4px;
	}

	.lens-label {
		margin-right: 8px;
		font-size: var(--lp-fs-sm);
		color: var(--lp-ink-2);
	}

	.lens button {
		min-height: 40px;
		padding: 0 14px;
		border: 1px solid var(--lp-rule);
		border-radius: 999px;
		background: transparent;
		color: var(--lp-ink);
		font-family: var(--lp-font);
		font-size: var(--lp-fs-sm);
		font-weight: 500;
		cursor: pointer;
		transition:
			background 0.18s ease,
			border-color 0.18s ease;
	}

	.lens button:hover {
		border-color: var(--lp-ink);
	}

	.lens button.on {
		border-color: var(--lp-ink);
		background: var(--lp-ink);
		color: var(--lp-paper);
	}

	/* Filas y bloques comparten la misma rejilla vertical: las filas van en el
	   flujo y los bloques, encima, colocados con `--row`. */
	.plot {
		position: relative;
		height: calc(var(--rows) * var(--row-h));
		margin-top: 12px;
		transition: height 0.6s cubic-bezier(0.65, 0, 0.35, 1);
	}

	.rows {
		margin: 0;
		padding: 0;
		list-style: none;
	}

	.row {
		display: grid;
		grid-template-columns: minmax(0, 1fr) auto;
		grid-template-rows: auto auto;
		align-content: center;
		column-gap: 16px;
		width: calc(var(--label-w) - 28px);
		height: var(--row-h);
		animation: row-in 0.45s ease both;
	}

	.row-name {
		font-size: var(--lp-fs-lead);
		font-stretch: 110%;
		font-weight: 620;
		line-height: 1.15;
	}

	.row-count {
		grid-row: 2;
		font-size: 13px;
		color: var(--lp-ink-2);
	}

	.row-value {
		grid-column: 2;
		grid-row: 1 / span 2;
		align-self: center;
		font-size: var(--lp-fs-lead);
		font-weight: 500;
		font-variant-numeric: tabular-nums;
	}

	@keyframes row-in {
		from {
			opacity: 0;
		}
	}

	.blocks {
		position: absolute;
		top: 0;
		bottom: 0;
		left: var(--label-w);
		right: 0;
	}

	.block {
		position: absolute;
		top: calc(var(--row) * var(--row-h) + var(--bar-top));
		left: calc(var(--x) * 100%);
		width: calc(var(--w) * 100% - 2px);
		height: var(--bar-h);
		overflow: hidden;
		border-radius: 3px;
		container-type: inline-size;
		transition:
			top 0.7s cubic-bezier(0.65, 0, 0.35, 1),
			left 0.7s cubic-bezier(0.65, 0, 0.35, 1),
			opacity 0.2s ease;
	}

	.block.dim {
		opacity: 0.45;
	}

	.block-label {
		display: block;
		padding: 0 7px;
		font-size: 12px;
		font-weight: 600;
		font-stretch: 92%;
		line-height: var(--bar-h);
		color: #fff;
		white-space: nowrap;
	}

	.block.dark-text .block-label {
		color: var(--lp-ink);
	}

	/* El símbolo solo entra si cabe entero; si no, el bloque va liso. */
	@container (max-width: 44px) {
		.block-label {
			display: none;
		}
	}

	.map-foot {
		display: flex;
		align-items: baseline;
		justify-content: space-between;
		gap: 12px 32px;
		flex-wrap: wrap;
		margin-top: 14px;
		padding-top: 14px;
		border-top: 1px solid var(--lp-rule);
	}

	.legend {
		display: flex;
		gap: 20px;
		margin: 0;
		padding: 0;
		list-style: none;
		font-size: var(--lp-fs-sm);
	}

	.legend li {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.legend i {
		width: 12px;
		height: 12px;
		border-radius: 3px;
	}

	.detail {
		margin: 0;
		font-size: var(--lp-fs-sm);
		color: var(--lp-ink-2);
		font-variant-numeric: tabular-nums;
	}

	@media (max-width: 1100px) {
		.map {
			--label-w: 250px;
		}
	}

	/* En estrecho, el nombre y el importe van encima de su barra. */
	@media (max-width: 760px) {
		.map {
			--label-w: 0px;
			--row-h: 74px;
			--bar-h: 28px;
			--bar-top: 40px;
			padding: 20px 16px 16px;
		}
		.row {
			width: 100%;
			align-content: start;
			padding-top: 6px;
			grid-template-columns: auto minmax(0, 1fr) auto;
			grid-template-rows: auto;
			column-gap: 10px;
			align-items: baseline;
		}
		.row-name {
			font-size: 16px;
		}
		.row-count {
			grid-row: 1;
			grid-column: 2;
		}
		.row-value {
			grid-column: 3;
			grid-row: 1;
			font-size: 16px;
		}
		.lens {
			width: 100%;
			flex-wrap: wrap;
		}
		.lens-label {
			width: 100%;
			margin: 0 0 4px;
		}
		.lens button {
			flex: 1 1 auto;
			padding: 0 10px;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.plot,
		.block {
			transition: none;
		}
		.row {
			animation: none;
		}
	}
</style>
