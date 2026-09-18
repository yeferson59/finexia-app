<script lang="ts">
	/**
	 * Registrar o editar un movimiento de efectivo.
	 *
	 * Al registrar se elige la cuenta —plataforma y moneda—, que es donde está el
	 * dinero; si esa plataforma aún no guardaba esa moneda, el primer depósito
	 * abre la cuenta. Al editar, la cuenta ya está decidida y solo se nombra:
	 * moverlo de cuenta es borrarlo y registrarlo de nuevo.
	 *
	 * El portafolio es secundario: solo decide dónde cuenta el dinero. Con uno
	 * solo no se pregunta. Con varios va al final, y elegir la cuenta propone el
	 * portafolio donde ya suma, para que un depósito no la parta en dos.
	 *
	 * El bolsillo solo se pregunta cuando la cuenta tiene alguno, y por defecto
	 * es la principal: una cuenta sin cajitas se comporta como siempre.
	 *
	 * El tipo va primero y como opciones a la vista, no en un desplegable, porque
	 * es la decisión que cambia la rentabilidad: la frase debajo dice cuál de los
	 * tres cuenta como ganancia.
	 */
	import { enhance } from '$app/forms';
	import Button from '$lib/ui/button.svelte';
	import Modal from '$lib/ui/modal.svelte';
	import DatePicker from '$lib/ui/date-picker.svelte';
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import { SUPPORTED_CURRENCIES } from '$lib/shared/currency';
	import { todayLocalDateString } from '$lib/shared/format/date';
	import {
		CASH_KIND_OPTIONS,
		cashAccountLabel,
		suggestCashPortfolio,
		type CashBalance,
		type CashKind,
		type CashMovement
	} from '../cash';
	import type { CashPocket } from '$lib/api/types';
	import CashChoice from './cash-choice.svelte';
	import CashMoneyInput from './cash-money-input.svelte';

	/** Lo que abre el formulario: un alta (quizá sobre un saldo) o una edición. */
	export type CashFormTarget =
		{ mode: 'create'; balance?: CashBalance } | { mode: 'edit'; movement: CashMovement };

	interface Props {
		target: CashFormTarget | null;
		portfolios: { id: string; name: string; isDefault: boolean }[];
		platforms: { id: string; name: string }[];
		/** Los saldos que ya hay, para proponer el portafolio de una cuenta existente. */
		balances: CashBalance[];
		/** Los bolsillos del usuario; la cuenta elegida toma los suyos. */
		pockets: CashPocket[];
		/** Moneda que se propone para un saldo nuevo: la de la cuenta. */
		currency: string;
		onClose: () => void;
	}

	let { target, portfolios, platforms, balances, pockets, currency, onClose }: Props = $props();

	interface Fields {
		kind: CashKind;
		portfolioId: string;
		sourceId: string;
		currency: string;
		/** Vacío es la cuenta principal. */
		pocketId: string;
		amount: string;
		fees: string;
		date: string;
		notes: string;
	}

	/** Cómo arranca el formulario: vacío, sobre un saldo o con el movimiento que se edita. */
	function initialFields(current: CashFormTarget | null): Fields {
		if (current?.mode === 'edit') {
			const m = current.movement;
			return {
				// Ni «otro» ni un dividendo llegan aquí —no son editables—; el
				// valor solo tiene que ser uno que este formulario sepa escribir.
				kind: m.kind === 'other' || m.kind === 'dividend' ? 'deposit' : m.kind,
				portfolioId: m.portfolioId,
				sourceId: m.sourceId,
				currency: m.currency,
				// Una edición se queda en el saldo en que ya está: cambiar de bolsillo
				// es mover, y mover son dos movimientos.
				pocketId: '',
				amount: String(parseFloat(m.amount) || ''),
				fees: parseFloat(m.fees) > 0 ? String(parseFloat(m.fees)) : '',
				date: m.date.slice(0, 10),
				notes: m.notes
			};
		}

		const b = current?.balance;
		const source = b?.sourceId ?? platforms[0]?.id ?? '';
		const code = b?.currency ?? currency;
		const fallback = (portfolios.find((p) => p.isDefault) ?? portfolios[0])?.id ?? '';
		return {
			kind: 'deposit',
			portfolioId: b?.portfolioId ?? suggestCashPortfolio(balances, source, code, fallback),
			sourceId: source,
			currency: code,
			pocketId: b?.pocketId ?? '',
			amount: '',
			fees: '',
			date: todayLocalDateString(),
			notes: ''
		};
	}

	/*
	 * Cada campo sale de lo que abre el formulario y se recalcula cuando se abre
	 * con otro. Son `$derived` reasignables: lo que escribe el usuario los pisa
	 * hasta la siguiente apertura, sin un `$effect` copiando valores a mano.
	 */
	const initial = $derived(initialFields(target));
	let kind = $derived(initial.kind);
	let portfolioId = $derived(initial.portfolioId);
	let sourceId = $derived(initial.sourceId);
	let currencyCode = $derived(initial.currency);
	let pocketId = $derived(initial.pocketId);
	let amount = $derived(initial.amount);
	let fees = $derived(initial.fees);
	let date = $derived(initial.date);
	let notes = $derived(initial.notes);
	let submitting = $state(false);
	let error = $state('');

	/* Cerrar limpia el error: la siguiente apertura no lo arrastra. */
	function close() {
		error = '';
		onClose();
	}

	/*
	 * Cambiar la cuenta propone dónde cuenta: si esa plataforma ya guarda esa
	 * moneda, el portafolio de ese saldo. Si no, se queda el que estaba.
	 */
	function chooseAccount(source: string, code: string) {
		sourceId = source;
		currencyCode = code;
		portfolioId = suggestCashPortfolio(balances, source, code, portfolioId);
		// Otra cuenta, otros bolsillos: el que estaba elegido ya no es de esta.
		pocketId = '';
	}

	/* Los bolsillos abiertos de la cuenta elegida. Sin ninguno no se pregunta. */
	const accountPockets = $derived(
		pockets.filter(
			(p) => p.sourceId === sourceId && p.currency === currencyCode && p.closedOn === null
		)
	);

	const editing = $derived(target?.mode === 'edit' ? target.movement : null);
	const hint = $derived(CASH_KIND_OPTIONS.find((o) => o.value === kind)?.hint ?? '');

	/* Con un solo portafolio no hay nada que elegir ni que nombrar. */
	const choosePortfolio = $derived(portfolios.length > 1);

	/* Lo que dice cada tecla bajo su nombre: hacia dónde va el dinero. */
	const KIND_DETAILS: Record<CashKind, string> = {
		deposit: 'Entra en la cuenta',
		withdrawal: 'Sale de la cuenta',
		interest: 'Lo que rindió'
	};

	const kinds = CASH_KIND_OPTIONS.map((o) => ({
		value: o.value,
		label: o.label,
		detail: KIND_DETAILS[o.value]
	}));

	/*
	 * Lo que guarda ahora el saldo elegido —plataforma, moneda y cajón—, para
	 * anotar un retiro sabiendo cuánto hay. `null` si todavía no existe: el
	 * primer depósito la abre.
	 */
	const held = $derived.by(() => {
		const rows = balances.filter(
			(b) =>
				b.sourceId === sourceId && b.currency === currencyCode && (b.pocketId ?? '') === pocketId
		);
		return rows.length === 0
			? null
			: rows.reduce((sum, b) => sum + (parseFloat(b.balance) || 0), 0);
	});
