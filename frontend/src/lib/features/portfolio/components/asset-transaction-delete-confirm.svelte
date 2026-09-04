<script lang="ts">
	/**
	 * Cuerpo de la confirmación de borrado de una transacción.
	 *
	 * Muestra la fila que se va a eliminar —tipo, fecha y total— porque el botón
	 * vive en una tabla donde todas las filas se parecen y el borrado no se puede
	 * deshacer. El diálogo lo pone el `Modal` del historial.
	 */
	import { enhance } from '$app/forms';
	import { invalidateAll } from '$app/navigation';
	import Button from '$lib/ui/button.svelte';
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
	La posición se recalcula con las transacciones que queden; si era la última, la cantidad pasa a 0.
	Esta acción no se puede deshacer.
</p>

{#if deleteError}
	<p class="error" role="alert">{deleteError}</p>
{/if}

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
	<div class="actions">
		<Button type="button" variant="ghost" onclick={onClose} disabled={isDeleting}>Cancelar</Button>
		<Button type="submit" variant="danger" loading={isDeleting}>Eliminar</Button>
	</div>
</form>

<style>
	.summary {
		margin: 0 0 0.6rem;
		color: var(--text);
		line-height: 1.6;
	}

	.summary strong {
		font-weight: 500;
	}

	.warning {
		margin: 0;
		color: var(--text-muted);
		font-size: 0.9rem;
		line-height: 1.6;
	}

	.error {
		margin: 1rem 0 0;
		padding: 0.6rem 0.85rem;
		border-left: 2px solid var(--red);
		background: rgba(224, 90, 90, 0.08);
		color: var(--red);
		font-size: 0.85rem;
	}

	.actions {
		display: flex;
		gap: 0.75rem;
		justify-content: flex-end;
		margin-top: 1.5rem;
	}
</style>
