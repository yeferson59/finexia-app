<script lang="ts">
	/*
	 * La cartera entera cortada en una sola barra, de mayor a menor.
	 *
	 * Sustituye a una torta de siete colores con su leyenda al lado. La leyenda
	 * repetía, activo por activo, el importe y el peso que la lista de abajo ya
	 * daba —dos lecturas del mismo dato, una peor— y los siete matices afirmaban
	 * siete grupos donde solo hay un ranking.
	 *
	 * Lo que la torta no podía decir y esta barra sí: los anchos son
	 * proporcionales al valor, así que el punto medio de la barra es la mitad
	 * del dinero. La marca del centro no calcula nada; está donde tiene que
	 * estar por construcción, y lo que se lee es cuántas franjas caben a su
	 * izquierda. Dos, y la cartera está en pocas manos; quince, y está repartida.
	 */
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatPercent } from '$lib/shared/format/percent';
	import { buildBand, halfValueCount, type AssetHoldingRow } from '../asset-holdings';

	let {
		rows,
		formatCurrency,
		active = $bindable(null)
	}: {
		rows: AssetHoldingRow[];
		formatCurrency: (value: number) => string;
		/**
		 * Ticker señalado, compartido con la lista de abajo: apuntar a una fila
		 * enciende su franja y al revés. Es lo que hace que la barra y la lista
		 * sean el mismo objeto y no dos gráficos del mismo dato.
		 */
		active?: string | null;
	} = $props();

	/*
	 * `pointerenter` también dispara al tocar, y en una pantalla táctil el
	 * `pointerleave` que lo apagaría puede no llegar nunca: la franja se quedaba
	 * encendida, el resto apagadas, y sin nada evidente que devolviera la barra
	 * a su estado. Donde no hay puntero, la barra no se señala; la lista de
	 * abajo dice lo mismo en texto.
	 */
	const canHover = typeof window !== 'undefined' && window.matchMedia('(hover: hover)').matches;

	function point(key: string | null) {
		if (canHover) active = key;
	}

	const segments = $derived(buildBand(rows));
	const half = $derived(halfValueCount(rows));

	/*
	 * Una franja lleva su ticker debajo solo si le cabe dentro. La medida es en
	 * píxeles y no en porcentaje porque un 6 % son setenta píxeles en un
	 * escritorio y veinte en un teléfono: con el umbral en porcentaje, la misma
	 * franja rotulaba de sobra en una pantalla y partía la etiqueta en la otra.
	 */
	let bandWidth = $state(0);

	function fits(label: string, width: number): boolean {
		return (width / 100) * bandWidth >= label.length * 7.2 + 10;
	}

	/** Qué dice la barra, en una frase. Cambia con lo que hay dentro. */
	const caption = $derived.by(() => {
		if (half === 0) return '';
		if (segments.length === 1) return 'Un solo activo: ahí está todo tu valor.';
		if (half === 1) return 'Tu mayor activo es, por sí solo, más de la mitad de tu valor.';

		return `La marca señala la mitad de tu valor: cae tras tus ${half} mayores activos.`;
	});
</script>

