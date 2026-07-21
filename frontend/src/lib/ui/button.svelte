<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils';

	type Variant = 'primary' | 'secondary' | 'ghost' | 'tertiary';
	type Size = 'sm' | 'md' | 'lg';

	interface Props {
		variant?: Variant;
		size?: Size;
		type?: 'button' | 'submit' | 'reset';
		disabled?: boolean;
		loading?: boolean;
		fullWidth?: boolean;
		onclick?: (event: MouseEvent) => void;
		class?: string;
		children?: Snippet;
	}

	let {
		variant = 'primary',
		size = 'md',
		type = 'submit',
		disabled = false,
		loading = false,
		fullWidth = false,
		onclick = undefined,
		class: className = '',
		children,
		...rest
	}: Props = $props();

	const variantClasses: Record<Variant, string> = {
		primary: 'btn-primary',
		secondary: 'btn-secondary',
		ghost: 'btn-ghost',
		tertiary: 'btn-tertiary'
	};

	const sizeClasses: Record<Size, string> = {
		sm: 'btn-sm',
		md: 'btn-md',
		lg: 'btn-lg'
	};
</script>

<button
	{type}
	{onclick}
	disabled={disabled || loading}
	class={cn(
		variantClasses[variant],
		sizeClasses[size],
		{ 'btn-loading': loading, 'btn-full-width': fullWidth },
		className
	)}
	{...rest}
>
	{#if loading}
		<span class="btn-spinner"></span>
	{/if}
	{@render children?.()}
</button>

<style>
	button {
		font-family: var(--font-body);
		border: none;
		cursor: pointer;
		border-radius: 8px;
		transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
		letter-spacing: 0.5px;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
		font-weight: 600;
		position: relative;
	}

	/* Primary variant */
	.btn-primary {
		background: var(--amber);
		color: #0d0800;
		box-shadow: 0 4px 16px rgba(212, 145, 42, 0.2);
	}

	.btn-primary:hover:not(:disabled) {
		background: var(--amber-light);
		transform: translateY(-2px);
		box-shadow: 0 6px 24px rgba(212, 145, 42, 0.35);
	}

	.btn-primary:active:not(:disabled) {
		transform: translateY(0);
	}

	/* Secondary variant */
	.btn-secondary {
		background: transparent;
		border: 1.5px solid var(--amber);
		color: var(--amber);
	}

	.btn-secondary:hover:not(:disabled) {
		background: var(--border);
		border-color: var(--amber-light);
		color: var(--amber-light);
	}

	/* Ghost variant */
	.btn-ghost {
		background: transparent;
		border: 1.5px solid rgba(236, 234, 229, 0.2);
		color: var(--text);
	}

	.btn-ghost:hover:not(:disabled) {
		background: var(--border);
		border-color: var(--amber);
		color: var(--amber);
	}

	/* Tertiary variant */
	.btn-tertiary {
		background: transparent;
		color: var(--amber);
		font-size: 0.9rem;
	}

	.btn-tertiary:hover:not(:disabled) {
		color: var(--amber-light);
	}

	/* Sizes */
	.btn-sm {
		padding: 0.5rem 1rem;
		font-size: 0.8rem;
	}

	.btn-md {
		padding: 0.875rem 1.5rem;
		font-size: 0.95rem;
	}

	.btn-lg {
		padding: 1rem 2rem;
		font-size: 1rem;
	}

	/* States */
	button:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.btn-loading {
		color: transparent;
	}

	.btn-full-width {
		width: 100%;
	}

	.btn-spinner {
		position: absolute;
		width: 16px;
		height: 16px;
		border: 2px solid rgba(212, 145, 42, 0.3);
		border-top-color: var(--amber);
		border-radius: 50%;
		animation: spin 0.8s linear infinite;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}
</style>
