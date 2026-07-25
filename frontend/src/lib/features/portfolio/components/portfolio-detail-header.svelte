<script lang="ts">
	let {
		name,
		description,
		onBack,
		onEdit,
		onAddAsset
	}: {
		name: string;
		description?: string | null;
		onBack: () => void;
		onEdit: () => void;
		onAddAsset: () => void;
	} = $props();
</script>

<header class="page-header">
	<div class="header-top">
		<div>
			<button onclick={onBack} class="btn-back" aria-label="Volver a portafolios">
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
			<h1 class="page-title">{name}</h1>
			<p class="page-subtitle">
				{description || 'Visión detallada de posiciones y asignación.'}
			</p>
		</div>
		<div class="header-actions">
			<button onclick={onEdit} class="btn-edit" aria-label="Editar portafolio">
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
			<button onclick={onAddAsset} class="btn-add-asset">
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
				Agregar Activo
			</button>
		</div>
	</div>
</header>

<style>
	.page-header {
		margin-bottom: 2rem;
		padding-bottom: 1.5rem;
		border-bottom: 1px solid var(--border);
	}

	.header-top {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 2rem;
		flex-wrap: wrap;
	}

	.header-top > div {
		display: flex;
		align-items: flex-start;
		gap: 1rem;
	}

	.page-title {
		margin: 0 0 0.5rem;
		font-size: 2.35rem;
		font-weight: 300;
		letter-spacing: -0.02em;
		color: var(--text);
		font-family: var(--font-display);
	}

	.page-subtitle {
		margin: 0;
		color: rgba(236, 234, 229, 0.62);
		font-size: 1rem;
	}

	.btn-back {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 40px;
		height: 40px;
		padding: 0;
		margin-right: 1rem;
		border: 1px solid rgba(212, 145, 42, 0.3);
		border-radius: 8px;
		background: var(--border);
		color: var(--amber);
		cursor: pointer;
		transition: all 0.3s ease;
	}

	.btn-back:hover {
		background: rgba(212, 145, 42, 0.2);
		border-color: rgba(212, 145, 42, 0.5);
		transform: translateX(-2px);
	}

	.header-actions {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		flex-wrap: wrap;
	}

	.btn-edit {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.75rem 1.25rem;
		border: 1px solid rgba(212, 145, 42, 0.4);
		border-radius: 10px;
		background: transparent;
		color: var(--amber);
		font-weight: 600;
		font-family: var(--font-body);
		font-size: 0.9rem;
		cursor: pointer;
		transition: all 0.3s ease;
		white-space: nowrap;
	}

	.btn-edit:hover {
		background: rgba(212, 145, 42, 0.12);
		border-color: rgba(212, 145, 42, 0.7);
	}

	.btn-add-asset {
		display: flex;
		align-items: center;
		gap: 0.6rem;
		padding: 0.85rem 1.5rem;
		border: none;
		border-radius: 10px;
		background: var(--amber);
		color: #0d0800;
		font-weight: 700;
		font-family: var(--font-body);
		font-size: 0.95rem;
		cursor: pointer;
		transition: all 0.3s ease;
		letter-spacing: 0.3px;
		white-space: nowrap;
	}

	.btn-add-asset:hover {
		transform: translateY(-2px);
		box-shadow: 0 10px 25px rgba(212, 145, 42, 0.25);
	}

	@media (max-width: 768px) {
		.page-title {
			font-size: 1.85rem;
		}

		.header-top {
			flex-direction: column;
		}

		.btn-add-asset {
			width: 100%;
			justify-content: center;
		}
	}
</style>
