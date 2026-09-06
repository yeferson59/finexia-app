<script lang="ts">
	/*
	 * Dónde tienes el activo.
	 *
	 * Era una tarjeta con sombra propia para un solo desplegable, y el caso sin
	 * plataformas era un recuadro de borde discontinuo, centrado, con un botón
	 * con un «+» dentro. Ahora es un campo más del bloque que lo contiene, y el
	 * caso vacío es una frase con un enlace: no hay nada que decidir ahí, solo un
	 * sitio al que ir.
	 */
	import { resolve } from '$app/paths';
	import type { Platform } from '$lib/api/types';

	let { platforms, selected = $bindable('') }: { platforms: Platform[]; selected?: string } =
		$props();
</script>

<div class="field">
	<label for="platformId">Plataforma</label>

	{#if platforms.length > 0}
		<select id="platformId" name="platformId" bind:value={selected} required>
			<option value="">Elige una plataforma</option>
			{#each platforms as platform (platform.id)}
				<option value={platform.id}>{platform.name}</option>
			{/each}
		</select>
	{:else}
		<p class="hint">
			Todavía no tienes ninguna plataforma registrada, y una posición vive en una de ellas.
			<a href={resolve('/dashboard/platforms/add')}>Crea la primera</a> y vuelve aquí.
		</p>
	{/if}
</div>
