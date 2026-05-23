<script lang="ts">
	import { goto } from '$app/navigation';
	import type { PageProps } from './$types';

	const { data }: PageProps = $props();

	let showDeleteConfirm = $state(false);
	let platformToDelete: string | null = $state(null);

	function viewDetails(id: string) {
		goto(`/dashboard/platforms/${id}`);
	}

	function addNewPlatform() {
		goto('/dashboard/platforms/add');
	}

	function editPlatform(id: string) {
		goto(`/dashboard/platforms/${id}`);
	}

	function confirmDelete(id: string) {
		platformToDelete = id;
		showDeleteConfirm = true;
	}

	function cancelDelete() {
		showDeleteConfirm = false;
		platformToDelete = null;
	}

	function getStatusColor(status: boolean) {
		return status === true ? '#2ecc71' : '#e74c3c';
	}
</script>

<svelte:head>
	<title>Plataformas de Inversión - FINEXIA</title>
	<meta name="description" content="Gestiona tus plataformas de inversión" />
</svelte:head>

<header class="page-header">
	<div class="header-top">
		<div>
			<h1 class="page-title">Plataformas de Inversión</h1>
			<p class="page-subtitle">Administra todas tus plataformas y corredurías en un solo lugar.</p>
		</div>
		<button onclick={addNewPlatform} class="btn-add-platform">
			<svg
				width="18"
				height="18"
				viewBox="0 0 24 24"
				fill="none"
				stroke="currentColor"
				stroke-width="2"
			>
				<path d="M12 5v14M5 12h14" />
			</svg>
			Agregar Plataforma
		</button>
	</div>
</header>

