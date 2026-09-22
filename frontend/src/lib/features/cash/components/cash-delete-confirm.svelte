<script lang="ts">
	/**
	 * Confirmación para borrar un movimiento.
	 *
	 * Dice qué pasa con el saldo, que es lo que el usuario no ve venir: borrar un
	 * depósito cuyo dinero ya salió en un retiro dejaría la cuenta en negativo, y
	 * el backend lo rechaza. Mejor avisarlo aquí que presentarlo como un error.
	 */
	import { enhance } from '$app/forms';
	import Button from '$lib/ui/button.svelte';
	import Modal from '$lib/ui/modal.svelte';
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import { formatCalendarDate } from '$lib/shared/format/date';
	import { OptimisticDialog, OptimisticList } from '$lib/shared/optimistic.svelte';
	import { cashAccountLabel, cashKindSign, formatCashKind, type CashMovement } from '../cash';

	interface Props {
		movement: CashMovement | null;
		/** Si la cuenta nombra también el portafolio. */
		showPortfolio?: boolean;
		/** El extracto: la fila se va al confirmar, sin esperar al servidor. */
		pending: OptimisticList<CashMovement>;
		onClose: () => void;
	}

	let { movement, showPortfolio = true, pending, onClose }: Props = $props();

	/* Si el servidor rechaza el borrado, la fila vuelve y el diálogo reaparece
	   con el motivo. */
	const dialog = new OptimisticDialog(() => movement);

	/* Cerrar limpia el error: el siguiente movimiento que se abra no lo arrastra. */
	function close() {
		dialog.reset();
		onClose();
	}

	const submit = dialog.submit({
		fallbackError: 'No pudimos borrar el movimiento.',
		apply: () => (movement ? pending.remove(movement.id) : undefined),
		onDone: close
	});

	const amount = $derived.by(() => {
		if (!movement) return '';
		const sign = cashKindSign(movement.kind);
		const prefix = sign > 0 ? '+' : sign < 0 ? '−' : '';
		return privacy.money(
			`${prefix}${formatCurrency(Math.abs(parseFloat(movement.amount) || 0), movement.currency)}`
		);
	});
</script>

<Modal
	open={movement !== null}
	hidden={dialog.hidden}
	title="Borrar movimiento"
	description="El saldo se recalcula sin él."
	size="sm"
	tone="danger"
	onClose={close}
>
	{#if movement}
		<form method="POST" action="?/delete" use:enhance={submit}>
			<input type="hidden" name="id" value={movement.id} />

			<!-- El movimiento como se ve en el extracto, para reconocerlo antes de
			     borrarlo: qué, cuánto, dónde y cuándo. -->
			<div class="receipt">
				<p class="receipt-kind">{formatCashKind(movement.kind)}</p>
				<p class="receipt-amount" class:income={movement.kind === 'interest'}>{amount}</p>
				<p class="receipt-meta">
					{cashAccountLabel(movement, showPortfolio)}, {formatCalendarDate(movement.date, {
						day: 'numeric',
						month: 'long',
						year: 'numeric'
					})}
				</p>
			</div>

			{#if movement.kind === 'deposit' || movement.kind === 'interest'}
				<p class="body">
					Si ese dinero ya salió en un retiro, no se podrá borrar hasta quitar el retiro.
				</p>
			{/if}

			{#if dialog.error}
				<p class="feedback error" role="alert">{dialog.error}</p>
			{/if}

			<div class="modal-actions">
				<Button type="button" variant="ghost" onclick={close} disabled={dialog.submitting}
					>Cancelar</Button
				>
				<Button type="submit" variant="danger" loading={dialog.submitting}>Borrar</Button>
			</div>
		</form>
	{/if}
</Modal>

<style>
	.receipt {
		display: grid;
		gap: 0.2rem;
		margin: 0 0 1rem;
		padding: 0.9rem 1rem;
		border: 1px solid var(--border);
		border-radius: 10px;
		background: rgba(255, 255, 255, 0.02);
	}

	.receipt p {
		margin: 0;
	}

	.receipt-kind {
		font-size: 0.8rem;
		color: var(--text-muted);
	}

	.receipt-amount {
		font-family: var(--font-mono);
		font-size: 1.35rem;
		letter-spacing: -0.02em;
		color: var(--text);
	}

	.receipt-amount.income {
		color: var(--green);
	}

	.receipt-meta {
		font-size: 0.8rem;
		color: var(--text-dim);
	}

	.body {
		margin: 0 0 1rem;
		font-size: 0.87rem;
		line-height: 1.55;
		color: var(--text-muted);
	}

	.feedback {
		margin: 0 0 1rem;
	}
</style>
