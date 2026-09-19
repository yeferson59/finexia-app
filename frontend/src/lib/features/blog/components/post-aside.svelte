<script lang="ts">
	/*
	 * El margen del artículo: la ficha y, en pantallas anchas, el índice.
	 *
	 * La ficha es lo que se consulta de un vistazo —cuándo, quién, cuánto se
	 * tarda— y va apilada, una cosa por línea, en vez de encadenada en una frase.
	 *
	 * El índice sigue la lectura: marca la sección en la que está el lector con
	 * el ámbar de la marca y el peso. Solo aparece con sitio para quedarse fijo
	 * al lado del texto; en el móvil los artículos se leen de arriba abajo y el
	 * índice solo empujaría el principio hacia abajo.
	 */
	import { resolve } from '$app/paths';
	import { formatPostDate, formatReadingTime, tagSlug, type PostHeading, type Post } from '../blog';

	interface Props {
		post: Post;
	}

	let { post }: Props = $props();

	let current = $state('');

	/*
	 * La sección actual es la última cuyo encabezado ya pasó bajo la cabecera
	 * fija. Se recalcula al hacer scroll en vez de con un IntersectionObserver
	 * porque las secciones son largas: entre dos encabezados no hay ninguno a
	 * la vista y el observador no tendría nada que decir.
	 */
	$effect(() => {
		const targets = post.headings
			.map((heading: PostHeading) => document.getElementById(heading.id))
			.filter((el): el is HTMLElement => el !== null);
		if (targets.length === 0) return;

		let frame = 0;
		const update = () => {
			frame = 0;
			const line = window.innerHeight * 0.3;
			let active = '';
			for (const el of targets) {
				if (el.getBoundingClientRect().top <= line) active = el.id;
			}
			current = active;
		};
		const schedule = () => {
			if (!frame) frame = requestAnimationFrame(update);
		};

		update();
		window.addEventListener('scroll', schedule, { passive: true });
		window.addEventListener('resize', schedule);
		return () => {
			window.removeEventListener('scroll', schedule);
			window.removeEventListener('resize', schedule);
			if (frame) cancelAnimationFrame(frame);
		};
	});
</script>

<aside class="post-aside" aria-label="Sobre este artículo">
	<dl class="facts">
		<div>
			<dt class="lp-sr-only">Publicado</dt>
			<dd><time datetime={post.date}>{formatPostDate(post.date)}</time></dd>
		</div>
		<div>
			<dt class="lp-sr-only">Autor</dt>
			<dd>{post.author}</dd>
		</div>
		<div>
			<dt class="lp-sr-only">Lectura</dt>
			<dd>{formatReadingTime(post.readingMinutes)}</dd>
		</div>
		{#if post.draft}
			<div>
				<dt class="lp-sr-only">Estado</dt>
				<dd><span class="draft">Borrador</span></dd>
			</div>
		{/if}
	</dl>

	<ul class="tags" aria-label="Etiquetas">
		{#each post.tags as tag (tag)}
			<li>
				<a href={resolve('/blog/tag/[tag]', { tag: tagSlug(tag) })}>{tag}</a>
			</li>
		{/each}
	</ul>

	{#if post.headings.length > 1}
		<nav class="toc" aria-labelledby="toc-title">
			<h2 id="toc-title">En este artículo</h2>
			<ol>
				{#each post.headings as heading (heading.id)}
					<li>
						<a href="#{heading.id}" aria-current={current === heading.id ? 'location' : undefined}
							>{heading.text}</a
						>
					</li>
				{/each}
			</ol>
		</nav>
	{/if}
</aside>

<style>
	.post-aside {
		position: sticky;
		top: calc(var(--blog-header) + 32px);
		align-self: start;
		font-size: var(--lp-fs-sm);
		line-height: 1.45;
		color: var(--lp-ink-2);
	}

	.facts {
		display: flex;
		flex-direction: column;
		gap: 2px;
		margin: 0;
	}

	.facts dd {
		margin: 0;
	}

	.facts time {
		font-weight: 600;
		color: var(--lp-ink);
	}

	.draft {
		display: inline-block;
		margin-top: 6px;
		padding: 1px 9px;
		border: 1px dashed currentColor;
		border-radius: 999px;
		font-size: 13px;
		font-weight: 600;
	}

	.tags {
		display: flex;
		flex-wrap: wrap;
		gap: 4px 14px;
		margin: 14px 0 0;
		padding: 0;
		list-style: none;
	}

	.tags a {
		color: var(--lp-ink-2);
		text-decoration: underline;
		text-decoration-color: var(--lp-rule);
		text-underline-offset: 3px;
	}

	.tags a:hover {
		color: var(--lp-ink);
		text-decoration-color: currentColor;
	}

	.toc {
		margin-top: 36px;
		padding-top: 20px;
		border-top: 1px solid var(--lp-rule);
	}

	.toc h2 {
		margin: 0 0 12px;
		font-size: var(--lp-fs-sm);
		font-weight: 600;
		color: var(--lp-ink);
	}

	.toc ol {
		margin: 0;
		padding: 0;
		list-style: none;
	}

	/* El filete de la izquierda es la regla por la que baja el marcador. */
	.toc a {
		display: block;
		padding: 6px 0 6px 14px;
		border-left: 2px solid var(--lp-rule);
		color: var(--lp-ink-2);
		text-wrap: pretty;
		transition:
			border-color 0.2s ease,
			color 0.2s ease;
	}

	.toc a:hover {
		color: var(--lp-ink);
	}

	.toc a[aria-current='location'] {
		border-left-color: var(--lp-jubilacion);
		font-weight: 600;
		color: var(--lp-ink);
	}

	@media (max-width: 1023px) {
		.post-aside {
			position: static;
		}
		.facts {
			flex-direction: row;
			flex-wrap: wrap;
			column-gap: 14px;
		}
		.draft {
			margin-top: 0;
		}
		.toc {
			display: none;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.toc a {
			transition: none;
		}
	}
</style>
