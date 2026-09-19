<script lang="ts">
	/*
	 * Las etiquetas del blog, con la que se está viendo marcada.
	 *
	 * «Todos» va primero y es un enlace como los demás: desde una etiqueta, la
	 * salida tiene que estar en el mismo sitio donde se entró.
	 *
	 * Van como texto y no como píldoras: son un filtro de un archivo, no botones.
	 * La activa lleva el subrayado grueso en tinta y el peso, que no dependen del
	 * color.
	 */
	import { resolve } from '$app/paths';
	import type { TagSummary } from '../blog';

	interface Props {
		tags: TagSummary[];
		/** El slug de la etiqueta activa; vacío en el índice completo. */
		active?: string;
	}

	let { tags, active = '' }: Props = $props();
</script>

{#if tags.length > 0}
	<nav class="tag-nav" aria-label="Etiquetas del blog">
		<a href={resolve('/blog')} class="tag" aria-current={active === '' ? 'page' : undefined}>
			Todos
		</a>
		{#each tags as tag (tag.slug)}
			<a
				href={resolve('/blog/tag/[tag]', { tag: tag.slug })}
				class="tag"
				aria-current={active === tag.slug ? 'page' : undefined}
			>
				{tag.label}<span class="count" aria-hidden="true">{tag.count}</span>
			</a>
		{/each}
	</nav>
{/if}

<style>
	.tag-nav {
		display: flex;
		flex-wrap: wrap;
		gap: 6px 26px;
	}

	.tag {
		display: inline-flex;
		align-items: baseline;
		gap: 5px;
		padding-block: 6px;
		font-size: 15px;
		color: var(--lp-ink-2);
		text-decoration: underline;
		text-decoration-color: transparent;
		text-decoration-thickness: 2px;
		text-underline-offset: 6px;
		transition:
			color 0.15s ease,
			text-decoration-color 0.15s ease;
	}

	.tag:hover {
		color: var(--lp-ink);
		text-decoration-color: var(--lp-rule);
	}

	.tag[aria-current='page'] {
		font-weight: 600;
		color: var(--lp-ink);
		text-decoration-color: var(--lp-ink);
	}

	.count {
		font-size: 12px;
		font-weight: 400;
		font-variant-numeric: tabular-nums;
		color: var(--lp-ink-2);
	}

	@media (prefers-reduced-motion: reduce) {
		.tag {
			transition: none;
		}
	}
</style>
