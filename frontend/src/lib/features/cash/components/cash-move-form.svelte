<script lang="ts">
	/**
	 * Mover dinero entre dos saldos del portafolio: entre los cajones de una
	 * cuenta —de la principal a un bolsillo, o al revés— y entre dos
	 * plataformas, que es el traslado de la app donde está el ahorro al bróker
	 * que va a gastarlo.
	 *
	 * No es un retiro y un depósito anotados a mano. Las dos patas van en una
	 * sola transacción, así que el dinero nunca está en los dos sitios ni en
	 * ninguno, y como se compensan —mismo día, sin comisión, y el mismo valor
	 * una vez aplicada la tasa— la rentabilidad del portafolio no se mueve: el
	 * dinero cambió de sitio, no entró ni salió.
	 *
	 * Cruzando monedas se piden los dos importes: lo que salió y lo que llegó.
	 * No la tasa —0,00025 no es un número que nadie tenga a mano, mientras que
	 * los dos importes están los dos en el extracto—, y así lo que se guarda es
	 * exactamente lo que llegó, sin los centavos que sobran o faltan al
	 * calcularlo desde una tasa. La tasa se enseña debajo, ya despejada y en la
	 * dirección en que se piensa: cuántos pesos por cada dólar.
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
	import { SUPPORTED_CURRENCIES } from '$lib/shared/currency';
	import { OptimisticDialog } from '$lib/shared/optimistic.svelte';
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
		/** Los bolsillos del usuario; cada cuenta toma los suyos. */
		pockets: CashPocket[];
		/** Las plataformas activas, que son los destinos posibles. */
		platforms: { id: string; name: string }[];
		onClose: () => void;
	}

	let { target, pockets, platforms, onClose }: Props = $props();

	/**
	 * Los cajones de una cuenta: la principal primero y luego sus bolsillos
	 * abiertos. Un bolsillo sin saldo también está, que es de donde sale un
	 * primer traslado.
	 *
	 * Los depósitos a plazo quedan fuera. Están cerrados hasta que vencen y el
	 * backend rechaza cualquier movimiento a mano sobre ellos, así que ofrecerlos
	 * solo sirve para que el traslado falle al enviarlo.
	 */
	function drawersOf(sourceId: string, currency: string) {
		return [
			{ id: '', name: 'Cuenta principal' },
			...pockets
				.filter(
					(p) =>
						p.sourceId === sourceId &&
						p.currency === currency &&
						p.kind !== 'fixed' &&
						p.closedOn === null
				)
				.map((p) => ({ id: p.id, name: p.name }))
		];
	}

	let from = $derived(target?.account.pocketId ?? '');

	/* A dónde va: plataforma, moneda y cajón. Arranca en la cuenta de la que
	   sale, que es el traslado entre cajones de siempre. */
	let toSourceId = $derived(target?.account.sourceId ?? '');
	let toCurrency = $derived(target?.account.currency ?? '');
	let to = $state('');

	let amount = $state('');
	let toAmount = $state('');
	let date = $derived(todayLocalDateString());
	let notes = $state('');
	/* Se cierra al pulsar y guarda de fondo; si el servidor lo rechaza, vuelve
	   con lo escrito y el motivo. */
	const dialog = new OptimisticDialog(() => target);

	const origin = $derived(
		target === null ? [] : drawersOf(target.account.sourceId, target.account.currency)
	);

	/* El traslado sale de la cuenta en la que estaba: otra plataforma, otra
	   moneda, o las dos. Es lo que decide si hace falta tasa y qué se puede
	   intercambiar. */
	const crossing = $derived(
		target !== null &&
			(toSourceId !== target.account.sourceId || toCurrency !== target.account.currency)
	);

	/* El destino nunca es el origen, pero «el origen» es plataforma, moneda y
	   cajón a la vez: la cuenta principal de otro bróker es otro sitio aunque
	   las dos manden el cajón vacío. */
	const destinations = $derived.by(() => {
		const drawers = drawersOf(toSourceId, toCurrency);

		return crossing ? drawers : drawers.filter((d) => d.id !== from);
	});

	/*
	 * El destino elegido siempre es uno de los que se ofrecen.
	 *
	 * Antes el desplegable arrancaba en una opción vacía de relleno que valía lo
	 * mismo que «Cuenta principal», así que elegir la cuenta principal no
	 * cambiaba nada y el botón de mover seguía apagado: desde un bolsillo no se
	 * podía devolver el dinero a la cuenta. Ahora arranca en el primer destino
	 * real y no hay ningún valor ambiguo.
	 */
	$effect(() => {
		if (!destinations.some((d) => d.id === to)) {
			to = destinations[0]?.id ?? '';
		}
	});

	/* Dentro de una moneda llega lo mismo que sale, y un importe de llegada
	   colgado de un traslado anterior sería una conversión que no hubo. */
	$effect(() => {
		if (toCurrency === target?.account.currency) toAmount = '';
	});

	function chooseOrigin(id: string) {
		from = id;
		if (to === id && !crossing) to = '';
	}

	/* Darle la vuelta: lo que era el destino pasa a ser el origen. Solo dentro de
	   una cuenta —el origen lo fija la tarjeta desde la que se abrió esto, así
	   que un traslado a otra plataforma no se puede invertir aquí—. */
	const canSwap = $derived(!crossing && destinations.length > 0);

	function swap() {
		if (!canSwap) return;
		[from, to] = [to, from];
	}

	function close() {
		dialog.reset();
		amount = '';
		toAmount = '';
		notes = '';
		onClose();
	}

	const submit = dialog.submit({
		fallbackError: 'No pudimos mover el dinero.',
		onDone: close
	});

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

	/* La tasa que sale de los dos importes, para poder reconocerla: se enseña
	   como «1 USD = 4.060,91 COP», que es como se dice, y no como el 0,00025 que
	   habría que escribir al revés. */
	const leaves = $derived(parseFloat(amount) || 0);
	const arrives = $derived(parseFloat(toAmount) || 0);
	const impliedRate = $derived(leaves > 0 && arrives > 0 ? leaves / arrives : 0);

	const money = (value: number, code = target?.account.currency ?? 'USD') =>
		privacy.money(formatCurrency(value, code));

	const fromName = $derived(origin.find((d) => d.id === from)?.name ?? '');
	const toPlatformName = $derived(platforms.find((p) => p.id === toSourceId)?.name ?? '');
