<script lang="ts">
	/**
	 * Un bloque de la página: a la izquierda qué canal es y a la derecha lo que
	 * se puede decidir sobre él.
	 *
	 * Sustituye a la tarjeta que envolvía cada canal. La página eran dos tarjetas
	 * iguales una al lado de la otra, así que el canal que todavía no existe
	 * ocupaba la mitad de la pantalla para decir que no existe. En un carril cada
	 * bloque ocupa lo que tiene que contar.
	 *
	 * Es el mismo carril que usa la página de configuración, y a propósito: son
	 * vecinas en el menú y se hojean igual. Aporta con `:global`, acotadas al
	 * bloque, las clases de formulario que usa su contenido (`.form-actions`,
	 * `.feedback`): son nombres genéricos que en una hoja global chocarían con el
	 * resto del panel.
	 */
	import type { Snippet } from 'svelte';

	interface Props {
		title: string;
		/** Qué es el canal, en una frase; va bajo el título. */
		description: string;
		children: Snippet;
	}

	let { title, description, children }: Props = $props();
</script>

<section class="channel">
	<div class="rail">
		<h2>{title}</h2>
		<p class="description">{description}</p>
	</div>

	<div class="controls">
		{@render children()}
	</div>
</section>

<style>
	.channel {
		display: grid;
		grid-template-columns: minmax(0, 17rem) minmax(0, 1fr);
		gap: 1rem 3rem;
		padding: 2.25rem 0;
		border-top: 1px solid var(--border-strong);
	}

	.rail {
		min-width: 0;
	}

	h2 {
		margin: 0;
		font-family: var(--font-display);
		font-size: 1.35rem;
		font-weight: 300;
		letter-spacing: -0.02em;
		color: var(--text);
	}

	.description {
		max-width: 34ch;
		margin: 0.5rem 0 0;
		font-size: 0.83rem;
		line-height: 1.55;
		color: var(--text-muted);
	}

	.controls {
		min-width: 0;
		max-width: 38rem;
	}

	/* El botón sin el halo ámbar que `ui/button` le pone: el ámbar de esta
	   página dice qué avisos están encendidos, y encendía también el botón. */
	.channel :global(.btn-primary) {
		box-shadow: none;
	}

	.channel :global(.form-actions) {
		margin-top: 1.5rem;
	}

	/* Los avisos son prosa con un filete del color de lo que dicen, no cajas de
	   alerta: es el idioma que ya hablan configuración y el panel. */
	.channel :global(.feedback) {
		max-width: 62ch;
		margin: 1rem 0 0;
		padding-left: 0.75rem;
		border-left: 2px solid;
		font-size: 0.83rem;
		line-height: 1.5;
	}

	.channel :global(.feedback.success) {
		border-color: var(--green);
		color: var(--green);
	}

	.channel :global(.feedback.error) {
		border-color: var(--red);
		color: var(--red);
	}

	@media (max-width: 900px) {
		.channel {
			grid-template-columns: minmax(0, 1fr);
			gap: 1.25rem;
		}

		.controls {
			max-width: none;
		}
	}
</style>
