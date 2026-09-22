<script lang="ts">
	/**
	 * Cuerpo de la confirmación de borrado de una transacción.
	 *
	 * Muestra la fila que se va a eliminar —tipo, fecha y total— porque el botón
	 * vive en una tabla donde todas las filas se parecen y el borrado no se puede
	 * deshacer. El diálogo lo pone el `Modal` del historial.
	 */
	import { enhance } from '$app/forms';
	import { OptimisticList, optimisticSubmit } from '$lib/shared/optimistic.svelte';
	import Button from '$lib/ui/button.svelte';
	import { formatCalendarDate } from '$lib/shared/format/date';
	import type { Transaction } from '$lib/api/types';
	import { TYPE_LABEL } from '../asset';

	let {
		transaction,
		pending,
		onRejected,
		onClose,
		formatAmount
	}: {
		transaction: Transaction;
		/** De donde la fila se va al confirmar, sin esperar al servidor. */
		pending: OptimisticList<Transaction>;
		/**
		 * El servidor rechazó el borrado. El diálogo ya se cerró —no queda nada
		 * que corregir en él—, así que el motivo lo enseña el historial, con la
		 * fila de vuelta en su sitio.
		 */
		onRejected: (message: string) => void;
		onClose: () => void;
		formatAmount: (value: number, currency: string) => string;
	} = $props();

	const submit = optimisticSubmit({
		fallbackError: 'No se pudo eliminar la transacción.',
		apply: () => {
			const undo = pending.remove(transaction.id);
			onClose();
			return undo;
		},
		onError: (message) => onRejected(message)
	});

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
{#if transaction.cashCredited}
	<p class="warning">
		Lo que abonó al efectivo de la plataforma también sale de ahí. Si ese dinero ya salió en un
		retiro, no se podrá borrar hasta quitar el retiro.
	</p>
{/if}

{#if transaction.cashPaid}
	<p class="warning">Lo que pagaste con el efectivo de la plataforma vuelve a ese saldo.</p>
{/if}

<form method="POST" action="?/deleteTransaction" use:enhance={submit}>
	<input type="hidden" name="txnId" value={transaction.id} />
	<div class="modal-actions">
		<Button type="button" variant="ghost" onclick={onClose}>Cancelar</Button>
		<Button type="submit" variant="danger">Eliminar</Button>
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
</style>
