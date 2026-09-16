<script lang="ts">
	/**
	 * Un depósito a tasa fija: abrirlo, y luego cancelarlo o borrarlo.
	 *
	 * Un CDT, una cajita a plazo, una tasa promocional amarrada noventa días. Es
	 * un bolsillo que conserva la tasa del día en que se abrió: lleva un solo
	 * depósito, una sola versión de su tasa y un vencimiento opcional. Abrirlo es
	 * `cash-deposit-open`; una vez abierto no se edita, así que lo que se ve aquí
	 * es su ficha —lo que vale, a qué tasa y cuánto le falta— y las dos salidas.
	 */
	import { enhance } from '$app/forms';
	import Button from '$lib/ui/button.svelte';
	import Modal from '$lib/ui/modal.svelte';
	import DatePicker from '$lib/ui/date-picker.svelte';
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import { formatCalendarDate, todayLocalDateString } from '$lib/shared/format/date';
	import type { CashPocket, CashRate } from '$lib/api/types';
	import { formatAnnualRate } from '../rates';
	import CashDepositOpen from './cash-deposit-open.svelte';
	import CashMoneyInput from './cash-money-input.svelte';
	import CashTermTrack from './cash-term-track.svelte';

	/** Abrir uno en una cuenta, o mirar el que ya está abierto. */
	export type CashDepositTarget =
		| {
				mode: 'create';
				sourceId: string;
				sourceName: string;
				currency: string;
				/** El portafolio que pone el dinero; un depósito es de uno solo. */
				portfolioId: string;
				portfolioName: string;
		  }
		| { mode: 'manage'; pocket: CashPocket; rate: CashRate | null; balance: number };

	interface Props {
		target: CashDepositTarget | null;
		onClose: () => void;
	}

	let { target, onClose }: Props = $props();

	const today = todayLocalDateString();

	const open = $derived(target?.mode === 'manage' ? target.pocket : null);

	const description = $derived.by(() => {
		if (target === null) return '';
		if (target.mode === 'create') {
			return `En ${target.sourceName || 'Sin plataforma'}, en ${target.currency}, con dinero de ${target.portfolioName}.`;
		}
		return `En ${target.pocket.sourceName || 'Sin plataforma'}, en ${target.pocket.currency}.`;
	});

	/* Cancelar: el día en que el dinero vuelve, y lo que cobra la entidad. */
	let closesOn = $state(today);
	let penalty = $state<string | number | null>('');
	let submitting = $state(false);
	let error = $state('');

	function close() {
		error = '';
		penalty = '';
		closesOn = today;
		onClose();
	}

	const money = (value: number, currency: string) => privacy.money(formatCurrency(value, currency));

	const longDay = (iso: string) =>
		formatCalendarDate(iso.slice(0, 10), { day: 'numeric', month: 'long', year: 'numeric' });
</script>

<Modal
	open={target !== null}
	title={open ? open.name : 'Nuevo depósito a plazo'}
	{description}
	size={open ? 'sm' : 'md'}
	onClose={close}
