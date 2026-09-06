<script lang="ts">
	/*
	 * La ficha del documento y el lector incrustado.
	 *
	 * El PDF pesa varios megas, así que no se carga al abrir la página: hasta que
	 * no se pide «Ver la guía aquí» solo está la ficha. Quien solo venía a
	 * descargarlo no paga esa descarga dos veces.
	 */
	import { asset } from '$app/paths';
	import { formatBytes, formatGeneratedAt } from '../guide';
	import { manual } from '../manual-meta';

	let open = $state(false);

	/*
	 * `asset()` es la contraparte de `resolve()` para lo que vive en `static/`:
	 * su tipo es la lista de archivos reales, así que renombrar el PDF rompe la
	 * compilación en vez de dejar un enlace roto en producción.
	 */
	const pdf = asset('/manual-usuario.pdf');
</script>

<section class="document">
	<div class="identity">
		<h2>Manual de Usuario de Finexia</h2>
		<!--
			En prosa y no en insignias: son cuatro datos que solo se miran una vez,
			para saber si el documento que vas a abrir es el de ahora. Repartidos en
			cuatro píldoras pedían más atención de la que valen.
		-->
		<p class="colophon">
			Versión {manual.version}, de {manual.date.toLocaleLowerCase('es')}. El PDF ocupa {formatBytes(
				manual.bytes
			)} y se generó el {formatGeneratedAt(manual.generatedAt)}.
		</p>
	</div>

	<div class="actions">
		<button
			type="button"
			class="read"
			aria-expanded={open}
			aria-controls="manual-reader"
			onclick={() => (open = !open)}
		>
			{open ? 'Ocultar la guía' : 'Ver la guía aquí'}
		</button>

		<span class="links">
			<a href={pdf} download="finexia-manual-de-usuario.pdf">Descargar PDF</a>
			<a href={pdf} target="_blank" rel="noopener">Abrir en pestaña nueva</a>
		</span>
	</div>
</section>

{#if open}
	<div class="reader" id="manual-reader">
		<iframe src={pdf} title="Manual de Usuario de Finexia"></iframe>
		<p class="fallback">
			¿No se ve el documento? Algunos navegadores móviles no muestran PDF incrustados:
			<a href={pdf} download="finexia-manual-de-usuario.pdf">descarga la guía</a>
			y ábrela con tu lector habitual.
		</p>
	</div>
{/if}

<style>
	/*
	 * Sin tarjeta y sin el icono de documento que llevaba delante: la página
	 * entera va de un documento, así que dibujarlo otra vez en un cuadrado ámbar
	 * no informaba de nada. Lo que queda es el filete, que es lo que separa los
	 * bloques en el resto del panel.
	 */
	.document {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 1.5rem 2.5rem;
		flex-wrap: wrap;
		padding-bottom: 1.75rem;
		border-bottom: 1px solid var(--border-strong);
	}

	.identity {
		min-width: 0;
	}

	/* Al tamaño del título de una sección, no al de la portada: el título de la
	   página ya está encima y dos serifas grandes seguidas compiten. */
	h2 {
		margin: 0;
		font-family: var(--font-body);
		font-size: 1rem;
		font-weight: 500;
		color: var(--text);
	}

	.colophon {
		max-width: 52ch;
		margin: 0.45rem 0 0;
		font-size: 0.83rem;
		line-height: 1.55;
		color: var(--text-muted);
	}

	.actions {
		display: flex;
		flex-direction: column;
		align-items: flex-start;
		gap: 0.85rem;
		flex-shrink: 0;
	}

	/*
	 * El único ámbar lleno de la página. Abrir el manual es a lo que se viene, y
	 * descargarlo o abrirlo aparte son variantes de lo mismo: no necesitan tres
	 * botones del mismo peso discutiéndose la atención, como tenían antes.
	 */
	.read {
		padding: 0.6rem 1.15rem;
		border: 1px solid var(--amber);
		border-radius: 7px;
		background: var(--amber);
		color: #0d0800;
		font-family: var(--font-body);
		font-size: 0.85rem;
		font-weight: 600;
		cursor: pointer;
		white-space: nowrap;
		transition: background 0.2s ease;
	}

	.read:hover {
		background: var(--amber-light);
		border-color: var(--amber-light);
	}

	.links {
		display: flex;
		flex-wrap: wrap;
		gap: 0.35rem 1.25rem;
	}

	.links a {
		font-size: 0.82rem;
		color: var(--text-muted);
		text-decoration: none;
		border-bottom: 1px solid var(--border-strong);
		padding-bottom: 1px;
		transition: color 0.2s ease;
	}

	.links a:hover {
		color: var(--text);
		border-color: var(--amber);
	}

	.reader {
		padding-top: 1.75rem;
		border-bottom: 1px solid var(--border-strong);
		padding-bottom: 1.75rem;
	}

	iframe {
		display: block;
		width: 100%;
		height: min(78vh, 900px);
		border: 1px solid var(--border);
		border-radius: 10px;
		background: #16181c;
	}

	.fallback {
		max-width: 62ch;
		margin: 0.85rem 0 0;
		font-size: 0.8rem;
		line-height: 1.6;
		color: var(--text-dim);
	}

	.fallback a {
		color: var(--amber);
	}

	@media (prefers-reduced-motion: reduce) {
		.read,
		.links a {
			transition: none;
		}
	}

	@media (max-width: 640px) {
		.actions {
			width: 100%;
		}

		.read {
			width: 100%;
		}
	}
</style>
