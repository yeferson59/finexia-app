<script lang="ts">
	/**
	 * Cómo se clasifica un activo del catálogo: una industria, o el desglose de
	 * las varias en las que está repartido.
	 *
	 * Vive en un componente propio porque el alta y la edición piden lo mismo y
	 * son once casillas: duplicarlas era duplicar también la regla que las
	 * gobierna, que es la parte que no puede desalinearse.
	 *
	 * Las dos formas son **excluyentes** y el formulario lo hace cumplir
	 * desactivando la que no está en uso, en vez de dejar que el backend
	 * conteste 400. Un campo desactivado no se envía, así que «desactivado»
	 * también significa «no lo mandes», que es exactamente lo que hace falta:
	 * poner un desglose borra la industria única y viceversa.
	 *
	 * Por qué existe el desglose: una acción y un ETF sectorial caben en una
	 * industria —Apple es tecnología, XLK es tecnología— y un fondo de mercado
	 * ancho no. Un VOO es el S&P 500 entero; ficharlo bajo «Tecnología» metería
	 * dos tercios de la posición en industrias en las que no está, y dejarlo sin
	 * clasificar haría que el panel contara la mayor posición de la cartera como
	 * trabajo pendiente que nadie puede hacer.
	 */
	import { SECTOR_OPTIONS, sectorWeightsTotal } from '$lib/shared/format/sector';
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

	const filled = $derived(
		SECTOR_OPTIONS.map((s) => ({ sector: s.value, weight: weights[s.value] })).filter(
			(w): w is { sector: string; weight: number } =>
				typeof w.weight === 'number' && Number.isFinite(w.weight)
		)
	);

	const weighted = $derived(filled.length > 0);
	const total = $derived(sectorWeightsTotal(filled));
</script>

<div class="field">
	<label for="{idPrefix}sector">Industria <span class="optional">(opcional)</span></label>
	<select id="{idPrefix}sector" name="sector" bind:value={sector} disabled={weighted}>
		<option value="">Sin clasificar</option>
		{#each SECTOR_OPTIONS as s (s.value)}
			<option value={s.value}>{s.label}</option>
		{/each}
	</select>
	<p class="hint">
		{#if weighted}
			El desglose de abajo ocupa su lugar: un activo lleva una industria o un reparto entre varias,
			nunca las dos.
		{:else}
			Es lo que reparte el panel por industria. Dejarlo sin clasificar no rompe nada: esa parte del
			patrimonio se cuenta aparte, con su propia etiqueta.
		{/if}
	</p>
</div>

<details class="breakdown" open={weighted}>
	<summary>Desglose por industrias <span class="optional">(fondos)</span></summary>

	<p class="hint">
		Para un ETF de mercado ancho, que está en todas a la vez. Se copian de la ficha del fondo, en
		porcentaje. <strong>No hace falta que sumen 100</strong>: el reparto normaliza sobre lo que
		haya, así que un desglose al 97 % sigue explicando el 100 % del dinero del fondo.
	</p>

	<div class="weights">
		{#each SECTOR_OPTIONS as s (s.value)}
			<div class="weight">
				<label for="{idPrefix}weight-{s.value}">{s.label}</label>
				<input
					id="{idPrefix}weight-{s.value}"
					type="number"
					name="weight.{s.value}"
					bind:value={weights[s.value]}
					min="0"
					max="100"
					step="any"
					placeholder="—"
					disabled={sector !== ''}
				/>
			</div>
		{/each}
	</div>

	{#if sector !== ''}
		<p class="hint">
			Para repartir este activo entre varias industrias, vuelve la de arriba a «Sin clasificar».
		</p>
	{:else if weighted}
		<p class="total" class:over={total > 100}>
			Suma: {formatPercent(total)}
			{#if total > 100}
				— por encima de 100 %, revisa las cifras
			{/if}
		</p>
	{/if}
</details>

<style>
	.breakdown {
		border: 1px solid var(--border);
		border-radius: 0.5rem;
		padding: 0.6rem 0.75rem;
	}

	.breakdown summary {
		cursor: pointer;
		font-size: 0.82rem;
		font-weight: 600;
	}

	.breakdown .hint {
		margin: 0.5rem 0 0;
	}

	.weights {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(9.5rem, 1fr));
		gap: 0.5rem 0.75rem;
		margin-top: 0.6rem;
	}

	.weight {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.5rem;
	}

	.weight label {
		font-size: 0.78rem;
		color: var(--text-dim);
	}

	.weight input {
		width: 5rem;
		text-align: right;
	}

	.total {
		margin: 0.6rem 0 0;
		font-size: 0.8rem;
		font-variant-numeric: tabular-nums;
		color: var(--text-dim);
	}

	.total.over {
		color: var(--red);
	}

	.hint {
		margin: 0.35rem 0 0;
		font-size: 0.78rem;
		line-height: 1.45;
		color: var(--text-dim);
	}
</style>
