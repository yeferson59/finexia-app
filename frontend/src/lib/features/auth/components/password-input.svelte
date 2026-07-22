<script lang="ts">
	import Input from '$lib/ui/input.svelte';
	import type { HTMLInputAttributes } from 'svelte/elements';

	interface Props {
		label: string;
		id?: string;
		name: string;
		placeholder?: string;
		value?: string;
		error?: string;
		autocomplete?: HTMLInputAttributes['autocomplete'];
		required?: boolean;
	}

	let {
		label,
		id,
		name,
		placeholder = '',
		value = $bindable(''),
		error,
		autocomplete,
		required = true
	}: Props = $props();

	let showPassword = $state(false);
</script>

<div class="password-wrapper">
	<Input
		{label}
		{id}
		{name}
		type={showPassword ? 'text' : 'password'}
		{placeholder}
		bind:value
		{error}
		{autocomplete}
		{required}
	/>
	<button
		type="button"
		class="password-toggle"
		onclick={() => (showPassword = !showPassword)}
		aria-label={showPassword ? 'Ocultar contraseña' : 'Mostrar contraseña'}
	>
		{#if showPassword}
			<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor">
				<path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path>
				<circle cx="12" cy="12" r="3"></circle>
			</svg>
		{:else}
			<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor">
				<path
					d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1 4.24 4.24"
				></path>
				<line x1="1" y1="1" x2="23" y2="23"></line>
			</svg>
		{/if}
	</button>
</div>

<style>
	.password-wrapper {
		position: relative;
		width: 100%;
	}

	.password-toggle {
		position: absolute;
		right: 0.75rem;
		top: 50%;
		margin-top: 1.125rem;
		transform: translateY(-50%);
		background: none;
		border: none;
		color: var(--text-secondary);
		cursor: pointer;
		transition: all 0.25s ease;
		padding: 0.5rem;
		display: flex;
		align-items: center;
		justify-content: center;
		width: 40px;
		height: 40px;
		z-index: 10;
	}

	.password-toggle:hover {
		color: var(--gold-primary);
		transform: translateY(-50%) scale(1.1);
	}

	.password-toggle svg {
		stroke-width: 2;
	}
</style>
