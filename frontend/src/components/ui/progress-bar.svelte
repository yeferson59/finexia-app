<script lang="ts">
	import { cn } from '$lib/utils';

	type Tone = 'amber' | 'success' | 'danger';

	interface Props {
		/** Fill percentage (0–100). */
		value: number;
		tone?: Tone;
		/** Caption shown below the bar. */
		label?: string;
		/** Accessible name for the progress bar. */
		ariaLabel?: string;
		class?: string;
	}

	let {
		value,
		tone = 'amber',
		label = '',
		ariaLabel = 'Progreso',
		class: className = ''
	}: Props = $props();

	const clamped = $derived(Math.max(0, Math.min(100, value)));
</script>

<div class={cn('progress', className)}>
	<div
		class="progress-track"
		role="progressbar"
		aria-label={ariaLabel}
		aria-valuenow={clamped}
		aria-valuemin={0}
		aria-valuemax={100}
	>
		<div class={cn('progress-fill', `progress-${tone}`)} style={`width: ${clamped}%`}></div>
	</div>
	{#if label}
		<p class="progress-label">{label}</p>
	{/if}
</div>

<style>
	.progress {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
		width: 100%;
	}

	.progress-track {
		height: 6px;
		border-radius: 999px;
		background: rgba(236, 234, 229, 0.12);
		overflow: hidden;
	}

	.progress-fill {
		height: 100%;
		border-radius: inherit;
		transition: width 0.4s ease;
	}

	.progress-amber {
		background: var(--amber);
	}

	.progress-success {
		background: var(--green);
	}

	.progress-danger {
		background: var(--red);
	}

	.progress-label {
		margin: 0;
		font-size: 0.75rem;
		color: var(--text-muted);
	}
</style>
