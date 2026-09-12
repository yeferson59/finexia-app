<script lang="ts">
	/*
	 * Los avisos de que la fecha de una operación no parece la suya (ver
	 * `tradeDateWarnings`). Van debajo del campo de fecha en los dos formularios
	 * que la piden —el alta de un activo y la transacción sobre una posición— y
	 * dejan guardar: son avisos, no errores.
	 *
	 * Recibe la operación sin «hoy» porque hoy es el del calendario de quien la
	 * escribe, y eso lo sabe el navegador, no el formulario.
	 */
	import { todayLocalDateString } from '$lib/shared/format/date';
	import { tradeDateWarnings, type TradeDateCheck } from '../asset';

	let { check }: { check: Omit<TradeDateCheck, 'today'> | null } = $props();

	const warnings = $derived(
		check ? tradeDateWarnings({ ...check, today: todayLocalDateString() }) : []
	);
</script>

{#each warnings as warning (warning)}
	<p class="date-warning" role="status">{warning}</p>
{/each}

<style>
	/* En el tono de lo que dice y sin caja, como el resto de avisos de campo. */
	.date-warning {
		margin: 0;
		font-size: 0.8rem;
		line-height: 1.5;
		color: var(--amber);
	}
</style>
