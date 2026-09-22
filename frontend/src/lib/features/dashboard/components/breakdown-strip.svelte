<script lang="ts">
	/*
	 * El patrimonio partido en tramos, justo debajo de la cifra.
	 *
	 * Sustituye a una barra por fila: aquellas medían cada fila contra el total
	 * en su propio carril, y había que comparar largos sueltos para ver cómo se
	 * repartía. Aquí las partes están una al lado de la otra y suman el ancho
	 * entero, que es lo que son. La tabla de debajo repite el tono de cada tramo
	 * en su fila, así que la franja no necesita leyenda.
	 *
	 * Es decorativa para el lector de pantalla: las mismas cifras están en la
	 * tabla, con su porcentaje escrito.
	 */
	import { formatPercent } from '$lib/shared/format/percent';
	import type { StripSegment } from '../breakdown';

	const { segments }: { segments: StripSegment[] } = $props();

	/* Debajo de esto el nombre no cabe sin partirse; el tramo se queda sin rótulo
	   y su fila de la tabla lo nombra. */
	const LABEL_MIN = 0.12;
</script>

<div class="strip" aria-hidden="true">
	<!-- Por posición y no por clave: al cambiar de corte los tramos se estiran
	     hasta su nuevo ancho en vez de desaparecer y volver a nacer. -->
	<div class="bars">
		{#each segments as segment, i (i)}
			<span class="bar tone-{segment.tone}" style="flex-grow: {segment.share}"></span>
		{/each}
	</div>
	<div class="labels">
		{#each segments as segment, i (i)}
			<span class="label" style="flex-grow: {segment.share}">
				{#if segment.share >= LABEL_MIN}
					<span class="name">{segment.label}</span>
					<span class="share">{formatPercent(segment.share * 100)}</span>
				{/if}
			</span>
		{/each}
	</div>
</div>

<style>
	.bars,
	.labels {
		display: flex;
		gap: 3px;
	}

	.bar,
	.label {
		flex-basis: 0;
		min-width: 0;
		transition: flex-grow 0.45s cubic-bezier(0.3, 0.7, 0.2, 1);
	}

	.bar {
		height: 14px;
		min-width: 3px;
		border-radius: 2px;
	}

	/* Los tonos los define el reparto, que los comparte con las muestras de la
	   tabla: tramo y fila tienen que ser el mismo color. */
	.tone-0 {
		background: var(--tone-0);
	}

	.tone-1 {
		background: var(--tone-1);
	}

	.tone-2 {
		background: var(--tone-2);
	}

	.tone-3 {
		background: var(--tone-3);
	}

	.tone-4 {
		background: var(--tone-4);
	}

	.labels {
		margin-top: 0.6rem;
	}

	.label {
		display: flex;
		flex-direction: column;
		gap: 0.1rem;
		padding-right: 0.5rem;
	}

	.name {
		overflow: hidden;
		font-size: 0.8rem;
		color: var(--text);
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.share {
		font-family: var(--font-figures);
		font-size: 0.78rem;
		font-variant-numeric: tabular-nums;
		color: var(--text-muted);
	}

	@media (prefers-reduced-motion: reduce) {
		.bar,
		.label {
			transition: none;
		}
	}

	/* En un teléfono los rótulos se apretaban hasta cortarse en dos letras: la
	   tabla de justo debajo ya dice cada nombre con su tono. */
	@media (max-width: 560px) {
		.labels {
			display: none;
		}
	}
</style>
