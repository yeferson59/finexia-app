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
	import { cashAccountLabel, formatCashKind, type CashMovement } from '../cash';

	interface Props {
		movement: CashMovement | null;
		onClose: () => void;
	}

	let { movement, onClose }: Props = $props();

	let submitting = $state(false);
	let error = $state('');

	$effect(() => {
		if (movement) error = '';
	});

	const amount = $derived(
		movement
			? privacy.money(formatCurrency(parseFloat(movement.amount) || 0, movement.currency))
			: ''
	);
</script>

<Modal
	open={movement !== null}
	title="Borrar movimiento"
	description="El saldo se recalcula sin él."
	size="sm"
	tone="danger"
	{onClose}
>
	{#if movement}
		<form
			method="POST"
			action="?/delete"
			use:enhance={() => {
				submitting = true;
				return async ({ result, update }) => {
					submitting = false;
					if (result.type === 'failure') {
						error = (result.data?.error as string) ?? 'No pudimos borrar el movimiento.';
						return;
					}
					await update();
					onClose();
				};
			}}
		>
			<input type="hidden" name="id" value={movement.id} />

			<p class="body">
				{formatCashKind(movement.kind)} de <strong>{amount}</strong> en {cashAccountLabel(
					movement
				)}.
				{#if movement.kind === 'deposit' || movement.kind === 'interest'}
					Si ese dinero ya salió en un retiro, no se podrá borrar hasta quitar el retiro.
				{/if}
			</p>

			{#if error}
				<p class="feedback error">{error}</p>
			{/if}

			<div class="modal-actions">
				<Button type="button" variant="ghost" onclick={onClose} disabled={submitting}
					>Cancelar</Button
				>
				<Button type="submit" variant="danger" loading={submitting}>Borrar</Button>
			</div>
		</form>
	{/if}
</Modal>

<style>
	.body {
		margin: 0 0 1rem;
		font-size: 0.9rem;
		line-height: 1.6;
		color: var(--text-muted);
	}

	.body strong {
		color: var(--text);
		font-weight: 500;
	}

	.feedback {
		margin: 0 0 1rem;
	}
</style>
