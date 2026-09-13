<script lang="ts">
	/**
	 * Cómo se clasifica un activo del catálogo: una industria, o el reparto entre
	 * las varias en las que está.
	 *
	 * Vive en un componente propio porque el alta y la edición piden lo mismo y
	 * son trece casillas: duplicarlas era duplicar también la regla que las
	 * gobierna, que es la parte que no puede desalinearse.
	 *
	 * Las dos formas son **excluyentes**, y el formulario lo dice con un selector
	 * de modo en vez de con campos desactivados. Antes estaban las dos a la vista
	 * y una se apagaba cuando la otra tenía algo, así que para pasar de una a
	 * otra había que adivinar qué vaciar. Ahora solo existe en la página la del
	 * modo elegido, y un campo que no está no se envía: guardar en «Una
	 * industria» borra el reparto y viceversa. Lo que se escribió en el otro modo
	 * sigue en el borrador —volver atrás no lo pierde— y un aviso dice qué se va
	 * a descartar antes de pulsar.
	 *
	 * Por qué existe el reparto: una acción y un ETF sectorial caben en una
	 * industria —Apple es tecnología, XLK es tecnología— y un fondo de mercado
	 * ancho no. Un VOO es el S&P 500 entero; ficharlo bajo «Tecnología» metería
	 * dos tercios de la posición en industrias en las que no está, y dejarlo sin
	 * clasificar haría que el panel contara la mayor posición de la cartera como
	 * trabajo pendiente que nadie puede hacer.
	 */
	import { untrack } from 'svelte';
	import { SECTOR_OPTIONS, formatSector, sectorWeightsTotal } from '$lib/shared/format/sector';
	import { formatPercent } from '$lib/shared/format/percent';

	interface Props {
		/** La industria única, o «» cuando no la tiene. */
		sector: string;
		/**
		 * Los pesos por industria, en porcentaje.
		 *
		 * Números y no texto porque es lo que escribe `bind:value` sobre un
		 * `input type="number"`: una casilla vacía vuelve como `null`, no como
		 * cadena vacía. El envío sigue siendo texto —lo serializa el `<form>`—,
		 * así que los decimales llegan al backend tal cual se escribieron.
		 */
		weights: Record<string, number | null>;
		/** Prefijo de los `id`, para que los dos formularios no colisionen. */
		idPrefix?: string;
	}

	let { sector = $bindable(''), weights = $bindable({}), idPrefix = '' }: Props = $props();

	type SectorOption = (typeof SECTOR_OPTIONS)[number];
	type Mode = 'single' | 'split';

	const MODES: { value: Mode; label: string; example: string }[] = [
		{ value: 'single', label: 'Una industria', example: 'Una acción o un ETF sectorial' },
		{ value: 'split', label: 'Varias industrias', example: 'Un fondo de mercado ancho' }
	];

	/*
	 * Renta fija y efectivo van aparte porque no son industrias: son la parte del
	 * fondo que no está en acciones. Separarlas es lo que permite explicar, junto
	 * a sus dos casillas, qué pasa si no se escriben.
	 */
	const RESERVES = new Set(['fixed_income', 'cash']);
	const industries = SECTOR_OPTIONS.filter((s) => !RESERVES.has(s.value));
	const reserves = SECTOR_OPTIONS.filter((s) => RESERVES.has(s.value));

	function weightOf(value: string): number | null {
		const weight = weights[value];
		return typeof weight === 'number' && Number.isFinite(weight) ? weight : null;
	}

	const filled = $derived(
		SECTOR_OPTIONS.flatMap((s) => {
			const weight = weightOf(s.value);
			return weight === null ? [] : [{ sector: s.value, weight }];
		})
	);

	// El modo se decide una vez, con lo que traiga el activo: reaccionar a los
	// pesos cambiaría de modo al vaciar la última casilla, en mitad de la edición.
	let mode = $state<Mode>(untrack(() => (sector === '' && filled.length > 0 ? 'split' : 'single')));

	const total = $derived(sectorWeightsTotal(filled));
	const over = $derived(total > 100);
	const hasEmptyWeight = $derived(filled.some((w) => w.weight <= 0));

	/*
	 * La barra se dibuja sobre 100 o sobre el total, lo que sea mayor: pasado de
	 * 100 la escala se estira para que el exceso quepa, y una marca enseña dónde
	 * estaba el límite.
	 */
	const scale = $derived(Math.max(100, total));
	const segments = $derived.by(() => {
		let at = 0;
		return filled.map((w) => {
			const weight = Math.max(0, w.weight);
			const segment = {
				sector: w.sector,
				start: (at / scale) * 100,
				width: (weight / scale) * 100
			};
			at += weight;
			return segment;
		});
	});

	/*
	 * La fila bajo el puntero o con el foco enciende su tramo de la barra. El
	 * puntero manda sobre el foco para que pasar por encima de otra fila mientras
	 * se escribe en una no deje dos tramos a medias.
	 */
	let hovered = $state<string | null>(null);
	let focused = $state<string | null>(null);
	const active = $derived(hovered ?? focused);
	const spotlight = $derived(active !== null && weightOf(active) !== null);
