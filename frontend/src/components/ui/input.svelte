<script lang="ts">
	import { cn } from '$lib/utils';

	interface Props {
		type?: string;
		placeholder?: string;
		value?: string;
		disabled?: boolean;
		error?: string;
		label?: string;
		id?: string;
		name?: string;
		required?: boolean;
		onchange?: (event: Event) => void;
		onfocus?: (event: Event) => void;
		class?: string;
	}

	let {
		type = 'text',
		placeholder = '',
		value = $bindable(''),
		disabled = false,
		error = '',
		label = '',
		id = '',
		name = '',
		required = false,
		onchange = undefined,
		onfocus = undefined,
		class: className = '',
		...rest
	}: Props = $props();

	let isFocused = $state(false);

	const handleBlur = () => {
		isFocused = false;
	};

	const handleFocus = (e: Event) => {
		isFocused = true;
		onfocus?.(e);
	};
</script>

<div class="input-wrapper">
	{#if label}
		<label for={id || name} class="input-label">
			{label}
			{#if required}
				<span class="text-red-400">*</span>
			{/if}
		</label>
	{/if}
	<input
		{type}
		{placeholder}
		bind:value
		{disabled}
		{id}
		{name}
		{required}
		{onchange}
		onfocus={handleFocus}
		onblur={handleBlur}
		class={cn('input-field', { 'input-focused': isFocused, 'input-error': !!error }, className)}
		{...rest}
	/>
	{#if error}
		<span class="input-error-text">{error}</span>
	{/if}
</div>

<style>
	.input-wrapper {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.input-label {
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--text);
		letter-spacing: 0.3px;
	}

	.input-field {
		width: 100%;
		padding: 0.875rem 1rem;
		padding-right: 3rem;
		border-radius: 8px;
		border: 1px solid rgba(212, 145, 42, 0.2);
		background: rgba(255, 255, 255, 0.03);
		color: var(--text);
		font-size: 0.95rem;
		font-family: var(--font-body);
		transition: all 0.25s ease;
		outline: none;
		box-sizing: border-box;
	}

	.input-field::placeholder {
		color: rgba(236, 234, 229, 0.4);
	}

	.input-field:hover:not(:disabled) {
		border-color: rgba(212, 145, 42, 0.35);
		background: rgba(255, 255, 255, 0.03);
	}

	.input-field:focus {
		border-color: var(--amber);
		background: rgba(255, 255, 255, 0.03);
		box-shadow: 0 0 0 3px var(--border);
	}

	.input-field.input-error {
		border-color: var(--red);
	}

	.input-field.input-error:focus {
		box-shadow: 0 0 0 3px rgba(224, 90, 90, 0.1);
	}

	.input-field:disabled {
		background: rgba(255, 255, 255, 0.03);
		color: rgba(236, 234, 229, 0.5);
		cursor: not-allowed;
	}

	.input-error-text {
		font-size: 0.8rem;
		color: var(--red);
		letter-spacing: 0.2px;
	}
</style>
