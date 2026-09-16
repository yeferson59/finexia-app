<script lang="ts">
	/**
	 * Un campo numérico con su unidad escrita dentro: la moneda de un importe, el
	 * «% E.A.» de una tasa. Estaba copiado en cada formulario del efectivo con
	 * rellenos distintos, y la unidad se montaba sobre la cifra en unos y no en
	 * otros.
	 *
	 * `size="lg"` es para la cifra que manda en el formulario —el importe de un
	 * movimiento, el de un depósito—: se escribe en la letra de las cifras y a un
	 * tamaño que se lee de lejos, como se leerá luego en la lista.
	 *
	 * Sin flechas: con `step="any"` suben de uno en uno, que en un importe o una
	 * tasa no sirve, y tapaban la unidad.
	 */
	import type { HTMLInputAttributes } from 'svelte/elements';

	interface Props extends Omit<HTMLInputAttributes, 'value' | 'size' | 'type'> {
		id: string;
		/** Lo que se escribe; al teclear llega como número. */
		value?: string | number | null;
		unit: string;
		size?: 'md' | 'lg';
	}

	let { id, value = $bindable(''), unit, size = 'md', ...rest }: Props = $props();
</script>

<div class="money" class:lg={size === 'lg'} style:--unit-width="{Math.max(unit.length, 2)}ch">
	<input
		{id}
		type="number"
		inputmode="decimal"
		step="any"
		min="0"
		{...rest}
		aria-describedby={[`${id}-unit`, rest['aria-describedby']].filter(Boolean).join(' ')}
		bind:value
	/>
	<span class="unit" id="{id}-unit">{unit}</span>
</div>

<style>
	.money {
		position: relative;
	}

	/* `.rail-fields` pinta el campo; aquí solo el sitio de la unidad y la cifra. */
	.money input[type='number'] {
		padding-right: calc(var(--unit-width) + 1.9rem);
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
		appearance: textfield;
		-moz-appearance: textfield;
	}

	.money input::-webkit-inner-spin-button,
	.money input::-webkit-outer-spin-button {
		-webkit-appearance: none;
		margin: 0;
	}

	.lg input[type='number'] {
		padding-top: 0.85rem;
		padding-bottom: 0.85rem;
		font-size: 1.45rem;
		letter-spacing: -0.02em;
	}

	.unit {
		position: absolute;
		top: 50%;
		right: 0.95rem;
		transform: translateY(-50%);
		font-size: 0.78rem;
		font-weight: 600;
		letter-spacing: 0.04em;
		color: var(--text-dim);
		pointer-events: none;
	}

	.lg .unit {
		font-size: 0.85rem;
	}
</style>
