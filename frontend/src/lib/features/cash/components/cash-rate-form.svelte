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
	 * - «Recalcular» tira los días ya calculados desde una fecha y los vuelve a
	 *   calcular sobre lo que la cuenta guarda hoy. Es lo que arregla un
	 *   movimiento anotado con fecha pasada, y solo aparece si hay días que
	 *   recalcular.
	 *
	 * La proyección usa el saldo de hoy y la tasa que se está escribiendo, neta de
	 * retención y por tramos, para compararla con lo que abona la
	 * entidad.
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
		annualFromNominal,
		CASH_RATE_FALLBACK,
		CASH_RECALCULATE_FALLBACK,
		cashAccountRate,
		describeCashAccountRate,
		effectiveAnnualRate,
		formatAnnualRate,
		NOMINAL_PERIODS,
		projectInterest,
		type CashRate,
		type RateTier
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

	type Mode = 'new' | 'edit' | 'end' | 'delete' | 'recalc';
	type Posting = 'daily' | 'monthly';

	interface Fields {
		mode: Mode;
		annualRatePct: string;
		withholdingPct: string;
		posting: Posting;
		tiers: TierRow[];
		date: string;
	}

	/** Una fila de tramo como la escribe el usuario. */
	interface TierRow {
		fromBalance: string;
		annualRatePct: string;
	}

	/** Los que acepta el backend en una versión. */
	const MAX_TIERS = 10;

	const today = todayLocalDateString();

	const status = $derived(
		target
			? cashAccountRate(
					rates,
					target.account.sourceId,
					target.account.currency,
					today,
					target.account.pocketId
				)
			: { current: null, upcoming: null, latest: null, accruedThrough: null }
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
	 * la tasa casi siempre conserva la retención, los tramos y la forma de abono— y
	 * la fecha de hoy.
	 */
	function initialFields(current: CashRateTarget | null, version: CashRate | null): Fields {
		return {
			mode: 'new',
			annualRatePct: current ? pctField(version?.annualRatePct) : '',
			withholdingPct: current ? pctField(version?.withholdingPct) : '',
			posting: (current && version?.posting === 'monthly' ? 'monthly' : 'daily') as Posting,
			tiers:
				current && version
					? version.tiers.map((tier) => ({
							fromBalance: String(parseFloat(tier.fromBalance)),
							annualRatePct: String(parseFloat(tier.annualRatePct))
						}))
					: [],
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
	let posting: Posting = $derived(initial.posting);
	let tiers: TierRow[] = $derived(initial.tiers);
	let effectiveFrom = $derived(initial.date);
	let endsOn = $derived(initial.date);
	/* Recalcular mira hacia atrás, así que arranca en el primer día del mes. */
	let recalcFrom = $derived(`${today.slice(0, 7)}-01`);
	let submitting = $state(false);
	let error = $state('');

	/*
	 * Las filas se reemplazan enteras en vez de mutarse: un `$derived`
	 * reasignable avisa cuando se reasigna, no cuando cambia algo por dentro.
	 */
	function setTier(index: number, field: keyof TierRow, value: string) {
		tiers = tiers.map((tier, i) => (i === index ? { ...tier, [field]: value } : tier));
	}

	function addTier(annualRatePct = '') {
		tiers = [...tiers, { fromBalance: '', annualRatePct }];
	}

	function removeTier(index: number) {
		tiers = tiers.filter((_, i) => i !== index);
	}

	/*
	 * El conversor: lo que dice el folleto de la entidad cuando publica una tasa
	 * nominal. No se envía — solo rellena el campo de la efectiva anual.
	 */
	let nominalPct = $state('');
	let nominalPeriods = $state(12);
	const converted = $derived(annualFromNominal(parseFloat(nominalPct) || 0, nominalPeriods));

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

	/* Sin días calculados no hay nada que recalcular. */
	const computed = $derived(status.accruedThrough);

	const modes = $derived<{ value: Mode; label: string }[]>(
		latest
			? [
					{ value: 'new', label: stopped ? 'Reanudar' : 'Cambiar tasa' },
					...(used ? [] : [{ value: 'edit' as const, label: 'Corregir' }]),
					...(stopped ? [] : [{ value: 'end' as const, label: 'Pausar' }]),
					...(used ? [] : [{ value: 'delete' as const, label: 'Borrar' }]),
					...(computed ? [{ value: 'recalc' as const, label: 'Recalcular' }] : [])
				]
			: []
	);

	const ACTIONS: Record<Mode, string> = {
		new: '?/createRate',
		edit: '?/updateRate',
		end: '?/endRate',
		delete: '?/deleteRate',
		recalc: '?/recalculateInterest'
	};

	const SUBMIT_LABELS: Record<Mode, string> = {
		new: 'Guardar tasa',
		edit: 'Guardar corrección',
		end: 'Pausar rentabilidad',
		delete: 'Borrar tasa',
		recalc: 'Recalcular intereses'
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
			case 'recalc':
				return `Los intereses están calculados hasta el ${longDate(computed ?? today)}. Desde el día que elijas se borran los abonos automáticos y se vuelven a calcular sobre lo que la cuenta guarda ahora. Úsalo si anotaste un depósito o un retiro con fecha pasada.`;
		}
	});

	const summary = $derived(describeCashAccountRate(status));

	/* Los campos numéricos entregan un número al escribir y una cadena al abrirse. */
	const rateValue = $derived(parseFloat(String(annualRatePct)) || 0);
	/* Los tramos que se pueden leer, en orden: una fila a medio escribir no cuenta. */
	const tierValues = $derived<RateTier[]>(
		tiers
			.map((tier) => ({
				fromBalance: parseFloat(String(tier.fromBalance)),
				annualRatePct: parseFloat(String(tier.annualRatePct)) || 0
			}))
			.filter((tier) => tier.fromBalance > 0)
			.sort((a, b) => a.fromBalance - b.fromBalance)
	);
	const projection = $derived(
		target
			? projectInterest(
					target.account.balance,
					rateValue,
					parseFloat(String(withholdingPct)) || 0,
					tierValues
				)
			: null
	);
	/*
	 * Con el saldo por encima del primer tramo, la cuenta ya no rinde la tasa de
	 * arriba sobre todo, y eso es lo que hay que decir: a cuánto rinde en conjunto.
	 */
	const blended = $derived(
		target && tierValues.length > 0 && target.account.balance > tierValues[0].fromBalance
			? effectiveAnnualRate(target.account.balance, rateValue, tierValues)
			: null
	);

	const money = (amount: number) =>
		privacy.money(formatCurrency(amount, target?.account.currency ?? 'USD'));
</script>

<Modal
	open={target !== null}
	title="Rentabilidad de la cuenta"
	description={target
		? `${target.account.sourceName || 'Sin plataforma'}${target.account.pocketName ? ` · ${target.account.pocketName}` : ''} · ${target.account.currency}${summary ? ` · ${summary}` : ''}`
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
						error =
							(result.data?.error as string) ??
							(mode === 'recalc' ? CASH_RECALCULATE_FALLBACK : CASH_RATE_FALLBACK);
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

			{#if mode === 'new' || mode === 'recalc'}
				<input type="hidden" name="sourceId" value={account.sourceId} />
				<input type="hidden" name="currency" value={account.currency} />
				<!-- Vacío es la cuenta principal: un bolsillo rinde su propia tasa. -->
				<input type="hidden" name="pocketId" value={account.pocketId ?? ''} />
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
					<fieldset class="field posting">
						<legend class="field-label">Cuándo lo abona</legend>
						<div class="options" style:--options="2">
							<label class="option" class:selected={posting === 'daily'}>
								<input type="radio" name="posting" value="daily" bind:group={posting} />
								Cada día
							</label>
							<label class="option" class:selected={posting === 'monthly'}>
								<input type="radio" name="posting" value="monthly" bind:group={posting} />
								Cada mes
							</label>
						</div>
						<p class="hint">
							{posting === 'monthly'
								? 'Se calcula igual todos los días y se abona todo junto el último día del mes, como hace la entidad.'
								: 'La cuenta recibe lo que rindió cada día, a la mañana siguiente.'}
						</p>
					</fieldset>
				</div>

				<details class="advanced">
					<summary>Opciones avanzadas</summary>
					<div class="advanced-body">
						<div class="pair">
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

						<fieldset class="tiers">
							<legend class="field-label">
								Tramos <span class="optional">(opcional)</span>
							</legend>
							<p class="hint">
								Si la entidad paga distinto según el saldo. La tasa de arriba rige desde cero y cada
								tramo desde el saldo que escribas. Un tramo al 0 % es un tope: desde ahí la cuenta
								no rinde.
							</p>
							{#each tiers as tier, i (i)}
								<div class="tier-row">
									<div class="with-unit">
										<label class="sr-only" for="cash-rate-tier-from-{i}">
											Desde qué saldo rige el tramo {i + 1}
										</label>
										<input
											id="cash-rate-tier-from-{i}"
											name="tierFromBalance"
											type="number"
											inputmode="decimal"
											step="any"
											min="0"
											placeholder="Desde"
											value={tier.fromBalance}
											oninput={(event) => setTier(i, 'fromBalance', event.currentTarget.value)}
										/>
										<span class="unit">{account.currency}</span>
									</div>
									<div class="with-unit">
										<label class="sr-only" for="cash-rate-tier-rate-{i}">
											Tasa efectiva anual del tramo {i + 1}
										</label>
										<input
											id="cash-rate-tier-rate-{i}"
											name="tierAnnualRatePct"
											type="number"
											inputmode="decimal"
											step="any"
											min="0"
											max="100"
											placeholder="Tasa"
											value={tier.annualRatePct}
											oninput={(event) => setTier(i, 'annualRatePct', event.currentTarget.value)}
										/>
										<span class="unit">% E.A.</span>
									</div>
									<Button type="button" variant="ghost" onclick={() => removeTier(i)}>
										Quitar<span class="sr-only"> el tramo {i + 1}</span>
									</Button>
								</div>
							{/each}
							<div class="tier-actions">
								<Button
									type="button"
									variant="ghost"
									disabled={tiers.length >= MAX_TIERS}
									onclick={() => addTier()}
								>
									Agregar tramo
								</Button>
								<Button
									type="button"
									variant="ghost"
									disabled={tiers.length >= MAX_TIERS}
									onclick={() => addTier('0')}
								>
									Agregar tope
								</Button>
							</div>
						</fieldset>

						<div class="converter">
							<p class="lead">
								¿La entidad publica una tasa nominal? Escríbela y la pasamos a efectiva anual.
							</p>
							<div class="converter-row">
								<div class="with-unit">
									<label class="sr-only" for="cash-rate-nominal">Tasa nominal</label>
									<input
										id="cash-rate-nominal"
										type="number"
										inputmode="decimal"
										step="any"
										min="0"
										placeholder="12"
										bind:value={nominalPct}
									/>
									<span class="unit">% N.A.</span>
								</div>
								<label class="sr-only" for="cash-rate-periods">Capitaliza</label>
								<select id="cash-rate-periods" bind:value={nominalPeriods}>
									{#each NOMINAL_PERIODS as period (period.value)}
										<option value={period.value}>{period.label}</option>
									{/each}
								</select>
								<Button
									type="button"
									variant="ghost"
									disabled={converted <= 0}
									onclick={() => (annualRatePct = String(Number(converted.toFixed(4))))}
								>
									{converted > 0 ? `Usar ${formatAnnualRate(converted.toFixed(4))}` : 'Convertir'}
								</Button>
							</div>
						</div>
					</div>
				</details>
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
			{:else if mode === 'recalc'}
				<div class="field">
					<span class="field-label">Recalcular desde</span>
					<DatePicker name="from" bind:value={recalcFrom} required />
				</div>
			{/if}

			<p class="hint">{hint}</p>

			{#if projection && (mode === 'new' || mode === 'edit')}
				<div class="projection" aria-live="polite">
					{#if rateValue <= 0}
						<p class="lead">Escribe la tasa para ver cuánto rendiría la cuenta.</p>
					{:else if account.balance > 0}
						<p class="lead">
							{#if blended !== null}
								Con los tramos, el saldo de hoy, {money(account.balance)}, rinde en conjunto un
								{formatAnnualRate(Number(blended.toFixed(2)))}, y rendiría
							{:else}
								Con el saldo de hoy, {money(account.balance)}, rendiría
							{/if}
						</p>
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
						Los intereses se calculan cada día sobre el saldo al cierre.
						{posting === 'monthly'
							? 'Con abono mensual se guardan hasta el último día del mes y entran todos juntos en un movimiento de intereses.'
							: 'Se abonan solos a la mañana siguiente, como movimientos de intereses.'}
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

	/* Lo que la entidad rara vez cambia queda plegado: la retención, los tramos y el
	   conversor de una tasa nominal. */
	.advanced {
		border: 1px solid var(--border);
		border-radius: 10px;
	}

	.advanced summary {
		padding: 0.7rem 0.95rem;
		font-size: 0.84rem;
		color: var(--text-muted);
		cursor: pointer;
	}

	.advanced summary:focus-visible {
		outline: 2px solid var(--amber);
		outline-offset: 2px;
	}

	.advanced-body {
		display: grid;
		gap: 0.9rem;
		padding: 0 0.95rem 0.95rem;
	}

	.posting {
		margin: 0;
		padding: 0;
		border: none;
	}

	.posting legend {
		margin-bottom: 0.45rem;
		padding: 0;
	}

	.tiers {
		display: grid;
		gap: 0.55rem;
		margin: 0;
		padding: 0;
		border: none;
	}

	.tiers legend {
		margin-bottom: 0.45rem;
		padding: 0;
	}

	.tier-row {
		display: grid;
		grid-template-columns: minmax(0, 1.4fr) minmax(0, 1fr) auto;
		align-items: center;
		gap: 0.5rem;
	}

	.tier-actions {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
	}

	.converter {
		display: grid;
		gap: 0.55rem;
		padding-top: 0.2rem;
		border-top: 1px solid var(--border);
	}

	.converter-row {
		display: grid;
		grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) auto;
		align-items: center;
		gap: 0.5rem;
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
		dl,
		.converter-row,
		.tier-row {
			grid-template-columns: 1fr;
		}

		.options {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}
	}
</style>
