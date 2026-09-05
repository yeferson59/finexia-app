<script lang="ts">
	/*
	 * Lo que se puede hacer con un movimiento, con su nombre escrito.
	 *
	 * Eran dos iconos de 13 píxeles cuyo único texto vivía en el `aria-label`,
	 * en una tabla de filas casi idénticas y con un borrado que se lleva la
	 * operación para siempre: había que acertar el lápiz correcto para editar y
	 * la papelera correcta para borrar.
	 *
	 * Cada botón completa su nombre con la fila a la que pertenece, en el texto
	 * que solo oye el lector de pantalla: en pantalla la columna ya lo dice, y
	 * fuera de ella «Eliminar» a secas se repite una vez por movimiento.
	 */
	import { formatCalendarDate } from '$lib/shared/format/date';
	import type { Transaction } from '$lib/api/types';
	import { TYPE_LABEL } from '../asset';

	let {
		transaction,
		selling,
		onEdit,
		onToggleSell,
		onDelete
	}: {
		transaction: Transaction;
		/** Este es el lote cuyo panel de venta está abierto. */
		selling: boolean;
		onEdit: (txn: Transaction) => void;
		onToggleSell: (txn: Transaction) => void;
		onDelete: (txn: Transaction) => void;
	} = $props();

	const date = $derived(
		formatCalendarDate(transaction.transactionDate, {
			year: 'numeric',
			month: 'short',
			day: 'numeric'
		})
	);

	/** Nombra la fila: «el Compra del 20 de feb de 2026». */
	const label = $derived(`el ${TYPE_LABEL[transaction.type] ?? transaction.type} del ${date}`);

	/** Solo se puede vender lo que entró: un dividendo no es una posición. */
	const isBuyLot = $derived(transaction.type === 'buy' || transaction.type === 'transfer_in');
</script>

<div class="actions">
	<button type="button" class="action" onclick={() => onEdit(transaction)}>
		Editar<span class="sr-only"> {label}</span>
	</button>
	<button type="button" class="action danger" onclick={() => onDelete(transaction)}>
		Eliminar<span class="sr-only"> {label}</span>
	</button>
	{#if isBuyLot}
		<button
			type="button"
			class="action sell"
			class:active={selling}
			onclick={() => onToggleSell(transaction)}
		>
			{selling ? 'Cancelar' : 'Vender'}<span class="sr-only"> este lote del {date}</span>
		</button>
	{/if}
</div>

<style>
	.actions {
		display: flex;
		justify-content: flex-end;
		gap: 0.9rem;
		white-space: nowrap;
	}

	.action {
		padding: 0;
		border: none;
		background: none;
		color: var(--text-muted);
		font-family: var(--font-body);
		font-size: 0.8rem;
		cursor: pointer;
		transition: color 0.15s ease;
	}

	.action:hover {
		color: var(--text);
		text-decoration: underline;
		text-underline-offset: 3px;
	}

	.action.danger:hover,
	.action.sell:hover,
	.action.sell.active {
		color: var(--red);
	}

	.sr-only {
		position: absolute;
		width: 1px;
		height: 1px;
		padding: 0;
		margin: -1px;
		overflow: hidden;
		clip-path: inset(50%);
		white-space: nowrap;
	}

	@media (prefers-reduced-motion: reduce) {
		.action {
			transition: none;
		}
	}

	/* Plegada, la fila ocupa todo el ancho y las acciones empiezan por la
	   izquierda, alineadas con el resto de la fila. */
	@media (max-width: 760px) {
		.actions {
			justify-content: flex-start;
		}
	}
</style>
