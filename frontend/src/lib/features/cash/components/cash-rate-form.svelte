<script lang="ts">
	/**
	 * La rentabilidad de una cuenta de efectivo: la tasa que le paga su plataforma.
	 *
	 * Sin tasa, el formulario solo la anota. Con tasa ofrece lo que se le puede
	 * hacer a la versión más reciente, a la vista y no en un menú, porque cada
	 * opción cambia la historia de otra manera:
	 *
	 * - «Cambiar tasa» anota una versión desde un día: la anterior termina la
	 *   víspera y los días pasados conservan la suya. Con la tasa ya terminada se
	 *   llama «Reanudar».
	 * - «Corregir» reescribe los valores sin tocar las fechas: es para un error al
	 *   escribirla, no para un cambio de la entidad.
	 * - «Pausar» dice desde qué día la cuenta deja de rendir.
	 * - «Borrar» quita la versión.
	 *
	 * La proyección usa el saldo de hoy y la tasa que se está escribiendo, neta de
	 * retención, para compararla con lo que abona la entidad.
	 */
	import { enhance } from '$app/forms';
	import Button from '$lib/ui/button.svelte';
	import Modal from '$lib/ui/modal.svelte';
	import DatePicker from '$lib/ui/date-picker.svelte';
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import { formatCalendarDate, todayLocalDateString } from '$lib/shared/format/date';
	import type { CashAccount } from '../cash';
	import {
		cashAccountRate,
		describeCashAccountRate,
		formatAnnualRate,
		projectInterest,
		type CashRate
	} from '../rates';

	/** La cuenta cuya tasa se gestiona. */
	export interface CashRateTarget {
		account: CashAccount;
	}

	interface Props {
		target: CashRateTarget | null;
		/** Todas las versiones de todas las cuentas; el formulario toma las de la suya. */
		rates: CashRate[];
		onClose: () => void;
	}

	let { target, rates, onClose }: Props = $props();

	type Mode = 'new' | 'edit' | 'end' | 'delete';

	interface Fields {
		mode: Mode;
		annualRatePct: string;
		withholdingPct: string;
		date: string;
	}

	const today = todayLocalDateString();

	const status = $derived(
		target
			? cashAccountRate(rates, target.account.sourceId, target.account.currency, today)
			: { current: null, upcoming: null, latest: null }
	);
	const latest = $derived(status.latest);

	/* La versión más reciente ya no rige: terminó antes de hoy. */
	const stopped = $derived(!!latest?.endedOn && latest.endedOn.slice(0, 10) < today);

	/* La versión más reciente aún no ha empezado. */
	const pending = $derived(!!latest && latest.effectiveFrom.slice(0, 10) > today);

	/** Un porcentaje del backend («9.25», «0») como valor del campo: vacío si es cero. */
	function pctField(pct: string | undefined): string {
		const value = parseFloat(pct ?? '');
		return value > 0 ? String(value) : '';
	}

	/*
	 * Cómo arranca al abrirse: con los valores de la versión más reciente —cambiar
	 * la tasa casi siempre conserva la retención— y la fecha de hoy.
	 */
	function initialFields(current: CashRateTarget | null, version: CashRate | null): Fields {
		return {
			mode: 'new',
			annualRatePct: current ? pctField(version?.annualRatePct) : '',
			withholdingPct: current ? pctField(version?.withholdingPct) : '',
			date: today
		};
	}

	/*
	 * `$derived` reasignables, como en el formulario de movimientos: lo que
	 * escribe el usuario los pisa hasta que se abre otra cuenta.
	 */
	const initial = $derived(initialFields(target, latest));
	let mode = $derived(initial.mode);
	let annualRatePct = $derived(initial.annualRatePct);
	let withholdingPct = $derived(initial.withholdingPct);
	let effectiveFrom = $derived(initial.date);
	let endsOn = $derived(initial.date);
	let submitting = $state(false);
	let error = $state('');

	/* Cerrar limpia el error: la siguiente apertura no lo arrastra. */
	function close() {
		error = '';
		onClose();
	}

	/*
	 * Con intereses ya calculados la tasa no se corrige ni se borra: esos días se
	 * ganaron a ella. Se cambia con una versión nueva, o se pausa.
	 */
	const used = $derived(!!latest?.accruedThrough);

	const modes = $derived<{ value: Mode; label: string }[]>(
		latest
			? [
					{ value: 'new', label: stopped ? 'Reanudar' : 'Cambiar tasa' },
					...(used ? [] : [{ value: 'edit' as const, label: 'Corregir' }]),
					...(stopped ? [] : [{ value: 'end' as const, label: 'Pausar' }]),
					...(used ? [] : [{ value: 'delete' as const, label: 'Borrar' }])
				]
			: []
	);

	const ACTIONS: Record<Mode, string> = {
		new: '?/createRate',
		edit: '?/updateRate',
		end: '?/endRate',
		delete: '?/deleteRate'
	};

	const SUBMIT_LABELS: Record<Mode, string> = {
		new: 'Guardar tasa',
		edit: 'Guardar corrección',
		end: 'Pausar rentabilidad',
		delete: 'Borrar tasa'
	};

	const longDate = (iso: string) =>
		formatCalendarDate(iso.slice(0, 10), { day: 'numeric', month: 'long', year: 'numeric' });

	/** Qué le pasa a la historia de la cuenta con cada opción. */
	const hint = $derived.by(() => {
		if (!latest) return 'Desde ese día, la cuenta rinde esta tasa.';

		switch (mode) {
			case 'new':
				if (pending) {
					return `Ya hay una tasa anotada desde el ${longDate(latest.effectiveFrom)}: la nueva tiene que empezar después. Para cambiar esa, usa «Corregir».`;
				}
				return stopped
					? 'Desde ese día la cuenta vuelve a rendir. La pausa se queda como estaba.'
					: `La tasa de ahora, ${formatAnnualRate(latest.annualRatePct)}, termina la víspera. Los días anteriores conservan la suya.`;
			case 'edit':
				return `Corrige la tasa anotada desde el ${longDate(latest.effectiveFrom)}, sin crear otra. Si la entidad cambió la tasa, usa «Cambiar tasa».`;
			case 'end':
				return 'Desde ese día la cuenta deja de rendir. Los días anteriores conservan su tasa.';
			case 'delete':
				return `Borra la tasa de ${formatAnnualRate(latest.annualRatePct)} anotada desde el ${longDate(latest.effectiveFrom)}. Si al empezar cerró otra, esa vuelve a regir.`;
		}
	});

	const summary = $derived(describeCashAccountRate(status));

	/* Los campos numéricos entregan un número al escribir y una cadena al abrirse. */
	const rateValue = $derived(parseFloat(String(annualRatePct)) || 0);
	const projection = $derived(
		target
			? projectInterest(target.account.balance, rateValue, parseFloat(String(withholdingPct)) || 0)
			: null
	);

	const money = (amount: number) =>
		privacy.money(formatCurrency(amount, target?.account.currency ?? 'USD'));
