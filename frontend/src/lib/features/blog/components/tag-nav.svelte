<script lang="ts">
	/*
	 * Las etiquetas del blog, con la que se está viendo marcada.
	 *
	 * «Todos» va primero y es un enlace como los demás: desde una etiqueta, la
	 * salida tiene que estar en el mismo sitio donde se entró.
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
				{tag.label}
				<span class="count" aria-hidden="true">{tag.count}</span>
			</a>
		{/each}
	</nav>
{/if}

<style>
	.tag-nav {
		display: flex;
		flex-wrap: wrap;
		gap: 10px;
		margin-bottom: 8px;
	}

	.tag {
		display: inline-flex;
		align-items: center;
		gap: 7px;
		padding: 7px 14px;
		border: 1px solid var(--lp-rule);
		border-radius: 999px;
		font-size: var(--lp-fs-sm);
		color: var(--lp-ink-2);
		transition:
			border-color 0.15s ease,
			background 0.15s ease,
			color 0.15s ease;
	}

	.tag:hover {
		border-color: var(--lp-ink);
		color: var(--lp-ink);
	}

	/* La activa en tinta llena: es el único estado que no depende del color. */
	.tag[aria-current='page'] {
		border-color: var(--lp-ink);
		background: var(--lp-ink);
		color: var(--lp-paper);
	}

	.count {
		font-size: 12px;
		font-variant-numeric: tabular-nums;
		opacity: 0.7;
	}
</style>
