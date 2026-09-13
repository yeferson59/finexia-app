<script lang="ts">
	/**
	 * Una fila del reparto entre industrias: la etiqueta y su peso.
	 *
	 * Vivía como snippet dentro de `sector-classification`, que con sus estilos
	 * pasó del presupuesto de tamaño. El puntero y el foco los guarda el padre,
	 * porque es él quien enciende el tramo de la barra que corresponde a la fila.
	 */
	interface Props {
		/** Clave del sector (`technology`, `cash`…). */
		value: string;
		label: string;
		/**
		 * El peso en porcentaje. Sin valor por defecto a propósito: la mayoría de
		 * las filas llegan sin peso, y `bind:` sobre `undefined` no admite uno.
		 */
		weight: number | null | undefined;
		/** Prefijo de los `id`, para que los dos formularios no colisionen. */
		idPrefix: string;
		active: boolean;
		onHover: (value: string | null) => void;
		onFocus: (value: string | null) => void;
	}

	let { value, label, weight = $bindable(), idPrefix, active, onHover, onFocus }: Props = $props();

	const filled = $derived(typeof weight === 'number' && Number.isFinite(weight));
</script>

<div
	class="row"
	class:filled
	class:active
	role="presentation"
	onpointerenter={() => onHover(value)}
	onpointerleave={() => onHover(null)}
	onfocusin={() => onFocus(value)}
	onfocusout={() => onFocus(null)}
>
	<label for="{idPrefix}weight-{value}">{label}</label>
	<span class="cell">
		<input
			id="{idPrefix}weight-{value}"
			type="number"
			name="weight.{value}"
			bind:value={weight}
			min="0"
			max="100"
			step="any"
			inputmode="decimal"
			placeholder="—"
			aria-invalid={filled && (weight ?? 0) <= 0}
		/>
		<span class="unit" aria-hidden="true">%</span>
	</span>
</div>

<style>
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

	@media (prefers-reduced-motion: reduce) {
		.row {
			transition: none;
		}
	}
</style>
