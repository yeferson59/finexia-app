<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils';

	type Padding = 'none' | 'sm' | 'md' | 'lg';
	type Variant = 'default' | 'elevated';

	interface Props {
		/** Surface style: `default` (dashboard) or `elevated` (deep shadow panel). */
		variant?: Variant;
		/** Inner padding scale. */
		padding?: Padding;
		/** Lift + highlight border on hover (use for clickable cards). */
		hover?: boolean;
		/** Renders a <button>; for navigation use goto(resolve(...)) inside. */
		onclick?: (event: MouseEvent) => void;
		/** Accessible label, required when `onclick` is set. */
		ariaLabel?: string;
		class?: string;
		children?: Snippet;
		[key: string]: unknown;
	}

	let {
		variant = 'default',
		padding = 'md',
		hover = false,
		onclick = undefined,
		ariaLabel = undefined,
		class: className = '',
		children,
		...rest
	}: Props = $props();

	const paddingClasses: Record<Padding, string> = {
		none: '',
		sm: 'card-p-sm',
		md: 'card-p-md',
		lg: 'card-p-lg'
	};

	const classes = $derived(
		cn(
			'card',
			`card-${variant}`,
			paddingClasses[padding],
			{ 'card-interactive': !!onclick || hover },
			className
		)
	);
</script>

{#if onclick}
	<button type="button" {onclick} aria-label={ariaLabel} class={classes} {...rest}>
		{@render children?.()}
	</button>
{:else}
	<div class={classes} {...rest}>
		{@render children?.()}
	</div>
{/if}

<style>
	.card {
		display: block;
		position: relative;
		width: 100%;
		text-align: inherit;
		color: var(--text);
		background: var(--surface);
		border: 1px solid var(--border-strong);
	}

	.card-default {
		border-radius: 14px;
		backdrop-filter: blur(10px);
	}

	.card-elevated {
		border-radius: 16px;
		backdrop-filter: blur(16px);
		box-shadow:
			0 20px 60px rgba(0, 0, 0, 0.3),
			inset 0 1px 0 rgba(255, 255, 255, 0.05);
	}

	/* Reset native button styling when used interactively */
	button.card {
		font: inherit;
		cursor: pointer;
	}

	.card-p-sm {
		padding: 1.35rem;
	}

	.card-p-md {
		padding: 2rem;
	}

	.card-p-lg {
		padding: 2.5rem;
	}

	.card-interactive {
		transition:
			background 0.3s ease,
			border-color 0.3s ease,
			transform 0.3s ease,
			box-shadow 0.3s ease;
	}

	.card-interactive:hover {
		background: var(--surface-2);
		border-color: rgba(212, 145, 42, 0.3);
		transform: translateY(-4px);
		box-shadow: 0 30px 80px rgba(0, 0, 0, 0.4);
	}

	@media (prefers-reduced-motion: reduce) {
		.card-interactive:hover {
			transform: none;
		}
	}

	@media (max-width: 768px) {
		.card-p-md {
			padding: 1.5rem;
		}

		.card-p-lg {
			padding: 1.75rem;
		}
	}
</style>
