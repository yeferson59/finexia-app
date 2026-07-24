<script lang="ts">
	import { enhance } from '$app/forms';

	let { platformName, onCancel }: { platformName: string; onCancel: () => void } = $props();

	let isDeleting = $state(false);
</script>

<div class="modal-overlay">
	<div class="modal-content">
		<h3>Confirmar eliminación</h3>
		<p>
			¿Estás seguro de que deseas eliminar <strong>{platformName}</strong>? Esta acción no se puede
			deshacer.
		</p>
		<div class="modal-actions">
			<button onclick={onCancel} class="btn btn-secondary"> Cancelar </button>
			<form
				method="POST"
				action="?/delete"
				use:enhance={() => {
					isDeleting = true;
					return async ({ update }) => {
						await update();
						isDeleting = false;
					};
				}}
			>
				<button type="submit" disabled={isDeleting} class="btn btn-danger">
					{#if isDeleting}
						<span class="spinner spinner-white"></span>
						Eliminando...
					{:else}
						Eliminar
					{/if}
				</button>
			</form>
		</div>
	</div>
</div>

<style>
	.modal-overlay {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.55);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 1000;
		backdrop-filter: blur(4px);
	}

	.modal-content {
		background: var(--surface);
		border: 1px solid rgba(212, 145, 42, 0.2);
		border-radius: 16px;
		padding: 2rem;
		max-width: 420px;
		width: 90%;
		box-shadow: 0 25px 50px rgba(0, 0, 0, 0.4);
	}

	.modal-content h3 {
		margin: 0 0 1rem;
		color: var(--text);
		font-size: 1.3rem;
		font-family: var(--font-body);
	}

	.modal-content p {
		margin: 0 0 1.5rem;
		color: rgba(236, 234, 229, 0.7);
		line-height: 1.6;
	}

	.modal-actions {
		display: flex;
		gap: 1rem;
		align-items: center;
	}

	.modal-actions form {
		flex: 1;
	}

	.modal-actions .btn-secondary {
		flex: 1;
	}

	.modal-actions .btn-danger {
		width: 100%;
	}

	.btn {
		padding: 0.75rem 1.5rem;
		border: none;
		border-radius: 8px;
		font-weight: 700;
		font-family: var(--font-body);
		font-size: 0.9rem;
		cursor: pointer;
		transition: all 0.3s ease;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
		letter-spacing: 0.3px;
	}

	.btn-secondary {
		background: transparent;
		color: var(--text);
		border: 1.5px solid rgba(212, 145, 42, 0.25);
	}

	.btn-secondary:hover {
		border-color: var(--amber);
		background: var(--border);
		color: var(--amber);
	}

	.btn-danger {
		background: var(--red);
		color: white;
	}

	.btn-danger:hover:not(:disabled) {
		box-shadow: 0 10px 25px rgba(224, 90, 90, 0.3);
	}

	.btn-danger:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.spinner {
		display: inline-block;
		width: 14px;
		height: 14px;
		border: 2px solid rgba(13, 8, 0, 0.3);
		border-top-color: #0d0800;
		border-radius: 50%;
		animation: spin 0.6s linear infinite;
	}

	.spinner-white {
		border-color: rgba(255, 255, 255, 0.3);
		border-top-color: white;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}

	@media (max-width: 768px) {
		.btn {
			width: 100%;
		}
	}
</style>
