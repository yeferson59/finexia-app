<script lang="ts">
	/*
	 * El índice del manual, repartido en bloques.
	 *
	 * Es lo que más pesa de la página, y a propósito: quien entra aquí no viene
	 * a mirar la ficha del PDF, viene a averiguar en qué capítulo está lo suyo.
	 *
	 * Los capítulos no son enlaces —el manual es un PDF, no páginas de la
	 * aplicación—, así que el número no decora: es con lo que se busca el
	 * capítulo una vez abierto el documento. Por eso va delante y en la
	 * tipografía de máquina, la misma que llevan las cifras del resto del panel.
	 */
	import { groupSections } from '../guide';
	import { manual } from '../manual-meta';

	const groups = groupSections(manual.sections);
</script>

<nav class="contents" aria-label="Índice del manual">
	{#each groups as group (group.label)}
		<section class="group">
			<h2>{group.label}</h2>

			<ol>
				{#each group.sections as section (section.number)}
					<li>
						<span class="num">{String(section.number).padStart(2, '0')}</span>
						<span class="title">{section.title}</span>
					</li>
				{/each}
			</ol>
		</section>
	{/each}
</nav>

<style>
	/*
	 * El mismo carril que la página de configuración: a la izquierda de qué va
	 * el bloque y a la derecha lo que contiene. Diecinueve capítulos en una
	 * columna se leen como una pared; en cuatro bloques rotulados, el ojo baja
	 * por los rótulos y solo se mete en el que le toca.
	 */
	.group {
		display: grid;
		/*
		 * Carril más estrecho que el de configuración: allí el carril lleva un
		 * título y un párrafo, y aquí dos palabras. A diecisiete rem el rótulo se
		 * quedaba a medio palmo de su propia lista y dejaban de leerse como una
		 * sola cosa.
		 */
		grid-template-columns: minmax(0, 12rem) minmax(0, 1fr);
		gap: 0.75rem 2.5rem;
		padding: 1.75rem 0;
		border-bottom: 1px solid var(--border);
	}

	.group:last-child {
		border-bottom: none;
	}

	/* En la tipografía de portada, como los bloques de configuración: son los
	   asideros de una página larga. */
	h2 {
		margin: 0;
		font-family: var(--font-display);
		font-size: 1.15rem;
		font-weight: 400;
		letter-spacing: -0.01em;
		color: var(--text);
	}

	ol {
		margin: 0;
		padding: 0;
		list-style: none;
		min-width: 0;
	}

	/*
	 * Sin filete entre capítulos: dentro de un bloque son un índice, no una
	 * tabla, y las rayas los volvían filas de datos. La separación la da el
	 * interlineado, como en el índice de cualquier libro.
	 */
	li {
		display: grid;
		grid-template-columns: 2.25rem minmax(0, 1fr);
		align-items: baseline;
		padding: 0.3rem 0;
		font-size: 0.9rem;
		line-height: 1.5;
		color: var(--text-muted);
	}

	/*
	 * El número en gris y no en ámbar: en esta aplicación el ámbar dice que algo
	 * está encendido, y un número de capítulo no lo está. Así el único ámbar de
	 * la página es el botón que abre el manual.
	 */
	.num {
		font-family: var(--font-mono);
		font-size: 0.72rem;
		font-weight: 500;
		letter-spacing: 0.04em;
		color: var(--text-dim);
	}

	.title {
		min-width: 0;
		overflow-wrap: anywhere;
	}

	@media (max-width: 900px) {
		.group {
			grid-template-columns: minmax(0, 1fr);
			gap: 0.85rem;
		}
	}
</style>
