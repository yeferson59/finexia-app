<script lang="ts">
	/*
	 * Un campo de las pantallas de acceso: etiqueta, control, ayuda y error.
	 *
	 * No usa `ui/input` porque ese es el del panel oscuro —su borde ámbar
	 * translúcido y su texto claro desaparecen sobre el papel—. Aquí el error se
	 * enlaza al control con `aria-describedby`, y no se marca lo obligatorio:
	 * en estas pantallas lo es todo.
	 */
	import type { Snippet } from 'svelte';
	import type { HTMLInputAttributes } from 'svelte/elements';

	interface Props {
		label: string;
		id: string;
		name: string;
		type?: HTMLInputAttributes['type'];
		value?: string;
		placeholder?: string;
		autocomplete?: HTMLInputAttributes['autocomplete'];
		inputmode?: HTMLInputAttributes['inputmode'];
		required?: boolean;
		error?: string;
		hint?: string;
		/** Una acción del campo, a la derecha de la etiqueta. */
		action?: Snippet;
		/** Algo dentro del control, a la derecha: el ojo de la contraseña. */
		trailing?: Snippet;
	}

	let {
		label,
		id,
		name,
		type = 'text',
		value = $bindable(''),
		placeholder,
		autocomplete,
		inputmode,
		required = true,
		error,
		hint,
		action,
		trailing
	}: Props = $props();

	const describedBy = $derived(
		[error ? `${id}-error` : '', hint ? `${id}-hint` : ''].filter(Boolean).join(' ') || undefined
	);
</script>

<div class="lp-field">
	<div class="lp-label-row">
		<label class="lp-label" for={id}>{label}</label>
		{@render action?.()}
	</div>

	<div class="control" class:with-trailing={!!trailing}>
		<input
			class="lp-input"
			{id}
			{name}
			{type}
			{placeholder}
			{autocomplete}
			{inputmode}
			{required}
			bind:value
			aria-invalid={error ? 'true' : undefined}
			aria-describedby={describedBy}
		/>
		{@render trailing?.()}
	</div>

	{#if hint}
		<p class="lp-hint" id="{id}-hint">{hint}</p>
	{/if}
	{#if error}
		<p class="lp-field-error" id="{id}-error">{error}</p>
	{/if}
</div>

<style>
	.control {
		position: relative;
	}

	.control.with-trailing :global(.lp-input) {
		padding-right: 52px;
	}
</style>