</script>

<Modal
	open={target !== null}
	title={editing ? 'Editar movimiento' : 'Registrar movimiento'}
	description={editing
		? cashAccountLabel(editing, choosePortfolio)
		: 'Anota el dinero que entra o sale de una cuenta.'}
	size="md"
	onClose={close}
>
	{#if target}
		<form
			method="POST"
			action={editing ? '?/update' : '?/create'}
			class="rail-fields"
			use:enhance={() => {
				submitting = true;
				return async ({ result, update }) => {
					submitting = false;
					if (result.type === 'failure') {
						error = (result.data?.error as string) ?? 'No pudimos guardar el movimiento.';
						return;
					}
					await update();
					close();
				};
			}}
		>
			{#if editing}
				<input type="hidden" name="id" value={editing.id} />
			{/if}

			<div class="field">
				<CashChoice
					legend="Qué pasó"
					name="kind"
					options={kinds}
					bind:value={kind}
					describedby="cash-kind-hint"
				/>
				<p class="hint" id="cash-kind-hint">{hint}</p>
			</div>

			<!-- La cuenta: plataforma y moneda van juntas porque juntas dicen dónde
			     está el dinero. Al editar ya está decidida y la nombra la cabecera. -->
			{#if !editing}
				<div class="account">
					<div class="pair">
						<div class="field">
							<label for="cash-source">Plataforma</label>
							<select
								id="cash-source"
								name="sourceId"
								bind:value={() => sourceId, (source) => chooseAccount(source, currencyCode)}
								required
							>
								{#each platforms as platform (platform.id)}
									<option value={platform.id}>{platform.name}</option>
								{/each}
							</select>
						</div>
						<div class="field">
							<label for="cash-currency">Moneda</label>
							<select
								id="cash-currency"
								name="currency"
								bind:value={() => currencyCode, (code) => chooseAccount(sourceId, code)}
								required
							>
								{#each SUPPORTED_CURRENCIES as code (code)}
									<option value={code}>{code}</option>
								{/each}
							</select>
						</div>
					</div>

					<!-- Solo cuando la cuenta tiene cajitas. Por defecto la principal, que
					     es donde cae todo mientras no haya ninguna. -->
					{#if accountPockets.length > 0}
						<div class="field">
							<label for="cash-pocket">Bolsillo</label>
							<select id="cash-pocket" name="pocketId" bind:value={pocketId}>
								<option value="">Cuenta principal</option>
								{#each accountPockets as pocket (pocket.id)}
									<option value={pocket.id}>{pocket.name}</option>
								{/each}
							</select>
						</div>
					{/if}

					<p class="held" aria-live="polite">
						{#if held === null}
							Cuenta nueva: este movimiento la abre.
						{:else}
							Ahora guarda <strong>{privacy.money(formatCurrency(held, currencyCode))}</strong>
						{/if}
					</p>
				</div>
			{/if}

			<!--
				El importe es la cifra que manda: a lo ancho y en la letra de las
				cifras. La comisión se descuenta de la misma cuenta, así que lleva la
				misma moneda; con intereses desaparece, porque se anotan netos.
			-->
			<div class="field">
				<label for="cash-amount">Importe</label>
				<CashMoneyInput
					id="cash-amount"
					name="amount"
					unit={currencyCode}
					size="lg"
					bind:value={amount}
					required
				/>
			</div>

			{#if kind !== 'interest' || (!editing && choosePortfolio)}
				<div class="pair">
					{#if kind !== 'interest'}
						<div class="field">
							<label for="cash-fees">Comisión <span class="optional">(opcional)</span></label>
							<CashMoneyInput id="cash-fees" name="fees" unit={currencyCode} bind:value={fees} />
						</div>
					{/if}
					{#if !editing && choosePortfolio}
						<div class="field">
							<label for="cash-portfolio">Portafolio</label>
							<select
								id="cash-portfolio"
								name="portfolioId"
								bind:value={portfolioId}
								aria-describedby="cash-portfolio-hint"
								required
							>
								{#each portfolios as portfolio (portfolio.id)}
									<option value={portfolio.id}>{portfolio.name}</option>
								{/each}
							</select>
						</div>
					{/if}
				</div>
			{/if}
			{#if !editing && choosePortfolio}
				<p class="hint portfolio-hint" id="cash-portfolio-hint">
					El portafolio dice en cuál suma este dinero. No cambia dónde está: sigue en la plataforma.
				</p>
			{:else if !editing}
				<input type="hidden" name="portfolioId" value={portfolioId} />
			{/if}

			<!-- Fila propia: día, mes y año no caben en media columna. -->
			<div class="field">
				<span class="field-label">Fecha</span>
				<DatePicker name="date" bind:value={date} required />
			</div>

			<div class="field">
				<label for="cash-notes">Nota <span class="optional">(opcional)</span></label>
				<textarea id="cash-notes" name="notes" rows="2" maxlength="500" bind:value={notes}
				></textarea>
			</div>

			{#if error}
				<p class="feedback error">{error}</p>
			{/if}

			<div class="modal-actions">
				<Button type="button" variant="ghost" onclick={close} disabled={submitting}>Cancelar</Button
				>
				<Button type="submit" loading={submitting}>
					{editing ? 'Guardar cambios' : 'Guardar movimiento'}
				</Button>
			</div>
		</form>
	{/if}
</Modal>

<style>
	/* La cuenta elegida, con lo que guarda, como un bloque: plataforma, moneda y
	   cajón son una sola respuesta a «dónde». */
	.account {
		display: grid;
		gap: 0.9rem;
		padding: 1rem;
		border: 1px solid var(--border);
		border-radius: 10px;
		background: rgba(255, 255, 255, 0.02);
	}

	.held {
		margin: 0;
		font-size: 0.82rem;
		color: var(--text-muted);
	}

	.held strong {
		font-family: var(--font-mono);
		font-weight: 400;
		color: var(--text);
	}

	.portfolio-hint {
		margin-top: -0.85rem;
	}

	.feedback {
		margin: 0;
	}
</style>
