<script lang="ts">
	/**
	 * Tarjeta de una sección de ajustes (perfil, seguridad, 2FA, sesiones…).
	 *
	 * Aporta el chrome que todas comparten —`Card` + `.section` + título— y, con
	 * `:global` acotado a esta tarjeta, las clases de formulario que su contenido
	 * usa (`.form-fields`, `.form-actions`, `.hint`, `.feedback`). Son nombres
	 * genéricos: en una hoja global colisionarían con el resto del dashboard, y
	 * copiados en cada sección acabarían divergiendo.
	 */
	import type { Snippet } from 'svelte';
	import Card from '$lib/ui/card.svelte';

	interface Props {
		title: string;
		children: Snippet;
	}

	let { title, children }: Props = $props();
</script>

<Card variant="elevated" padding="none">
	<div class="section">
		<h2 class="section-title">{title}</h2>
		{@render children()}
	</div>
</Card>

<style>
	.section {
		padding: 1.5rem;
	}

	.section-title {
		margin: 0 0 1.5rem;
		font-size: 1rem;
		font-weight: 600;
		color: var(--text);
		letter-spacing: 0.3px;
	}

	.section :global(.form-fields) {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.section :global(.form-actions) {
		margin-top: 1.5rem;
		display: flex;
		justify-content: flex-end;
	}

	.section :global(.hint) {
		margin: 0;
		font-size: 0.8rem;
		color: rgba(236, 234, 229, 0.5);
		line-height: 1.65;
	}

	.section :global(.feedback) {
		margin: 0.875rem 0 0;
		font-size: 0.835rem;
		padding: 0.6rem 0.875rem;
		border-radius: 6px;
	}

	.section :global(.feedback.success) {
		background: rgba(74, 222, 128, 0.08);
		border: 1px solid rgba(74, 222, 128, 0.25);
		color: #4ade80;
	}

	.section :global(.feedback.error) {
		background: rgba(224, 90, 90, 0.08);
		border: 1px solid rgba(224, 90, 90, 0.25);
		color: var(--red, #e05a5a);
	}
</style>
