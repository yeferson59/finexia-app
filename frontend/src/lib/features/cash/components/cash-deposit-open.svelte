<script lang="ts">
	/**
	 * Abrir un depósito a tasa fija: el dinero, la tasa y el plazo a la vez,
	 * porque son una misma cosa y la tasa no se le da después.
	 *
	 * La fecha de apertura puede estar en el pasado. En un depósito no se anotan
	 * intereses a mano, así que no hay nada que contar dos veces: al guardarlo,
	 * Finexia calcula los días que ya ganó y los abona como rendimiento.
	 *
	 * Lo que se quiere saber antes de guardar va abajo, con la línea del plazo:
	 * lo que ya ganó, lo que rendirá y en qué queda al vencer.
	 */
	import { enhance } from '$app/forms';
	import Button from '$lib/ui/button.svelte';
	import DatePicker from '$lib/ui/date-picker.svelte';
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import { formatCalendarDate, todayLocalDateString } from '$lib/shared/format/date';
	import { formatAnnualRate } from '../rates';
	import { addCalendarDays, daysBetween, projectInterestOverDays } from '../deposits';
	import CashChoice from './cash-choice.svelte';
	import CashMoneyInput from './cash-money-input.svelte';
	import CashTermTrack from './cash-term-track.svelte';

	interface Props {
		sourceId: string;
		currency: string;
		/** El portafolio que pone el dinero; un depósito es de uno solo. */
		portfolioId: string;
		onClose: () => void;
	}

	let { sourceId, currency, portfolioId, onClose }: Props = $props();

	const today = todayLocalDateString();

	/* Los plazos que cotizan las entidades, y las dos salidas: una fecha
	   cualquiera, y un depósito que no vence. */
	const TERMS = [
		...[30, 60, 90, 180, 360].map((days) => ({ value: String(days), label: `${days} días` })),
		{ value: 'custom', label: 'Otra fecha' },
		{ value: 'none', label: 'Sin plazo' }
	];

	let term = $state('90');
	let openedOn = $state(today);
	let customMaturity = $state(addCalendarDays(today, 90));
	let amount = $state<string | number | null>('');
	let name = $state('');
	let annualRatePct = $state<string | number | null>('');
	let withholdingPct = $state<string | number | null>('');
	let posting = $state<'daily' | 'at_maturity'>('daily');
	let submitting = $state(false);
	let error = $state('');

	/* El día en que vence, según el plazo elegido; vacío si no tiene. */
	const maturesOn = $derived(
		term === 'none'
			? ''
			: term === 'custom'
				? customMaturity
				: addCalendarDays(openedOn, Number(term))
	);

	/* Sin vencimiento no hay último día en que abonarlo todo: se ve y se envía
	   «cada día», y lo elegido vuelve si se le pone plazo otra vez. */
	const postingShown = $derived(maturesOn ? posting : 'daily');

	const termDays = $derived(maturesOn ? daysBetween(openedOn, maturesOn) : 0);

	/* Los días que ya pasaron: lo que Finexia abona en cuanto se guarda. */
	const elapsed = $derived(Math.max(0, daysBetween(openedOn, today)));

	const principal = $derived(parseFloat(String(amount)) || 0);
	const rate = $derived(parseFloat(String(annualRatePct)) || 0);
	const withheld = $derived(parseFloat(String(withholdingPct)) || 0);

	const atMaturity = $derived(projectInterestOverDays(principal, rate, termDays, withheld));
	const alreadyEarned = $derived(projectInterestOverDays(principal, rate, elapsed, withheld));

	const money = (value: number) => privacy.money(formatCurrency(value, currency));

	const longDay = (iso: string) =>
		formatCalendarDate(iso.slice(0, 10), { day: 'numeric', month: 'long', year: 'numeric' });
</script>

<form
	method="POST"
	action="?/openDeposit"
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
			onClose();
		};
	}}