<section class="panel table-panel">
	<header class="table-head">
		<h2>Tus Plataformas</h2>
		<p class="platform-count">
			{data.platforms.length} plataforma{data.platforms.length !== 1 ? 's' : ''}
		</p>
	</header>

	{#if data.platforms.length === 0}
		<div class="empty-state">
			<svg
				width="64"
				height="64"
				viewBox="0 0 24 24"
				fill="none"
				stroke="currentColor"
				stroke-width="1.5"
			>
				<rect x="3" y="3" width="18" height="18" rx="2" />
				<path d="M3 9h18" />
				<path d="M9 3v18" />
			</svg>
			<h3>No hay plataformas registradas</h3>
			<p>Comienza agregando tu primera plataforma de inversión</p>
			<button onclick={addNewPlatform} class="btn-empty-action">Agregar Plataforma</button>
		</div>
	{:else}
		<div class="platforms-grid">
			{#each data.platforms as platform (platform.id)}
				<div class="platform-card">
					<div class="card-header">
						<div class="card-title-section">
							<h3 class="platform-name">{platform.name}</h3>
							<span class="platform-type">{platform.sourceType}</span>
						</div>
						<div class="status-badge" style="--status-color: {getStatusColor(platform.isActive)}">
							{platform.isActive ? 'Activo' : 'Inactivo'}
						</div>
					</div>

					<div class="card-stats">
						<div class="stat-item">
							<span class="stat-label">Inversiones</span>
							<span class="stat-value">{platform.investments ?? 0}</span>
						</div>
						<div class="stat-item">
							<span class="stat-label">Valor Total</span>
							<span class="stat-value">{platform.totalValue ?? 0}</span>
						</div>
					</div>

					<div class="card-actions">
						<button
							onclick={() => viewDetails(platform.id)}
							class="action-btn view-btn"
							aria-label={`Ver detalles de ${platform.name}`}
						>
							<svg
								width="16"
								height="16"
								viewBox="0 0 24 24"
								fill="none"
								stroke="currentColor"
								stroke-width="2"
							>
								<path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
								<circle cx="12" cy="12" r="3" />
							</svg>
							Ver
						</button>
						<button
							onclick={() => editPlatform(platform.id)}
							class="action-btn edit-btn"
							aria-label={`Editar ${platform.name}`}
						>
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
						<button
							onclick={() => confirmDelete(platform.id)}
							class="action-btn delete-btn"
							aria-label={`Eliminar ${platform.name}`}
						>
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
					</div>
				</div>
			{/each}
		</div>
	{/if}
</section>

{#if showDeleteConfirm && platformToDelete}
	<div class="modal-overlay">
		<div class="modal-content">
			<h3>Confirmar eliminación</h3>
			<p>¿Estás seguro de que deseas eliminar esta plataforma? Esta acción no se puede deshacer.</p>
			<div class="modal-actions">
				<button onclick={cancelDelete} class="btn btn-secondary">Cancelar</button>
				<button onclick={() => console.log('delete')} class="btn btn-danger">Eliminar</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.page-header {
		margin-bottom: 2rem;
		padding-bottom: 1.5rem;
		border-bottom: 1px solid rgba(212, 175, 55, 0.1);
	}

	.header-top {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 2rem;
		flex-wrap: wrap;
	}

	.page-title {
		margin: 0 0 0.5rem;
		font-size: 2.35rem;
		font-weight: 700;
		letter-spacing: 0.5px;
		color: #d4af37;
		font-family: 'Poppins', system-ui, sans-serif;
	}

	.page-subtitle {
		margin: 0;
		color: rgba(224, 224, 224, 0.62);
	}

	.btn-add-platform {
		display: flex;
		align-items: center;
		gap: 0.6rem;
		padding: 0.85rem 1.5rem;
		border: none;
		border-radius: 10px;
		background: linear-gradient(135deg, #d4af37, #e8c547);
		color: #0f1419;
		font-weight: 700;
		font-family: 'Poppins', system-ui, sans-serif;
		font-size: 0.95rem;
		cursor: pointer;
		transition: all 0.3s ease;
		letter-spacing: 0.3px;
		white-space: nowrap;
	}

	.btn-add-platform:hover {
		transform: translateY(-2px);
		box-shadow: 0 10px 25px rgba(212, 175, 55, 0.25);
	}

	.panel {
		border: 1px solid rgba(212, 175, 55, 0.15);
		border-radius: 16px;
		background: linear-gradient(135deg, rgba(26, 31, 46, 0.9) 0%, rgba(32, 39, 56, 0.9) 100%);
		box-shadow:
			0 20px 60px rgba(0, 0, 0, 0.3),
			inset 0 1px 0 rgba(255, 255, 255, 0.05);
		backdrop-filter: blur(16px);
	}

	.table-panel {
		padding: 1.5rem;
	}

	.table-head {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 1.5rem;
		border-bottom: 1px solid rgba(212, 175, 55, 0.1);
		padding-bottom: 1rem;
	}

	.table-head h2 {
		margin: 0;
		color: #e0e0e0;
		font-size: 1.15rem;
	}

	.platform-count {
		margin: 0;
		color: rgba(224, 224, 224, 0.6);
		font-size: 0.9rem;
	}

	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 1rem;
		padding: 3rem 2rem;
		text-align: center;
		color: rgba(224, 224, 224, 0.6);
	}

	.empty-state svg {
		color: rgba(212, 175, 55, 0.4);
	}

	.empty-state h3 {
		margin: 0;
		color: #e0e0e0;
		font-size: 1.2rem;
	}

	.empty-state p {
		margin: 0 0 0.5rem;
		color: rgba(224, 224, 224, 0.6);
	}

	.btn-empty-action {
		margin-top: 0.5rem;
		padding: 0.75rem 1.5rem;
		border: 1.5px solid #d4af37;
		border-radius: 8px;
		background: transparent;
		color: #d4af37;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.3s ease;
	}

	.btn-empty-action:hover {
		background: rgba(212, 175, 55, 0.1);
		transform: translateY(-2px);
	}

	.platforms-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
		gap: 1.5rem;
		animation: fade-in 0.4s ease-out;
	}

	.platform-card {
		border: 1px solid rgba(212, 175, 55, 0.15);
		border-radius: 12px;
		background: rgba(15, 20, 25, 0.4);
		padding: 1.5rem;
		transition: all 0.3s ease;
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.platform-card:hover {
		border-color: rgba(212, 175, 55, 0.3);
		background: rgba(15, 20, 25, 0.6);
		box-shadow: 0 10px 30px rgba(212, 175, 55, 0.1);
	}

	.card-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		gap: 1rem;
	}

	.card-title-section {
		display: flex;
		flex-direction: column;
		gap: 0.35rem;
	}

	.platform-name {
		margin: 0;
		color: #e0e0e0;
		font-size: 1.1rem;
		font-weight: 700;
	}

	.platform-type {
		color: rgba(212, 175, 55, 0.75);
		font-size: 0.8rem;
		text-transform: uppercase;
		letter-spacing: 0.4px;
		font-weight: 600;
	}

	.status-badge {
		padding: 0.4rem 0.8rem;
		border-radius: 6px;
		background: var(--status-color, #d4af37);
		color: #0f1419;
		font-size: 0.75rem;
		font-weight: 700;
		white-space: nowrap;
		text-transform: uppercase;
		letter-spacing: 0.3px;
	}

	.card-stats {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1rem;
		padding: 1rem;
		background: rgba(212, 175, 55, 0.08);
		border-radius: 8px;
	}

	.stat-item {
		display: flex;
		flex-direction: column;
		gap: 0.35rem;
	}

	.stat-label {
		font-size: 0.8rem;
		color: rgba(224, 224, 224, 0.6);
		text-transform: uppercase;
		letter-spacing: 0.3px;
		font-weight: 600;
	}

	.stat-value {
		color: #d4af37;
		font-size: 1.1rem;
		font-weight: 700;
		font-family: 'Courier New', monospace;
	}

	.card-actions {
		display: flex;
		gap: 0.5rem;
		margin-top: auto;
	}

	.action-btn {
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.35rem;
		padding: 0.65rem;
		border: 1px solid rgba(212, 175, 55, 0.2);
		border-radius: 8px;
		background: transparent;
		color: #e0e0e0;
		font-size: 0.8rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.3s ease;
		white-space: nowrap;
	}

	.action-btn:hover {
		border-color: rgba(212, 175, 55, 0.5);
		color: #d4af37;
		background: rgba(212, 175, 55, 0.1);
	}

	.delete-btn:hover {
		border-color: #e74c3c;
		color: #e74c3c;
		background: rgba(231, 76, 60, 0.1);
	}

	.modal-overlay {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.5);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 1000;
		backdrop-filter: blur(4px);
	}

	.modal-content {
		background: linear-gradient(135deg, rgba(26, 31, 46, 0.95) 0%, rgba(32, 39, 56, 0.95) 100%);
		border: 1px solid rgba(212, 175, 55, 0.2);
		border-radius: 16px;
		padding: 2rem;
		max-width: 400px;
		box-shadow: 0 25px 50px rgba(0, 0, 0, 0.3);
	}

	.modal-content h3 {
		margin: 0 0 1rem;
		color: #e0e0e0;
		font-size: 1.3rem;
	}

	.modal-content p {
		margin: 0 0 1.5rem;
		color: rgba(224, 224, 224, 0.7);
		line-height: 1.6;
	}

	.modal-actions {
		display: flex;
		gap: 1rem;
	}

	.btn {
		flex: 1;
		padding: 0.75rem 1.5rem;
		border: none;
		border-radius: 8px;
		font-weight: 700;
		font-size: 0.9rem;
		cursor: pointer;
		transition: all 0.3s ease;
		font-family: 'Poppins', system-ui, sans-serif;
	}

	.btn-secondary {
		border: 1.5px solid rgba(212, 175, 55, 0.25);
		background: transparent;
		color: #e0e0e0;
	}

	.btn-secondary:hover {
		border-color: #d4af37;
		color: #d4af37;
		background: rgba(212, 175, 55, 0.1);
	}

	.btn-danger {
		background: #e74c3c;
		color: white;
	}

	.btn-danger:hover {
		background: #c0392b;
		box-shadow: 0 10px 25px rgba(231, 76, 60, 0.3);
	}

	@keyframes fade-in {
		from {
			opacity: 0;
			transform: translateY(10px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	@media (max-width: 768px) {
		.page-title {
			font-size: 1.85rem;
		}

		.header-top {
			flex-direction: column;
		}

		.btn-add-platform {
			width: 100%;
		}

		.platforms-grid {
			grid-template-columns: 1fr;
		}

		.table-head {
			flex-direction: column;
			align-items: flex-start;
			gap: 0.5rem;
		}

		.action-btn {
			font-size: 0.75rem;
			padding: 0.5rem;
		}

		.action-btn svg {
			width: 14px;
			height: 14px;
		}
	}
</style>
