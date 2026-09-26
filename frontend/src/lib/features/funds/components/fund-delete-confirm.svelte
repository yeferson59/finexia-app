<script lang="ts">
	/**
	 * Eliminar un fondo.
	 *
	 * Un fondo que ya nadie guarda se deja de seguir sin más: se van sus valores
	 * anotados. Uno que todavía está en algún portafolio se lleva también sus
	 * posiciones, con todas sus compras, ventas y el efectivo que movieron, y eso
	 * no se deshace: por eso lo dice con cifras y pide marcar una casilla.
	 */
	import { enhance } from '$app/forms';
	import Button from '$lib/ui/button.svelte';
	import Modal from '$lib/ui/modal.svelte';
	import { OptimisticDialog } from '$lib/shared/optimistic.svelte';
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import type { Fund } from '../funds';

	interface Props {
		fund: Fund | null;
		/** Cuántos aportes y retiros tiene, si se conocen. */
		movements: number;
		onClose: () => void;
	}

	let { fund, movements, onClose }: Props = $props();

	let confirmed = $state(false);

	const dialog = new OptimisticDialog(() => fund);

	function close() {
		dialog.reset();
		confirmed = false;
		onClose();
	}

	const handler = dialog.submit({
		fallbackError: 'No pudimos eliminar el fondo.',
		onDone: close
	});

	const held = $derived((fund?.positions.length ?? 0) > 0);
	/* Lo que se pierde, contado: «su posición, 4 movimientos y 12 valores anotados». */
	const lost = $derived.by(() => {
		if (!fund) return '';
		const n = fund.positions.length;
		const parts = [n === 1 ? 'su posición' : `sus ${n} posiciones`];
		if (movements > 0) parts.push(`${movements} ${movements === 1 ? 'movimiento' : 'movimientos'}`);
		if (fund.marks > 0)
			parts.push(`${fund.marks} ${fund.marks === 1 ? 'valor anotado' : 'valores anotados'}`);
		return parts.length === 1 ? parts[0] : `${parts.slice(0, -1).join(', ')} y ${parts.at(-1)}`;
	});

	const portfolios = $derived(
		fund ? [...new Set(fund.positions.map((p) => p.portfolioName))].join(', ') : ''
	);
</script>

<Modal
	open={fund !== null}
	hidden={dialog.hidden}
	title={fund ? `Eliminar ${fund.name}` : 'Eliminar fondo'}
	description={held
		? 'El fondo y todo su historial desaparecen de Finexia.'
		: 'Ningún portafolio guarda este fondo. Sus valores anotados se borran con él.'}
	size="sm"
	tone="danger"
	onClose={close}
>
	{#if fund}
		<form method="POST" action="?/deleteFund" use:enhance={handler}>
			<input type="hidden" name="id" value={fund.assetId} />

			{#if held}
				<div class="receipt">
					<p>
						Vale hoy <strong
							>{privacy.money(formatCurrency(parseFloat(fund.value) || 0, fund.currency))}</strong
						>
						y está en {portfolios}.
					</p>
					<p>
						Se borran {lost}. Si algún retiro abonó efectivo o algún aporte salió de él, ese
						efectivo vuelve a como estaba.
					</p>
					<p>
						Si solo sacaste el dinero, no lo elimines: anota un retiro con «Retirar todo» y la
						ganancia queda en tu historial.
					</p>
				</div>

				<label class="check">
					<input type="checkbox" name="withPositions" bind:checked={confirmed} />
					<span>Entiendo que se borran también sus posiciones y su historial.</span>
				</label>
			{/if}

			{#if dialog.error}
				<p class="feedback error" role="alert">{dialog.error}</p>
			{/if}

			<div class="modal-actions">
				<Button type="button" variant="ghost" onclick={close}>Cancelar</Button>
				<Button
					type="submit"
					variant="danger"
					loading={dialog.submitting}
					disabled={held && !confirmed}
				>
					Eliminar fondo
				</Button>
			</div>
		</form>
	{/if}
</Modal>

<style>
	.receipt {
		display: grid;
		gap: 0.6rem;
		margin-bottom: 1rem;
		font-size: 0.87rem;
		line-height: 1.5;
		color: var(--text-muted);
	}

	.receipt p {
		margin: 0;
	}

	.receipt strong {
		color: var(--text);
		font-family: var(--font-mono);
		font-weight: 500;
	}

	.check {
		display: flex;
		align-items: flex-start;
		gap: 0.6rem;
		margin-bottom: 0.5rem;
		font-size: 0.87rem;
		color: var(--text);
		cursor: pointer;
	}

	.check input {
		margin-top: 0.2rem;
		accent-color: var(--red);
	}

	.feedback {
		margin: 0 0 0.5rem;
	}
</style>
