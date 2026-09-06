<script lang="ts">
	/*
	 * Los archivos que puedes llevarte, y qué trae cada uno.
	 *
	 * Eran tres tarjetas con una píldora «XLSX», un título y un botón del ancho
	 * de la tarjeta: mucha superficie para tres enlaces, y ninguna pista de qué
	 * había dentro de cada archivo hasta abrirlo. Aquí cada uno se explica en
	 * una línea y el formato viaja en el propio enlace, que es donde importa
	 * saber qué va a aterrizar en la carpeta de descargas.
	 */
	import { resolve } from '$app/paths';
	import { REPORT_DOWNLOADS } from '../reports';
</script>

<section class="downloads" aria-labelledby="downloads">
	<h2 id="downloads">Llévate los datos</h2>

	<ul>
		{#each REPORT_DOWNLOADS as report (report.type)}
			<li>
				<div class="what">
					<h3>{report.title}</h3>
					<p>{report.description}</p>
				</div>
				<a
					class="download"
					href={resolve(`/dashboard/reports/download?type=${report.type}`)}
					aria-label="Descargar {report.title} en {report.format}"
				>
					Descargar <span class="format">{report.format}</span>
				</a>
			</li>
		{/each}
	</ul>
</section>

<style>
	.downloads {
		padding: 2rem 0 0;
	}

	h2 {
		margin: 0 0 1.1rem;
		font-family: var(--font-body);
		font-size: 1.05rem;
		font-weight: 500;
		color: var(--text);
	}

	ul {
		max-width: 52rem;
		margin: 0;
		padding: 0;
		list-style: none;
	}

	li {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		justify-content: space-between;
		gap: 0.75rem 1.5rem;
		padding: 0.9rem 0;
		border-bottom: 1px solid var(--border);
	}

	li:last-child {
		border-bottom: none;
	}

	.what {
		min-width: 0;
	}

	h3 {
		margin: 0;
		font-size: 0.92rem;
		font-weight: 500;
		color: var(--text);
	}

	.what p {
		max-width: 62ch;
		margin: 0.2rem 0 0;
		font-size: 0.8rem;
		line-height: 1.45;
		color: var(--text-muted);
	}

	/*
	 * En el tono de la marca pero sin el bloque de antes: es un enlace a un
	 * archivo, no la acción principal de la página.
	 */
	.download {
		flex-shrink: 0;
		display: inline-flex;
		align-items: baseline;
		gap: 0.45rem;
		padding: 0.5rem 0.9rem;
		border: 1px solid var(--border-strong);
		border-radius: 9px;
		color: var(--text);
		font-size: 0.85rem;
		font-weight: 500;
		text-decoration: none;
		white-space: nowrap;
		transition:
			border-color 0.2s ease,
			background 0.2s ease;
	}

	.download:hover {
		border-color: rgba(212, 145, 42, 0.5);
		background: var(--panel);
	}

	.format {
		font-family: var(--font-mono);
		font-size: 0.68rem;
		letter-spacing: 0.04em;
		color: var(--amber);
	}

	@media (prefers-reduced-motion: reduce) {
		.download {
			transition: none;
		}
	}
</style>
