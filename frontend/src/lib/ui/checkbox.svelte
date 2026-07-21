<script lang="ts">
	import { cn } from '$lib/utils';

	interface Props {
		label?: string;
		id?: string;
		name?: string;
		checked?: boolean;
		disabled?: boolean;
		onchange?: (event: Event) => void;
		class?: string;
	}

	let {
		label = '',
		id = '',
		name = '',
		checked = $bindable(false),
		disabled = false,
		onchange = undefined,
		class: className = '',
		...rest
	}: Props = $props();
</script>

<div class="checkbox-wrapper">
	<input
		type="checkbox"
		{id}
		{name}
		bind:checked
		{disabled}
		{onchange}
		class={cn('checkbox-input', className)}
		{...rest}
	/>
	{#if label}
		<label for={id || name} class="checkbox-label">
			{label}
		</label>
	{/if}
</div>

<style>
	.checkbox-wrapper {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		cursor: pointer;
	}

	.checkbox-input {
		appearance: none;
		width: 20px;
		height: 20px;
		border: 1.5px solid rgba(212, 145, 42, 0.3);
		border-radius: 6px;
		background: rgba(255, 255, 255, 0.03);
		cursor: pointer;
		transition: all 0.25s ease;
		position: relative;
		flex-shrink: 0;
	}

	.checkbox-input:hover:not(:disabled) {
		border-color: rgba(212, 145, 42, 0.5);
	}

	.checkbox-input:checked {
		background: var(--amber);
		border-color: var(--amber);
	}

	.checkbox-input:checked::after {
		content: '✓';
		position: absolute;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
		color: #0d0800;
		font-size: 0.875rem;
		font-weight: 700;
	}

	.checkbox-input:focus {
		outline: none;
		box-shadow: 0 0 0 3px var(--border);
	}

	.checkbox-input:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.checkbox-label {
		font-size: 0.8rem;
		color: var(--text);
		cursor: pointer;
		transition: color 0.25s ease;
		letter-spacing: 0.2px;
		font-weight: 500;
	}

	.checkbox-input:hover:not(:disabled) ~ .checkbox-label {
		color: var(--amber);
	}
</style>
