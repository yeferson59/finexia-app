<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/shared/css';

	type Size = 'sm' | 'md';

	interface Props {
		/** What is missing, in one short sentence. */
		title: string;
		/** Optional line explaining how the state fills in. */
		description?: string;
		/** Decorative mark above the title; rendered `aria-hidden`. */
		icon?: Snippet;
		/** Next step, e.g. a link to the screen that creates the missing data. */
		action?: Snippet;
		/** `sm` for inline slots inside a card, `md` for a whole panel. */
		size?: Size;
		/** Dashed outline, for the placeholder of a chart or table. */
		bordered?: boolean;
		class?: string;
	}

	let {
		title,
		description = '',
		icon,
		action,
		size = 'md',
		bordered = false,
		class: className = ''
	}: Props = $props();
</script>

<div class={cn('empty-state', `empty-${size}`, { 'empty-bordered': bordered }, className)}>
	{#if icon}
		<span class="empty-icon" aria-hidden="true">{@render icon()}</span>
	{/if}
	<p class="empty-title">{title}</p>
	{#if description}
		<p class="empty-description">{description}</p>
	{/if}
	{#if action}
		<div class="empty-action">{@render action()}</div>
	{/if}
</div>

<style>
	.empty-state {
		display: flex;
		flex: 1;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 0.6rem;
		text-align: center;
		color: var(--text-muted);
	}

	.empty-sm {
		padding: 2rem 1rem;
	}

	.empty-md {
		padding: 3rem 2rem;
	}

	.empty-bordered {
		border: 1px dashed var(--border);
		border-radius: 8px;
	}

	.empty-icon {
		display: grid;
		place-items: center;
		color: var(--text-dim);
		margin-bottom: 0.15rem;
	}

	.empty-title {
		margin: 0;
		font-size: 0.9rem;
		font-weight: 500;
		color: var(--text);
	}

	.empty-sm .empty-title {
		font-size: 0.85rem;
	}

	.empty-description {
		margin: 0;
		max-width: 42ch;
		font-size: 0.8rem;
		font-weight: 300;
		line-height: 1.6;
		color: var(--text-dim);
	}

	.empty-action {
		margin-top: 0.5rem;
	}
</style>
