<script lang="ts">
	/*
	 * Cuánto se movió la cuenta para llegar a la cifra de la cabecera: los dos
	 * meses extremos y las tres medidas de riesgo.
	 *
	 * Eran doce mosaicos idénticos repartidos en tres bloques con antetítulo en
	 * versalitas, y lo que cada uno medía vivía en un `title`: en un móvil no se
	 * abre, con el teclado tampoco, y sin eso «Ratio de Sharpe −0,01» es un
	 * número sin idioma. Ahora la explicación es una columna de la tabla, que es
	 * la mitad del dato y no una ayuda opcional.
	 *
	 * Las seis métricas que faltan —rentabilidad, anualizada, periodo, capital,
	 * valor y ganancia— están en la cabecera de la página: no se repiten aquí.
	 */
	import { UNAVAILABLE, type KeyStat } from '../reports';

	let { stats }: { stats: KeyStat[] } = $props();
</script>

<section class="movement" aria-labelledby="movement">
	<h2 id="movement">Cómo se movió</h2>

	<table>
		<thead>
			<tr>
				<th scope="col">Medida</th>
				<th scope="col" class="num">Valor</th>
				<th scope="col">Qué mide</th>
			</tr>
		</thead>
		<tbody>
			{#each stats as stat (stat.label)}
				<tr>
					<th scope="row">{stat.label}</th>
					<td class="num value {stat.tone ?? 'neutral'}" class:missing={stat.value === UNAVAILABLE}>
						{stat.value}
						{#if stat.detail}
							<span class="detail">{stat.detail}</span>
						{/if}
					</td>
					<td class="meaning">
						{stat.hint}
						{#if stat.note}
							<!-- El reparo va con la explicación y no bajo la cifra: alineado a la
							     derecha ocupaba tres renglones en bandera y no había por dónde
							     empezar a leerlo. -->
							<span class="note">{stat.note}</span>
						{/if}
					</td>
				</tr>
			{/each}
		</tbody>
	</table>
</section>

<style>
	.movement {
		padding: 2rem 0;
		border-bottom: 1px solid var(--border);
	}

	h2 {
		margin: 0 0 1.1rem;
		font-family: var(--font-body);
		font-size: 1.05rem;
		font-weight: 500;
		color: var(--text);
	}

	table {
		width: 100%;
		border-collapse: collapse;
	}

	thead th {
		padding: 0 0.75rem 0.7rem;
		border-bottom: 1px solid var(--border);
		font-size: 0.75rem;
		font-weight: 400;
		color: var(--text-dim);
		text-align: left;
	}

	thead th.num {
		text-align: right;
	}

	thead th:first-child {
		padding-left: 0;
	}

	tbody th,
	tbody td {
		padding: 0.9rem 0.75rem;
		border-bottom: 1px solid var(--border);
		font-size: 0.85rem;
		font-weight: 400;
		color: var(--text);
		text-align: left;
		vertical-align: top;
	}

	tbody th:first-child {
		width: 12rem;
		padding-left: 0;
	}

	tbody tr:last-child th,
	tbody tr:last-child td {
		border-bottom: none;
	}

	.num {
		text-align: right;
	}

	.value {
		width: 11rem;
		font-family: var(--font-mono);
		font-size: 0.95rem;
		font-weight: 600;
		font-variant-numeric: tabular-nums;
		white-space: nowrap;
	}

	.value.up {
		color: var(--green);
	}

	.value.down {
		color: var(--red);
	}

	/* Un `N/A` no es una cifra: se aparta del blanco para que no se lea como una
	   medida más, y la columna de al lado dice qué historial le falta. */
	.value.missing {
		font-weight: 400;
		color: var(--text-dim);
	}

	/* El mes al que pertenece el máximo, bajo su cifra: era la mitad derecha de
	   un «+1,7% · Oct 2025» que obligaba a leer dos datos en una línea. */
	.detail {
		display: block;
		margin-top: 0.25rem;
		font-family: var(--font-body);
		font-size: 0.75rem;
		font-weight: 400;
		line-height: 1.4;
		color: var(--text-muted);
		white-space: normal;
	}

	.note {
		display: block;
		margin-top: 0.3rem;
		font-size: 0.78rem;
		color: var(--text-dim);
	}

	.meaning {
		max-width: 62ch;
		line-height: 1.5;
		color: var(--text-muted);
	}

	/*
	 * Debajo de esto la explicación no cabe al lado de la cifra. La fila se
	 * pliega: medida y valor arriba, lo que mide debajo cruzando las dos
	 * columnas.
	 */
	@media (max-width: 860px) {
		thead {
			display: none;
		}

		tbody tr {
			display: grid;
			grid-template-columns: minmax(0, 1fr) auto;
			column-gap: 1rem;
			padding: 1rem 0;
			border-bottom: 1px solid var(--border);
		}

		tbody tr:last-child {
			border-bottom: none;
		}

		tbody th,
		tbody td {
			width: auto;
			padding: 0;
			border: none;
		}

		.value {
			width: auto;
		}

		.meaning {
			grid-column: 1 / -1;
			margin-top: 0.55rem;
			font-size: 0.8rem;
		}
	}
</style>