</script>

<fieldset class="classification">
	<legend class="field-label">Industria <span class="optional">(opcional)</span></legend>

	<div class="modes">
		{#each MODES as m (m.value)}
			<label class="mode" class:on={mode === m.value}>
				<input type="radio" name="{idPrefix}sector-mode" value={m.value} bind:group={mode} />
				<span class="mode-label">{m.label}</span>
				<span class="mode-example">{m.example}</span>
			</label>
		{/each}
	</div>

	{#if mode === 'single'}
		<select id="{idPrefix}sector" name="sector" bind:value={sector} aria-label="Industria">
			<option value="">Sin clasificar</option>
			{#each SECTOR_OPTIONS as s (s.value)}
				<option value={s.value}>{s.label}</option>
			{/each}
		</select>
		<p class="hint">
			Es lo que reparte el panel por industria. Dejarlo sin clasificar no rompe nada: esa parte del
			patrimonio se cuenta aparte, con su propia etiqueta.
		</p>

		{#if filled.length > 0}
			<p class="feedback warning">
				Al guardar se descarta el reparto entre industrias que hay escrito. Vuelve a «Varias
				industrias» para conservarlo.
			</p>
		{/if}
	{:else}
		<div class="meter">
			<div class="strip" class:spotlight aria-hidden="true">
				{#each segments as seg (seg.sector)}
					<span
						class="segment"
						class:on={active === seg.sector}
						style:left="{seg.start}%"
						style:width="max(2px, calc({seg.width}% - 2px))"
					></span>
				{/each}
				{#if over}
					<span class="excess" style:left="{(100 / scale) * 100}%"></span>
				{/if}
			</div>
			<output class="total" class:over class:empty={filled.length === 0}>
				{formatPercent(total)}
			</output>
		</div>

		<p class="status" class:over aria-live="polite">
			{#if over}
				Sobran {formatPercent(total - 100)}: un fondo no pasa de 100%. Revisa las cifras.
			{:else if hasEmptyWeight}
				Un peso en 0 no reparte nada. Deja esa casilla en blanco.
			{:else if filled.length === 0}
				Copia los pesos de la ficha del fondo, en porcentaje.
			{:else if 100 - total < 0.05}
				El reparto cubre el fondo entero.
			{:else}
				Faltan {formatPercent(100 - total)}. No hace falta completarlo: se reparte en proporción
				entre las industrias que tienen peso.
			{/if}
		</p>

		<div class="rows">
			{#each industries as option (option.value)}
				{@render row(option)}
			{/each}
		</div>

		<p class="hint reserves-hint">
			Si el fondo tiene bonos o caja, escríbelos aquí: lo que no se anota se reparte entre las
			industrias de arriba.
		</p>

		<div class="rows">
			{#each reserves as option (option.value)}
				{@render row(option)}
			{/each}
		</div>

		{#if sector !== ''}
			<p class="feedback warning">
				Al guardar, {formatSector(sector)} deja de ser la industria de este activo.
			</p>
		{/if}
	{/if}
</fieldset>

{#snippet row(option: SectorOption)}
	{@const weight = weightOf(option.value)}
	<div
		class="row"
		class:filled={weight !== null}
		class:active={active === option.value}
		role="presentation"
		onpointerenter={() => (hovered = option.value)}
		onpointerleave={() => (hovered = null)}
		onfocusin={() => (focused = option.value)}
		onfocusout={() => (focused = null)}
	>
		<label for="{idPrefix}weight-{option.value}">{option.label}</label>
		<span class="cell">
			<input
				id="{idPrefix}weight-{option.value}"
				type="number"
				name="weight.{option.value}"
				bind:value={weights[option.value]}
				min="0"
				max="100"
				step="any"
				inputmode="decimal"
				placeholder="—"
				aria-invalid={weight !== null && weight <= 0}
			/>
			<span class="unit" aria-hidden="true">%</span>
		</span>
	</div>
{/snippet}

<style>
	.classification {
		container-type: inline-size;
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
		min-width: 0;
	}

	.classification legend {
		padding: 0;
		margin-bottom: 0.45rem;
	}

	/* Las dos formas, como un interruptor: el mismo idioma que las pestañas del
	   reparto del panel. El radio real sigue ahí para el teclado y el lector. */
	.modes {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 0.2rem;
		padding: 0.2rem;
		border: 1px solid var(--border-strong);
		border-radius: 10px;
	}

	.mode {
		position: relative;
		display: flex;
		flex-direction: column;
		gap: 0.1rem;
		padding: 0.55rem 0.8rem;
		border-radius: 8px;
		cursor: pointer;
		transition: background 0.15s ease;
	}

	.mode input {
		position: absolute;
		inset: 0;
		margin: 0;
		opacity: 0;
		cursor: pointer;
	}

	.mode:hover {
		background: var(--surface-2);
	}

	.mode.on {
		background: var(--surface-3);
		box-shadow: inset 0 0 0 1px rgba(212, 145, 42, 0.35);
	}

	.mode:has(input:focus-visible) {
		outline: 2px solid var(--amber);
		outline-offset: 2px;
	}

	.mode-label {
		font-size: 0.87rem;
		font-weight: 500;
		color: var(--text-muted);
	}

	.mode.on .mode-label {
		color: var(--text);
	}

	.mode-example {
		font-size: 0.75rem;
		color: var(--text-dim);
	}

	/*
	 * La barra se queda arriba mientras se baja por las filas: se escribe
	 * «Efectivo» mirando cuánto falta, no desplazándose para comprobarlo. El
	 * fondo es el del modal, del que hereda la variable.
	 */
	.meter {
		position: sticky;
		top: 0;
		z-index: 1;
		display: flex;
		align-items: center;
		gap: 1rem;
		margin-top: 0.35rem;
		padding-block: 0.5rem;
		background: var(--modal-surface, var(--bg));
	}

	/*
	 * Una sola tinta, como el resto de repartos del panel: la barra dice cuánto
	 * está escrito y cuánto falta, y cada tramo es de quien lo encienda desde su
	 * fila. El rayado es lo que falta: no es caja, es lo que se reparte en
	 * proporción.
	 */
	.strip {
		position: relative;
		flex: 1;
		height: 12px;
		border-radius: 4px;
		background-color: rgba(255, 255, 255, 0.025);
		background-image: repeating-linear-gradient(
			-45deg,
			rgba(255, 255, 255, 0.08) 0 1.5px,
			transparent 1.5px 6px
		);
		overflow: hidden;
	}

	.segment {
		position: absolute;
		top: 0;
		bottom: 0;
		border-radius: 2px;
		background: var(--amber);
		transition:
			left 0.2s ease,
			width 0.2s ease,
			opacity 0.15s ease,
			background 0.15s ease;
	}

	/* Opaco y no con `opacity`: a través de un tramo translúcido se veía el
	   rayado de lo que falta, y el resto del fondo parecía sin escribir. */
	.strip.spotlight .segment {
		background: color-mix(in srgb, var(--amber) 38%, #0e0f11);
	}

	.strip .segment.on {
		background: var(--amber-light);
	}

	.excess {
		position: absolute;
		top: 0;
		right: 0;
		bottom: 0;
		border-left: 2px solid var(--text);
		background: var(--red);
	}

	.total {
		min-width: 4.5rem;
		font-family: var(--font-mono);
		font-size: 1rem;
		font-variant-numeric: tabular-nums;
		text-align: right;
		color: var(--text);
	}

	.total.empty {
		color: var(--text-dim);
	}

	.total.over,
	.status.over {
		color: var(--red);
	}

	.status {
		margin: -0.35rem 0 0.25rem;
		font-size: 0.83rem;
		line-height: 1.5;
		color: var(--text-muted);
	}

	.rows {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		column-gap: 1.75rem;
	}

	.row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.75rem;
		padding: 0.3rem 0.4rem 0.3rem 0.5rem;
		border-bottom: 1px solid var(--border);
		border-radius: 4px 4px 0 0;
		transition: background 0.15s ease;
	}

	.row.active {
		background: var(--surface-2);
	}

	.row label {
		min-width: 0;
		font-size: 0.87rem;
		color: var(--text-dim);
		cursor: pointer;
		overflow-wrap: anywhere;
	}

	.row.filled label {
		color: var(--text);
	}

	.cell {
		position: relative;
		flex: none;
		width: 6.25rem;
	}

	/* Más compacto que un campo suelto: son trece filas y tienen que caber a la
	   vista junto a la barra. */
	.cell input {
		padding: 0.4rem 1.65rem 0.4rem 0.5rem;
		border-color: rgba(212, 145, 42, 0.12);
		font-family: var(--font-mono);
		font-size: 0.85rem;
		font-variant-numeric: tabular-nums;
		text-align: right;
		appearance: textfield;
	}

	.cell input::-webkit-inner-spin-button,
	.cell input::-webkit-outer-spin-button {
		appearance: none;
		margin: 0;
	}

	.row.filled .cell input {
		border-color: rgba(212, 145, 42, 0.3);
	}

	/* Por encima de la regla de arriba, que si no le quitaba el borde de foco a
	   las filas con peso. */
	.row .cell input:focus {
		border-color: var(--amber);
	}

	.row .cell input[aria-invalid='true'] {
		border-color: var(--red);
	}

	.unit {
		position: absolute;
		top: 50%;
		right: 0.6rem;
		transform: translateY(-50%);
		font-family: var(--font-mono);
		font-size: 0.8rem;
		color: var(--text-dim);
		pointer-events: none;
	}

	.reserves-hint {
		margin-top: 0.5rem;
	}

	@container (max-width: 30rem) {
		.rows {
			grid-template-columns: minmax(0, 1fr);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.mode,
		.segment,
		.row {
			transition: none;
		}
	}
</style>