>
	<input type="hidden" name="portfolioId" value={portfolioId} />
	<input type="hidden" name="sourceId" value={sourceId} />
	<input type="hidden" name="currency" value={currency} />
	<!-- El vencimiento viaja en un solo campo. Con un plazo en días sale de la
	     cuenta; con «Otra fecha» lo pone el selector, que lleva ese mismo nombre. -->
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
		<CashMoneyInput
			id="cash-deposit-amount"
			name="amount"
			unit={currency}
			size="lg"
			bind:value={amount}
			required
		/>
	</div>

	<div class="pair">
		<div class="field">
			<label for="cash-deposit-rate">Tasa</label>
			<CashMoneyInput
				id="cash-deposit-rate"
				name="annualRatePct"
				unit="% E.A."
				max="100"
				bind:value={annualRatePct}
				required
			/>
		</div>
		<div class="field">
			<label for="cash-deposit-withholding">
				Retención <span class="optional">(opcional)</span>
			</label>
			<CashMoneyInput
				id="cash-deposit-withholding"
				name="withholdingPct"
				unit="%"
				max="99.99"
				placeholder="0"
				bind:value={withholdingPct}
			/>
		</div>
	</div>

	<CashChoice legend="Plazo" options={TERMS} bind:value={term} columns={4} />

	<div class="field">
		<span class="field-label">Lo abriste el</span>
		<DatePicker name="openedOn" bind:value={openedOn} required />
		<p class="hint">
			Si lo abriste hace días, pon esa fecha: Finexia calcula lo que ya ganó. En un depósito no se
			anotan intereses a mano, así que nada se cuenta dos veces.
		</p>
	</div>

	{#if term === 'custom'}
		<div class="field">
			<span class="field-label">Vence el</span>
			<DatePicker name="maturesOn" bind:value={customMaturity} required />
		</div>
	{/if}

	<div class="field">
		<CashChoice
			legend="Cuándo abona"
			name="posting"
			options={[
				{ value: 'daily', label: 'Cada día' },
				{ value: 'at_maturity', label: 'Todo al vencer', disabled: !maturesOn }
			]}
			bind:value={() => postingShown, (value) => (posting = value)}
			describedby="cash-deposit-posting-hint"
		/>
		<p class="hint" id="cash-deposit-posting-hint">
			{#if postingShown === 'at_maturity'}
				El saldo se queda en el capital y los intereses entran de golpe el día que vence. Cuadra con
				un extracto que solo enseña el capital.
			{:else}
				El saldo sube cada día. Lo que rinde al final es lo mismo; cambia cuándo se ve.
			{/if}
		</p>
	</div>

	<!-- Las cifras que se quieren saber antes de guardar, sobre la línea del plazo. -->
	<div class="preview" aria-live="polite">
		{#if maturesOn && termDays > 0}
			<CashTermTrack {openedOn} {maturesOn} {today} caption="vence el {longDay(maturesOn)}" />
		{:else}
			<p class="lead">Sin plazo: rinde hasta que lo canceles.</p>
		{/if}

		{#if principal > 0 && rate > 0}
			<dl class="figures">
				{#if elapsed > 0}
					<div>
						<dt>Ya ganó, en {elapsed} {elapsed === 1 ? 'día' : 'días'}</dt>
						<dd class="gain">+{money(alreadyEarned)}</dd>
					</div>
				{/if}
				{#if termDays > 0}
					<div>
						<dt>Rendirá en {termDays} días, neto</dt>
						<dd class="gain">+{money(atMaturity)}</dd>
					</div>
					<div>
						<dt>Al vencer tendrás</dt>
						<dd>{money(principal + atMaturity)}</dd>
					</div>
				{:else}
					<div>
						<dt>Rinde</dt>
						<dd>{formatAnnualRate(rate)}</dd>
					</div>
				{/if}
			</dl>
		{:else}
			<p class="lead">Escribe el importe y la tasa para ver lo que rendirá.</p>
		{/if}
	</div>

	{#if error}
		<p class="feedback error">{error}</p>
	{/if}

	<div class="modal-actions">
		<Button type="button" variant="ghost" onclick={onClose}>Cancelar</Button>
		<Button type="submit" loading={submitting}>Abrir depósito</Button>
	</div>
</form>

<style>
	.preview {
		display: grid;
		gap: 1rem;
		padding: 1rem;
		border: 1px solid var(--border);
		border-radius: 10px;
		background: rgba(255, 255, 255, 0.02);
		--ring: #131416;
	}

	.lead {
		margin: 0;
		font-size: 0.84rem;
		color: var(--text-muted);
	}

	.figures {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(8.5rem, 1fr));
		gap: 0.75rem 1rem;
		margin: 0;
	}

	.figures div {
		display: grid;
		gap: 0.2rem;
		min-width: 0;
	}

	dt {
		font-size: 0.74rem;
		color: var(--text-dim);
	}

	dd {
		margin: 0;
		font-family: var(--font-mono);
		font-size: 0.98rem;
		font-variant-numeric: tabular-nums;
		color: var(--text);
		overflow-wrap: anywhere;
	}

	dd.gain {
		color: var(--green);
	}

	.feedback {
		margin: 0;
	}
</style>
