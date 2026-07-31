<script lang="ts">
	/** Canal de alertas dentro del dashboard: anunciado, todavía no disponible. */
	import NotificationSection from './notification-section.svelte';
</script>

<NotificationSection
	title="Alertas en la app"
	description="Notificaciones en tiempo real dentro del dashboard"
	class="coming-soon-section"
>
	{#snippet icon()}
		<svg
			width="18"
			height="18"
			viewBox="0 0 24 24"
			fill="none"
			stroke="currentColor"
			stroke-width="2"
		>
			<path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"></path>
			<path d="M13.73 21a2 2 0 0 1-3.46 0"></path>
		</svg>
	{/snippet}

	<div class="coming-soon-badge">
		<span class="badge-dot"></span>
		<span>Próximamente</span>
	</div>
	<p class="coming-soon-hint">
		Las alertas en tiempo real para cambios de precio y eventos de portafolio estarán disponibles
		pronto.
	</p>
</NotificationSection>

<style>
	.coming-soon-badge {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.4rem 0.875rem;
		border-radius: 20px;
		background: rgba(212, 145, 42, 0.07);
		border: 1px solid rgba(212, 145, 42, 0.2);
		font-size: 0.8rem;
		color: var(--amber);
		font-weight: 500;
		width: fit-content;
	}

	.badge-dot {
		width: 6px;
		height: 6px;
		border-radius: 50%;
		background: var(--amber);
		flex-shrink: 0;
	}

	.coming-soon-hint {
		margin: 0;
		font-size: 0.8rem;
		color: rgba(236, 234, 229, 0.45);
		line-height: 1.65;
	}
</style>
