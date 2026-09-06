<script lang="ts">
	/*
	 * Maqueta de `/dashboard/reports`, la ficha de resultados: la cifra de
	 * cabecera, la matriz de rentabilidad mes a mes, las medidas de movimiento y
	 * los archivos.
	 *
	 * Antes enseñaba la página que había hasta el rediseño —doce mosaicos de un
	 * año, una tarjeta de estadísticas y tres botones de descarga— y prometía
	 * una pantalla que ya no existe.
	 */
	import {
		TOUR_KEY_STATS,
		TOUR_MONTHS,
		TOUR_PROJECTION,
		TOUR_RECORD,
		TOUR_REPORTS,
		TOUR_RETURNS,
		tourReturnBackground
	} from '../product-tour';

	/* Coma decimal, como el resto de cifras de la landing y del dashboard. */
	const fmt = (value: number | null) =>
		value === null ? '–' : `${value > 0 ? '+' : ''}${value.toFixed(1).replace('.', ',')}%`;
</script>

<div class="reports">
	<div class="mk-card">
		<div class="record">
			<div class="record-label">{TOUR_RECORD.label}</div>
			<div class="record-value">{TOUR_RECORD.value}</div>
			<div class="record-span">{TOUR_RECORD.span}</div>
			<div class="record-money mk-up">{TOUR_RECORD.money}</div>
		</div>

		<div class="mk-title matrix-title">Rentabilidad mes a mes</div>

		<div class="matrix">
			<div class="mrow mrow-head">
				<span class="year">Año</span>
				{#each TOUR_MONTHS as month (month)}
					<span>{month}</span>
				{/each}
				<span class="total">Total</span>
			</div>

			{#each TOUR_RETURNS as row (row.year)}
				<div class="mrow">
					<span class="year">{row.year}</span>
					{#each row.values as value, index (`${row.year}-${TOUR_MONTHS[index]}`)}
						<span class="cell" style:background-color={tourReturnBackground(value)}>
							{fmt(value)}
						</span>
					{/each}
					<span class="total" class:mk-up={row.total >= 0} class:mk-dn={row.total < 0}>
						{fmt(row.total)}
					</span>
				</div>
			{/each}
		</div>
	</div>

	<div class="bottom">
		<div class="mk-card">
			<div class="mk-title">Cómo se movió</div>
			<div class="stats">
				{#each TOUR_KEY_STATS as stat (stat.label)}
					<div class="mk-row stat-row">
						<span class="stat-label">
							{stat.label}
							{#if stat.detail}
								<em>{stat.detail}</em>
							{/if}
						</span>
						<b
							class="mk-num"
							class:mk-up={stat.value.startsWith('+')}
							class:mk-dn={stat.value.startsWith('−')}
						>
							{stat.value}
						</b>
					</div>
				{/each}
			</div>
		</div>

		<div class="right">
			<div class="mk-card">
				<div class="proj-top">
					<span class="mk-title">Si el ritmo se mantiene</span>
					<span class="mk-pill">{TOUR_PROJECTION.rate}</span>
				</div>
				<div class="proj">
					{#each TOUR_PROJECTION.columns as column (column.period)}
						<div class="proj-col">
							<span class="proj-period">{column.period}</span>
							<span class="proj-value">{column.value}</span>
							<span class="proj-accrued" class:mk-up={column.accrued.startsWith('+')}>
								{column.accrued}
							</span>
						</div>
					{/each}
				</div>
			</div>

			<div class="mk-card">
				<div class="mk-title">Llévate los datos</div>
				<div class="downloads">
					{#each TOUR_REPORTS as report (report.title)}
						<div class="mk-row dl">
							<span class="dl-what">
								<b>{report.title}</b>
								<em>{report.description}</em>
							</span>
							<span class="dl-btn">Descargar <i>{report.format}</i></span>
						</div>
					{/each}
				</div>
			</div>
		</div>
	</div>
</div>

<style>
	.reports {
		display: flex;
		flex-direction: column;
		gap: 14px;
	}

	/* La cifra de cabecera, en el mismo registro que la de la página real: una
	   etiqueta que la nombra, el número, y el periodo en prosa. */
	.record-label {
		font-size: 10.5px;
		color: var(--text-muted);
	}

	.record-value {
		margin-top: 3px;
		font-family: var(--font-mono);
		font-size: 26px;
		font-weight: 600;
		letter-spacing: -0.03em;
		line-height: 1;
	}

	.record-value {
		color: var(--text);
	}

	/* La cifra grande en blanco y la ganancia en verde, como en la página: el
	   color lo lleva el dinero, no el porcentaje que lo explica. */
	.record-span,
	.record-money {
		margin-top: 7px;
		font-size: 10px;
		line-height: 1.45;
		color: var(--text-muted);
	}

	.record-money {
		margin-top: 3px;
	}

	.record {
		padding-bottom: 13px;
		margin-bottom: 13px;
		border-bottom: 1px solid var(--border);
	}

	.matrix-title {
		margin-top: 0;
		font-size: 13px;
	}

	.matrix {
		margin-top: 10px;
	}

	/* Año + doce meses + total. La del total es fija para que las dos filas
	   cierren a la misma altura, como en la ficha real. */
	.mrow {
		display: grid;
		grid-template-columns: 30px repeat(12, minmax(0, 1fr)) 40px;
		align-items: center;
		gap: 3px;
		padding: 3px 0;
		font-family: var(--font-mono);
		font-size: 9px;
		font-variant-numeric: tabular-nums;
		text-align: center;
	}

	.mrow-head {
		padding-bottom: 5px;
		border-bottom: 1px solid var(--border);
		font-family: var(--font-body);
		font-size: 8.5px;
		color: var(--text-dim);
	}

	.year {
		text-align: left;
		color: var(--text);
	}

	.mrow-head .year {
		color: var(--text-dim);
	}

	/* El tinte va en la ficha y no en la celda: tres meses verdes seguidos se
	   fundían en una sola banda. */
	.cell {
		padding: 4px 0;
		border-radius: 3px;
		color: var(--text-muted);
	}

	.total {
		padding-left: 5px;
		border-left: 1px solid var(--border);
		text-align: right;
		font-weight: 600;
	}

	.mrow-head .total {
		border-left: none;
		font-weight: 400;
	}

	.bottom {
		display: grid;
		grid-template-columns: minmax(0, 1fr) minmax(0, 1.15fr);
		gap: 10px;
	}

	/* La proyección y las descargas comparten columna: cinco medidas a la
	   izquierda pesan lo mismo que estos dos bloques juntos. */
	.right {
		display: grid;
		grid-template-rows: auto minmax(0, 1fr);
		gap: 10px;
		min-width: 0;
	}

	.proj-top {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 10px;
	}

	.proj-top .mk-title {
		margin-top: 0;
		font-size: 13px;
	}

	/* La misma tabla girada de la ficha real: un año por columna, el valor
	   proyectado arriba y lo que se acumula desde hoy debajo. */
	.proj {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 8px;
		margin-top: 11px;
		padding-top: 9px;
		border-top: 1px solid var(--border);
	}

	.proj-col {
		display: flex;
		flex-direction: column;
		gap: 3px;
		text-align: right;
	}

	.proj-period {
		font-family: var(--font-mono);
		font-size: 8.5px;
		color: var(--text-dim);
	}

	.proj-value {
		font-family: var(--font-mono);
		font-size: 11.5px;
		font-variant-numeric: tabular-nums;
		color: var(--text);
	}

	.proj-accrued {
		font-family: var(--font-mono);
		font-size: 9.5px;
		color: var(--text-muted);
	}

	.stats,
	.downloads {
		margin-top: 10px;
	}

	/* Cinco medidas a la izquierda, y a la derecha una proyección y tres
	   archivos: las alturas se ajustan con el aire de cada fila para que las dos
	   columnas cierren juntas. */
	.stat-row,
	.dl {
		grid-template-columns: minmax(0, 1fr) auto;
		padding: 13px 0;
		font-size: 11px;
	}

	.dl {
		padding: 8px 0;
	}

	.stat-row:first-child,
	.dl:first-child {
		border-top: none;
	}

	.stat-label,
	.dl-what {
		display: flex;
		flex-direction: column;
		gap: 2px;
		min-width: 0;
	}

	/* Lo que mide cada cifra no cabe aquí; el mes en el que cayó, sí, y es lo
	   que la ficha real pone bajo el valor. */
	.stat-label em,
	.dl-what em {
		font-style: normal;
		font-size: 9.5px;
		color: var(--text-dim);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.stat-row b {
		font-size: 11.5px;
		font-weight: 600;
	}

	.dl-what b {
		font-size: 11px;
		font-weight: 500;
		color: var(--text);
	}

	.dl-btn {
		flex-shrink: 0;
		display: inline-flex;
		align-items: baseline;
		gap: 5px;
		padding: 4px 9px;
		border: 1px solid var(--border-strong);
		border-radius: 6px;
		font-size: 10px;
		color: var(--text);
		white-space: nowrap;
	}

	.dl-btn i {
		font-style: normal;
		font-family: var(--font-mono);
		font-size: 8px;
		letter-spacing: 0.06em;
		color: var(--amber);
	}

	@media (max-width: 700px) {
		/*
		 * Doce meses no caben: se enseña el medio año más reciente, que es la
		 * parte de la matriz que se mira primero. Los meses son los hijos 2 a 13
		 * de cada fila.
		 */
		.mrow {
			grid-template-columns: 30px repeat(6, minmax(0, 1fr)) 40px;
		}

		.mrow > :nth-child(n + 2):nth-child(-n + 7) {
			display: none;
		}

		.bottom {
			grid-template-columns: minmax(0, 1fr);
		}
	}
</style>
