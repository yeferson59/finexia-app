<script lang="ts">
	/**
	 * Un depósito a tasa fija: abrirlo, y luego cancelarlo o borrarlo.
	 *
	 * Un CDT, una cajita a plazo, una tasa promocional amarrada noventa días. Es
	 * un bolsillo que conserva la tasa del día en que se abrió: lleva un solo
	 * depósito, una sola versión de su tasa y un vencimiento opcional. Por eso el
	 * formulario pide el dinero, el plazo y la tasa a la vez —son una misma
	 * cosa— y por eso, una vez abierto, no se edita: se cancela.
	 *
	 * La fecha de apertura puede estar en el pasado. En un depósito no se anotan
	 * intereses a mano, así que no hay nada que contar dos veces: al guardarlo,
	 * Finexia calcula los días que ya ganó y los abona como rendimiento.
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
	import { addCalendarDays, daysBetween, projectInterestOverDays } from '../deposits';

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

	const open = $derived(target?.mode === 'manage' ? target.pocket : null);

	const currency = $derived(
		target === null ? 'USD' : target.mode === 'manage' ? target.pocket.currency : target.currency
	);

	const account = $derived(
		target === null
			? ''
			: target.mode === 'manage'
				? `${target.pocket.sourceName || 'Sin plataforma'} · ${currency}`
				: `${target.sourceName || 'Sin plataforma'} · ${currency} · ${target.portfolioName}`
	);

	/*
	 * Los plazos que cotizan las entidades, y las dos salidas: una fecha
	 * cualquiera, y un depósito que no vence. `$derived` reasignable, como en el
	 * resto de los formularios: lo que se elige pisa el valor inicial hasta la
	 * siguiente apertura.
	 */
	const TERMS = [30, 60, 90, 180, 360];

	let term = $derived<string>('90');
	let openedOn = $derived(todayLocalDateString());
	let customMaturity = $derived(addCalendarDays(todayLocalDateString(), 90));
	let amount = $state('');
	let name = $state('');
	let annualRatePct = $state('');
	let withholdingPct = $state('');
	let posting = $derived<'daily' | 'at_maturity'>('daily');

	/* El día en que vence, según el plazo elegido; vacío si no tiene. */
	const maturesOn = $derived(
		term === 'none'
			? ''
			: term === 'custom'
				? customMaturity
				: addCalendarDays(openedOn, Number(term))
	);

	const termDays = $derived(maturesOn ? daysBetween(openedOn, maturesOn) : 0);

	/* Los días que ya pasaron: lo que Finexia abona en cuanto se guarda. */
	const elapsed = $derived(Math.max(0, daysBetween(openedOn, todayLocalDateString())));

	const money = (value: number) => privacy.money(formatCurrency(value, currency));

	const principal = $derived(parseFloat(amount) || 0);
	const rate = $derived(parseFloat(annualRatePct) || 0);
	const withheld = $derived(parseFloat(withholdingPct) || 0);

	const atMaturity = $derived(projectInterestOverDays(principal, rate, termDays, withheld));
	const alreadyEarned = $derived(projectInterestOverDays(principal, rate, elapsed, withheld));

	const day = (iso: string) =>
		formatCalendarDate(iso.slice(0, 10), { day: 'numeric', month: 'short', year: 'numeric' });

	/* Cancelar: el día en que el dinero vuelve, y lo que cobra la entidad. */
	let closesOn = $derived(todayLocalDateString());
	let penalty = $state('');

	let submitting = $state(false);
	let error = $state('');

	/* Lo escrito se va con la ventana: la siguiente que se abra parte en blanco. */
	function close() {
		error = '';
		amount = name = annualRatePct = withholdingPct = penalty = '';
		onClose();
	}

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
				error = (result.data?.error as string) ?? 'No pudimos guardar el depósito.';
				return;
			}
			await update();
			close();
		};
	}
</script>

<Modal
	open={target !== null}
	title={open ? open.name : 'Nuevo depósito a tasa fija'}
	description={account}
	size="sm"
	onClose={close}
