<script lang="ts">
	/**
	 * Una sección de ajustes: a la izquierda qué es y a la derecha los controles.
	 *
	 * Sustituye a la tarjeta que envolvía cada sección. La página era una rejilla
	 * de ocho tarjetas de dos columnas, así que un párrafo de dos líneas se
	 * estiraba hasta el alto del formulario de al lado y el orden de lectura iba
	 * en zigzag. En un carril, el ojo recorre solo los títulos hasta encontrar el
	 * recado que trae a esta página.
	 *
	 * El carril tiene además la ventaja de que los campos dejan de medir mil
	 * píxeles de ancho: una contraseña no se escribe en un campo de esa longitud.
	 *
	 * Aporta con `:global`, acotadas a la sección, las clases de formulario que
	 * usa su contenido (`.form-fields`, `.form-actions`, `.hint`, `.feedback`).
	 * Son nombres genéricos: en una hoja global colisionarían con el resto del
	 * dashboard, y copiados en cada sección acabarían divergiendo.
	 */
	import type { Snippet } from 'svelte';

	interface Props {
		title: string;
		/** Qué hace la sección, en una o dos frases; va bajo el título. */
		description?: string;
		id?: string;
		/** Contenido extra del carril, bajo la descripción. */
		aside?: Snippet;
		children: Snippet;
	}

	let { title, description = '', id = undefined, aside, children }: Props = $props();
</script>

<section class="section" {id}>
	<div class="rail">
		<h3>{title}</h3>
		{#if description}
			<p class="description">{description}</p>
		{/if}
		{#if aside}
			{@render aside()}
		{/if}
	</div>

	<div class="controls">
		{@render children()}
	</div>
</section>

<style>
	.section {
		display: grid;
		grid-template-columns: minmax(0, 17rem) minmax(0, 1fr);
		gap: 1rem 3rem;
		padding: 1.75rem 0;
		border-bottom: 1px solid var(--border);
	}

	/* El último de un grupo no lleva filete: lo pone el grupo siguiente. */
	.section:last-child {
		border-bottom: none;
		padding-bottom: 0.5rem;
	}

	.rail {
		min-width: 0;
	}

	h3 {
		margin: 0;
		font-family: var(--font-body);
		font-size: 1rem;
		font-weight: 500;
		color: var(--text);
	}

	.description {
		max-width: 40ch;
		margin: 0.45rem 0 0;
		font-size: 0.83rem;
		line-height: 1.55;
		color: var(--text-muted);
	}

	/* Los controles no se estiran a lo ancho de la página: un campo de texto de
	   mil píxeles no ayuda a escribir en él. */
	.controls {
		min-width: 0;
		max-width: 34rem;
	}

	.section :global(.form-fields) {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	/*
	 * Los botones de una fila —cerrar una sesión, desconectar una aplicación,
	 * rotar o borrar un token— son el mismo gesto sobre la misma clase de cosa,
	 * así que se ven igual. Callados hasta que se apuntan: son irreversibles,
	 * pero no son lo que se viene a hacer a esta página. Vivían copiados en tres
	 * secciones con tres aspectos distintos.
	 */
	.section :global(.row-action) {
		flex-shrink: 0;
		padding: 0.35rem 0.7rem;
		border: 1px solid var(--border-strong);
		border-radius: 7px;
		background: transparent;
		font-family: var(--font-body);
		font-size: 0.78rem;
		color: var(--text-muted);
		cursor: pointer;
		transition:
			color 0.2s ease,
			border-color 0.2s ease;
	}

	.section :global(.row-action:hover:not(:disabled)) {
		border-color: rgba(212, 145, 42, 0.5);
		color: var(--text);
	}

	.section :global(.row-action.danger:hover:not(:disabled)) {
		border-color: var(--red);
		color: var(--red);
	}

	.section :global(.row-action:disabled) {
		cursor: default;
		opacity: 0.6;
	}

	@media (prefers-reduced-motion: reduce) {
		.section :global(.row-action) {
			transition: none;
		}
	}

	/*
	 * El botón de una sección no es el de la portada: sin el halo ámbar que
	 * `ui/button` le pone, que en una página con seis formularios encendía seis
	 * luces. Es el mismo camino que ya tomaron el listado de portafolios y la
	 * ficha de un activo con sus enlaces de acción.
	 */
	.section :global(.btn-primary) {
		box-shadow: none;
	}

	.section :global(.form-actions) {
		margin-top: 1.25rem;
		display: flex;
		justify-content: flex-start;
	}

	.section :global(.hint) {
		margin: 0;
		font-size: 0.83rem;
		line-height: 1.55;
		color: var(--text-muted);
	}

	/* Los avisos son prosa con un filete del color de lo que dicen, no cajas de
	   alerta: es el idioma que ya hablan el panel y la vista de un activo. */
	.section :global(.feedback) {
		max-width: 62ch;
		margin: 0.85rem 0 0;
		padding-left: 0.75rem;
		border-left: 2px solid;
		font-size: 0.83rem;
		line-height: 1.5;
	}

	.section :global(.feedback.success) {
		border-color: var(--green);
		color: var(--green);
	}

	.section :global(.feedback.error) {
		border-color: var(--red);
		color: var(--red);
	}

	.section :global(.feedback.warning) {
		border-color: rgba(212, 145, 42, 0.45);
		color: var(--amber);
	}

	@media (max-width: 900px) {
		.section {
			grid-template-columns: minmax(0, 1fr);
			gap: 1rem;
		}

		.controls {
			max-width: none;
		}
	}
</style>
