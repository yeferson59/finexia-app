<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils';

	type Tone = 'default' | 'positive' | 'negative' | 'highlight';
	type Align = 'left' | 'right';

	interface Props {
		/** Small label above the value. */
		label: string;
		/** The value. Use `children` instead for rich content. */
		value?: string | number;
		/** Inline suffix after the value, e.g. "anual". */
		unit?: string;
		tone?: Tone;
		align?: Align;
		class?: string;
		/** Overrides `value`/`unit` with custom markup. */
		children?: Snippet;
	}

	let {
		label,
		value = undefined,
		unit = '',
		tone = 'default',
		align = 'left',
		class: className = '',
		children
	}: Props = $props();
</script>

<div class={cn('stat', `stat-${align}`, className)}>
	<span class="stat-label">{label}</span>
	<p class={cn('stat-value', `stat-${tone}`)}>
		{#if children}
			{@render children()}
		{:else}
			{value}
		{/if}
		{#if unit}
			<span class="stat-unit">{unit}</span>
		{/if}
	</p>
</div>

<style>
	.stat {
		display: flex;
		flex-direction: column;
		gap: 0.4rem;
		min-width: 0;
	}

	.stat-right {
		text-align: right;
	}

	.stat-right .stat-value {
		justify-content: flex-end;
	}

	.stat-label {
		font-family: var(--font-mono);
		font-size: 0.625rem;
		color: var(--text-dim);
		text-transform: uppercase;
		letter-spacing: 0.12em;
		font-weight: 500;
	}

	.stat-value {
		display: flex;
		align-items: baseline;
		gap: 0.4rem;
		min-width: 0;
		font-family: var(--font-mono);
		font-size: 1.15rem;
		font-weight: 600;
		color: var(--text);
		margin: 0;
		font-variant-numeric: tabular-nums;
		overflow-wrap: anywhere;
	}

	.stat-positive {
		color: var(--green);
	}

	.stat-negative {
		color: var(--red);
	}

	.stat-highlight {
		color: var(--amber-light);
	}

	.stat-unit {
		font-size: 0.7rem;
		font-weight: 400;
		color: var(--text-dim);
	}
</style>
