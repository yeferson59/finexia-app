<script lang="ts">
	/**
	 * Aportar a un fondo, retirar de él, y corregir o borrar lo que ya se anotó.
	 *
	 * En un fondo por saldo solo se habla de dinero. Las unidades las calcula
	 * Finexia al valor del último saldo anotado antes del día del movimiento, que
	 * es lo que separa lo aportado de lo ganado. Si ese saldo es viejo, «Saldo
	 * justo antes» lo hace exacto: es lo que tenía el fondo al cierre del día
	 * anterior.
	 *
	 * En un fondo por unidades se escribe lo que trae el extracto: cuántas
	 * unidades entraron o salieron y a qué valor de unidad. «Retirar todo» vende
	 * todas las que guarda la posición.
	 *
	 * Corregir un movimiento no cambia de qué posición es ni hacia dónde fue: eso
	 * es borrarlo y anotar otro. Borrar o corregir recalcula todo lo que viene
	 * después.
	 */
	import { enhance } from '$app/forms';
	import Button from '$lib/ui/button.svelte';
	import DatePicker from '$lib/ui/date-picker.svelte';
	import Modal from '$lib/ui/modal.svelte';
	import { OptimisticDialog } from '$lib/shared/optimistic.svelte';
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import { formatCalendarDate, todayLocalDateString } from '$lib/shared/format/date';
	import type { Fund, FundMovement } from '../funds';
	import FundMovementHistory from './fund-movement-history.svelte';

	interface Option {
		id: string;
		name: string;
	}

	interface Props {
		fund: Fund | null;
		/** Los movimientos del fondo abierto, del más reciente. */
		movements: FundMovement[];
		portfolios: Option[];
		platforms: Option[];
		onClose: () => void;
	}

	let { fund, movements, portfolios, platforms, onClose }: Props = $props();

	const today = todayLocalDateString();

	const byBalance = $derived(fund?.tracking === 'balance');

	/* El movimiento que se está corrigiendo; `null` anota uno nuevo. */
	let editing = $state<FundMovement | null>(null);

	let kind = $state<'contribution' | 'withdrawal'>('contribution');
	let date = $derived.by(() => {
		void fund;
		return today;
	});
	let amount = $state<string | number | null>('');
	let units = $state<string | number | null>('');
	let unitValue = $state<string | number | null>('');
	let fees = $state<string | number | null>('');
	let balanceBefore = $state<string | number | null>('');
	let notes = $state('');
	let all = $state(false);

	/* Por defecto, donde ya está el fondo: casi siempre se aporta al mismo sitio. */
	let portfolioId = $derived(fund?.positions[0]?.portfolioId ?? portfolios[0]?.id ?? '');
	let sourceId = $derived(fund?.positions[0]?.sourceId ?? platforms[0]?.id ?? '');
	let entryId = $derived(fund?.positions[0]?.entryId ?? '');

	const canWithdraw = $derived((fund?.positions.length ?? 0) > 0);

	const dialog = new OptimisticDialog(() => fund);

	function clear() {
		editing = null;
		amount = '';
		units = '';
		unitValue = '';
		fees = '';
		balanceBefore = '';
		notes = '';
		all = false;
		kind = 'contribution';
		date = today;
	}

	function close() {
		dialog.reset();
		clear();
		onClose();
	}

	/* Un NUMERIC del backend sin los ceros de relleno: «10150.00000000» es 10150. */
	const plain = (v: string) => (v.includes('.') ? v.replace(/\.?0+$/, '') : v);

	/* Lleva el movimiento al formulario, con lo que dijo cuando se anotó. */
	function edit(m: FundMovement) {
		dialog.reset();
		editing = m;
		kind = m.kind;
		date = m.date.slice(0, 10);
		amount = plain(m.amount);
		units = plain(m.units);
		unitValue = plain(m.unitValue);
		fees = parseFloat(m.fees) > 0 ? plain(m.fees) : '';
		notes = m.notes;
		all = m.all;
	}

	const handler = dialog.submit({
		fallbackError: 'No pudimos guardar el movimiento.',
		onDone: close
	});

	const action = $derived(
		editing ? '?/updateMovement' : kind === 'contribution' ? '?/contribute' : '?/withdraw'
	);

	const money = (value: number) =>
		fund ? privacy.money(formatCurrency(value, fund.currency)) : String(value);

	const day = (iso: string) =>
		formatCalendarDate(iso, { day: 'numeric', month: 'short', year: 'numeric' });

	const position = $derived(fund?.positions.find((p) => p.entryId === entryId) ?? null);

	/* En un fondo por unidades, lo que suman: lo que el extracto llama «valor». */
	const unitsTyped = $derived(
		!editing && kind === 'withdrawal' && all && position
			? parseFloat(position.units)
			: parseFloat(String(units)) || 0
	);
	const total = $derived(unitsTyped * (parseFloat(String(unitValue)) || 0));

	/* Retirar todo propone lo que vale la posición: casi siempre es lo que llegó. */
	function takeAll() {
		if (!byBalance || !all || !position || !fund?.unitValue) return;
		if (!amount) amount = (parseFloat(position.units) * parseFloat(fund.unitValue)).toFixed(2);
	}
