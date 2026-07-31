<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import Card from '$lib/ui/card.svelte';
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatSourceType, type Platform } from '../platforms';
	import PlatformEditForm from './platform-edit-form.svelte';
	import PlatformDeleteDialog from './platform-delete-dialog.svelte';

	let { platform }: { platform: Platform } = $props();

	function formatDate(dateString: string): string {
		return new Date(dateString).toLocaleDateString('es-CO', {
			year: 'numeric',
			month: 'long',
			day: 'numeric'
		});
	}

	function formatCurrency(value: string): string {
		return privacy.money(
			'$' +
				new Intl.NumberFormat('es-CO', {
					minimumFractionDigits: 2,
					maximumFractionDigits: 2
				}).format(parseFloat(value) || 0)
		);
	}

	let isEditing = $state(false);
	let submitSuccess = $state(false);
	let showDeleteConfirm = $state(false);

	function goBack() {
		goto(resolve('/dashboard/platforms'));
	}
</script>

<div class="page-container">
	<header class="page-header">
		<div class="header-top">
			<button onclick={goBack} class="btn-back" aria-label="Volver">
				<svg
					width="20"
					height="20"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
				>
					<path d="M19 12H5M12 19l-7-7 7-7" />
				</svg>
			</button>
			<div class="header-content">
				<h1 class="page-title">{platform.name}</h1>
				<p class="page-subtitle">{formatSourceType(platform.sourceType)}</p>
			</div>
			<div class="header-actions">
				<div
					class="status-badge"
					style="--status-color: {platform.isActive ? 'var(--green)' : 'var(--red)'}"
				>
					{platform.isActive ? 'Activo' : 'Inactivo'}
				</div>
				{#if !isEditing}
					<button onclick={() => (isEditing = true)} class="btn-edit">
						<svg
							width="16"
							height="16"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							stroke-width="2"
						>
							<path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" />
							<path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
						</svg>
						Editar
					</button>
					<button onclick={() => (showDeleteConfirm = true)} class="btn-delete">
						<svg
							width="16"
							height="16"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							stroke-width="2"
						>
							<polyline points="3 6 5 6 21 6" />
							<path
								d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"
							/>
						</svg>
						Eliminar
					</button>
				{/if}
			</div>
		</div>
	</header>

	<div class="main-content">
		{#if !isEditing}
			<!-- View mode -->
			<Card variant="elevated" padding="none">
				<div class="panel-body">
					<h2 class="section-title">Información de la Plataforma</h2>

					<div class="info-group">
						<div class="info-item">
							<span class="info-label">Tipo de Plataforma</span>
							<span class="info-value">{formatSourceType(platform.sourceType)}</span>
						</div>
						<div class="info-item">
							<span class="info-label">Registrada el</span>
							<span class="info-value">{formatDate(platform.createdAt)}</span>
						</div>
					</div>

					{#if platform.description}
						<div class="info-description">
							<h3>Descripción</h3>
							<p>{platform.description}</p>
						</div>
					{/if}
				</div>
			</Card>

			<Card variant="elevated" padding="none">
				<div class="panel-body">
					<h2 class="section-title">Resumen de Inversiones</h2>
					<div class="stats-grid">
						<div class="stat-card">
							<div class="stat-icon">
								<svg width="24" height="24" viewBox="0 0 24 24" fill="currentColor">
									<path
										d="M19 3H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zM9 17H7v-7h2v7zm4 0h-2V7h2v10zm4 0h-2v-4h2v4z"
									></path>
								</svg>
							</div>
							<div class="stat-content">
								<span class="stat-label">Posiciones</span>
								<span class="stat-value">{platform.investments}</span>
							</div>
						</div>
						<div class="stat-card">
							<div class="stat-icon">
								<svg width="24" height="24" viewBox="0 0 24 24" fill="currentColor">
									<path
										d="M11.8 10.9c-2.27-.59-3-1.2-3-2.15 0-1.09 1.01-1.85 2.7-1.85 1.78 0 2.44.85 2.5 2.1h2.21c-.07-1.72-1.12-3.3-3.21-3.81V3h-3v2.16c-1.94.42-3.5 1.68-3.5 3.61 0 2.31 1.91 3.46 4.7 4.13 2.5.6 3 1.48 3 2.41 0 .69-.49 1.79-2.7 1.79-2.06 0-2.87-.92-2.98-2.1h-2.2c.12 2.19 1.76 3.42 3.68 3.83V21h3v-2.15c1.95-.37 3.5-1.5 3.5-3.55 0-2.84-2.43-3.81-4.7-4.4z"
									></path>
								</svg>
							</div>
							<div class="stat-content">
								<span class="stat-label">Total Invertido</span>
								<span class="stat-value">{formatCurrency(platform.totalValue)}</span>
							</div>
						</div>
					</div>
				</div>
			</Card>
		{:else}
			<PlatformEditForm
				{platform}
				onCancel={() => (isEditing = false)}
				onSaved={() => {
					submitSuccess = true;
					isEditing = false;
					setTimeout(() => (submitSuccess = false), 3000);
				}}
			/>
		{/if}

		{#if submitSuccess}
			<p class="success-msg">✓ Plataforma actualizada correctamente</p>
		{/if}
	</div>
</div>

<!-- Delete confirmation modal -->
{#if showDeleteConfirm}
	<PlatformDeleteDialog platformName={platform.name} onCancel={() => (showDeleteConfirm = false)} />
{/if}

<style>
	.page-container {
		display: flex;
		flex-direction: column;
		gap: 2rem;
	}

	.page-header {
		padding-bottom: 1.5rem;
		border-bottom: 1px solid var(--border);
	}

	.header-top {
		display: flex;
		align-items: center;
		gap: 1rem;
		flex-wrap: wrap;
	}

	.btn-back {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 40px;
		height: 40px;
		border: none;
		border-radius: 8px;
		background: transparent;
		color: var(--amber);
		cursor: pointer;
		transition: all 0.3s ease;
		flex-shrink: 0;
	}

	.btn-back:hover {
		background: var(--border);
		transform: translateX(-2px);
	}

	.header-content {
		flex: 1;
		min-width: 180px;
	}

	.page-title {
		margin: 0 0 0.3rem;
		font-size: 1.85rem;
		font-weight: 300;
		color: var(--text);
		font-family: var(--font-display);
		letter-spacing: -0.02em;
	}

	.page-subtitle {
		margin: 0;
		color: rgba(236, 234, 229, 0.6);
		font-size: 0.95rem;
	}

	.header-actions {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		flex-wrap: wrap;
	}

	.status-badge {
		padding: 0.4rem 0.9rem;
		border-radius: 6px;
		background: var(--status-color, var(--amber));
		color: #0d0800;
		font-size: 0.8rem;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.3px;
		white-space: nowrap;
	}

	.btn-edit,
	.btn-delete {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.65rem 1.1rem;
		border-radius: 8px;
		font-weight: 600;
		font-size: 0.875rem;
		cursor: pointer;
		transition: all 0.3s ease;
		white-space: nowrap;
		font-family: var(--font-body);
	}

	.btn-edit {
		border: 1.5px solid rgba(212, 145, 42, 0.3);
		background: transparent;
		color: var(--amber);
	}

	.btn-edit:hover {
		border-color: var(--amber);
		background: var(--border);
	}

	.btn-delete {
		border: 1.5px solid rgba(224, 90, 90, 0.3);
		background: transparent;
		color: var(--red);
	}

	.btn-delete:hover {
		border-color: var(--red);
		background: rgba(224, 90, 90, 0.08);
	}

	.main-content {
		display: flex;
		flex-direction: column;
		gap: 2rem;
	}

	.panel-body {
		padding: 1.75rem;
	}

	.section-title {
		margin: 0 0 1.5rem;
		font-size: 1.15rem;
		font-weight: 400;
		color: var(--text);
		font-family: var(--font-display);
	}

	.info-group {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 1.5rem;
		margin-bottom: 1.5rem;
	}

	.info-item {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.info-label {
		font-size: 0.8rem;
		color: rgba(236, 234, 229, 0.6);
		text-transform: uppercase;
		letter-spacing: 0.3px;
		font-weight: 600;
	}

	.info-value {
		color: var(--text);
		font-size: 0.95rem;
	}

	.info-description {
		margin-top: 1.5rem;
		padding-top: 1.5rem;
		border-top: 1px solid var(--border);
	}

	.info-description h3 {
		margin: 0 0 0.75rem;
		color: var(--text);
		font-size: 0.95rem;
		font-weight: 600;
	}

	.info-description p {
		margin: 0;
		color: rgba(236, 234, 229, 0.75);
		line-height: 1.6;
		font-size: 0.9rem;
	}

	.stats-grid {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1.5rem;
	}

	.stat-card {
		display: flex;
		gap: 1rem;
		padding: 1.25rem;
		border-radius: 12px;
		background: var(--border);
		border: 1px solid var(--border-strong);
		align-items: center;
	}

	.stat-icon {
		width: 44px;
		height: 44px;
		border-radius: 10px;
		background: rgba(212, 145, 42, 0.12);
		border: 1px solid rgba(212, 145, 42, 0.2);
		color: var(--amber);
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	.stat-content {
		display: flex;
		flex-direction: column;
		gap: 0.35rem;
	}

	.stat-label {
		font-size: 0.8rem;
		color: rgba(236, 234, 229, 0.6);
		text-transform: uppercase;
		letter-spacing: 0.3px;
	}

	.stat-value {
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
		color: var(--amber);
		font-size: 1.25rem;
		font-weight: 700;
	}

	.success-msg {
		color: var(--green);
		font-size: 0.9rem;
		font-weight: 600;
		text-align: center;
		padding: 0.75rem;
		border-radius: 8px;
		background: rgba(34, 201, 126, 0.1);
		border: 1px solid rgba(34, 201, 126, 0.2);
		margin: 0;
	}

	@media (max-width: 768px) {
		.page-title {
			font-size: 1.5rem;
		}

		.header-actions {
			width: 100%;
		}

		.stats-grid {
			grid-template-columns: 1fr;
		}

		.info-group {
			grid-template-columns: 1fr;
		}
	}
</style>
