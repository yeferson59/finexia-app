<script lang="ts">
	/**
	 * El mapa de rendimiento: cada cuenta con dinero es un bloque, su ancho es lo
	 * que guarda y su alto, la tasa a la que rinde hoy.
	 *
	 * El área del bloque es lo que gana, así que se lee sin hacer cuentas dónde
	 * trabaja el dinero y dónde está parado: lo parado es ancho sin altura, un
	 * suelo gris pegado a la base. Van de la tasa más alta a la más baja, y la
	 * escalera que sale dice también cuánto del efectivo rinde por encima de la
	 * media.
	 *
	 * Un solo tono —el verde de los intereses— para lo que rinde y el gris del
	 * texto apagado para lo parado: la leyenda de abajo lo nombra, y cada bloque
	 * dice lo suyo al pasar por encima o al llegar con el teclado. La lista de
	 * cuentas de debajo es la misma información en forma de tabla.
	 */
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import { formatAnnualRate } from '../rates';
	import { cashRateScale, type CashYieldBlock, type CashYieldMap } from '../yield';

	interface Props {
		map: CashYieldMap;
		/** La tasa media de lo que rinde, ponderada por el valor. */
		average: number;
		/** Abre lo que rinde el cajón de un bloque: su tasa, o la ficha del depósito. */
		onSelect: (block: CashYieldBlock) => void;
	}

	let { map, average, onSelect }: Props = $props();

	const scale = $derived(cashRateScale(map.top));
	const flat = $derived(map.top <= 0);

	let plotWidth = $state(0);
	/* El bloque señalado y dónde cae: su borde izquierdo y su ancho en el lienzo. */
	let active = $state<{ block: CashYieldBlock; left: number; width: number } | null>(null);

	const GAP = 2;
	/* El ancho en píxeles de un bloque, para decidir si su etiqueta cabe entera. */
	const widthOf = (block: CashYieldBlock) =>
		block.share * Math.max(plotWidth - GAP * (map.blocks.length - 1), 0);

	const percent = new Intl.NumberFormat('es-CO', { style: 'percent', maximumFractionDigits: 0 });

	const ratePct = (pct: number) =>
		`${pct.toLocaleString('es-CO', { maximumFractionDigits: pct < 10 ? 2 : 1 })}%`;

	const money = (amount: number, currency: string) =>
		privacy.money(formatCurrency(amount, currency));

	const where = (block: CashYieldBlock) =>
		block.pocketName ? block.pocketName : `Cuenta en ${block.currency}`;

	/* Lo que cabe bajo un bloque: el nombre de la plataforma y el del cajón, a
	   unos 6,4 px por carácter en la letra de la etiqueta. */
	const nameFits = (block: CashYieldBlock) =>
		widthOf(block) >= Math.max(block.sourceName.length, where(block).length) * 6.4 + 10;

	function describe(block: CashYieldBlock): string {
		const rate =
			block.pct > 0
				? `rinde ${formatAnnualRate(Number(block.pct.toFixed(2)))}${block.fixed ? ', tasa fija' : ''}`
				: 'no rinde';
		const action = block.fixed ? 'Ver el depósito' : block.pct > 0 ? 'Ver su tasa' : 'Darle tasa';
		return `${action}. ${block.sourceName}, ${where(block)}: ${money(block.balance, block.currency)}, ${percent.format(block.share)} del efectivo, ${rate}`;
	}

	/* El ancho de la etiqueta, para saber a qué lado del bloque cabe. */
	const TIP_WIDTH = 250;

	/*
	 * Al lado del bloque y no encima: encima de uno alto se salía del lienzo y
	 * tapaba el titular. Donde haya sitio —a la derecha primero—, y dentro del
	 * lienzo, arriba, si el bloque lo ocupa casi entero.
	 */
	function tipSide(left: number, width: number): 'right' | 'left' | 'inside' {
		if (plotWidth - (left + width) >= TIP_WIDTH) return 'right';
		if (left >= TIP_WIDTH) return 'left';
		return 'inside';
	}

	function show(event: Event, block: CashYieldBlock) {
		// El `<li>`, no el botón: el botón se mide contra su bloque y el bloque,
		// contra el lienzo, que es donde se coloca la etiqueta.
		const el = (event.currentTarget as HTMLElement).parentElement;
		if (el) active = { block, left: el.offsetLeft, width: el.offsetWidth };
	}
</script>

