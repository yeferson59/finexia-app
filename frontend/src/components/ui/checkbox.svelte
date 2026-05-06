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
		border: 1.5px solid rgba(212, 175, 55, 0.3);
		border-radius: 6px;
		background: rgba(26, 31, 46, 0.5);
		cursor: pointer;
		transition: all 0.25s ease;
		position: relative;
		flex-shrink: 0;
	}

	.checkbox-input:hover:not(:disabled) {
		border-color: rgba(212, 175, 55, 0.5);
	}

	.checkbox-input:checked {
		background: #d4af37;
		border-color: #d4af37;
	}

	.checkbox-input:checked::after {
		content: '✓';
		position: absolute;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
		color: #0f1419;
		font-size: 0.875rem;
		font-weight: 700;
	}

	.checkbox-input:focus {
		outline: none;
		box-shadow: 0 0 0 3px rgba(212, 175, 55, 0.1);
	}

	.checkbox-input:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.checkbox-label {
		font-size: 0.8rem;
		color: #e0e0e0;
		cursor: pointer;
		transition: color 0.25s ease;
		letter-spacing: 0.2px;
		font-weight: 500;
	}

	.checkbox-input:hover:not(:disabled) ~ .checkbox-label {
		color: #d4af37;
	}
</style>
