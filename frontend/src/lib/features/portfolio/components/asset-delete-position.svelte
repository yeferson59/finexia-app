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
	import { formatUnits } from '../asset';

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

	function entryDate(entry: Holding): string {
		return formatCalendarDate(entry.entryDate, {
			year: 'numeric',
			month: 'long',
			day: 'numeric'
		});
	}

	function close() {
		confirming = null;
		deleteError = '';
	}
</script>

<section class="remove" aria-labelledby="remove-title">
	<h2 class="title" id="remove-title">Quitar esta posición</h2>
	<p class="hint">
		Se va del portafolio con todas sus transacciones, y no se puede deshacer. Para corregir una
		cifra sin perder el historial, edita el movimiento que la trajo.
	</p>

	{#each entries as entry (entry.id)}
		<div class="entry">
			<p class="detail">
				{formatUnits(parseFloat(entry.quantity) || 0, entry.assetType)}, {formatAmount(
					(parseFloat(entry.quantity) || 0) * (parseFloat(entry.price) || 0),
					entry.costCurrency
				)} de coste, desde el {entryDate(entry)}
			</p>
			<button type="button" class="remove-entry" onclick={() => (confirming = entry)}>
				Eliminar
			</button>
		</div>
	{/each}
</section>

<Modal open={!!confirming} title="Eliminar posición" onClose={close} size="sm">
	{#if confirming}
		<p class="summary">
			<strong>{confirming.ticker}</strong>, {formatUnits(
				parseFloat(confirming.quantity) || 0,
				confirming.assetType
			)} desde el {entryDate(confirming)}
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
	/*
	 * El borrado no es una zona aparte con su propio fondo rojo: es el último
	 * apartado de la ficha, escrito en el mismo tono que los demás. El rojo se
	 * guarda para el botón que lo hace y para el diálogo que lo confirma, que
	 * es donde de verdad hay algo que perder.
	 */
	.remove {
		padding: 2rem 0 0;
	}

	.title {
		margin: 0;
		font-family: var(--font-body);
		font-size: 1.05rem;
		font-weight: 500;
		color: var(--text);
	}

	.hint {
		max-width: 64ch;
		margin: 0.5rem 0 0;
		font-size: 0.85rem;
		line-height: 1.5;
		color: var(--text-muted);
	}

	.entry {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		justify-content: space-between;
		gap: 0.75rem 1rem;
		margin-top: 1.1rem;
		padding-top: 1.1rem;
		border-top: 1px solid var(--border);
	}

	.detail {
		margin: 0;
		font-size: 0.85rem;
		color: var(--text-muted);
		font-variant-numeric: tabular-nums;
	}

	.remove-entry {
		flex-shrink: 0;
		padding: 0.5rem 1.1rem;
		border: 1px solid var(--border-strong);
		border-radius: 9px;
		background: none;
		color: var(--red);
		font-family: var(--font-body);
		font-size: 0.85rem;
		font-weight: 600;
		cursor: pointer;
		transition:
			border-color 0.2s ease,
			background 0.2s ease;
	}

	.remove-entry:hover {
		border-color: var(--red);
		background: rgba(224, 90, 90, 0.1);
	}

	@media (prefers-reduced-motion: reduce) {
		.remove-entry {
			transition: none;
		}
	}

	.summary {
		margin: 0 0 0.6rem;
		font-size: 0.9rem;
		color: var(--text);
	}

	.warning {
		margin: 0 0 1.1rem;
		font-size: 0.85rem;
		color: var(--text-muted);
		line-height: 1.5;
	}

	.error {
		margin: 0 0 1rem;
		padding-left: 0.75rem;
		border-left: 2px solid var(--red);
		color: var(--red);
		font-size: 0.85rem;
		line-height: 1.5;
	}

	.modal-actions {
		display: flex;
		gap: 0.75rem;
		justify-content: flex-end;
	}
</style>