>
	{#if target?.mode === 'create'}
		<CashDepositOpen
			sourceId={target.sourceId}
			currency={target.currency}
			portfolioId={target.portfolioId}
			onClose={close}
		/>
	{:else if target?.mode === 'manage' && open}
		<form
			method="POST"
			action="?/closeDeposit"
			class="rail-fields"
			use:enhance={() => {
				submitting = true;
				return async ({ result, update }) => {
					submitting = false;
					if (result.type === 'failure') {
						error = (result.data?.error as string) ?? 'No pudimos guardar el depósito.';
						return;
					}
					await update();
					close();
				};
			}}
		>
			<input type="hidden" name="id" value={open.id} />

			<!-- La ficha: lo que vale hoy en grande, y debajo a qué tasa y cuánto
			     le queda, que es lo que se le pregunta a un depósito. -->
			<div class="sheet">
				<p class="worth">
					<span class="worth-label">Vale hoy</span>
					<span class="worth-amount">{money(target.balance, open.currency)}</span>
				</p>
				<dl class="terms">
					<div>
						<dt>Tasa fija</dt>
						<dd>{target.rate ? formatAnnualRate(target.rate.annualRatePct) : 'sin tasa'}</dd>
					</div>
					<div>
						<dt>Abierto el</dt>
						<dd>{longDay(open.openedOn)}</dd>
					</div>
					<div>
						<dt>Vence</dt>
						<dd>{open.maturesOn ? longDay(open.maturesOn) : 'sin plazo'}</dd>
					</div>
				</dl>
				{#if open.maturesOn && !open.closedOn}
					<CashTermTrack openedOn={open.openedOn} maturesOn={open.maturesOn} {today} />
				{/if}
			</div>

			<p class="note">
				Un depósito no admite depósitos, retiros ni intereses a mano, y su tasa es la del día en que
				se abrió. Al vencer, su saldo vuelve solo a la cuenta principal.
			</p>

			{#if open.closedOn}
				<p class="note">Se cerró el {longDay(open.closedOn)}.</p>
			{:else}
				<fieldset class="cancel">
					<legend class="field-label">Cancelarlo antes de tiempo</legend>
					<!-- Uno debajo del otro: día, mes y año no caben en media columna. -->
					<div class="field">
						<span class="field-label quiet">El día</span>
						<DatePicker name="closesOn" bind:value={closesOn} required />
					</div>
					<div class="field penalty">
						<label for="cash-deposit-penalty" class="quiet">
							Penalidad <span class="optional">(opcional)</span>
						</label>
						<CashMoneyInput
							id="cash-deposit-penalty"
							name="penalty"
							unit={open.currency}
							placeholder="0"
							bind:value={penalty}
						/>
					</div>
					<p class="hint">
						La tasa termina la víspera, se abona lo que llevaba ganado y el resto vuelve a la cuenta
						principal. Lo que se quede la entidad cuenta como pérdida, no como dinero que sacaste.
					</p>
				</fieldset>
			{/if}

			<!-- Borrar lee solo el id, así que sale de este mismo formulario con su
			     propia acción, sin pedir la fecha de cancelación. -->
			<div class="danger-zone">
				<p class="danger-note">
					Borrarlo lo quita entero —el depósito, sus intereses y su tasa—, como si nunca hubiera
					existido. Es para algo que anotaste mal; si el dinero sí estuvo ahí, cancélalo.
				</p>
				<button
					type="submit"
					class="danger-link"
					formaction="?/deletePocket"
					formnovalidate
					disabled={submitting}
				>
					Borrar depósito
				</button>
			</div>

			{#if error}
				<p class="feedback error">{error}</p>
			{/if}

			<div class="modal-actions">
				<Button type="button" variant="ghost" onclick={close}>Cerrar</Button>
				{#if !open.closedOn}
					<Button type="submit" loading={submitting}>Cancelar depósito</Button>
				{/if}
			</div>
		</form>
	{/if}
</Modal>

<style>
	.sheet {
		display: grid;
		gap: 1rem;
		padding: 1rem;
		border: 1px solid var(--border);
		border-radius: 10px;
		background: rgba(255, 255, 255, 0.02);
		--ring: #131416;
	}

	.worth {
		display: grid;
		gap: 0.2rem;
		margin: 0;
	}

	.worth-label {
		font-size: 0.78rem;
		color: var(--text-dim);
	}

	.worth-amount {
		font-family: var(--font-mono);
		font-size: 1.6rem;
		letter-spacing: -0.03em;
		color: var(--text);
	}

	.terms {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 0.75rem;
		margin: 0;
	}

	.terms div {
		display: grid;
		gap: 0.15rem;
		min-width: 0;
	}

	dt {
		font-size: 0.74rem;
		color: var(--text-dim);
	}

	dd {
		margin: 0;
		font-size: 0.84rem;
		color: var(--text);
	}

	.terms div:first-child dd {
		font-family: var(--font-mono);
		color: var(--green);
	}

	.note {
		margin: 0;
		font-size: 0.8rem;
		line-height: 1.55;
		color: var(--text-dim);
	}

	.cancel {
		display: grid;
		gap: 0.9rem;
		min-width: 0;
		margin: 0;
		padding: 0;
		border: none;
	}

	.cancel legend {
		margin-bottom: 0.9rem;
		padding: 0;
	}

	.quiet {
		font-weight: 400;
		color: var(--text-muted);
	}

	.penalty {
		max-width: 16rem;
	}

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

	/* Rojo solo en el texto y el borde: es la salida que no se deshace, no un
	   botón más del pie. */
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

	.feedback {
		margin: 0;
	}

	@media (max-width: 520px) {
		.terms {
			grid-template-columns: minmax(0, 1fr);
		}
	}
</style>
