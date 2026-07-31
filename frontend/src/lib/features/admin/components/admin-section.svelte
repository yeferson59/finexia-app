<script lang="ts">
	/**
	 * Bloque «título + tabla» de las pantallas de administración.
	 *
	 * Interno de la feature. Aporta el chrome compartido por los tres listados de
	 * la pantalla de usuarios y, con `:global` acotado a este bloque, las clases
	 * de celda que usan sus tablas. Son nombres genéricos (`.cell-email`,
	 * `.row-error`…): sueltos en una hoja global chocarían con otras tablas del
	 * dashboard.
	 */
	import type { Snippet } from 'svelte';
	import Card from '$lib/ui/card.svelte';

	interface Props {
		title: string;
		children: Snippet;
	}

	let { title, children }: Props = $props();
</script>

<section class="section">
	<h2 class="section-title">{title}</h2>
	<Card padding="none">
		{@render children()}
	</Card>
</section>

<style>
	.section {
		margin-bottom: 1.75rem;
	}

	.section-title {
		font-family: var(--font-mono);
		font-size: 0.7rem;
		font-weight: 500;
		text-transform: uppercase;
		letter-spacing: 0.12em;
		color: var(--text-dim);
		margin: 0 0 0.75rem 0;
	}

	.section :global(.cell-name) {
		color: var(--text) !important;
		font-weight: 500;
		white-space: nowrap;
	}

	.section :global(.cell-email) {
		font-family: var(--font-mono);
		font-size: 0.8rem;
	}

	.section :global(.cell-date) {
		white-space: nowrap;
		font-family: var(--font-mono);
		font-size: 0.8rem;
	}

	.section :global(.cell-actions) {
		text-align: right;
		white-space: nowrap;
	}

	.section :global(.action-row) {
		display: flex;
		align-items: center;
		justify-content: flex-end;
		gap: 0.25rem;
	}

	.section :global(.delete-label) {
		color: var(--red);
	}

	.section :global(.row-error) {
		font-size: 0.75rem;
		color: var(--red);
		margin: 0.25rem 0 0;
		text-align: right;
	}

	.section :global(.empty-state) {
		text-align: center;
		padding: 3rem;
		color: var(--text-dim);
		font-size: 0.9rem;
	}
</style>
