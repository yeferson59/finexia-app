<script lang="ts">
	/**
	 * Un bloque de administración: qué contiene, cómo está y su tabla.
	 *
	 * Interno de la feature. Sustituye a las dos piezas que había —una tarjeta
	 * con título en versalitas para las tablas de usuarios y otra tarjeta para
	 * las del catálogo—, que se habían separado en el aspecto de las mismas
	 * celdas: el correo a 0,8 rem aquí y el ticker a 0,85 allá, dos rejillas de
	 * acciones y dos formas de pintar una fila con error.
	 *
	 * Dos decisiones de forma viven aquí y valen para las cinco tablas:
	 *
	 * - No hay tarjeta. Una tabla ya trae sus propias líneas; meterla en una caja
	 *   con borde y esquinas redondeadas es dibujar el mismo rectángulo dos
	 *   veces. La tabla se apoya en la página entre dos filetes, como un libro
	 *   de cuentas.
	 * - El bloque abre con una frase que dice qué le pasa a lo que contiene, no
	 *   con una etiqueta que repite el nombre de la tabla.
	 */
	import type { Snippet } from 'svelte';

	interface Props {
		/** Qué contiene el bloque. Es el encabezado de la tabla que va debajo. */
		title: string;
		/** Cómo está eso hoy, en una frase. */
		summary?: string;
		/** Acciones del bloque, alineadas con el título. */
		actions?: Snippet;
		/** Bajo la tabla y fuera de sus filetes: la paginación. */
		footer?: Snippet;
		children: Snippet;
	}

	let { title, summary = '', actions, footer, children }: Props = $props();
</script>

<section class="block">
	<header class="head">
		<div class="text">
			<h2>{title}</h2>
			{#if summary}
				<p class="summary">{summary}</p>
			{/if}
		</div>
		{#if actions}
			<div class="actions">{@render actions()}</div>
		{/if}
	</header>

	<div class="body">
		{@render children()}
	</div>

	{#if footer}
		<div class="foot">{@render footer()}</div>
	{/if}
</section>

<style>
	.block {
		margin-bottom: 3rem;
	}

	.head {
		display: flex;
		align-items: flex-end;
		justify-content: space-between;
		flex-wrap: wrap;
		gap: 0.75rem 2rem;
		margin-bottom: 1.1rem;
	}

	.text {
		min-width: 0;
	}

	/*
	 * Con la tipografía de la portada, como el título de un grupo de
	 * configuración. Antes era mono de 0,7 rem en versalitas: un `<h2>` que
	 * parecía la etiqueta de un campo, y en una página con tres tablas no había
	 * forma de ver dónde empezaba cada una.
	 */
	h2 {
		margin: 0;
		font-family: var(--font-display);
		font-size: 1.45rem;
		font-weight: 300;
		letter-spacing: -0.02em;
		color: var(--text);
	}

	.summary {
		max-width: 68ch;
		margin: 0.45rem 0 0;
		font-size: 0.875rem;
		line-height: 1.55;
		color: var(--text-muted);
	}

	.actions {
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}

	/* El filete de arriba lo pone la cabecera de la tabla; este cierra la
	   última fila, que sin él se queda flotando sobre la página. */
	.body {
		border-top: 1px solid var(--border-strong);
		border-bottom: 1px solid var(--border-strong);
	}

	.foot {
		padding-top: 0.9rem;
	}

	/* --- El idioma de las celdas, para las cinco tablas -------------------- */

	/*
	 * Van con `:global` porque las celdas se escriben en la tabla que usa el
	 * bloque y llevan su scope, no el de aquí. Los nombres son genéricos
	 * (`.cell-key`, `.row-error`), así que quedan acotados al bloque para no
	 * chocar con las tablas del resto del panel.
	 */

	/* Lo que nombra la fila: un ticker, un par de divisas. Va en mono porque es
	   un identificador, y en el color del texto normal: en ámbar estaba cien
	   veces por pantalla, y un acento que sale en todas las filas deja de
	   acentuar nada. */
	.block :global(.cell-key) {
		font-family: var(--font-mono);
		font-size: 0.82rem;
		font-weight: 600;
		color: var(--text) !important;
		white-space: nowrap;
	}

	.block :global(.cell-name) {
		color: var(--text) !important;
		font-weight: 500;
	}

	.block :global(.cell-email) {
		font-family: var(--font-mono);
		font-size: 0.8rem;
	}

	/* La edad en palabras. `aged` es lo que lleva más tiempo del que debería:
	   el ámbar apagado se lee como un ámbar que perdió el calor. */
	.block :global(.cell-age) {
		white-space: nowrap;
	}

	.block :global(.cell-age.aged) {
		color: var(--stale);
	}

	.block :global(.cell-actions) {
		text-align: right;
		white-space: nowrap;
	}

	.block :global(.row-actions) {
		display: flex;
		align-items: center;
		justify-content: flex-end;
		gap: 1rem;
	}

	.block :global(.row-error) {
		margin: 0.3rem 0 0;
		font-size: 0.75rem;
		color: var(--red);
	}

	.block :global(.cell-actions .row-error) {
		text-align: right;
	}

	/* Una fila que ya no cuenta: baneada. Se apaga en vez de teñirse de rojo,
	   que es el color de lo que va a romperse ahora. */
	.block :global(tr.row-muted td) {
		color: var(--text-dim);
	}

	.block :global(tr.row-muted .cell-name),
	.block :global(tr.row-muted .cell-key) {
		color: var(--text-dim) !important;
	}

	/*
	 * El editor de una fila: un número que se puede escribir encima, no un campo
	 * con caja permanente. Con veinte filas, veinte recuadros dibujados hacían
	 * que la columna del precio pesara más que la del precio de verdad; aquí el
	 * campo se lee como el resto de la cifra y solo se enmarca cuando lo señalas
	 * o lo escribes.
	 */
	.block :global(.edit) {
		display: flex;
		align-items: center;
		justify-content: flex-end;
		gap: 0.75rem;
	}

	.block :global(.edit-input) {
		width: 7.5rem;
		padding: 0.35rem 0.5rem;
		border: 1px solid transparent;
		border-radius: 6px;
		background: transparent;
		color: var(--text-muted);
		font-family: var(--font-mono);
		font-size: 0.82rem;
		font-variant-numeric: tabular-nums;
		text-align: right;
		transition:
			border-color 0.15s ease,
			color 0.15s ease;
	}

	.block :global(tr:hover .edit-input) {
		border-color: var(--border-strong);
	}

	.block :global(.edit-input:focus) {
		border-color: var(--amber);
		background: rgba(255, 255, 255, 0.04);
		color: var(--text);
	}

	.block :global(.edit-input.invalid) {
		border-color: var(--red);
		color: var(--red);
	}

	/* El acuse y el error ocupan el mismo hueco, bajo el campo. */
	.block :global(.row-note) {
		margin: 0.3rem 0 0;
		font-size: 0.75rem;
		text-align: right;
		color: var(--green);
	}

	.block :global(.sr-only) {
		position: absolute;
		width: 1px;
		height: 1px;
		padding: 0;
		margin: -1px;
		overflow: hidden;
		clip-path: inset(50%);
		white-space: nowrap;
		border: 0;
	}

	@media (prefers-reduced-motion: reduce) {
		.block :global(.edit-input) {
			transition: none;
		}
	}

	@media (max-width: 640px) {
		h2 {
			font-size: 1.25rem;
		}
	}
</style>
