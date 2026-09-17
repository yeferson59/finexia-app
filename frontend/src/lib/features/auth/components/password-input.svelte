<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { HTMLInputAttributes } from 'svelte/elements';
	import AuthField from './auth-field.svelte';

	interface Props {
		label: string;
		id: string;
		name: string;
		placeholder?: string;
		value?: string;
		error?: string;
		hint?: string;
		autocomplete?: HTMLInputAttributes['autocomplete'];
		/** Una acción del campo, a la derecha de la etiqueta. */
		action?: Snippet;
	}

	let {
		label,
		id,
		name,
		placeholder,
		value = $bindable(''),
		error,
		hint,
		autocomplete,
		action
	}: Props = $props();

	let showPassword = $state(false);
</script>

<AuthField
	{label}
	{id}
	{name}
	type={showPassword ? 'text' : 'password'}
	{placeholder}
	{autocomplete}
	{error}
	{hint}
	{action}
	bind:value
>
	{#snippet trailing()}
		<button
			type="button"
			class="toggle"
			onclick={() => (showPassword = !showPassword)}
			aria-label={showPassword ? 'Ocultar contraseña' : 'Mostrar contraseña'}
			aria-controls={id}
		>
			<svg
				width="20"
				height="20"
				viewBox="0 0 24 24"
				fill="none"
				stroke="currentColor"
				stroke-width="1.8"
				stroke-linecap="round"
				stroke-linejoin="round"
				aria-hidden="true"
			>
				<path d="M2 12s3.6-7 10-7 10 7 10 7-3.6 7-10 7-10-7-10-7z" />
				<circle cx="12" cy="12" r="3" />
				{#if !showPassword}
					<path d="m3 3 18 18" />
				{/if}
			</svg>
		</button>
	{/snippet}
</AuthField>

<style>
	.toggle {
		position: absolute;
		top: 50%;
		right: 3px;
		display: grid;
		place-items: center;
		width: 44px;
		height: 44px;
		padding: 0;
		border: none;
		border-radius: 6px;
		background: transparent;
		color: var(--lp-ink-2);
		cursor: pointer;
		transform: translateY(-50%);
	}

	.toggle:hover {
		color: var(--lp-ink);
	}
</style>
