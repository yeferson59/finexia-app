<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import Card from '$lib/ui/card.svelte';
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency as formatMoney } from '$lib/shared/format/money';
	import PlatformStatsGrid from './platform-stats-grid.svelte';
	import { FALLBACK_CURRENCY } from '$lib/shared/currency';
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

	// La moneda del total la informa el backend: es la suma de posiciones
	// compradas en varias monedas, convertida a la de la cuenta, y etiquetarla
	// siempre con "$" le ponía el símbolo equivocado.
	const currency = $derived(platform.displayCurrency || FALLBACK_CURRENCY);

	// Posiciones que entraron al total sin convertir por no haber tasa.
	const unconverted = $derived(platform.positionsUnconverted ?? 0);

	// Ausente no es cero: un backend anterior a estas métricas no dice que la
	// plataforma no haya ganado nada, dice que no lo sabe. Se calla en vez de
	// afirmar.
	const gain = $derived(platform.gainLoss !== undefined ? parseFloat(platform.gainLoss) : null);

	function formatCurrency(value: string): string {
		return privacy.money(formatMoney(parseFloat(value) || 0, currency));
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
					<PlatformStatsGrid {platform} {unconverted} {gain} {formatCurrency} />
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

		.info-group {
			grid-template-columns: 1fr;
		}
	}
</style>