{#if segments.length > 0}
	<section class="concentration" aria-labelledby="concentration-title">
		<h2 id="concentration-title">Cómo está repartido</h2>

		<!--
			El dibujo es `aria-hidden` y la lista de abajo es la superficie
			accesible: dice los mismos activos, en el mismo orden, con su peso en
			texto. Repetirlos aquí obligaría a recorrer dos veces lo mismo.
		-->
		<div class="plot" aria-hidden="true">
			<div class="band" bind:clientWidth={bandWidth}>
				{#each segments as segment (segment.key)}
					<span
						class="segment"
						class:dimmed={active !== null && active !== segment.key}
						style="width: {segment.width.toFixed(3)}%; --segment: {segment.color}"
						role="presentation"
						onpointerenter={() => point(segment.key)}
						onpointerleave={() => point(null)}
					></span>
				{/each}

				{#if segments.length > 1 && half > 0}
					<span class="half-mark"></span>
				{/if}
			</div>

			<div class="tags">
				{#each segments as segment (segment.key)}
					<span class="tag" style="width: {segment.width.toFixed(3)}%">
						{#if fits(segment.label, segment.width)}
							{segment.label}
						{/if}
					</span>
				{/each}
			</div>
		</div>

		<p class="caption">
			{#if active}
				{@const on = segments.find((segment) => segment.key === active)}
				{#if on}
					<strong>{on.label}</strong>{#if on.assets > 1}, {on.assets} activos{/if}:
					{privacy.money(formatCurrency(on.value))} ({formatPercent(on.percent)} del total)
				{/if}
			{:else}
				{caption}
			{/if}
		</p>
	</section>
{/if}

<style>
	.concentration {
		padding: 2rem 0;
		border-bottom: 1px solid var(--border);
	}

	h2 {
		margin: 0 0 1.4rem;
		font-family: var(--font-body);
		font-size: 1.05rem;
		font-weight: 500;
		color: var(--text);
	}

	/* Sitio de sobra arriba para la marca, que sobresale de la barra. */
	.plot {
		padding-top: 0.65rem;
	}

	.band {
		position: relative;
		display: flex;
		height: 26px;
	}

	.segment {
		display: block;
		height: 100%;
		min-width: 0;
		background: var(--segment);
		transition: filter 0.15s ease;
	}

	/*
	 * La separación se pinta dentro de la franja y no como hueco: un `gap` de
	 * dos píxeles entre treinta franjas se lleva sesenta del ancho, y entonces
	 * los anchos ya no son proporcionales al dinero.
	 */
	.segment:not(:last-child) {
		box-shadow: inset -1px 0 0 var(--bg);
	}

	.segment.dimmed {
		filter: saturate(0.3) brightness(0.5);
	}

	/*
	 * La marca de la mitad. No hay cálculo detrás: los anchos son proporcionales
	 * al valor, así que la mitad de la barra es la mitad del dinero.
	 *
	 * Sale por arriba y lleva un vértice, para que no se lea como una separación
	 * más entre franjas: las separaciones son oscuras y no salen de la barra.
	 */
	.half-mark {
		position: absolute;
		top: -0.55rem;
		bottom: -0.25rem;
		left: 50%;
		width: 1px;
		background: var(--text);
		/* Halo oscuro: la línea cruza franjas claras y oscuras, y sin él se
		   pierde justo sobre las primeras, que son las que más importan. */
		box-shadow: 0 0 0 1px rgba(8, 9, 10, 0.75);
		transform: translateX(-50%);
	}

	.half-mark::before {
		content: '';
		position: absolute;
		top: -1px;
		left: 50%;
		width: 0;
		height: 0;
		border-left: 3.5px solid transparent;
		border-right: 3.5px solid transparent;
		border-top: 5px solid var(--text);
		transform: translateX(-50%);
	}

	/*
	 * Los rótulos van fuera de la barra y no dentro: dentro tenían que ir en
	 * oscuro sobre el color, que solo funciona mientras la franja siga siendo
	 * clara, y a partir de la sexta ya no lo es.
	 */
	.tags {
		display: flex;
		margin-top: 0.45rem;
	}

	.tag {
		min-width: 0;
		padding-right: 0.4rem;
		font-family: var(--font-mono);
		font-size: 0.68rem;
		letter-spacing: 0.02em;
		color: var(--text-muted);
		white-space: nowrap;
		overflow: hidden;
	}

	.caption {
		max-width: 62ch;
		margin: 1rem 0 0;
		min-height: 1.4em;
		font-size: 0.85rem;
		line-height: 1.4;
		color: var(--text-muted);
	}

	.caption strong {
		font-family: var(--font-mono);
		font-weight: 600;
		color: var(--text);
	}

	@media (prefers-reduced-motion: reduce) {
		.segment {
			transition: none;
		}
	}
</style>
