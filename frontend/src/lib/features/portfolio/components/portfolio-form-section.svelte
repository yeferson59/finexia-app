<script lang="ts">
	/**
	 * Un bloque del alta de un portafolio: a la izquierda qué se decide aquí y a
	 * la derecha los campos.
	 *
	 * Sustituye a las tres tarjetas apiladas que tenía el formulario, cada una
	 * con su sombra y su leyenda EN VERSALITAS ÁMBAR. Es el mismo carril que la
	 * página de configuración: siete controles seguidos son una lista larga, y en
	 * tres bloques rotulados se sabe en cuál está lo que se busca.
	 *
	 * Aporta con `:global`, acotadas al bloque, las clases de formulario que usan
	 * sus hijos —el selector de riesgo y los campos de la meta viven en sus
	 * propios componentes—. Son nombres genéricos: en una hoja global chocarían
	 * con el resto del panel, y copiados en cada componente acabarían divergiendo,
	 * que es justo lo que le pasaba a este formulario.
	 */
	import type { Snippet } from 'svelte';

	interface Props {
		title: string;
		/** Qué se decide en el bloque, en una o dos frases. */
		description: string;
		children: Snippet;
	}

	let { title, description, children }: Props = $props();
</script>

<section class="section">
	<div class="rail">
		<h2>{title}</h2>
		<p>{description}</p>
	</div>

	<div class="fields">
		{@render children()}
	</div>
</section>

<style>
	.section {
		display: grid;
		grid-template-columns: minmax(0, 19rem) minmax(0, 1fr);
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

	.rail p {
		max-width: 38ch;
		margin: 0.6rem 0 0;
		font-size: 0.83rem;
		line-height: 1.6;
		color: var(--text-muted);
	}

	/* Los campos no se estiran a lo ancho de la página: el nombre de un
	   portafolio no se escribe en una caja de mil píxeles. */
	.fields {
		display: flex;
		flex-direction: column;
		gap: 1.35rem;
		min-width: 0;
		max-width: 34rem;
	}

	/* Dos campos cortos comparten fila; por debajo de 640 se apilan. */
	.section :global(.pair) {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 1.35rem;
	}

	.section :global(.field) {
		display: flex;
		flex-direction: column;
		gap: 0.45rem;
		min-width: 0;
	}

	.section :global(.field > label),
	.section :global(.field-label) {
		font-size: 0.87rem;
		font-weight: 500;
		color: var(--text);
	}

	/* Se marca lo opcional, no lo obligatorio: de siete controles casi todos lo
	   son, así que los asteriscos rojos señalaban la pantalla entera. */
	.section :global(.optional) {
		font-weight: 400;
		color: var(--text-dim);
	}

	/*
	 * `dp-select` fuera: son los tres huecos —día, mes, año— del selector de
	 * fecha de `ui/date-picker`, que trae su propio ancho. Con el de aquí, los
	 * tres se estiraban al 100 % y el de en medio se quedaba sin sitio.
	 */
	.section :global(input[type='text']),
	.section :global(input[type='number']),
	.section :global(select:not(.dp-select)),
	.section :global(textarea) {
		width: 100%;
		padding: 0.8rem 0.95rem;
		border: 1px solid rgba(212, 145, 42, 0.2);
		border-radius: 8px;
		background: rgba(255, 255, 255, 0.03);
		color: var(--text);
		font-family: var(--font-body);
		font-size: 0.92rem;
		box-sizing: border-box;
		transition: border-color 0.2s ease;
	}

	.section :global(input::placeholder),
	.section :global(textarea::placeholder) {
		color: var(--text-dim);
	}

	.section :global(input:hover:not(:disabled)),
	.section :global(select:not(.dp-select):hover:not(:disabled)),
	.section :global(textarea:hover:not(:disabled)) {
		border-color: rgba(212, 145, 42, 0.35);
	}

	/* El anillo de foco lo pinta la regla global de `layout.css`, la misma en
	   toda la aplicación; aquí solo se enciende el borde. */
	.section :global(input:focus),
	.section :global(select:not(.dp-select):focus),
	.section :global(textarea:focus) {
		border-color: var(--amber);
	}

	.section :global(input:disabled),
	.section :global(select:not(.dp-select):disabled),
	.section :global(textarea:disabled) {
		color: var(--text-muted);
		cursor: not-allowed;
	}

	.section :global(textarea) {
		resize: vertical;
		min-height: 5.5rem;
		line-height: 1.55;
	}

	.section :global(.hint) {
		margin: 0;
		font-size: 0.8rem;
		line-height: 1.5;
		color: var(--text-muted);
	}

	@media (prefers-reduced-motion: reduce) {
		.section :global(input),
		.section :global(select),
		.section :global(textarea) {
			transition: none;
		}
	}

	@media (max-width: 900px) {
		.section {
			grid-template-columns: minmax(0, 1fr);
			gap: 1.5rem;
		}

		.fields {
			max-width: none;
		}
	}

	@media (max-width: 640px) {
		.section :global(.pair) {
			grid-template-columns: minmax(0, 1fr);
		}
	}
</style>
