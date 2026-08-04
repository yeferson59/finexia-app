<script lang="ts">
	/*
	 * Visor del manual: la ficha del documento y el PDF incrustado.
	 *
	 * El PDF pesa varios megas, así que no se carga al abrir la página: hasta que
	 * no se pide «Ver la guía aquí» solo hay una portada estática. Quien solo
	 * venía a descargarlo no paga esa descarga dos veces.
	 */
	import { asset } from '$app/paths';
	import Badge from '$lib/ui/badge.svelte';
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

<section class="viewer">
	<header class="doc">
		<span class="doc-icon" aria-hidden="true">
			<svg
				width="26"
				height="26"
				viewBox="0 0 24 24"
				fill="none"
				stroke="currentColor"
				stroke-width="1.6"
				stroke-linecap="round"
				stroke-linejoin="round"
			>
				<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
				<path d="M14 2v6h6" />
				<path d="M8 13h8M8 17h5" />
			</svg>
		</span>

		<div class="doc-text">
			<h2>Manual de Usuario de Finexia</h2>
			<!-- Sin separadores «·»: al envolverse dejaban un punto colgando al
			     final de la línea. La separación la da el `gap`. -->
			<p class="doc-meta">
				<Badge tone="amber">v{manual.version}</Badge>
				<span>{manual.date}</span>
				<span>PDF · {formatBytes(manual.bytes)}</span>
				<span>Generado el {formatGeneratedAt(manual.generatedAt)}</span>
			</p>
		</div>

		<div class="doc-actions">
			<button type="button" class="action primary" onclick={() => (open = !open)}>
				{open ? 'Ocultar la guía' : 'Ver la guía aquí'}
			</button>
			<a class="action" href={pdf} target="_blank" rel="noopener">Abrir en pestaña nueva</a>
			<a class="action" href={pdf} download="finexia-manual-de-usuario.pdf"> Descargar PDF </a>
		</div>
	</header>

	{#if open}
		<div class="frame">
			<iframe src={pdf} title="Manual de Usuario de Finexia"></iframe>
			<p class="fallback">
				¿No se ve el documento? Algunos navegadores móviles no muestran PDF incrustados:
				<a href={pdf} download="finexia-manual-de-usuario.pdf">descarga la guía</a>
				y ábrela con tu lector habitual.
			</p>
		</div>
	{/if}
</section>

<style>
	.viewer {
		background: var(--surface);
		border: 1px solid var(--border-strong);
		border-radius: 14px;
		backdrop-filter: blur(10px);
		overflow: hidden;
	}

	.doc {
		display: flex;
		align-items: flex-start;
		gap: 1.25rem;
		padding: 1.75rem;
	}

	.doc-icon {
		display: grid;
		place-items: center;
		width: 48px;
		height: 48px;
		flex-shrink: 0;
		border-radius: 12px;
		border: 1px solid rgba(212, 145, 42, 0.25);
		background: rgba(212, 145, 42, 0.09);
		color: var(--amber-light);
	}

	.doc-text {
		flex: 1;
		min-width: 0;
	}

	.doc-text h2 {
		margin: 0 0 0.5rem;
		font-family: var(--font-display);
		font-size: 1.25rem;
		font-weight: 500;
		letter-spacing: -0.01em;
		color: var(--text);
	}

	.doc-meta {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 0.4rem 1.1rem;
		margin: 0;
		font-size: 0.78rem;
		color: var(--text-dim);
	}

	.doc-actions {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
		flex-shrink: 0;
	}

	.action {
		display: inline-flex;
		align-items: center;
		padding: 0.6rem 1.1rem;
		border: 1px solid var(--border-strong);
		border-radius: 6px;
		background: transparent;
		color: var(--text);
		font-family: var(--font-body);
		font-size: 0.82rem;
		font-weight: 600;
		text-decoration: none;
		cursor: pointer;
		white-space: nowrap;
		transition:
			background 0.2s ease,
			border-color 0.2s ease,
			color 0.2s ease;
	}

	.action:hover {
		background: rgba(212, 145, 42, 0.06);
		border-color: rgba(212, 145, 42, 0.4);
		color: var(--amber-light);
	}

	.action.primary {
		background: var(--amber);
		border-color: var(--amber);
		color: #0d0800;
	}

	.action.primary:hover {
		background: var(--amber-light);
		border-color: var(--amber-light);
		color: #0d0800;
	}

	.frame {
		border-top: 1px solid var(--border);
		padding: 1.25rem 1.75rem 1.75rem;
	}

	iframe {
		width: 100%;
		height: min(78vh, 900px);
		border: 1px solid var(--border);
		border-radius: 10px;
		background: #16181c;
	}

	.fallback {
		margin: 0.85rem 0 0;
		font-size: 0.78rem;
		line-height: 1.6;
		color: var(--text-dim);
	}

	.fallback a {
		color: var(--amber-light);
	}

	@media (max-width: 900px) {
		.doc {
			flex-wrap: wrap;
			padding: 1.35rem;
		}

		.doc-actions {
			width: 100%;
		}

		.action {
			flex: 1;
			justify-content: center;
		}

		.frame {
			padding: 1rem 1.35rem 1.35rem;
		}
	}
</style>
