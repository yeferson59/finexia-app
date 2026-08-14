<script lang="ts">
	/**
	 * Confirmación de borrado de una transacción. Muestra la fila que se va a
	 * eliminar —tipo, fecha y total— porque el botón vive en una tabla donde
	 * todas las filas se parecen, y el borrado no se puede deshacer.
	 *
	 * Cierra e invalida por su cuenta, igual que el modal de edición.
	 */
	import { enhance } from '$app/forms';
	import { invalidateAll } from '$app/navigation';
	import { formatCalendarDate } from '$lib/shared/format/date';
	import type { Transaction } from '$lib/api/types';
	import { TYPE_LABEL } from '../asset';

	let {
		transaction,
		onClose,
		formatAmount
	}: {
		transaction: Transaction;
		onClose: () => void;
		formatAmount: (value: number, currency: string) => string;
	} = $props();

	let isDeleting = $state(false);
	let deleteError = $state('');

	const total = $derived(
		(parseFloat(transaction.quantity) || 0) * (parseFloat(transaction.price) || 0)
	);
</script>

<div
	class="modal-backdrop"
	role="button"
	tabindex="0"
	aria-label="Cerrar modal"
	onclick={onClose}
	onkeydown={(e) => e.key === 'Enter' && onClose()}
></div>
<div class="modal" role="dialog" aria-modal="true" aria-label="Eliminar transacción">
	<h3>Eliminar transacción</h3>

	<p class="summary">
		<strong>{TYPE_LABEL[transaction.type] ?? transaction.type}</strong>
		del {formatCalendarDate(transaction.transactionDate, {
			year: 'numeric',
			month: 'short',
			day: 'numeric'
		})}
		por {formatAmount(total, transaction.currency)}
	</p>
	<p class="warning">
		La posición se recalcula con las transacciones que queden; si era la última, la cantidad pasa a
		0. Esta acción no se puede deshacer.
	</p>

	{#if deleteError}
		<p class="error">{deleteError}</p>
	{/if}

	<div class="modal-actions">
		<button type="button" onclick={onClose} class="btn btn-secondary">Cancelar</button>
		<form
			method="POST"
			action="?/deleteTransaction"
			use:enhance={() => {
				isDeleting = true;
				deleteError = '';
				return async ({ result, update }) => {
					await update({ reset: false });
					isDeleting = false;
					const data =
						result.type === 'success'
							? (result.data as { success?: boolean; error?: string } | undefined)
							: undefined;
					if (data?.success) {
						onClose();
						await invalidateAll();
					} else {
						deleteError = data?.error ?? 'No se pudo eliminar la transacción.';
					}
				};
			}}
		>
			<input type="hidden" name="txnId" value={transaction.id} />
			<button type="submit" disabled={isDeleting} class="btn btn-danger">
				{isDeleting ? 'Eliminando...' : 'Eliminar'}
			</button>
		</form>
	</div>
</div>

<style>
	.modal-backdrop {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.55);
		backdrop-filter: blur(4px);
		z-index: 1000;
	}

	.modal {
		position: fixed;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
		z-index: 1001;
		width: min(420px, 90vw);
		background: var(--surface);
		border: 1px solid rgba(212, 145, 42, 0.2);
		border-radius: 16px;
		padding: 2rem;
		box-shadow: 0 25px 50px rgba(0, 0, 0, 0.4);
	}

	.modal h3 {
		margin: 0 0 1rem;
		color: var(--text);
		font-size: 1.3rem;
		font-family: var(--font-body);
	}

	.summary {
		margin: 0 0 0.75rem;
		color: var(--text);
		line-height: 1.6;
	}

	.warning {
		margin: 0 0 1.5rem;
		color: rgba(236, 234, 229, 0.7);
		font-size: 0.9rem;
		line-height: 1.6;
	}

	.error {
		margin: 0 0 1rem;
		color: var(--red);
		font-size: 0.85rem;
	}

	.modal-actions {
		display: flex;
		gap: 1rem;
		align-items: center;
	}

	.modal-actions form {
		flex: 1;
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
		flex: 1;
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
		width: 100%;
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
</style>
