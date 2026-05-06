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

	const handleBlur = (e: Event) => {
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
		color: #e0e0e0;
		letter-spacing: 0.3px;
	}

	.input-field {
		width: 100%;
		padding: 0.875rem 1rem;
		padding-right: 3rem;
		border-radius: 8px;
		border: 1px solid rgba(212, 175, 55, 0.2);
		background: rgba(26, 31, 46, 0.5);
		color: #e0e0e0;
		font-size: 0.95rem;
		font-family: 'Lato', system-ui, sans-serif;
		transition: all 0.25s ease;
		outline: none;
		box-sizing: border-box;
	}

	.input-field::placeholder {
		color: rgba(224, 224, 224, 0.4);
	}

	.input-field:hover:not(:disabled) {
		border-color: rgba(212, 175, 55, 0.35);
		background: rgba(26, 31, 46, 0.7);
	}

	.input-field:focus {
		border-color: #d4af37;
		background: rgba(26, 31, 46, 0.9);
		box-shadow: 0 0 0 3px rgba(212, 175, 55, 0.1);
	}

	.input-field.input-error {
		border-color: #e74c3c;
	}

	.input-field.input-error:focus {
		box-shadow: 0 0 0 3px rgba(231, 76, 60, 0.1);
	}

	.input-field:disabled {
		background: rgba(26, 31, 46, 0.3);
		color: rgba(224, 224, 224, 0.5);
		cursor: not-allowed;
	}

	.input-error-text {
		font-size: 0.8rem;
		color: #e74c3c;
		letter-spacing: 0.2px;
	}
</style>