</script>

<Modal
	open={target !== null}
	hidden={dialog.hidden}
	title="Mover dinero"
	description={target
		? `Desde ${target.account.sourceName || 'Sin plataforma'} en ${target.account.currency}, dentro de ${target.portfolioName}.`
		: ''}
	size="sm"
	onClose={close}
>
	{#if target}
		<form method="POST" action="?/move" class="rail-fields" use:enhance={submit}>
			<input type="hidden" name="portfolioId" value={target.portfolioId} />
			<input type="hidden" name="sourceId" value={target.account.sourceId} />
			<input type="hidden" name="currency" value={target.account.currency} />
			<input type="hidden" name="toSourceId" value={toSourceId} />
			<input type="hidden" name="toCurrency" value={toCurrency} />

			<!-- De dónde sale y a dónde va, uno encima del otro como el trayecto que
			     son, con la vuelta a mano entre los dos. -->
			<div class="route">
				<div class="field">
					<label for="cash-move-from">De</label>
					<select id="cash-move-from" name="fromPocketId" bind:value={() => from, chooseOrigin}>
						{#each origin as drawer (drawer.id)}
							<option value={drawer.id}>{drawer.name}</option>
						{/each}
					</select>
				</div>
				<button
					type="button"
					class="swap"
					onclick={swap}
					disabled={!canSwap}
					aria-label="Intercambiar origen y destino"
				>
					<svg viewBox="0 0 16 16" aria-hidden="true">
						<path d="M5.5 2.5v11M3 11l2.5 2.5L8 11M10.5 13.5v-11M8 5l2.5-2.5L13 5" />
					</svg>
				</button>
				<div class="field">
					<label for="cash-move-to">A</label>
					<div class="where">
						<select id="cash-move-to-platform" bind:value={toSourceId} aria-label="Plataforma">
							{#each platforms as platform (platform.id)}
								<option value={platform.id}>{platform.name}</option>
							{/each}
						</select>
						<select id="cash-move-to-currency" bind:value={toCurrency} aria-label="Moneda">
							{#each SUPPORTED_CURRENCIES as code (code)}
								<option value={code}>{code}</option>
							{/each}
						</select>
					</div>
					<select id="cash-move-to" name="toPocketId" bind:value={to}>
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

			{#if toCurrency !== target.account.currency}
				<div class="field">
					<label for="cash-move-arrives">Llegan</label>
					<CashMoneyInput
						id="cash-move-arrives"
						name="toAmount"
						unit={toCurrency}
						size="lg"
						bind:value={toAmount}
						required
						aria-describedby="cash-move-arrives-hint"
					/>
					<p class="hint" id="cash-move-arrives-hint">
						{#if impliedRate > 0}
							Son <strong>1 {toCurrency} = {money(impliedRate, target.account.currency)}</strong>.
						{:else}
							Lo que de verdad entró en la otra cuenta, según tu extracto.
						{/if}
					</p>
				</div>
			{/if}

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
				{#if crossing}
					El dinero sigue dentro de {target.portfolioName}, así que tu rentabilidad no se mueve:
					solo cambia dónde está y a qué tasa rinde. Una vez en {toPlatformName ||
						'la otra plataforma'}, una compra ahí puede pagarse con él.
				{:else}
					El dinero sigue en la misma plataforma, así que tu rentabilidad no se mueve: solo cambia a
					qué tasa rinde.
				{/if}
			</p>

			{#if dialog.error}
				<p class="feedback error" role="alert">{dialog.error}</p>
			{/if}

			<div class="modal-actions">
				<Button type="button" variant="ghost" onclick={close}>Cancelar</Button>
				<Button type="submit" loading={dialog.submitting} disabled={destinations.length === 0}>
					Mover dinero
				</Button>
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

	/* Plataforma y moneda comparten renglón: juntas nombran la cuenta a la que
	   llega, y el cajón va debajo porque es una elección dentro de ella. */
	.where {
		display: grid;
		grid-template-columns: minmax(0, 1fr) auto;
		gap: 0.35rem;
		margin-bottom: 0.35rem;
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
