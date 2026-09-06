<script lang="ts">
	import { cn } from '$lib/shared/css';
	import type { HTMLInputAttributes } from 'svelte/elements';

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
		autocomplete?: HTMLInputAttributes['autocomplete'];
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
</script>

<div class="field input-wrapper">
	{#if label}
		<label for={id || name} class="field-label">
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
		{onfocus}
		class={cn('field-control', { 'input-error': !!error }, className)}
		{...rest}
	/>
	{#if error}
		<span class="input-error-text">{error}</span>
	{/if}
</div>

<style>
	/*
	 * El contenedor, la etiqueta y el aspecto del campo son los de cualquier
	 * formulario del panel y viven en `routes/layout.css`: este componente los
	 * pedía con sus propios nombres y su propia copia del CSS, medio píxel más
	 * grande, así que un `<Input>` y un `<select>` suelto del mismo formulario no
	 * medían igual.
	 *
	 * `input-wrapper` se queda como gancho: dos secciones de configuración lo
	 * usan para meter el campo en su propia fila.
	 */
	.field-control.input-error {
		border-color: var(--red);
	}

	.input-error-text {
		font-size: 0.8rem;
		color: var(--red);
		letter-spacing: 0.2px;
	}
</style>
