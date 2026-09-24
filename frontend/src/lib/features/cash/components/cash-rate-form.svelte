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
	 *   llama «Reanudar». El día puede ser pasado —la tasa que la cuenta ya rendía
	 *   antes de anotarla— y los intereses desde entonces se calculan al guardar;
	 *   si ya había días calculados desde ese día, se rehacen con la tasa nueva.
	 * - «Cambiar fecha» mueve el día en que empieza la versión más reciente: la
	 *   anterior rige hasta la víspera del día nuevo. Es para un cambio que la
	 *   entidad anunció para un día y aplicó otro, o que se anotó con la fecha
	 *   equivocada. Si ya generó intereses, sus días se rehacen desde el primero
	 *   que toca el cambio.
	 * - «Corregir» reescribe los valores sin tocar las fechas: es para un error al
	 *   escribirla, no para un cambio de la entidad.
	 * - «Pausar» dice desde qué día la cuenta deja de rendir.
	 * - «Borrar» quita la versión.
	 * - «Recalcular» tira los días ya calculados desde una fecha y los vuelve a
	 *   calcular sobre lo que la cuenta guarda hoy, hasta el último que ya estaba
	 *   calculado. Es lo que arregla un
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
	import { OptimisticDialog } from '$lib/shared/optimistic.svelte';
	import { todayLocalDateString } from '$lib/shared/format/date';
	import type { CashAccount } from '../cash';
	import {
		CASH_RATE_FALLBACK,
		CASH_RECALCULATE_FALLBACK,
		cashAccountRate,
		effectiveAnnualRate,
		projectInterest,
		type CashRate,
		type RateTier
	} from '../rates';
	import { cashRateLines } from '../yield';
	import { cashRecalcSent, type CashRecalcSent } from '../interest';
	import { cashRateHistory, type CashRateMode } from '../rate-history';
	import { cashRecalc } from '../recalc.svelte';
	import type { CashRecalculation } from '$lib/api/types';
	import CashChoice from './cash-choice.svelte';
	import CashMoneyInput from './cash-money-input.svelte';
	import CashRateAdvanced, { type TierRow } from './cash-rate-advanced.svelte';
	import CashRateProjection from './cash-rate-projection.svelte';

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

	type Mode = CashRateMode;
	type Posting = 'daily' | 'monthly';

	interface Fields {
		mode: Mode;
		annualRatePct: string;
		withholdingPct: string;
		posting: Posting;
		tiers: TierRow[];
		date: string;
	}

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
	/* Los campos numéricos entregan un número al escribir y una cadena al abrirse. */
	let annualRatePct: string | number | null = $derived(initial.annualRatePct);
	let withholdingPct: string | number | null = $derived(initial.withholdingPct);
	let posting: Posting = $derived(initial.posting);
	let tiers: TierRow[] = $derived(initial.tiers);
	let effectiveFrom = $derived(initial.date);
	/* Mover arranca en el día en que empieza hoy la versión, no en hoy. */
	let moveTo = $derived(latest ? latest.effectiveFrom.slice(0, 10) : today);
	let endsOn = $derived(initial.date);
	/* Recalcular mira hacia atrás, así que arranca en el primer día del mes. */
	let recalcFrom = $derived(`${today.slice(0, 7)}-01`);
	/* Se cierra al pulsar y guarda de fondo; si el servidor lo rechaza, vuelve
	   con lo escrito y el motivo. */
	const dialog = new OptimisticDialog(() => target);

	/* Cerrar limpia el error: la siguiente apertura no lo arrastra. */
	function close() {
		dialog.reset();
		onClose();
	}

	/*
	 * Cómo estaba la cuenta al pedir el recálculo. Se toma al enviar porque,
	 * cuando la página se refresca, el diálogo ya se cerró y los saldos son los
	 * nuevos: es contra esto que se ve si cambió algo.
	 */
	let recalcSent: CashRecalcSent | null = null;

	const submit = dialog.submit({
		fallbackError: () => (mode === 'recalc' ? CASH_RECALCULATE_FALLBACK : CASH_RATE_FALLBACK),
		apply: () => {
			recalcSent =
				mode === 'recalc' && target ? cashRecalcSent(target.account, rates, recalcFrom) : null;
		},
		onDone: close,
		onSaved: (data) => {
			const result = data?.recalculation as CashRecalculation | undefined;
			if (recalcSent && result) cashRecalc.show({ ...recalcSent, result });
			recalcSent = null;
		}
	});

	/*
	 * Con intereses ya calculados la tasa no se corrige ni se borra: esos días se
	 * ganaron a ella. Se cambia con una versión nueva, se pausa, o se mueve su
	 * inicio rehaciendo sus días.
	 */
	const used = $derived(!!latest?.accruedThrough);

	/* Sin días calculados no hay nada que recalcular. */
	const computed = $derived(status.accruedThrough);

	const modes = $derived<{ value: Mode; label: string }[]>(
		latest
			? [
					{ value: 'new', label: stopped ? 'Reanudar' : 'Cambiar tasa' },
					{ value: 'move' as const, label: 'Cambiar fecha' },
					...(used ? [] : [{ value: 'edit' as const, label: 'Corregir' }]),
					...(stopped ? [] : [{ value: 'end' as const, label: 'Pausar' }]),
					...(used ? [] : [{ value: 'delete' as const, label: 'Borrar' }]),
					...(computed ? [{ value: 'recalc' as const, label: 'Recalcular' }] : [])
				]
			: []
	);

	const ACTIONS: Record<Mode, string> = {
		new: '?/createRate',
		move: '?/rescheduleRate',
		edit: '?/updateRate',
		end: '?/endRate',
		delete: '?/deleteRate',
		recalc: '?/recalculateInterest'
	};

	const SUBMIT_LABELS: Record<Mode, string> = {
		new: 'Guardar tasa',
		move: 'Guardar fecha',
		edit: 'Guardar corrección',
		end: 'Pausar rentabilidad',
		delete: 'Borrar tasa',
		recalc: 'Recalcular intereses'
	};

	/* El día desde el que rige la tasa en el modo abierto: anotarla o moverla. */
	const startDay = $derived(mode === 'move' ? moveTo : effectiveFrom);

	/* Qué le pasa a la historia de la cuenta con la opción abierta. */
	const history = $derived(
		cashRateHistory({ mode, latest, used, stopped, pending, computed, startDay, today })
	);

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

	/* La tasa de ahora, en la cabecera: la cifra y lo que la matiza. */
	const current = $derived(cashRateLines(status, money));

	const title = $derived(
		target
			? `${target.account.sourceName || 'Sin plataforma'}, ${target.account.pocketName || `cuenta en ${target.account.currency}`}`
			: ''
	);
