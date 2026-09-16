<script lang="ts">
	/**
	 * Mover dinero entre los saldos de una misma cuenta: de la cuenta principal a
	 * un bolsillo, o al revés.
	 *
	 * No es un retiro y un depósito anotados a mano. Las dos patas van en una
	 * sola transacción, así que el dinero nunca está en los dos sitios ni en
	 * ninguno, y como se compensan exactamente —mismo importe, mismo día, sin
	 * comisión— la rentabilidad del portafolio no se mueve: el dinero cambió de
	 * cajón, no entró ni salió.
	 *
	 * El portafolio no se pregunta: el dinero se mueve dentro del portafolio en
	 * que ya está. Moverlo a otro portafolio sería un retiro y un depósito, y
	 * esos sí cambian de dónde viene la rentabilidad.
	 */
	import { enhance } from '$app/forms';
	import Button from '$lib/ui/button.svelte';
	import Modal from '$lib/ui/modal.svelte';
	import DatePicker from '$lib/ui/date-picker.svelte';
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import { todayLocalDateString } from '$lib/shared/format/date';
	import type { CashPocket } from '$lib/api/types';
	import type { CashAccount } from '../cash';

	/** La cuenta cuyos saldos se mueven, y de qué cajón sale por defecto. */
	export interface CashMoveTarget {
		account: CashAccount;
		/** El portafolio en que se mueve; con varios, el del saldo mayor. */
		portfolioId: string;
		portfolioName: string;
	}

	interface Props {
		target: CashMoveTarget | null;
		/** Los bolsillos del usuario; la cuenta toma los suyos. */
		pockets: CashPocket[];
		onClose: () => void;
	}

	let { target, pockets, onClose }: Props = $props();

	/**
	 * Los cajones de la cuenta: la principal primero y luego sus bolsillos
	 * abiertos. Un bolsillo sin saldo también está, que es de donde sale un
	 * primer traslado.
	 */
	const drawers = $derived(
		target === null
			? []
			: [
					{ id: '', name: 'Cuenta principal' },
					...pockets
						.filter(
							(p) =>
								p.sourceId === target.account.sourceId &&
								p.currency === target.account.currency &&
								p.closedOn === null
						)
						.map((p) => ({ id: p.id, name: p.name }))
				]
	);

	let from = $derived(target?.account.pocketId ?? '');
	let to = $derived('');
	let amount = $state('');
	let date = $derived(todayLocalDateString());
	let notes = $state('');
	let submitting = $state(false);
	let error = $state('');

	/* El destino nunca es el origen: mover el dinero a donde ya está no hace
	   nada, y el backend lo rechaza igual. */
	const destinations = $derived(drawers.filter((d) => d.id !== from));

	function chooseOrigin(id: string) {
		from = id;
		if (to === id) to = '';
	}

	function close() {
		error = '';
		amount = '';
		notes = '';
		onClose();
	}

	/* Lo que guarda el cajón del que sale, para no pedir más de lo que hay. */
	const available = $derived.by(() => {
		if (!target) return 0;
		const drawer =
			from === ''
				? target.account
				: (target.account.pockets.find((p) => p.pocketId === from) ?? null);

		return (
			drawer?.balances
				.filter((b) => b.portfolioId === target.portfolioId)
				.reduce((sum, b) => sum + (parseFloat(b.balance) || 0), 0) ?? 0
		);
	});

	const money = (value: number) =>
		privacy.money(formatCurrency(value, target?.account.currency ?? 'USD'));
</script>

<Modal
	open={target !== null}
	title="Mover dinero"
	description={target
		? `${target.account.sourceName || 'Sin plataforma'} · ${target.account.currency} · ${target.portfolioName}`
		: ''}
	size="sm"
	onClose={close}
>
	{#if target}
		<form
			method="POST"
			action="?/move"
			class="rail-fields"
			use:enhance={() => {
				submitting = true;
				return async ({ result, update }) => {
					submitting = false;
					if (result.type === 'failure') {
						error = (result.data?.error as string) ?? 'No pudimos mover el dinero.';
						return;
					}
					await update();
					close();
				};
			}}
		>
			<input type="hidden" name="portfolioId" value={target.portfolioId} />
			<input type="hidden" name="sourceId" value={target.account.sourceId} />
			<input type="hidden" name="currency" value={target.account.currency} />

			<div class="pair">
				<div class="field">
					<label for="cash-move-from">De</label>
					<select id="cash-move-from" name="fromPocketId" bind:value={() => from, chooseOrigin}>
						{#each drawers as drawer (drawer.id)}
							<option value={drawer.id}>{drawer.name}</option>
						{/each}
					</select>
				</div>
				<div class="field">
					<label for="cash-move-to">A</label>
					<select id="cash-move-to" name="toPocketId" bind:value={to} required>
						<option value="" disabled>Elige el destino</option>
						{#each destinations as drawer (drawer.id)}
							<option value={drawer.id}>{drawer.name}</option>
						{/each}
					</select>
				</div>
			</div>

			<div class="field">
				<label for="cash-move-amount">Importe</label>
				<div class="with-unit">
					<input
						id="cash-move-amount"
						name="amount"
						type="number"
						inputmode="decimal"
						step="any"
						min="0"
						bind:value={amount}
						required
						aria-describedby="cash-move-amount-hint"
					/>
					<span class="unit">{target.account.currency}</span>
				</div>
				<p class="hint" id="cash-move-amount-hint">
					Ahí hay {money(available)}. El dinero sigue en la misma plataforma, así que tu
					rentabilidad no se mueve: solo cambia a qué tasa rinde.
				</p>
			</div>

			<div class="field">
				<span class="field-label">Fecha</span>
				<DatePicker name="date" bind:value={date} required />
			</div>

			<div class="field">
				<label for="cash-move-notes">Nota <span class="optional">(opcional)</span></label>
				<input
					id="cash-move-notes"
					name="notes"
					type="text"
					maxlength="500"
					placeholder="Para el viaje"
					bind:value={notes}
				/>
			</div>

			{#if error}
				<p class="feedback error">{error}</p>
			{/if}

			<div class="actions">
				<Button type="button" variant="ghost" onclick={close}>Cancelar</Button>
				<Button type="submit" disabled={submitting || to === ''}>Mover</Button>
			</div>
		</form>
	{/if}
</Modal>

<style>
	.pair {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 0.9rem;
	}

	.with-unit {
		position: relative;
		display: flex;
		align-items: center;
	}

	.with-unit input {
		width: 100%;
		padding-right: 3.4rem;
	}

	.unit {
		position: absolute;
		right: 0.85rem;
		font-size: 0.78rem;
		color: var(--text-dim);
		pointer-events: none;
	}

	.optional {
		color: var(--text-dim);
		font-weight: 400;
	}

	.actions {
		display: flex;
		justify-content: flex-end;
		gap: 0.6rem;
		margin-top: 0.4rem;
	}

	@media (max-width: 520px) {
		.pair {
			grid-template-columns: 1fr;
		}
	}
</style>
