<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils';

	interface Props {
		/** Small mono uppercase label above the title. */
		eyebrow?: string;
		/** Heading text (omit and use `children` for custom content). */
		title?: string;
		/** Removes the bottom divider. */
		divider?: boolean;
		class?: string;
		/** Custom title-area content (overrides `eyebrow`/`title`). */
		children?: Snippet;
		/** Right-aligned content: a pill, link, or button. */
		action?: Snippet;
	}

	let {
		eyebrow = '',
		title = '',
		divider = true,
		class: className = '',
		children,
		action
	}: Props = $props();
</script>

<div class={cn('card-header', { 'card-header-divider': divider }, className)}>
	<div class="card-header-text">
		{#if children}
			{@render children()}
		{:else}
			{#if eyebrow}
				<p class="card-eyebrow">{eyebrow}</p>
			{/if}
			{#if title}
				<h2 class="card-title">{title}</h2>
			{/if}
		{/if}
	</div>

	{#if action}
		<div class="card-header-action">
			{@render action()}
		</div>
	{/if}
</div>

<style>
	.card-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		gap: 1rem;
		margin-bottom: 1.5rem;
	}

	.card-header-divider {
		padding-bottom: 1.25rem;
		border-bottom: 1px solid var(--border);
	}

	.card-header-text {
		min-width: 0;
	}

	.card-eyebrow {
		font-family: var(--font-mono);
		font-size: 0.625rem;
		font-weight: 500;
		letter-spacing: 0.18em;
		text-transform: uppercase;
		color: var(--text-dim);
		margin: 0 0 0.4rem 0;
	}

	.card-title {
		font-family: var(--font-display);
		font-size: 1.15rem;
		font-weight: 500;
		letter-spacing: -0.01em;
		color: var(--text);
		margin: 0;
	}

	.card-header-action {
		flex-shrink: 0;
		white-space: nowrap;
	}
</style>
