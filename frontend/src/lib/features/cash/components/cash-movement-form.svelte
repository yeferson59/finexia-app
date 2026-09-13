<script lang="ts">
	/**
	 * Registrar o editar un movimiento de efectivo.
	 *
	 * Al registrar se elige sobre qué saldo cae —portafolio, plataforma y
	 * moneda—; si esa plataforma aún no guardaba esa moneda para ese portafolio,
	 * el primer depósito abre el saldo. Al editar, el saldo ya está decidido y
	 * solo se nombra: moverlo de cuenta es borrarlo y registrarlo de nuevo.
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
		/** Moneda que se propone para un saldo nuevo: la de la cuenta. */
		currency: string;
		onClose: () => void;
	}

	let { target, portfolios, platforms, currency, onClose }: Props = $props();

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
		return {
			kind: 'deposit',
			portfolioId:
				b?.portfolioId ?? (portfolios.find((p) => p.isDefault) ?? portfolios[0])?.id ?? '',
			sourceId: b?.sourceId ?? platforms[0]?.id ?? '',
			currency: b?.currency ?? currency,
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

	const editing = $derived(target?.mode === 'edit' ? target.movement : null);
	const hint = $derived(CASH_KIND_OPTIONS.find((o) => o.value === kind)?.hint ?? '');
</script>

<Modal
	open={target !== null}
	title={editing ? 'Editar movimiento' : 'Registrar movimiento'}
	description={editing
		? cashAccountLabel(editing)
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

			{#if !editing}
				<div class="pair">
					<div class="field">
						<label for="cash-source">Plataforma</label>
						<select id="cash-source" name="sourceId" bind:value={sourceId} required>
							{#each platforms as platform (platform.id)}
								<option value={platform.id}>{platform.name}</option>
							{/each}
						</select>
					</div>
					<div class="field">
						<label for="cash-portfolio">Portafolio</label>
						<select id="cash-portfolio" name="portfolioId" bind:value={portfolioId} required>
							{#each portfolios as portfolio (portfolio.id)}
								<option value={portfolio.id}>{portfolio.name}</option>
							{/each}
						</select>
					</div>
				</div>
			{/if}

			<div class="pair">
				<div class="field">
					<label for="cash-amount">Importe</label>
					<input
						id="cash-amount"
						name="amount"
						type="number"
						inputmode="decimal"
						step="any"
						min="0"
						bind:value={amount}
						required
					/>
				</div>
				{#if editing}
					<div class="field">
						<span class="field-label">Moneda</span>
						<p class="fixed">{editing.currency}</p>
					</div>
				{:else}
					<div class="field">
						<label for="cash-currency">Moneda</label>
						<select id="cash-currency" name="currency" bind:value={currencyCode} required>
							{#each SUPPORTED_CURRENCIES as code (code)}
								<option value={code}>{code}</option>
							{/each}
						</select>
					</div>
				{/if}
			</div>

			<div class="pair">
				<div class="field">
					<span class="field-label">Fecha</span>
					<DatePicker name="date" bind:value={date} required />
				</div>
				{#if kind !== 'interest'}
					<div class="field">
						<label for="cash-fees">Comisión <span class="optional">(opcional)</span></label>
						<input
							id="cash-fees"
							name="fees"
							type="number"
							inputmode="decimal"
							step="any"
							min="0"
							bind:value={fees}
						/>
					</div>
				{/if}
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

	.fixed {
		margin: 0;
		padding: 0.8rem 0;
		font-family: var(--font-mono);
		font-size: 0.9rem;
		color: var(--text-muted);
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