>
	{#if target?.mode === 'create'}
		<form method="POST" action="?/openDeposit" class="rail-fields" use:enhance={handler}>
			<input type="hidden" name="portfolioId" value={target.portfolioId} />
			<input type="hidden" name="sourceId" value={target.sourceId} />
			<input type="hidden" name="currency" value={target.currency} />
			<!-- El día del vencimiento viaja en un solo campo. Con un plazo en días
			     sale de la cuenta y no se escribe; con «Otra fecha» lo pone el
			     selector de abajo, que lleva ese mismo nombre. -->
			{#if term !== 'custom'}
				<input type="hidden" name="maturesOn" value={maturesOn} />
			{/if}

			<div class="field">
				<label for="cash-deposit-name">Nombre</label>
				<input
					id="cash-deposit-name"
					name="name"
					type="text"
					maxlength="100"
					placeholder="CDT 90 días"
					bind:value={name}
					required
				/>
			</div>

			<div class="field">
				<label for="cash-deposit-amount">Importe</label>
				<div class="with-unit">
					<input
						id="cash-deposit-amount"
						name="amount"
						type="number"
						inputmode="decimal"
						step="any"
						min="0"
						bind:value={amount}
						required
					/>
					<span class="unit">{target.currency}</span>
				</div>
			</div>

			<div class="field">
				<span class="field-label">Lo abriste el</span>
				<DatePicker name="openedOn" bind:value={openedOn} required />
				<p class="hint">
					Si lo abriste hace días, pon esa fecha: Finexia calcula lo que ya ganó. En un depósito no
					se anotan intereses a mano, así que nada se cuenta dos veces.
				</p>
			</div>

			<div class="pair">
				<div class="field">
					<label for="cash-deposit-term">Plazo</label>
					<select id="cash-deposit-term" bind:value={term}>
						{#each TERMS as days (days)}
							<option value={String(days)}>{days} días</option>
						{/each}
						<option value="custom">Otra fecha</option>
						<option value="none">Sin plazo</option>
					</select>
				</div>
				<div class="field">
					<span class="field-label">Vence el</span>
					{#if term === 'custom'}
						<DatePicker name="maturesOn" bind:value={customMaturity} required />
					{:else if maturesOn}
						<p class="readout">{day(maturesOn)}</p>
					{:else}
						<p class="readout">Hasta que lo canceles</p>
					{/if}
				</div>
			</div>

			<div class="pair">
				<div class="field">
					<label for="cash-deposit-rate">Tasa</label>
					<div class="with-unit">
						<input
							id="cash-deposit-rate"
							name="annualRatePct"
							type="number"
							inputmode="decimal"
							step="any"
							min="0"
							max="100"
							bind:value={annualRatePct}
							required
						/>
						<span class="unit">% E.A.</span>
					</div>
				</div>
				<div class="field">
					<label for="cash-deposit-withholding">
						Retención <span class="optional">(opcional)</span>
					</label>
					<div class="with-unit">
						<input
							id="cash-deposit-withholding"
							name="withholdingPct"
							type="number"
							inputmode="decimal"
							step="any"
							min="0"
							max="99.99"
							placeholder="0"
							bind:value={withholdingPct}
						/>
						<span class="unit">%</span>
					</div>
				</div>
			</div>

			<div class="field">
				<label for="cash-deposit-posting">Abono</label>
				<select id="cash-deposit-posting" name="posting" bind:value={posting}>
					<option value="daily">Cada día</option>
					<option value="at_maturity" disabled={!maturesOn}>Todo al vencer</option>
				</select>
				<p class="hint">
					{#if posting === 'at_maturity'}
						El saldo se queda en el capital y los intereses entran de golpe el día del vencimiento.
						Cuadra con un extracto que solo enseña el capital.
					{:else}
						El saldo sube cada día. Lo que rinde al final es lo mismo; cambia cuándo se ve.
					{/if}
				</p>
			</div>

			<!-- Las dos cifras que se quieren saber antes de guardar: lo que ya ganó
			     y lo que rendirá en total. -->
			{#if principal > 0 && rate > 0}
				<p class="preview">
					{#if elapsed > 0}
						Al guardarlo tendrá <strong>{money(alreadyEarned)}</strong> de intereses, los
						{elapsed}
						{elapsed === 1 ? 'día' : 'días'} que lleva abierto.<br />
					{/if}
					{#if termDays > 0}
						Al vencer, en {termDays} días, habrá rendido
						<strong>{money(atMaturity)}</strong> netos: {money(principal + atMaturity)} en total a {formatAnnualRate(
							rate
						)}.
					{:else}
						Rinde {formatAnnualRate(rate)} hasta que lo canceles.
					{/if}
				</p>
			{/if}

			{#if error}
				<p class="feedback error">{error}</p>
			{/if}

			<div class="actions">
				<Button type="button" variant="ghost" onclick={close}>Cancelar</Button>
				<Button type="submit" disabled={submitting}>Abrir depósito</Button>
			</div>
		</form>
	{:else if target?.mode === 'manage' && open}
		<div class="terms">
			<p class="line">
				<span>Tasa</span>
				<strong>{target.rate ? formatAnnualRate(target.rate.annualRatePct) : 'sin tasa'}</strong>
			</p>
			<p class="line">
				<span>Abierto el</span>
				<strong>{day(open.openedOn)}</strong>
			</p>
			<p class="line">
				<span>Vence</span>
				<strong>{open.maturesOn ? day(open.maturesOn) : 'sin plazo'}</strong>
			</p>
			<p class="line">
				<span>Vale hoy</span>
				<strong>{money(target.balance)}</strong>
			</p>
		</div>

		<p class="note">
			Un depósito no admite depósitos, retiros ni intereses a mano, y su tasa es la del día en que
			se abrió. Al vencer, su saldo vuelve solo a la cuenta principal.
		</p>

		{#if open.closedOn}
			<p class="note">Se cerró el {day(open.closedOn)}.</p>
		{:else}
			<form method="POST" action="?/closeDeposit" class="rail-fields" use:enhance={handler}>
				<input type="hidden" name="id" value={open.id} />

				<div class="pair">
					<div class="field">
						<span class="field-label">Cancelarlo el</span>
						<DatePicker name="closesOn" bind:value={closesOn} required />
					</div>
					<div class="field">
						<label for="cash-deposit-penalty">
							Penalidad <span class="optional">(opcional)</span>
						</label>
						<div class="with-unit">
							<input
								id="cash-deposit-penalty"
								name="penalty"
								type="number"
								inputmode="decimal"
								step="any"
								min="0"
								placeholder="0"
								bind:value={penalty}
							/>
							<span class="unit">{open.currency}</span>
						</div>
					</div>
				</div>

				<p class="hint">
					La tasa termina la víspera, se abona lo que llevaba ganado y el resto vuelve a la cuenta
					principal. Lo que se quede la entidad cuenta como pérdida, no como dinero que sacaste.
				</p>

				{#if error}
					<p class="feedback error">{error}</p>
				{/if}

				<div class="actions">
					<Button type="button" variant="ghost" onclick={close}>Cerrar</Button>
					<Button type="submit" disabled={submitting}>Cancelar depósito</Button>
				</div>
			</form>
		{/if}

		<form method="POST" action="?/deletePocket" class="danger" use:enhance={handler}>
			<input type="hidden" name="id" value={open.id} />
			<p class="danger-note">
				Borrarlo lo quita entero —el depósito, sus intereses y su tasa—, como si nunca hubiera
				existido. Es para algo que anotaste mal; si el dinero sí estuvo ahí, cancélalo.
			</p>
			<Button type="submit" variant="ghost" disabled={submitting}>Borrar depósito</Button>
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
		padding-right: 4.2rem;
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

	.readout {
		margin: 0;
		padding: 0.6rem 0;
		font-size: 0.9rem;
		color: var(--text);
	}

	.preview {
		margin: 0;
		padding: 0.75rem 0.9rem;
		border: 1px solid rgba(34, 201, 126, 0.25);
		border-radius: 8px;
		font-size: 0.82rem;
		line-height: 1.6;
		color: var(--text-muted);
	}

	.preview strong {
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
		color: var(--green);
	}

	.terms {
		margin: 0 0 1rem;
	}

	.terms .line {
		display: flex;
		align-items: baseline;
		justify-content: space-between;
		gap: 1rem;
		margin: 0;
		padding: 0.4rem 0;
		font-size: 0.85rem;
		color: var(--text-muted);
	}

	.terms .line + .line {
		border-top: 1px solid var(--border);
	}

	.terms strong {
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
		font-weight: 500;
		color: var(--text);
	}

	.note {
		margin: 0 0 1.2rem;
		font-size: 0.8rem;
		line-height: 1.55;
		color: var(--text-dim);
	}

	.actions {
		display: flex;
		justify-content: flex-end;
		gap: 0.6rem;
		margin-top: 0.4rem;
	}

	.danger {
		margin-top: 1.4rem;
		padding-top: 1.1rem;
		border-top: 1px solid var(--border);
		text-align: right;
	}

	.danger-note {
		margin: 0 0 0.6rem;
		font-size: 0.8rem;
		line-height: 1.5;
		color: var(--text-muted);
		text-align: left;
	}

	@media (max-width: 520px) {
		.pair {
			grid-template-columns: 1fr;
		}
	}
</style>
