<script lang="ts">
	/*
	 * Elección del nivel de riesgo.
	 *
	 * Eran tres cajas con borde y fondo, una debajo de otra: tres tarjetas para
	 * tres palabras. Ahora son filas con filete, como las opciones de correo de
	 * notificaciones y las sesiones abiertas de configuración, y la fila entera
	 * es la etiqueta del radio.
	 */
	import type { Risk } from '$lib/api/types';

	let {
		risks,
		selected = $bindable(''),
		disabled = false
	}: { risks: Risk[]; selected?: string; disabled?: boolean } = $props();
</script>

<fieldset class="risk">
	<legend class="field-label">Nivel de riesgo</legend>

	<div class="options">
		{#each risks as risk (risk.id)}
			<label class="option">
				<input
					type="radio"
					name="riskId"
					value={risk.id}
					bind:group={selected}
					{disabled}
					required
					aria-describedby="risk-{risk.id}"
				/>
				<span class="text">
					<span class="name">{risk.name}</span>
					<span class="description" id="risk-{risk.id}">{risk.description}</span>
				</span>
			</label>
		{/each}
	</div>
</fieldset>

<style>
	.risk {
		min-width: 0;
		margin: 0;
		padding: 0;
		border: none;
	}

	.field-label {
		padding: 0;
		font-size: 0.87rem;
		font-weight: 500;
		color: var(--text);
	}

	.options {
		margin-top: 0.45rem;
	}

	/* El radio delante y la fila entera como etiqueta: puestos en columna se ve
	   de un vistazo cuál está marcado. */
	.option {
		display: grid;
		grid-template-columns: auto minmax(0, 1fr);
		align-items: start;
		gap: 0.9rem;
		padding: 0.85rem 0;
		border-bottom: 1px solid var(--border);
		cursor: pointer;
	}

	.option:last-child {
		border-bottom: none;
	}

	.option input[type='radio'] {
		width: 18px;
		height: 18px;
		margin: 0.1rem 0 0;
		cursor: pointer;
	}

	.option input[type='radio']:disabled {
		cursor: not-allowed;
		opacity: 0.6;
	}

	.text {
		min-width: 0;
	}

	.name {
		display: block;
		font-size: 0.9rem;
		font-weight: 500;
		color: var(--text);
	}

	.description {
		display: block;
		max-width: 46ch;
		margin-top: 0.2rem;
		font-size: 0.8rem;
		line-height: 1.55;
		color: var(--text-muted);
	}
</style>
