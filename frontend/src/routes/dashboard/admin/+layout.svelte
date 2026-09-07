<script lang="ts">
	import type { LayoutProps } from './$types';

	let { children }: LayoutProps = $props();
</script>

<div class="admin">
	{@render children()}
</div>

<style>
	/*
	 * Lo que comparten las cuatro pantallas de administración y solo ellas.
	 *
	 * `display: contents` para no meter una caja más en la columna del panel: ni
	 * el color heredado ni las reglas `:global` de aquí dependen de que este div
	 * genere su propia caja.
	 */
	.admin {
		display: contents;

		/*
		 * El único color propio del área.
		 *
		 * La paleta del producto no tenía forma de decir «esto envejeció»: el
		 * ámbar ya significa otra cosa aquí —lo que escribe el feed público— y los
		 * tres grises del texto son jerarquía, no estado. `--stale` es ese mismo
		 * ámbar sin calor, que es exactamente lo que quiere decir: algo que
		 * estuvo al día. Conserva la suficiente carga de color para no confundirse
		 * con `--text-muted`, que es gris y está a su lado en la misma fila, y da
		 * 6,3:1 contra el fondo.
		 *
		 * Vive aquí y no en `routes/layout.css` porque solo tiene sentido frente a
		 * un dato mantenido a mano; en el panel de un usuario no hay nada que se
		 * ponga rancio.
		 */
		--stale: #b08b4a;
	}

	/*
	 * Una acción que no es la de la pantalla: banear una cuenta, revocar una
	 * invitación, guardar un precio, importar una hoja. Son texto, no botones con
	 * borde, y valen igual en una fila de tabla que junto al título de la página.
	 *
	 * Es el mismo gesto —y el mismo aspecto— que las acciones de fila de
	 * configuración. Con dos botones enmarcados por fila, veinte filas eran
	 * cuarenta cajas compitiendo con los datos.
	 */
	.admin :global(.row-action) {
		padding: 0;
		border: none;
		background: none;
		font-family: var(--font-body);
		font-size: 0.85rem;
		color: var(--text-muted);
		text-decoration: none;
		cursor: pointer;
		transition: color 0.15s ease;
	}

	/* Dentro de una tabla se lee un punto más pequeño: van dos por fila y
	   compiten con la cifra de al lado. */
	.admin :global(td .row-action) {
		font-size: 0.8rem;
	}

	.admin :global(.row-action:hover:not(:disabled)) {
		color: var(--text);
		text-decoration: underline;
		text-underline-offset: 3px;
	}

	/* Lo irreversible solo se tiñe cuando ya lo estás apuntando. */
	.admin :global(.row-action.danger:hover:not(:disabled)) {
		color: var(--red);
	}

	.admin :global(.row-action:disabled) {
		color: var(--text-dim);
		cursor: default;
	}

	/* Los nombres de columna de un archivo de importación son literales: van en
	   mono, como cualquier identificador de estas pantallas. */
	.admin :global(.hint code) {
		font-family: var(--font-mono);
		font-size: 0.78rem;
		color: var(--text);
	}

	@media (prefers-reduced-motion: reduce) {
		.admin :global(.row-action) {
			transition: none;
		}
	}
</style>
