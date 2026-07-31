<script lang="ts">
	/**
	 * Panel del centro de reportes: el marco de cristal que comparten los cuatro
	 * bloques de la página, con cabecera opcional.
	 *
	 * Interno de la feature. `.empty-text` va con `:global` acotado al panel
	 * porque el texto de «sin datos» lo escribe cada bloque en su propio scope.
	 */
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/shared/css';

	interface Props {
		/** `div` para los paneles que solo alojan un mensaje. */
		tag?: 'article' | 'div';
		class?: string;
		/** Sin título no se pinta la cabecera. */
		title?: string;
		/** Píldora a la derecha del título (el año, en el calendario). */
		badge?: string;
		children: Snippet;
	}

	let {
		tag = 'article',
		class: className = '',
		title = '',
		badge = '',
		children
	}: Props = $props();
</script>

<svelte:element this={tag} class={cn('panel', className)}>
	{#if title}
		<div class="section-head">
			<h2>{title}</h2>
			{#if badge}
				<span>{badge}</span>
			{/if}
		</div>
	{/if}
	{@render children()}
</svelte:element>

<style>
	.panel {
		border: 1px solid var(--border-strong);
		border-radius: 16px;
		background: var(--surface);
		box-shadow:
			0 20px 60px rgba(0, 0, 0, 0.3),
			inset 0 1px 0 rgba(255, 255, 255, 0.05);
		backdrop-filter: blur(16px);
	}

	.section-head {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 0.9rem;
	}

	.section-head h2 {
		margin: 0;
		font-size: 1rem;
		color: var(--text);
	}

	.section-head span {
		font-size: 0.75rem;
		padding: 0.25rem 0.6rem;
		border-radius: 999px;
		background: rgba(212, 145, 42, 0.12);
		color: var(--amber-light);
	}

	.panel :global(.empty-text) {
		margin: 0;
		font-size: 0.85rem;
		color: rgba(236, 234, 229, 0.45);
	}
</style>
