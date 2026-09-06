<script lang="ts">
	/**
	 * Uno de los dos bloques de la página: cómo entras y qué tienes conectado.
	 *
	 * El resumen no es un adorno: es la única forma de saber si algo pide
	 * atención sin recorrer los formularios uno a uno. El estado vivía repartido
	 * en insignias dentro de ocho tarjetas, y desde ninguna se veía otra.
	 */
	import type { Snippet } from 'svelte';

	interface Props {
		title: string;
		/** Cómo está hoy lo que contiene el grupo, en una frase. */
		summary: string;
		children: Snippet;
	}

	let { title, summary, children }: Props = $props();
</script>

<section class="group">
	<h2>{title}</h2>
	<p class="summary">{summary}</p>
	{@render children()}
</section>

<style>
	.group {
		padding-top: 2.5rem;
		border-top: 1px solid var(--border-strong);
	}

	/*
	 * En la tipografía de la portada, como el título de la página: son los dos
	 * puntos de referencia de una página larga, y desde arriba tienen que
	 * distinguirse del título de cada sección.
	 */
	h2 {
		margin: 0;
		font-family: var(--font-display);
		font-size: 1.6rem;
		font-weight: 300;
		letter-spacing: -0.02em;
		color: var(--text);
	}

	.summary {
		max-width: 64ch;
		margin: 0.6rem 0 0;
		font-size: 0.9rem;
		line-height: 1.55;
		color: var(--text-muted);
	}
</style>
