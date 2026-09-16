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
	import CashMoneyInput from './cash-money-input.svelte';

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

	/* Darle la vuelta: lo que era el destino pasa a ser el origen. */
	function swap() {
		if (to === '') return;
		[from, to] = [to, from];
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

	const fromName = $derived(drawers.find((d) => d.id === from)?.name ?? '');
</script>

<Modal
	open={target !== null}
	title="Mover dinero"
	description={target
		? `Entre los cajones de ${target.account.sourceName || 'Sin plataforma'} en ${target.account.currency}, dentro de ${target.portfolioName}.`
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

			<!-- De dónde sale y a dónde va, uno encima del otro como el trayecto que
			     son, con la vuelta a mano entre los dos. -->
			<div class="route">
				<div class="field">
					<label for="cash-move-from">De</label>
					<select id="cash-move-from" name="fromPocketId" bind:value={() => from, chooseOrigin}>
						{#each drawers as drawer (drawer.id)}
							<option value={drawer.id}>{drawer.name}</option>
						{/each}
					</select>
				</div>
				<button
					type="button"
					class="swap"
					onclick={swap}
					disabled={to === ''}
					aria-label="Intercambiar origen y destino"
				>
					<svg viewBox="0 0 16 16" aria-hidden="true">
						<path d="M5.5 2.5v11M3 11l2.5 2.5L8 11M10.5 13.5v-11M8 5l2.5-2.5L13 5" />
					</svg>
				</button>
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
				<CashMoneyInput
					id="cash-move-amount"
					name="amount"
					unit={target.account.currency}
					size="lg"
					bind:value={amount}
					required
					aria-describedby="cash-move-amount-hint"
				/>
				<div class="available">
					<p class="hint" id="cash-move-amount-hint">
						En {fromName} hay <strong>{money(available)}</strong>.
					</p>
					{#if available > 0}
						<button type="button" class="all" onclick={() => (amount = String(available))}>
							Mover todo
						</button>
					{/if}
				</div>
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

			<p class="hint rule">
				El dinero sigue en la misma plataforma, así que tu rentabilidad no se mueve: solo cambia a
				qué tasa rinde.
			</p>

			{#if error}
				<p class="feedback error">{error}</p>
			{/if}

			<div class="modal-actions">
				<Button type="button" variant="ghost" onclick={close}>Cancelar</Button>
				<Button type="submit" loading={submitting} disabled={to === ''}>Mover dinero</Button>
			</div>
		</form>
	{/if}
</Modal>

<style>
	.route {
		display: grid;
		grid-template-columns: minmax(0, 1fr);
		gap: 0.35rem;
		padding: 1rem;
		border: 1px solid var(--border);
		border-radius: 10px;
		background: rgba(255, 255, 255, 0.02);
	}

	.swap {
		justify-self: center;
		display: grid;
		place-items: center;
		width: 2.25rem;
		height: 2.25rem;
		margin: 0.1rem 0 -0.3rem;
		padding: 0;
		border: 1px solid var(--border-strong);
		border-radius: 50%;
		background: var(--bg);
		color: var(--text-muted);
		cursor: pointer;
	}

	.swap:hover:not(:disabled) {
		border-color: var(--amber);
		color: var(--amber-light);
	}

	.swap:disabled {
		cursor: default;
		opacity: 0.45;
	}

	.swap svg {
		width: 0.95rem;
		height: 0.95rem;
		fill: none;
		stroke: currentColor;
		stroke-width: 1.5;
		stroke-linecap: round;
		stroke-linejoin: round;
	}

	.available {
		display: flex;
		flex-wrap: wrap;
		align-items: baseline;
		justify-content: space-between;
		gap: 0.25rem 1rem;
	}

	.available strong {
		font-family: var(--font-mono);
		font-weight: 400;
		color: var(--text);
	}

	.all {
		padding: 0.2rem 0.5rem;
		border: none;
		border-radius: 6px;
		background: transparent;
		font: inherit;
		font-size: 0.8rem;
		color: var(--amber);
		cursor: pointer;
	}

	.all:hover {
		background: rgba(212, 145, 42, 0.1);
		color: var(--amber-light);
	}

	.rule {
		font-size: 0.8rem;
		color: var(--text-dim);
	}

	.feedback {
		margin: 0;
	}
</style>
