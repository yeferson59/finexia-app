<script lang="ts">
	/**
	 * Abrir, renombrar y borrar un bolsillo de una cuenta.
	 *
	 * Un bolsillo es una subcuenta de la cuenta —la «cajita» del banco, el
	 * subsaldo del bróker—, no otra plataforma: su dinero sigue contando en la
	 * plataforma y en sus cifras. Lo propio del bolsillo es su tasa, que se le da
	 * desde su línea, como a cualquier cuenta.
	 *
	 * Lo único que se cambia de un bolsillo es el nombre. Lo que guarda se mueve
	 * con movimientos, y cuándo rinde se mueve con su tasa.
	 */
	import { enhance } from '$app/forms';
	import Button from '$lib/ui/button.svelte';
	import Modal from '$lib/ui/modal.svelte';
	import type { CashPocket } from '$lib/api/types';

	/** Abrir uno en una cuenta, o editar el que ya existe. */
	export type CashPocketTarget =
		| { mode: 'create'; sourceId: string; sourceName: string; currency: string }
		| { mode: 'edit'; pocket: CashPocket };

	interface Props {
		target: CashPocketTarget | null;
		onClose: () => void;
	}

	let { target, onClose }: Props = $props();

	const editing = $derived(target?.mode === 'edit' ? target.pocket : null);

	/* Como en el resto de los formularios: `$derived` reasignable, así que lo que
	   se escribe pisa el valor inicial hasta la siguiente apertura. */
	let name = $derived(editing?.name ?? '');
	let submitting = $state(false);
	let error = $state('');

	function close() {
		error = '';
		onClose();
	}

	const account = $derived(
		target === null
			? ''
			: target.mode === 'edit'
				? `Dentro de ${target.pocket.sourceName || 'Sin plataforma'}, en ${target.pocket.currency}.`
				: `Dentro de ${target.sourceName || 'Sin plataforma'}, en ${target.currency}.`
	);

	/* Un bolsillo con movimientos no se borra: se perderían. Vaciarlo no basta,
	   porque su historia es lo que se llevaría por delante. */
	const canDelete = $derived(editing !== null && editing.movements === 0);

	function handler() {
		submitting = true;
		return async ({
			result,
			update
		}: {
			result: { type: string; data?: Record<string, unknown> };
			update: () => Promise<void>;
		}) => {
			submitting = false;
			if (result.type === 'failure') {
				error = (result.data?.error as string) ?? 'No pudimos guardar el bolsillo.';
				return;
			}
			await update();
			close();
		};
	}
</script>

<Modal
	open={target !== null}
	title={editing ? 'Editar bolsillo' : 'Nuevo bolsillo'}
	description={account}
	size="sm"
	onClose={close}
>
	{#if target}
		<form
			method="POST"
			action={editing ? '?/renamePocket' : '?/createPocket'}
			class="rail-fields"
			use:enhance={handler}
		>
			{#if target.mode === 'edit'}
				<input type="hidden" name="id" value={target.pocket.id} />
			{:else}
				<input type="hidden" name="sourceId" value={target.sourceId} />
				<input type="hidden" name="currency" value={target.currency} />
			{/if}

			<div class="field">
				<label for="cash-pocket-name">Nombre</label>
				<input
					id="cash-pocket-name"
					name="name"
					type="text"
					maxlength="100"
					placeholder="Viajes"
					bind:value={name}
					required
					aria-describedby="cash-pocket-name-hint"
				/>
				<p class="hint" id="cash-pocket-name-hint">
					Como lo llama tu entidad: «Cajita», «Bolsillo», «Meta». Su dinero sigue sumando en la
					plataforma.
				</p>
			</div>

			<!-- Borrar sale del mismo formulario, con su propia acción: lo único que
			     lee es el id. Va encima del pie para no quedar debajo de él. -->
			{#if editing}
				<div class="danger-zone">
					{#if canDelete}
						<p class="danger-note">
							Todavía no tiene movimientos, así que se puede quitar sin perder nada.
						</p>
						<button
							type="submit"
							class="danger-link"
							formaction="?/deletePocket"
							formnovalidate
							disabled={submitting}
						>
							Borrar bolsillo
						</button>
					{:else}
						<p class="danger-note">
							Tiene {editing.movements}
							{editing.movements === 1 ? 'movimiento' : 'movimientos'}, así que no se puede borrar:
							su historia se iría con él. Bórralos primero si lo anotaste por error.
						</p>
					{/if}
				</div>
			{/if}

			{#if error}
				<p class="feedback error">{error}</p>
			{/if}

			<div class="modal-actions">
				<Button type="button" variant="ghost" onclick={close}>Cancelar</Button>
				<Button type="submit" loading={submitting}>
					{editing ? 'Guardar nombre' : 'Crear bolsillo'}
				</Button>
			</div>
		</form>
	{/if}
</Modal>

<style>
	.danger-zone {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		justify-content: space-between;
		gap: 0.5rem 1rem;
		padding-top: 1rem;
		border-top: 1px solid var(--border);
	}

	.danger-note {
		flex: 1 1 16rem;
		margin: 0;
		font-size: 0.8rem;
		line-height: 1.5;
		color: var(--text-dim);
	}

	/* Rojo solo en el texto: es la salida que no se deshace, no un botón más. */
	.danger-link {
		padding: 0.4rem 0.7rem;
		border: 1px solid rgba(224, 90, 90, 0.35);
		border-radius: 7px;
		background: transparent;
		font: inherit;
		font-size: 0.82rem;
		color: var(--red);
		cursor: pointer;
	}

	.danger-link:hover:not(:disabled) {
		border-color: var(--red);
		background: rgba(224, 90, 90, 0.08);
	}

	.danger-link:disabled {
		cursor: default;
		opacity: 0.5;
	}
</style>