</script>

<Modal
	open={fund !== null}
	hidden={dialog.hidden}
	title={fund ? `Aportes y retiros de ${fund.name}` : 'Aportes y retiros'}
	description={byBalance
		? 'Escribe el dinero que entró o salió; Finexia calcula cuánto del saldo es aporte y cuánto es rendimiento.'
		: 'Escribe las unidades que entraron o salieron y su valor de unidad, como en tu extracto.'}
	size="md"
	onClose={close}
>
	{#if fund}
		<form method="POST" {action} class="rail-fields" use:enhance={handler}>
			<input type="hidden" name="id" value={fund.assetId} />
			<input type="hidden" name="tracking" value={fund.tracking} />

			{#if editing}
				<input type="hidden" name="txnId" value={editing.txnId} />
				<p class="editing">
					Corrigiendo el {editing.kind === 'contribution' ? 'aporte' : 'retiro'} del {day(
						editing.date
					)} · {editing.portfolioName} · {editing.sourceName}
					<button type="button" class="link" onclick={clear}>Anotar uno nuevo</button>
				</p>
			{:else}
				<fieldset class="choice">
					<legend class="visually-hidden">Qué quieres anotar</legend>
					<label class:on={kind === 'contribution'}>
						<input type="radio" value="contribution" bind:group={kind} />
						<span>Aporte</span>
					</label>
					<label class:on={kind === 'withdrawal'} class:off={!canWithdraw}>
						<input type="radio" value="withdrawal" bind:group={kind} disabled={!canWithdraw} />
						<span>Retiro</span>
					</label>
				</fieldset>

				{#if kind === 'contribution'}
					{#if portfolios.length > 1 || platforms.length > 1}
						<div class="pair">
							<div class="field">
								<label for="fund-mv-portfolio">Portafolio</label>
								<select id="fund-mv-portfolio" name="portfolioId" bind:value={portfolioId}>
									{#each portfolios as p (p.id)}
										<option value={p.id}>{p.name}</option>
									{/each}
								</select>
							</div>
							<div class="field">
								<label for="fund-mv-source">Plataforma</label>
								<select id="fund-mv-source" name="sourceId" bind:value={sourceId}>
									{#each platforms as p (p.id)}
										<option value={p.id}>{p.name}</option>
									{/each}
								</select>
							</div>
						</div>
					{:else}
						<input type="hidden" name="portfolioId" value={portfolioId} />
						<input type="hidden" name="sourceId" value={sourceId} />
					{/if}
				{:else if fund.positions.length > 1}
					<div class="field">
						<label for="fund-mv-entry">Sale de</label>
						<select id="fund-mv-entry" name="entryId" bind:value={entryId}>
							{#each fund.positions as p (p.entryId)}
								<option value={p.entryId}>{p.portfolioName} · {p.sourceName}</option>
							{/each}
						</select>
					</div>
				{:else}
					<input type="hidden" name="entryId" value={entryId} />
				{/if}
			{/if}

			{#if byBalance}
				<div class="pair">
					<div class="field">
						<label for="fund-mv-amount">{kind === 'contribution' ? 'Aportaste' : 'Retiraste'}</label
						>
						<input
							id="fund-mv-amount"
							name="amount"
							type="number"
							inputmode="decimal"
							step="any"
							min="0"
							bind:value={amount}
							required
						/>
					</div>
					<div class="field">
						<span class="field-label">El día</span>
						<DatePicker name="date" bind:value={date} required />
					</div>
				</div>
			{:else}
				<div class="pair">
					<div class="field">
						<label for="fund-mv-units">Unidades</label>
						<input
							id="fund-mv-units"
							name="units"
							type="number"
							inputmode="decimal"
							step="any"
							min="0"
							bind:value={units}
							disabled={!editing && kind === 'withdrawal' && all}
							placeholder={!editing && kind === 'withdrawal' && all && position
								? position.units
								: ''}
							required={editing !== null || kind === 'contribution' || !all}
						/>
					</div>
					<div class="field">
						<label for="fund-mv-unit-value">Valor de unidad</label>
						<input
							id="fund-mv-unit-value"
							name="unitValue"
							type="number"
							inputmode="decimal"
							step="any"
							min="0"
							placeholder={fund.unitValue ? plain(fund.unitValue) : ''}
							bind:value={unitValue}
							required
						/>
					</div>
				</div>
				<div class="field">
					<span class="field-label">El día</span>
					<DatePicker name="date" bind:value={date} required />
				</div>
				<p class="hint total" aria-live="polite">
					{#if total > 0}
						{kind === 'contribution' ? 'Aportaste' : 'Retiraste'} {money(total)}
					{:else}
						Unidades × valor de unidad: lo que el extracto llama valor del movimiento.
					{/if}
				</p>
			{/if}

			{#if kind === 'contribution'}
				{#if byBalance && !editing}
					<details class="more">
						<summary>Saldo justo antes <span class="optional">(opcional)</span></summary>
						<div class="field">
							<label for="fund-mv-before">Saldo al cierre del día anterior</label>
							<input
								id="fund-mv-before"
								name="balanceBefore"
								type="number"
								inputmode="decimal"
								step="any"
								min="0"
								bind:value={balanceBefore}
							/>
							<p class="hint">
								Sin él, el aporte entra al valor del último saldo que anotaste. Con él, entra al
								valor exacto de ese día.
							</p>
						</div>
					</details>
				{/if}
			{:else}
				<div class="pair">
					<div class="field">
						<label for="fund-mv-fees"
							>Se quedó la entidad <span class="optional">(opcional)</span></label
						>
						<input
							id="fund-mv-fees"
							name="fees"
							type="number"
							inputmode="decimal"
							step="any"
							min="0"
							placeholder="0"
							bind:value={fees}
						/>
					</div>
					{#if byBalance || !editing}
						<label class="check">
							<input type="checkbox" name="all" bind:checked={all} onchange={takeAll} />
							<span>Retirar todo</span>
						</label>
					{/if}
				</div>
				<p class="hint">
					Lo que se queda la entidad —una penalidad, un impuesto— cuenta como pérdida.
					{byBalance
						? '«Retirar todo» vacía la posición aunque el importe no cuadre al centavo con el último saldo.'
						: '«Retirar todo» vende todas las unidades que guarda la posición.'}
				</p>
			{/if}

			<div class="field">
				<label for="fund-mv-notes">Nota <span class="optional">(opcional)</span></label>
				<input id="fund-mv-notes" name="notes" type="text" maxlength="500" bind:value={notes} />
			</div>

			{#if dialog.error}
				<p class="feedback error" role="alert">{dialog.error}</p>
			{/if}

			<div class="modal-actions">
				<Button type="button" variant="ghost" onclick={close}>Cerrar</Button>
				<Button type="submit" loading={dialog.submitting}>
					{editing
						? 'Guardar corrección'
						: kind === 'contribution'
							? 'Anotar aporte'
							: 'Anotar retiro'}
				</Button>
			</div>
		</form>

		<FundMovementHistory {fund} {movements} editingId={editing?.txnId ?? null} onEdit={edit} />
	{/if}
</Modal>

<style>
	.visually-hidden {
		position: absolute;
		width: 1px;
		height: 1px;
		overflow: hidden;
		clip: rect(0 0 0 0);
		white-space: nowrap;
	}

	.choice {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 0.5rem;
		margin: 0;
		padding: 0;
		border: 0;
	}

	.choice label {
		display: flex;
		align-items: center;
		gap: 0.6rem;
		padding: 0.7rem 0.9rem;
		border: 1px solid var(--border);
		border-radius: 9px;
		font-size: 0.87rem;
		color: var(--text-muted);
		cursor: pointer;
	}

	.choice label.on {
		border-color: var(--amber);
		color: var(--text);
		background: rgba(212, 145, 42, 0.06);
	}

	.choice label.off {
		cursor: not-allowed;
		opacity: 0.5;
	}

	.choice input,
	.check input {
		accent-color: var(--amber);
	}

	.check {
		display: flex;
		align-items: center;
		gap: 0.6rem;
		align-self: end;
		padding-bottom: 0.8rem;
		font-size: 0.87rem;
		color: var(--text);
		cursor: pointer;
	}

	.more summary {
		font-size: 0.87rem;
		color: var(--text-muted);
		cursor: pointer;
	}

	.more .field {
		margin-top: 0.75rem;
	}

	.feedback {
		margin: 0;
	}

	.editing {
		display: flex;
		flex-wrap: wrap;
		align-items: baseline;
		gap: 0.35rem 0.75rem;
		margin: 0;
		padding: 0.6rem 0.8rem;
		border: 1px solid var(--amber);
		border-radius: 9px;
		background: rgba(212, 145, 42, 0.06);
		font-size: 0.85rem;
		color: var(--text);
	}

	.link {
		padding: 0;
		border: 0;
		background: none;
		font: inherit;
		font-size: 0.82rem;
		color: var(--amber);
		text-decoration: underline;
		cursor: pointer;
	}

	.total {
		margin: 0;
		font-variant-numeric: tabular-nums;
	}
</style>
