<script lang="ts">
	/**
	 * El carril: a la izquierda qué se decide en el bloque, a la derecha dónde se
	 * decide.
	 *
	 * Es la forma de configuración, notificaciones, el alta de portafolio y la
	 * importación desde Excel. Cada una tenía su copia —`settings-section`,
	 * `notification-section`, `portfolio-form-section`, `import-section`— y las
	 * cuatro habían empezado a separarse: el carril medía 17rem en tres y 19rem
	 * en la cuarta, el aviso de error llevaba dos márgenes distintos, los botones
	 * de fila tres aspectos, y los estilos de los campos estaban escritos dos
	 * veces con dos tamaños de letra. Aquí están una sola vez.
	 *
	 * Aporta con `:global`, acotadas al bloque, las clases que solo tienen sentido
	 * dentro de un carril (`.form-fields`, `.form-actions`, `.row-action`, y a qué
	 * distancia queda un aviso). El aspecto de los campos, las ayudas y los avisos
	 * es el mismo dentro y fuera de un carril, así que vive en
	 * `routes/layout.css`; `fields` solo enciende aquí la clase que lo trae.
	 */
	import type { Snippet } from 'svelte';

	interface Props {
		title: string;
		/** Qué se decide en el bloque, en una o dos frases. */
		description?: string;
		id?: string;
		/**
		 * `2` para un bloque de primer nivel, con la tipografía de la portada;
		 * `3` cuando vive dentro de un grupo que ya puso su `<h2>`.
		 */
		level?: 2 | 3;
		/** Dónde va el filete que lo separa del bloque vecino. */
		divider?: 'top' | 'bottom' | 'none';
		/**
		 * Ancho máximo del contenido. Los campos no se estiran a lo ancho de la
		 * página: el nombre de un portafolio no se escribe en una caja de mil
		 * píxeles. `none` para lo que sí pide el ancho entero, como una tabla.
		 */
		contentMax?: string;
		/** Apila el contenido como columna de campos y le aporta su CSS. */
		fields?: boolean;
		/** Contenido extra del carril, bajo la descripción. */
		aside?: Snippet;
		children: Snippet;
	}

	let {
		title,
		description = '',
		id = undefined,
		level = 2,
		divider = 'top',
		contentMax = '34rem',
		fields = false,
		aside,
		children
	}: Props = $props();
</script>

<section
	{id}
	class="rail-section"
	class:divider-top={divider === 'top'}
	class:divider-bottom={divider === 'bottom'}
	class:nested={level === 3}
	style="--rail-content-max: {contentMax}"
>
	<div class="rail">
		<svelte:element this={`h${level}`} class="title">{title}</svelte:element>
		{#if description}
			<p class="description">{description}</p>
		{/if}
		{#if aside}
			{@render aside()}
		{/if}
	</div>

	<div class="content" class:rail-fields={fields}>
		{@render children()}
	</div>
</section>

<style>
	.rail-section {
		display: grid;
		grid-template-columns: minmax(0, 17rem) minmax(0, 1fr);
		gap: 1rem 3rem;
		padding: 2.25rem 0;
	}

	.divider-top {
		border-top: 1px solid var(--border-strong);
	}

	.divider-bottom {
		border-bottom: 1px solid var(--border);
	}

	/* Dentro de un grupo se respira menos: el grupo ya puso el aire de arriba. */
	.nested {
		padding: 1.75rem 0;
	}

	/* El último de un grupo no lleva filete: lo pone el grupo siguiente. */
	.divider-bottom:last-child {
		padding-bottom: 0.5rem;
		border-bottom: none;
	}

	.rail {
		min-width: 0;
	}

	.title {
		margin: 0;
		font-family: var(--font-display);
		font-size: 1.35rem;
		font-weight: 300;
		letter-spacing: -0.02em;
		color: var(--text);
	}

	.nested .title {
		font-family: var(--font-body);
		font-size: 1rem;
		font-weight: 500;
		letter-spacing: normal;
	}

	.description {
		max-width: 38ch;
		margin: 0.5rem 0 0;
		font-size: 0.83rem;
		line-height: 1.55;
		color: var(--text-muted);
	}

	.content {
		min-width: 0;
		max-width: var(--rail-content-max);
	}

	/* --- Lo que comparte el contenido de cualquier bloque ------------------- */

	.rail-section :global(.form-fields) {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.rail-section :global(.form-actions) {
		display: flex;
		justify-content: flex-start;
		margin-top: 1.25rem;
	}

	/*
	 * El aspecto de los campos, las ayudas y los avisos vive en
	 * `routes/layout.css`: son los mismos dentro y fuera de un carril, y aquí
	 * eran una copia más de las cinco que había. De ellos, lo único que es de
	 * este bloque es a qué distancia queda el aviso de lo que tiene encima.
	 */
	.rail-section :global(.feedback) {
		margin: 0.85rem 0 0;
	}

	/*
	 * Los botones de una fila —cerrar una sesión, desconectar una aplicación,
	 * rotar o borrar un token— son el mismo gesto sobre la misma clase de cosa,
	 * así que se ven igual. Callados hasta que se apuntan: son irreversibles,
	 * pero no son lo que se viene a hacer a la página.
	 */
	.rail-section :global(.row-action) {
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

	.rail-section :global(.row-action:hover:not(:disabled)) {
		border-color: rgba(212, 145, 42, 0.5);
		color: var(--text);
	}

	.rail-section :global(.row-action.danger:hover:not(:disabled)) {
		border-color: var(--red);
		color: var(--red);
	}

	.rail-section :global(.row-action:disabled) {
		cursor: default;
		opacity: 0.6;
	}

	/*
	 * El botón de un bloque no es el de la portada: sin el halo ámbar que le pone
	 * `ui/button`, que en una página con seis formularios encendía seis luces.
	 */
	.rail-section :global(.btn-primary) {
		box-shadow: none;
	}

	@media (prefers-reduced-motion: reduce) {
		.rail-section :global(.row-action) {
			transition: none;
		}
	}

	@media (max-width: 900px) {
		.rail-section {
			grid-template-columns: minmax(0, 1fr);
			gap: 1.25rem;
		}

		.content {
			max-width: none;
		}
	}
</style>
