<script lang="ts">
	/**
	 * Tarjeta de un canal de notificación: icono, título, descripción y contenido.
	 *
	 * Interna de la feature. Como en `settings`, las clases de nombre genérico
	 * que usa su contenido (`.feedback`, `.form-actions`) se declaran aquí con
	 * `:global` acotado a la tarjeta: en una hoja global chocarían con el resto
	 * del dashboard.
	 */
	import type { Snippet } from 'svelte';
	import Card from '$lib/ui/card.svelte';

	interface Props {
		title: string;
		description: string;
		/** Icono del canal, dibujado por quien usa la sección. */
		icon: Snippet;
		/** Clase extra del contenedor interno (p. ej. el canal aún no disponible). */
		class?: string;
		children: Snippet;
	}

	let { title, description, icon, class: className = '', children }: Props = $props();
</script>

<Card variant="elevated" padding="none">
	<div class="section {className}">
		<div class="section-header">
			<div class="section-icon">
				{@render icon()}
			</div>
			<div>
				<h2 class="section-title">{title}</h2>
				<p class="section-desc">{description}</p>
			</div>
		</div>
		{@render children()}
	</div>
</Card>

<style>
	.section {
		padding: 1.5rem;
	}

	.section-header {
		display: flex;
		align-items: flex-start;
		gap: 0.875rem;
		margin-bottom: 1.5rem;
		padding-bottom: 1.25rem;
		border-bottom: 1px solid rgba(212, 145, 42, 0.1);
	}

	.section-icon {
		width: 36px;
		height: 36px;
		border-radius: 8px;
		background: rgba(212, 145, 42, 0.1);
		border: 1px solid rgba(212, 145, 42, 0.2);
		color: var(--amber);
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	.section-title {
		margin: 0 0 0.2rem;
		font-size: 0.95rem;
		font-weight: 600;
		color: var(--text);
	}

	.section-desc {
		margin: 0;
		font-size: 0.8rem;
		color: rgba(236, 234, 229, 0.5);
	}

	/* Variante del canal que aún no está disponible: sin formulario, su
	   contenido se reparte en columna. */
	.section.coming-soon-section {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.section :global(.form-actions) {
		margin-top: 1.5rem;
		display: flex;
		justify-content: flex-end;
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
