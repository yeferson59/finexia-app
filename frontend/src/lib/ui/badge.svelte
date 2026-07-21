<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils';

	type Tone = 'neutral' | 'amber' | 'success' | 'warning' | 'danger' | 'info';
	type Size = 'sm' | 'md';

	interface Props {
		/** Color intent. */
		tone?: Tone;
		size?: Size;
		/** Fully rounded pill (true) vs. squared tag (false). */
		pill?: boolean;
		/** Mono uppercase styling (status-pill look). */
		uppercase?: boolean;
		class?: string;
		children?: Snippet;
	}

	let {
		tone = 'neutral',
		size = 'sm',
		pill = true,
		uppercase = true,
		class: className = '',
		children
	}: Props = $props();
</script>

<span
	class={cn(
		'badge',
		`badge-${tone}`,
		`badge-${size}`,
		{ 'badge-pill': pill, 'badge-uppercase': uppercase },
		className
	)}
>
	{@render children?.()}
</span>

<style>
	.badge {
		display: inline-flex;
		align-items: center;
		gap: 0.35rem;
		font-family: var(--font-mono);
		font-weight: 600;
		letter-spacing: 0.06em;
		white-space: nowrap;
		border: 1px solid transparent;
		border-radius: 4px;
		line-height: 1;
	}

	.badge-uppercase {
		text-transform: uppercase;
	}

	.badge-pill {
		border-radius: 999px;
	}

	/* Sizes */
	.badge-sm {
		font-size: 0.625rem;
		padding: 0.25rem 0.6rem;
	}

	.badge-md {
		font-size: 0.72rem;
		padding: 0.35rem 0.8rem;
	}

	/* Tones */
	.badge-neutral {
		color: var(--text-muted);
		background: var(--surface-2);
		border-color: var(--border);
	}

	.badge-amber {
		color: var(--amber-light);
		background: rgba(212, 145, 42, 0.15);
		border-color: rgba(212, 145, 42, 0.3);
	}

	.badge-success {
		color: var(--green);
		background: rgba(34, 201, 126, 0.15);
		border-color: rgba(34, 201, 126, 0.3);
	}

	.badge-warning {
		color: var(--amber-light);
		background: rgba(241, 196, 15, 0.15);
		border-color: rgba(241, 196, 15, 0.3);
	}

	.badge-danger {
		color: var(--red);
		background: rgba(224, 90, 90, 0.15);
		border-color: rgba(224, 90, 90, 0.3);
	}

	.badge-info {
		color: #6b8cef;
		background: rgba(107, 140, 239, 0.15);
		border-color: rgba(107, 140, 239, 0.3);
	}
</style>