</script>

<Modal
	open={target !== null}
	hidden={dialog.hidden}
	title="Tasa de la cuenta"
	description={title}
	size="md"
	onClose={close}
>
	{#if target}
		{@const account = target.account}
		<form method="POST" action={ACTIONS[mode]} class="rail-fields" use:enhance={submit}>
			<!-- Lo que rinde hoy, antes de tocar nada: es lo que se viene a mirar
			     más veces de las que se viene a cambiar. -->
			<div class="now" class:earning={current?.earning}>
				<span class="key" aria-hidden="true"></span>
				<div>
					<p class="now-rate">{current?.rate ?? 'Sin tasa todavía'}</p>
					<p class="now-detail">
						{#if current?.detail}
							{current.detail}
						{:else if !current}
							Anota la que te paga la entidad y Finexia abonará los intereses solo.
						{:else}
							Sobre {money(account.balance)} de saldo
						{/if}
					</p>
				</div>
			</div>

			{#if latest}
				<CashChoice legend="Qué quieres hacer" options={modes} bind:value={mode} />
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
						<CashMoneyInput
							id="cash-rate-annual"
							name="annualRatePct"
							unit="% E.A."
							max="100"
							size="lg"
							bind:value={annualRatePct}
							aria-describedby="cash-rate-annual-hint"
							required
						/>
						<p class="hint" id="cash-rate-annual-hint">
							La que publica la entidad. Un APY en dólares es la misma cifra.
						</p>
					</div>
					<div class="field">
						<CashChoice
							legend="Cuándo lo abona"
							name="posting"
							options={[
								{ value: 'daily', label: 'Cada día' },
								{ value: 'monthly', label: 'Cada mes' }
							]}
							bind:value={posting}
							describedby="cash-rate-posting-hint"
						/>
						<p class="hint" id="cash-rate-posting-hint">
							{posting === 'monthly'
								? 'Se abona todo junto el último día del mes, como hace la entidad.'
								: 'La cuenta recibe lo que rindió cada día, a la mañana siguiente.'}
						</p>
					</div>
				</div>

				<CashRateAdvanced
					currency={account.currency}
					bind:withholdingPct
					bind:tiers
					onConvert={(pct) => (annualRatePct = pct)}
				/>
			{/if}

			{#if mode === 'new'}
				<div class="field">
					<span class="field-label">Rige desde</span>
					<DatePicker name="effectiveFrom" bind:value={effectiveFrom} required />
					<p class="hint">
						Puede ser un día pasado, si la cuenta ya rendía esta tasa, o uno futuro, si la entidad
						anunció el cambio.
					</p>
				</div>
			{:else if mode === 'move'}
				<div class="field">
					<span class="field-label">Rige desde</span>
					<DatePicker name="effectiveFrom" bind:value={moveTo} required />
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

			{#if history.recomputes}
				<input type="hidden" name="recompute" value="true" />
			{/if}

			<p
				class="consequence"
				class:warn={mode === 'delete' || mode === 'recalc' || history.recomputes}
			>
				{history.hint}
			</p>

			{#if projection && (mode === 'new' || mode === 'edit')}
				<CashRateProjection
					balance={account.balance}
					currency={account.currency}
					rate={rateValue}
					{projection}
					{blended}
					{posting}
				/>
			{/if}

			{#if dialog.error}
				<p class="feedback error" role="alert">{dialog.error}</p>
			{/if}

			<div class="modal-actions">
				<Button type="button" variant="ghost" onclick={close} disabled={dialog.submitting}
					>Cancelar</Button
				>
				<Button
					type="submit"
					variant={mode === 'delete' ? 'danger' : 'primary'}
					loading={dialog.submitting}
				>
					{SUBMIT_LABELS[mode]}
				</Button>
			</div>
		</form>
	{/if}
</Modal>

<style>
	/* La tasa de ahora, con la marca de la fila de la cuenta: verde llena
	   mientras rinde, un aro gris cuando no. */
	.now {
		display: flex;
		align-items: baseline;
		gap: 0.7rem;
		padding: 0.9rem 1rem;
		border: 1px solid var(--border);
		border-radius: 10px;
		background: rgba(255, 255, 255, 0.02);
	}

	.key {
		flex-shrink: 0;
		width: 8px;
		height: 8px;
		border: 1.5px solid var(--text-dim);
		border-radius: 50%;
		transform: translateY(-0.1rem);
	}

	.earning .key {
		border-color: var(--green);
		background: var(--green);
	}

	.now p {
		margin: 0;
	}

	.now-rate {
		font-family: var(--font-mono);
		font-size: 1.15rem;
		color: var(--text);
	}

	.now-detail {
		margin-top: 0.15rem !important;
		font-size: 0.8rem;
		line-height: 1.45;
		color: var(--text-dim);
	}

	/* Lo que le pasa a la historia de la cuenta con la opción elegida: la línea
	   que hay que leer antes de guardar, con el filete de los avisos. */
	.consequence {
		margin: 0;
		padding-left: 0.75rem;
		border-left: 2px solid var(--border-strong);
		font-size: 0.83rem;
		line-height: 1.55;
		color: var(--text-muted);
	}

	.consequence.warn {
		border-left-color: rgba(212, 145, 42, 0.55);
	}

	.feedback {
		margin: 0;
	}
</style>
