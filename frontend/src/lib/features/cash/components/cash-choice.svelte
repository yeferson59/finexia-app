<script lang="ts" generics="T extends string">
	/**
	 * Una elección a la vista: radios con aspecto de teclas, para las decisiones
	 * que cambian lo que pasa —qué tipo de movimiento, qué se le hace a una tasa,
	 * a cuántos días va un depósito—. Un desplegable escondía justo la opción que
	 * había que leer antes de elegir.
	 *
	 * Cada opción puede llevar una segunda línea que dice qué hace. El radio sigue
	 * ahí para el teclado y el lector; lo que se ve es la etiqueta.
	 */
	interface Option {
		value: T;
		label: string;
		/** Qué hace, en pocas palabras, bajo la etiqueta. */
		detail?: string;
		disabled?: boolean;
	}

	interface Props {
		legend: string;
		/** El nombre del campo, si viaja con el formulario. */
		name?: string;
		options: Option[];
		value: T;
		/** Columnas en ancho; en estrecho bajan a dos. */
		columns?: number;
		/** Quién explica la elección, debajo. */
		describedby?: string;
	}

	let { legend, name, options, value = $bindable(), columns, describedby }: Props = $props();

	const group = $props.id();
</script>

<fieldset class="choice" aria-describedby={describedby}>
	<legend class="field-label">{legend}</legend>
	<div class="keys" style:--columns={columns ?? options.length}>
		{#each options as option (option.value)}
			<label class="key" class:selected={value === option.value} class:disabled={option.disabled}>
				<input
					type="radio"
					name={name ?? group}
					value={option.value}
					disabled={option.disabled}
					bind:group={value}
				/>
				<span class="label">{option.label}</span>
				{#if option.detail}
					<span class="detail">{option.detail}</span>
				{/if}
			</label>
		{/each}
	</div>
</fieldset>

<style>
	.choice {
		min-width: 0;
		margin: 0;
		padding: 0;
		border: none;
	}

	legend {
		margin-bottom: 0.5rem;
		padding: 0;
	}

	.keys {
		display: grid;
		grid-template-columns: repeat(var(--columns), minmax(0, 1fr));
		gap: 0.4rem;
	}

	.key {
		position: relative;
		display: flex;
		flex-direction: column;
		justify-content: center;
		gap: 0.1rem;
		min-height: 2.75rem;
		padding: 0.55rem 0.75rem;
		border: 1px solid var(--border-strong);
		border-radius: 8px;
		background: rgba(255, 255, 255, 0.02);
		color: var(--text-muted);
		cursor: pointer;
		transition:
			border-color 0.15s ease,
			background-color 0.15s ease,
			color 0.15s ease;
	}

	.key:hover:not(.disabled) {
		border-color: rgba(212, 145, 42, 0.4);
		color: var(--text);
	}

	.key input {
		position: absolute;
		opacity: 0;
		pointer-events: none;
	}

	.key:has(input:focus-visible) {
		outline: 2px solid var(--amber);
		outline-offset: 2px;
	}

	/* Elegida: el borde ámbar de las acciones y un fondo apenas tibio. La marca
	   de la esquina lo dice también sin color. */
	.key.selected {
		border-color: var(--amber);
		background: rgba(212, 145, 42, 0.08);
		color: var(--text);
	}

	.key.selected::after {
		content: '';
		position: absolute;
		top: 0.5rem;
		right: 0.5rem;
		width: 6px;
		height: 6px;
		border-radius: 50%;
		background: var(--amber);
	}

	.key.disabled {
		cursor: not-allowed;
		opacity: 0.45;
	}

	.label {
		font-size: 0.88rem;
		font-weight: 500;
		line-height: 1.25;
	}

	.detail {
		font-size: 0.74rem;
		line-height: 1.3;
		color: var(--text-dim);
	}

	.selected .detail {
		color: var(--text-muted);
	}

	@media (max-width: 640px) {
		.keys {
			grid-template-columns: repeat(min(var(--columns), 2), minmax(0, 1fr));
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.key {
			transition: none;
		}
	}
</style>
