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
	 * El tipo va primero y como opciones a la vista, no en un desplegable, porque
	 * es la decisión que cambia la rentabilidad: la frase debajo dice cuál de los
	 * tres cuenta como ganancia.
	 */
	import { enhance } from '$app/forms';
	import Button from '$lib/ui/button.svelte';
	import Modal from '$lib/ui/modal.svelte';
	import DatePicker from '$lib/ui/date-picker.svelte';
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

	/** Lo que abre el formulario: un alta (quizá sobre un saldo) o una edición. */
	export type CashFormTarget =
		{ mode: 'create'; balance?: CashBalance } | { mode: 'edit'; movement: CashMovement };

	interface Props {
		target: CashFormTarget | null;
		portfolios: { id: string; name: string; isDefault: boolean }[];
		platforms: { id: string; name: string }[];
		/** Los saldos que ya hay, para proponer el portafolio de una cuenta existente. */
		balances: CashBalance[];
		/** Moneda que se propone para un saldo nuevo: la de la cuenta. */
		currency: string;
		onClose: () => void;
	}

	let { target, portfolios, platforms, balances, currency, onClose }: Props = $props();

	interface Fields {
		kind: CashKind;
		portfolioId: string;
		sourceId: string;
		currency: string;
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
				kind: m.kind === 'other' ? 'deposit' : m.kind,
				portfolioId: m.portfolioId,
				sourceId: m.sourceId,
				currency: m.currency,
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
	}

	const editing = $derived(target?.mode === 'edit' ? target.movement : null);
	const hint = $derived(CASH_KIND_OPTIONS.find((o) => o.value === kind)?.hint ?? '');

	/* Con un solo portafolio no hay nada que elegir ni que nombrar. */
	const choosePortfolio = $derived(portfolios.length > 1);
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

			<fieldset class="kinds">
				<legend class="field-label">Qué pasó</legend>
				<div class="options">
					{#each CASH_KIND_OPTIONS as option (option.value)}
						<label class="option" class:selected={kind === option.value}>
							<input type="radio" name="kind" value={option.value} bind:group={kind} />
							{option.label}
						</label>
					{/each}
				</div>
				<p class="hint">{hint}</p>
			</fieldset>

			<!-- La cuenta: plataforma y moneda van juntas porque juntas dicen dónde
			     está el dinero. Al editar ya está decidida y la nombra la cabecera. -->
			{#if !editing}
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
			{/if}

			<!--
				Los dos importes juntos, con la moneda escrita dentro. Al editar no hay
				selector que la diga, y como texto suelto en media columna no se
				alineaba con nada. La comisión se descuenta de la misma cuenta, así
				que lleva la misma moneda.

				Con intereses la comisión desaparece pero el importe no se ensancha:
				cambiar de tipo no mueve el campo que se está escribiendo.
			-->
			<div class="pair">
				<div class="field">
					<label for="cash-amount">Importe</label>
					<div class="with-unit">
						<input
							id="cash-amount"
							name="amount"
							type="number"
							inputmode="decimal"
							step="any"
							min="0"
							bind:value={amount}
							aria-describedby="cash-amount-unit"
							required
						/>
						<span class="unit" id="cash-amount-unit">{currencyCode}</span>
					</div>
				</div>
				{#if kind !== 'interest'}
					<div class="field">
						<label for="cash-fees">Comisión <span class="optional">(opcional)</span></label>
						<div class="with-unit">
							<input
								id="cash-fees"
								name="fees"
								type="number"
								inputmode="decimal"
								step="any"
								min="0"
								bind:value={fees}
								aria-describedby="cash-fees-unit"
							/>
							<span class="unit" id="cash-fees-unit">{currencyCode}</span>
						</div>
					</div>
				{/if}
			</div>

			<!-- Fila propia: día, mes y año no caben en media columna, y el año se
			     montaba sobre la comisión. -->
			<div class="field">
				<span class="field-label">Fecha</span>
				<DatePicker name="date" bind:value={date} required />
			</div>

			{#if !editing}
				{#if choosePortfolio}
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
						<p class="hint" id="cash-portfolio-hint">
							En cuál suma este dinero. No cambia dónde está: sigue en la plataforma.
						</p>
					</div>
				{:else}
					<input type="hidden" name="portfolioId" value={portfolioId} />
				{/if}
			{/if}

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
	.kinds {
		margin: 0;
		padding: 0;
		border: none;
	}

	.kinds legend {
		margin-bottom: 0.45rem;
		padding: 0;
	}

	.options {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 0.5rem;
		margin-bottom: 0.55rem;
	}

	.option {
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 0.65rem 0.5rem;
		border: 1px solid rgba(212, 145, 42, 0.2);
		border-radius: 8px;
		font-size: 0.88rem;
		color: var(--text-muted);
		cursor: pointer;
		transition:
			border-color 0.2s ease,
			color 0.2s ease;
	}

	.option:hover {
		border-color: rgba(212, 145, 42, 0.35);
	}

	/* El radio queda para el teclado y el lector; lo que se ve es la etiqueta. */
	.option input {
		position: absolute;
		opacity: 0;
		pointer-events: none;
	}

	.option:has(input:focus-visible) {
		outline: 2px solid var(--amber);
		outline-offset: 2px;
	}

	.option.selected {
		border-color: var(--amber);
		color: var(--text);
		background: rgba(212, 145, 42, 0.08);
	}

	.with-unit {
		position: relative;
	}

	/* Sitio a la derecha para la moneda. Sin flechas de número: con `step="any"`
	   suben de uno en uno, que en un importe no sirve, y tapaban la moneda. */
	.with-unit input {
		padding-right: 3.75rem;
		appearance: textfield;
		-moz-appearance: textfield;
	}

	.with-unit input::-webkit-inner-spin-button,
	.with-unit input::-webkit-outer-spin-button {
		-webkit-appearance: none;
		margin: 0;
	}

	.unit {
		position: absolute;
		top: 50%;
		right: 0.95rem;
		transform: translateY(-50%);
		font-family: var(--font-mono);
		font-size: 0.78rem;
		letter-spacing: 0.04em;
		color: var(--text-dim);
		pointer-events: none;
	}

	.feedback {
		margin: 0;
	}

	@media (max-width: 640px) {
		.pair {
			grid-template-columns: 1fr;
		}
	}
</style>