</script>

<Modal
	open={target !== null}
	title="Rentabilidad de la cuenta"
	description={target
		? `${target.account.sourceName || 'Sin plataforma'} · ${target.account.currency}${summary ? ` · ${summary}` : ''}`
		: ''}
	size="md"
	onClose={close}
>
	{#if target}
		{@const account = target.account}
		<form
			method="POST"
			action={ACTIONS[mode]}
			class="rail-fields"
			use:enhance={() => {
				submitting = true;
				return async ({ result, update }) => {
					submitting = false;
					if (result.type === 'failure') {
						error = (result.data?.error as string) ?? 'No pudimos guardar la tasa.';
						return;
					}
					await update();
					close();
				};
			}}
		>
			{#if latest}
				<fieldset class="modes">
					<legend class="field-label">Qué quieres hacer</legend>
					<div class="options" style:--options={modes.length}>
						{#each modes as option (option.value)}
							<label class="option" class:selected={mode === option.value}>
								<input type="radio" name="rateMode" value={option.value} bind:group={mode} />
								{option.label}
							</label>
						{/each}
					</div>
				</fieldset>
			{/if}

			{#if mode === 'new'}
				<input type="hidden" name="sourceId" value={account.sourceId} />
				<input type="hidden" name="currency" value={account.currency} />
			{:else if latest}
				<input type="hidden" name="id" value={latest.id} />
			{/if}

			{#if mode === 'new' || mode === 'edit'}
				<div class="pair">
					<div class="field">
						<label for="cash-rate-annual">Tasa efectiva anual</label>
						<div class="with-unit">
							<input
								id="cash-rate-annual"
								name="annualRatePct"
								type="number"
								inputmode="decimal"
								step="any"
								min="0"
								max="100"
								bind:value={annualRatePct}
								aria-describedby="cash-rate-annual-unit cash-rate-annual-hint"
								required
							/>
							<span class="unit" id="cash-rate-annual-unit">% E.A.</span>
						</div>
						<p class="hint" id="cash-rate-annual-hint">
							La que publica la entidad. Un APY en dólares es la misma cifra.
						</p>
					</div>
					<div class="field">
						<label for="cash-rate-withholding">
							Retención <span class="optional">(opcional)</span>
						</label>
						<div class="with-unit">
							<input
								id="cash-rate-withholding"
								name="withholdingPct"
								type="number"
								inputmode="decimal"
								step="any"
								min="0"
								max="99.99"
								bind:value={withholdingPct}
								aria-describedby="cash-rate-withholding-unit cash-rate-withholding-hint"
							/>
							<span class="unit" id="cash-rate-withholding-unit">%</span>
						</div>
						<p class="hint" id="cash-rate-withholding-hint">
							Si la entidad te la descuenta, lo que rinde es neto.
						</p>
					</div>
				</div>
			{/if}

			{#if mode === 'new'}
				<div class="field">
					<span class="field-label">Rige desde</span>
					<DatePicker name="effectiveFrom" bind:value={effectiveFrom} required />
				</div>
			{:else if mode === 'end'}
				<div class="field">
					<span class="field-label">Deja de rendir desde</span>
					<DatePicker name="endsOn" bind:value={endsOn} required />
				</div>
			{/if}

			<p class="hint">{hint}</p>

			{#if projection && (mode === 'new' || mode === 'edit')}
				<div class="projection" aria-live="polite">
					{#if rateValue <= 0}
						<p class="lead">Escribe la tasa para ver cuánto rendiría la cuenta.</p>
					{:else if account.balance > 0}
						<p class="lead">Con el saldo de hoy, {money(account.balance)}, rendiría</p>
						<dl>
							<div>
								<dt>Al día</dt>
								<dd>≈ {money(projection.day)}</dd>
							</div>
							<div>
								<dt>En 30 días</dt>
								<dd>≈ {money(projection.month)}</dd>
							</div>
							<div>
								<dt>En un año</dt>
								<dd>≈ {money(projection.year)}</dd>
							</div>
						</dl>
					{:else}
						<p class="lead">Cuando la cuenta tenga saldo, aquí verás cuánto rinde.</p>
					{/if}
					<p class="note">
						Los intereses se calculan cada día sobre el saldo al cierre y se abonan solos a la
						mañana siguiente, como movimientos de intereses.
					</p>
				</div>
			{/if}

			{#if error}
				<p class="feedback error">{error}</p>
			{/if}

			<div class="modal-actions">
				<Button type="button" variant="ghost" onclick={close} disabled={submitting}>Cancelar</Button
				>
				<Button type="submit" loading={submitting}>{SUBMIT_LABELS[mode]}</Button>
			</div>
		</form>
	{/if}
</Modal>

<style>
	.modes {
		margin: 0;
		padding: 0;
		border: none;
	}

	.modes legend {
		margin-bottom: 0.45rem;
		padding: 0;
	}

	.options {
		display: grid;
		grid-template-columns: repeat(var(--options, 4), minmax(0, 1fr));
		gap: 0.5rem;
	}

	.option {
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 0.6rem 0.5rem;
		border: 1px solid rgba(212, 145, 42, 0.2);
		border-radius: 8px;
		font-size: 0.86rem;
		text-align: center;
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

	/* Sitio a la derecha para la unidad, y sin flechas: suben de uno en uno, que
	   en una tasa no sirve, y la tapaban. */
	.with-unit input {
		padding-right: 4.25rem;
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

	/* Lo que rendiría, en verde como los intereses de la lista de movimientos. */
	.projection {
		display: grid;
		gap: 0.65rem;
		padding: 0.9rem 1rem;
		border: 1px solid var(--border);
		border-radius: 10px;
		background: var(--surface);
	}

	.lead {
		margin: 0;
		font-size: 0.84rem;
		color: var(--text-muted);
	}

	dl {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 0.75rem;
		margin: 0;
	}

	dl div {
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
		font-family: var(--font-mono);
		font-size: 0.95rem;
		font-variant-numeric: tabular-nums;
		color: var(--green);
		overflow-wrap: anywhere;
	}

	.note {
		margin: 0;
		font-size: 0.76rem;
		line-height: 1.45;
		color: var(--text-dim);
	}

	.feedback {
		margin: 0;
	}

	@media (max-width: 640px) {
		.pair,
		dl {
			grid-template-columns: 1fr;
		}

		.options {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}
	}
</style>
