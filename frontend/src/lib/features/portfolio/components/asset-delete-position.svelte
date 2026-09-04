<script lang="ts">
	/**
	 * Eliminar una posición del portafolio.
	 *
	 * Vive aparte del borrado de transacciones porque no es lo mismo: aquello
	 * quita una operación y deja que la posición se recalcule con las que
	 * queden; esto quita la posición **y todo su historial**, porque la clave
	 * foránea cascadea. La confirmación existe para decir eso con un número,
	 * cuando se sabe cuál es.
	 *
	 * Una posición por plataforma: el mismo ticker comprado en dos brókers son
	 * dos filas, y se borran por separado. Por eso el botón se repite en vez de
	 * ofrecer un «eliminar activo» que borraría más de lo que el usuario ve.
	 */
	import { enhance } from '$app/forms';
	import { goto, invalidateAll } from '$app/navigation';
	import Modal from '$lib/ui/modal.svelte';
	import Button from '$lib/ui/button.svelte';
	import { resolve } from '$app/paths';
	import { formatCalendarDate } from '$lib/shared/format/date';
	import type { Holding } from '$lib/api/types';

	let {
		portfolioId,
		entries,
		transactionsCount,
		formatAmount
	}: {
		portfolioId: string;
		entries: Holding[];
		/**
		 * Transacciones del ticker en este portafolio. Solo describe una entrada
		 * concreta cuando hay una sola; con varias se suman todas y el diálogo
		 * deja de dar la cifra en vez de dar una que no es.
		 */
		transactionsCount: number;
		formatAmount: (value: number, currency: string) => string;
	} = $props();

	let confirming = $state<Holding | null>(null);
	let isDeleting = $state(false);
	let deleteError = $state('');

	const knownCount = $derived(entries.length === 1 ? transactionsCount : null);

	function entryLabel(entry: Holding): string {
		return `${entry.costCurrency} · ${formatCalendarDate(entry.entryDate, {
			year: 'numeric',
			month: 'short',
			day: 'numeric'
		})}`;
	}

	function close() {
		confirming = null;
		deleteError = '';
	}
</script>

<section class="danger-zone">
	<h3 class="danger-title">Eliminar posición</h3>
	<p class="danger-hint">
		Quita la posición del portafolio junto con todas sus transacciones. No se puede deshacer.
	</p>

	{#each entries as entry (entry.id)}
		<div class="entry-row">
			<div>
				<p class="entry-label">{entryLabel(entry)}</p>
				<p class="entry-detail">
					{parseFloat(entry.quantity).toLocaleString('es-CO', { maximumFractionDigits: 8 })} unidades
					· {formatAmount(
						(parseFloat(entry.quantity) || 0) * (parseFloat(entry.price) || 0),
						entry.costCurrency
					)} de coste
				</p>
			</div>
			<button type="button" class="btn-delete" onclick={() => (confirming = entry)}>
				Eliminar
			</button>
		</div>
	{/each}
</section>

<Modal open={!!confirming} title="Eliminar posición" onClose={close} size="sm">
	{#if confirming}
		<p class="summary">
			<strong>{confirming.ticker}</strong> — {entryLabel(confirming)}
		</p>
		<p class="warning">
			{#if knownCount !== null}
				Se eliminarán también sus
				<strong>{knownCount} {knownCount === 1 ? 'transacción' : 'transacciones'}</strong>.
			{:else}
				Se eliminarán también todas las transacciones de esta posición.
			{/if}
			Esta acción no se puede deshacer.
		</p>

		{#if deleteError}
			<p class="error" role="alert">{deleteError}</p>
		{/if}

		<form
			method="POST"
			action="?/deleteEntry"
			use:enhance={() => {
				isDeleting = true;
				const wasLast = entries.length === 1;

				return async ({ result }) => {
					isDeleting = false;

					if (result.type === 'success' && result.data?.success) {
						close();
						// La última entrada deja esta ficha sin nada que mostrar, así
						// que la salida es el portafolio; con más de una, la página
						// sigue teniendo sentido y basta recargarla.
						if (wasLast) {
							await goto(resolve('/dashboard/portfolios/[id]', { id: portfolioId }));
						} else {
							await invalidateAll();
						}
						return;
					}

					deleteError =
						(result.type === 'success' && (result.data?.error as string)) ||
						'No se pudo eliminar la posición.';
				};
			}}
		>
			<input type="hidden" name="entryId" value={confirming.id} />
			<div class="modal-actions">
				<Button type="button" variant="ghost" onclick={close} disabled={isDeleting}>Cancelar</Button
				>
				<Button type="submit" variant="danger" loading={isDeleting}>Eliminar posición</Button>
			</div>
		</form>
	{/if}
</Modal>

<style>
	.danger-zone {
		margin-top: 1.5rem;
		padding: 1.25rem;
		border: 1px solid rgba(224, 90, 90, 0.25);
		border-radius: 12px;
		background: rgba(224, 90, 90, 0.04);
	}

	.danger-title {
		margin: 0 0 0.35rem;
		font-size: 1rem;
		font-weight: 600;
		color: rgba(224, 90, 90, 0.9);
	}

	.danger-hint {
		margin: 0 0 1rem;
		font-size: 0.85rem;
		color: rgba(236, 234, 229, 0.5);
	}

	.entry-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
		padding: 0.75rem 0;
		border-top: 1px solid rgba(224, 90, 90, 0.15);
	}

	.entry-label {
		margin: 0;
		font-size: 0.9rem;
		font-weight: 600;
		color: var(--text);
	}

	.entry-detail {
		margin: 0.2rem 0 0;
		font-size: 0.8rem;
		color: rgba(236, 234, 229, 0.5);
		font-variant-numeric: tabular-nums;
	}

	.btn-delete {
		flex-shrink: 0;
		padding: 0.5rem 1rem;
		border: 1.5px solid rgba(224, 90, 90, 0.35);
		border-radius: 8px;
		background: transparent;
		color: rgba(224, 90, 90, 0.9);
		font-size: 0.85rem;
		font-weight: 600;
		font-family: var(--font-body);
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.btn-delete:hover {
		background: rgba(224, 90, 90, 0.12);
		border-color: rgba(224, 90, 90, 0.6);
	}

	.summary {
		margin: 0 0 0.6rem;
		font-size: 0.9rem;
		color: var(--text);
	}

	.warning {
		margin: 0 0 1.1rem;
		font-size: 0.85rem;
		color: rgba(236, 234, 229, 0.6);
		line-height: 1.5;
	}

	.error {
		margin: 0 0 1rem;
		padding: 0.6rem 0.9rem;
		border-radius: 6px;
		background: rgba(224, 90, 90, 0.1);
		border: 1px solid rgba(224, 90, 90, 0.3);
		color: rgba(224, 90, 90, 0.9);
		font-size: 0.85rem;
	}

	.modal-actions {
		display: flex;
		gap: 0.75rem;
		justify-content: flex-end;
	}
</style>