<div class="map" class:flat>
	<div class="plot">
		<div class="scale" aria-hidden="true">
			{#each scale.ticks as tick (tick)}
				<span class="tick" style:bottom="{(tick / scale.max) * 100}%">
					<span class="tick-label">{ratePct(tick)}</span>
				</span>
			{/each}
		</div>

		<div class="field" bind:clientWidth={plotWidth}>
			{#if average > 0 && !flat}
				<span class="mean" style:bottom="{(average / scale.max) * 100}%" aria-hidden="true">
					<span class="mean-label">media {ratePct(average)}</span>
				</span>
			{/if}

			<ol class="blocks" aria-label="Cuánto rinde cada cuenta">
				{#each map.blocks as block, index (block.key)}
					<li
						class="block"
						class:idle={block.pct <= 0}
						style:flex-grow={block.share}
						style:--h={block.pct / scale.max}
						style:--i={Math.min(index, 8)}
					>
						<button
							type="button"
							class="hit"
							aria-label={describe(block)}
							onclick={() => onSelect(block)}
							onpointerenter={(event) => show(event, block)}
							onpointerleave={() => (active = null)}
							onfocus={(event) => show(event, block)}
							onblur={() => (active = null)}
						>
							<span class="bar"></span>
						</button>
						{#if block.pct > 0 && widthOf(block) >= 52}
							<span class="cap" aria-hidden="true">{ratePct(block.pct)}</span>
						{/if}
					</li>
				{/each}
			</ol>

			{#if active}
				{@const block = active.block}
				{@const side = tipSide(active.left, active.width)}
				<div
					class="tip"
					aria-hidden="true"
					style:left={side === 'right' ? `${active.left + active.width + 10}px` : null}
					style:right={side === 'left'
						? `${plotWidth - active.left + 10}px`
						: side === 'inside'
							? '0.5rem'
							: null}
				>
					<p class="tip-rate">
						{block.pct > 0 ? formatAnnualRate(Number(block.pct.toFixed(2))) : 'No rinde'}
						{#if block.fixed}<span class="tip-fixed">fija</span>{/if}
					</p>
					<p class="tip-name">{block.sourceName}, {where(block)}</p>
					<p class="tip-line">
						{money(block.balance, block.currency)}
						<span class="tip-dim">{percent.format(block.share)} del efectivo</span>
					</p>
					{#if block.yearly > 0}
						<p class="tip-line tip-dim">≈ {money(block.yearly, block.currency)} al año</p>
					{/if}
					<p class="tip-hint">
						{block.fixed
							? 'Abre el depósito'
							: block.pct > 0
								? 'Abre su tasa'
								: 'Ábrelo para darle tasa'}
					</p>
				</div>
			{/if}
		</div>
	</div>

	<ol class="names" aria-hidden="true">
		{#each map.blocks as block (block.key)}
			<li style:flex-grow={block.share}>
				{#if nameFits(block)}
					<span class="name-source">{block.sourceName}</span>
					<span class="name-drawer">{where(block)}</span>
				{/if}
			</li>
		{/each}
	</ol>

	<ul class="legend">
		{#if map.earningShare > 0}
			<li><span class="swatch earning"></span>Rinde {percent.format(map.earningShare)}</li>
		{/if}
		{#if map.idleShare > 0.0005}
			<li><span class="swatch idle"></span>Parado {percent.format(map.idleShare)}</li>
		{/if}
	</ul>
</div>

<style>
	.map {
		--plot-h: 11.5rem;
		--gutter: 2.75rem;
		--work: #22c97e;
		--work-wash: rgba(34, 201, 126, 0.1);
		--rest: rgba(128, 123, 116, 0.55);

		min-width: 0;
	}

	.map.flat {
		--plot-h: 2.25rem;
	}

	.plot {
		display: grid;
		grid-template-columns: var(--gutter) minmax(0, 1fr);
		height: var(--plot-h);
	}

	/* Las marcas de la escala: filetes finos de lado a lado y su cifra en el
	   margen, con el mismo gris que cualquier eje del panel. */
	.scale {
		position: relative;
		grid-column: 1 / -1;
		grid-row: 1;
		pointer-events: none;
	}

	.tick {
		position: absolute;
		right: 0;
		left: var(--gutter);
		height: 0;
		border-top: 1px solid var(--border);
	}

	.tick:first-child {
		border-top-color: var(--border-strong);
	}

	.tick-label {
		position: absolute;
		right: calc(100% + 0.6rem);
		top: 0;
		transform: translateY(-50%);
		font-family: var(--font-mono);
		font-size: 0.68rem;
		white-space: nowrap;
		color: var(--text-dim);
	}

	.field {
		position: relative;
		grid-column: 2;
		grid-row: 1;
		min-width: 0;
	}

	.blocks {
		display: flex;
		align-items: flex-end;
		gap: 2px;
		height: 100%;
		margin: 0;
		padding: 0;
		list-style: none;
	}

	.block {
		position: relative;
		flex-basis: 0;
		min-width: 3px;
		height: 100%;
	}

	/* El bloque entero es la diana, no solo lo pintado: un suelo de 6 px no se
	   acierta, y la columna de aire encima de él sí. */
	.hit {
		display: flex;
		align-items: flex-end;
		width: 100%;
		height: 100%;
		padding: 0;
		border: none;
		background: none;
		cursor: pointer;
	}

	/* Un área: el lavado del verde y su borde de arriba a 2 px, con la esquina
	   redonda solo arriba, que es donde termina el dato. */
	.bar {
		width: 100%;
		height: max(calc(var(--h) * 100%), 4px);
		border-top: 2px solid var(--work);
		border-radius: 4px 4px 0 0;
		background: var(--work-wash);
		transition: background-color 0.15s ease;
	}

	.hit:hover .bar,
	.hit:focus-visible .bar {
		background: rgba(34, 201, 126, 0.2);
	}

	/* Lo parado no tiene altura: es el suelo, del grosor justo para verse. */
	.idle .bar {
		height: 6px;
		border-top: none;
		border-radius: 2px 2px 0 0;
		background: var(--rest);
	}

	.idle .hit:hover .bar,
	.idle .hit:focus-visible .bar {
		background: rgba(159, 153, 146, 0.8);
	}

	.cap {
		position: absolute;
		left: 50%;
		bottom: calc(max(calc(var(--h) * 100%), 4px) + 0.3rem);
		transform: translateX(-50%);
		font-family: var(--font-mono);
		font-size: 0.72rem;
		white-space: nowrap;
		color: var(--text);
		pointer-events: none;
	}

	.mean {
		position: absolute;
		right: 0;
		left: 0;
		z-index: 1;
		height: 0;
		border-top: 1px solid rgba(236, 234, 229, 0.35);
		pointer-events: none;
	}

	.mean-label {
		position: absolute;
		right: 0;
		bottom: 0.2rem;
		padding: 0 0 0 0.4rem;
		background: var(--bg);
		font-size: 0.7rem;
		white-space: nowrap;
		color: var(--text-muted);
	}

	.tip {
		position: absolute;
		top: 0;
		z-index: 2;
		display: grid;
		gap: 0.15rem;
		min-width: 11rem;
		max-width: 15rem;
		padding: 0.6rem 0.75rem;
		border: 1px solid var(--border-strong);
		border-radius: 8px;
		background: #121315;
		box-shadow: 0 12px 32px rgba(0, 0, 0, 0.5);
		pointer-events: none;
	}

	.tip p {
		margin: 0;
	}

	.tip-rate {
		display: flex;
		align-items: baseline;
		gap: 0.4rem;
		font-family: var(--font-mono);
		font-size: 0.95rem;
		color: var(--text);
	}

	.tip-name {
		font-size: 0.78rem;
		color: var(--text-muted);
	}

	.tip-line {
		display: flex;
		flex-wrap: wrap;
		gap: 0 0.5rem;
		font-family: var(--font-mono);
		font-size: 0.76rem;
		color: var(--text);
	}

	.tip-hint {
		margin-top: 0.3rem !important;
		padding-top: 0.35rem;
		border-top: 1px solid var(--border);
		font-size: 0.72rem;
		color: var(--text-dim);
	}

	.tip-dim,
	.tip-fixed {
		font-family: var(--font-body);
		color: var(--text-dim);
	}

	.names {
		display: flex;
		gap: 2px;
		margin: 0.55rem 0 0 var(--gutter);
		padding: 0;
		list-style: none;
	}

	.names li {
		display: flex;
		flex-direction: column;
		flex-basis: 0;
		min-width: 3px;
		padding-right: 0.4rem;
		line-height: 1.3;
	}

	.names span {
		font-size: 0.72rem;
		white-space: nowrap;
		color: var(--text-dim);
	}

	.names .name-source {
		color: var(--text-muted);
	}

	.legend {
		display: flex;
		flex-wrap: wrap;
		gap: 0.3rem 1.25rem;
		margin: 0.9rem 0 0 var(--gutter);
		padding: 0;
		list-style: none;
		font-size: 0.78rem;
		color: var(--text-muted);
	}

	.legend li {
		display: flex;
		align-items: center;
		gap: 0.45rem;
	}

	.swatch {
		width: 0.8rem;
		height: 0.55rem;
		border-radius: 2px 2px 0 0;
	}

	.swatch.earning {
		border-top: 2px solid var(--work);
		background: rgba(34, 201, 126, 0.3);
	}

	.swatch.idle {
		height: 0.3rem;
		background: var(--rest);
	}

	/* El único movimiento de la pantalla: los bloques suben una vez desde la
	   base, de la tasa más alta a la más baja, y llevan la vista a la escalera. */
	@media (prefers-reduced-motion: no-preference) {
		.bar {
			transform-origin: bottom center;
			animation: rise 0.6s cubic-bezier(0.2, 0.8, 0.2, 1) both;
			animation-delay: calc(var(--i) * 55ms);
		}

		.cap {
			animation: appear 0.3s ease-out both;
			animation-delay: calc(var(--i) * 55ms + 0.35s);
		}
	}

	@keyframes rise {
		from {
			transform: scaleY(0);
		}
	}

	@keyframes appear {
		from {
			opacity: 0;
		}
	}

	@media (max-width: 640px) {
		.map {
			--plot-h: 9rem;
			--gutter: 2.4rem;
		}
	}
</style>
