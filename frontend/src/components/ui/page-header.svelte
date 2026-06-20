<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils';

	interface Props {
		/** Small mono uppercase label above the title. */
		eyebrow?: string;
		/** Main page heading. */
		title?: string;
		/** id for the heading, e.g. to target with aria-labelledby. */
		titleId?: string;
		/** Supporting line under the title. */
		subtitle?: string;
		/** Removes the bottom divider. */
		divider?: boolean;
		class?: string;
		/** Custom title content (overrides `title`); useful for emphasis markup. */
		children?: Snippet;
		/** Right-aligned actions, e.g. a primary button. */
		actions?: Snippet;
	}

	let {
		eyebrow = '',
		title = '',
		titleId = undefined,
		subtitle = '',
		divider = true,
		class: className = '',
		children,
		actions
	}: Props = $props();
</script>

<header class={cn('page-header', { 'page-header-divider': divider }, className)}>
	<div class="page-header-text">
		{#if eyebrow}
			<p class="page-eyebrow">{eyebrow}</p>
		{/if}
		{#if children}
			<h1 id={titleId} class="page-title">{@render children()}</h1>
		{:else if title}
			<h1 id={titleId} class="page-title">{title}</h1>
		{/if}
		{#if subtitle}
			<p class="page-subtitle">{subtitle}</p>
		{/if}
	</div>

	{#if actions}
		<div class="page-header-actions">
			{@render actions()}
		</div>
	{/if}
</header>

<style>
	.page-header {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 2rem;
		flex-wrap: wrap;
		margin-bottom: 2.5rem;
	}

	.page-header-divider {
		padding-bottom: 1.75rem;
		border-bottom: 1px solid var(--border);
	}

	.page-header-text {
		min-width: 0;
	}

	.page-eyebrow {
		font-family: var(--font-mono);
		font-size: 0.6875rem;
		font-weight: 500;
		letter-spacing: 0.2em;
		text-transform: uppercase;
		color: var(--amber);
		margin: 0 0 0.75rem 0;
	}

	.page-title {
		font-family: var(--font-display);
		font-size: clamp(2rem, 4vw, 2.75rem);
		font-weight: 300;
		line-height: 1.05;
		letter-spacing: -0.02em;
		color: var(--text);
		margin: 0 0 0.6rem 0;
	}

	/* Italic amber emphasis for <em> inside the title (matches dashboard greeting). */
	.page-title :global(em) {
		font-style: italic;
		font-weight: 500;
		color: var(--amber-light);
	}

	.page-subtitle {
		font-size: 0.95rem;
		font-weight: 300;
		color: var(--text-muted);
		margin: 0;
	}

	.page-header-actions {
		flex-shrink: 0;
		display: flex;
		gap: 0.75rem;
		align-items: center;
	}

	@media (max-width: 768px) {
		.page-title {
			font-size: 2rem;
		}
	}
</style>
