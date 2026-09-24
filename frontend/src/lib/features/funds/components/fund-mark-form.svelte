<script lang="ts">
	/**
	 * Actualizar el valor de un fondo: el valor de unidad de un día, leído del
	 * extracto —o el saldo, en un fondo que se sigue por saldo—, y la lista de
	 * los que ya se escribieron.
	 *
	 * En un fondo por saldo, lo que se compara de un saldo a otro no es el saldo:
	 * un aporte lo sube sin que el fondo haya ganado nada. Es el valor de unidad
	 * que Finexia lleva por dentro, así que la rentabilidad de cada saldo se ve en
	 * la lista después de guardarlo.
	 *
	 * La fecha es la del extracto, no la de hoy. Una marca con fecha pasada
	 * corrige la gráfica de crecimiento desde ese día, así que la ganancia cae el
	 * día en que pasó y no el día en que se anotó. Otra marca el mismo día
	 * reemplaza a la que había.
	 */
	import { enhance } from '$app/forms';
	import Button from '$lib/ui/button.svelte';
	import DatePicker from '$lib/ui/date-picker.svelte';
	import Modal from '$lib/ui/modal.svelte';
	import { OptimisticDialog } from '$lib/shared/optimistic.svelte';
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import { formatSignedPercent } from '$lib/shared/format/percent';
	import { formatCalendarDate, todayLocalDateString } from '$lib/shared/format/date';
	import {
		canLinkFund,
		markBefore,
		markChange,
		marksWithChange,
		type Fund,
		type FundMark
	} from '../funds';
	import FundLink from './fund-link.svelte';
	import FundMarksPaste from './fund-marks-paste.svelte';

	interface Props {
		fund: Fund | null;
		/** Las marcas del fondo abierto, tal como las trae la página. */
		marks: FundMark[];
		onClose: () => void;
	}

	let { fund, marks, onClose }: Props = $props();

	const today = todayLocalDateString();

	let date = $derived.by(() => {
		void fund;
		return today;
	});
	let unitValue = $state<string | number | null>('');

	const dialog = new OptimisticDialog(() => fund);

	function close() {
		dialog.reset();
		unitValue = '';
		onClose();
	}

	const handler = dialog.submit({
		fallbackError: 'No pudimos guardar el valor.',
		onDone: close
	});

	const history = $derived(marksWithChange(marks));

	const byBalance = $derived(fund?.tracking === 'balance');

	/* Contra qué se compara el valor que se escribe: la marca anterior a su día. */
	const previous = $derived(markBefore(marks, date));
	const typed = $derived(parseFloat(String(unitValue)) || 0);
	const change = $derived(
		!byBalance && previous && typed > 0
			? markChange(previous, { date, unitValue: String(typed) })
			: null
	);
	const replaces = $derived(marks.some((m) => m.date.slice(0, 10) === date));

	const price = (value: string) =>
		fund ? privacy.money(formatCurrency(parseFloat(value) || 0, fund.currency, 6)) : value;

	const money = (value: string) =>
		fund ? privacy.money(formatCurrency(parseFloat(value) || 0, fund.currency)) : value;

	const day = (iso: string) =>
		formatCalendarDate(iso, { day: 'numeric', month: 'short', year: 'numeric' });

	/* Un fondo que ya nadie guarda se puede dejar de seguir; con posiciones, no. */
	const canDelete = $derived(fund !== null && fund.positions.length === 0);
</script>

<Modal
	open={fund !== null}
	hidden={dialog.hidden}
	title={fund ? `Actualizar ${fund.name}` : 'Actualizar valor'}
	description={byBalance
		? 'Escribe el saldo que muestra tu app, con su fecha.'
		: fund?.publicFund
			? 'El valor de unidad llega solo desde la Superfinanciera. Si tu extracto dice otro, escríbelo: el tuyo manda.'
			: 'Escribe el valor de unidad que trae tu extracto, con su fecha.'}
	size="md"
	onClose={close}
>
	{#if fund}
		<form method="POST" action="?/saveMark" class="rail-fields" use:enhance={handler}>
			<input type="hidden" name="id" value={fund.assetId} />

			<div class="pair">
				<div class="field">
					<label for="fund-mark-value">{byBalance ? 'Saldo' : 'Valor de unidad'}</label>
					<input
						id="fund-mark-value"
						name={byBalance ? 'balance' : 'unitValue'}
						type="number"
						inputmode="decimal"
						step="any"
						min="0"
						bind:value={unitValue}
						required
					/>
				</div>
				<div class="field">
					<span class="field-label">Día del extracto</span>
					<DatePicker name="date" bind:value={date} required />
				</div>
			</div>

			<p class="hint" aria-live="polite">
				{#if change}
					<span class:gain={change.pct >= 0} class:loss={change.pct < 0}>
						{formatSignedPercent(change.pct, 2)}
					</span>
					en {change.days}
					{change.days === 1 ? 'día' : 'días'}, desde el valor del {day(previous?.date ?? '')}.
				{:else if byBalance}
					El saldo al cierre de ese día, con los aportes y retiros de ese día ya dentro. Si es de un
					día pasado, la gráfica de crecimiento se corrige desde ese día.
				{:else}
					Si es de un día pasado, la gráfica de crecimiento se corrige desde ese día.
				{/if}
				{#if replaces}
					Ya hay un valor ese día: este lo reemplaza.
				{/if}
			</p>

			{#if dialog.error}
				<p class="feedback error" role="alert">{dialog.error}</p>
			{/if}

			<div class="modal-actions">
				<Button type="button" variant="ghost" onclick={close}>Cerrar</Button>
				<Button type="submit" loading={dialog.submitting}>Guardar valor</Button>
			</div>
		</form>

		<section class="history" aria-labelledby="fund-marks-title">
			<h3 id="fund-marks-title">{byBalance ? 'Saldos anotados' : 'Valores anotados'}</h3>
			{#if byBalance && history.length > 1}
				<p class="hint legend">
					El porcentaje es lo que rindió desde el saldo anterior, sin contar aportes ni retiros.
				</p>
			{/if}
			{#if history.length === 0}
				<p class="hint">Todavía no hay ninguno: el fondo vale lo que costó.</p>
			{:else}
				<ul>
					{#each history as { mark, change: delta } (mark.date)}
						<li>
							<span class="when">
								{day(mark.date)}
								{#if mark.source === 'public'}
									<span class="source" title="Publicado por la Superintendencia Financiera"
										>SFC</span
									>
								{/if}
							</span>
							<span class="price">
								{byBalance && mark.balance ? money(mark.balance) : price(mark.unitValue)}
							</span>
							<span
								class="delta"
								class:gain={delta && delta.pct >= 0}
								class:loss={delta && delta.pct < 0}
							>
								{delta ? formatSignedPercent(delta.pct, 2) : '—'}
							</span>
							<form method="POST" action="?/deleteMark" use:enhance>
								<input type="hidden" name="id" value={fund.assetId} />
								<input type="hidden" name="date" value={mark.date.slice(0, 10)} />
								<button
									type="submit"
									class="remove"
									aria-label="Borrar el valor del {day(mark.date)}"
								>
									Borrar
								</button>
							</form>
						</li>
					{/each}
				</ul>
			{/if}
		</section>

		<FundMarksPaste assetId={fund.assetId} {byBalance} />

		{#if fund.publicFund || canLinkFund(fund)}
			<FundLink {fund} />
		{/if}

		{#if canDelete}
			<form method="POST" action="?/deleteFund" class="danger-zone" use:enhance={handler}>
				<input type="hidden" name="id" value={fund.assetId} />
				<p class="danger-note">
					Ningún portafolio guarda este fondo. Puedes dejar de seguirlo; sus valores se borran con
					él.
				</p>
				<button type="submit" class="danger-link">Dejar de seguirlo</button>
			</form>
		{/if}
	{/if}
</Modal>

<style>
	.gain {
		color: var(--green);
	}

	.loss {
		color: var(--red);
	}

	.feedback {
		margin: 0;
	}

	.history {
		margin-top: 1.5rem;
		padding-top: 1.25rem;
		border-top: 1px solid var(--border);
	}

	.history h3 {
		margin: 0 0 0.75rem;
		font-size: 0.78rem;
		font-weight: 500;
		letter-spacing: 0.04em;
		text-transform: uppercase;
		color: var(--text-dim);
	}

	.legend {
		margin-bottom: 0.6rem;
	}

	.history ul {
		display: grid;
		gap: 0.35rem;
		max-height: 16rem;
		margin: 0;
		padding: 0;
		overflow-y: auto;
		list-style: none;
	}

	.history li {
		display: grid;
		grid-template-columns: minmax(6rem, 1fr) minmax(6rem, 1.2fr) 5rem auto;
		align-items: center;
		gap: 0.75rem;
		padding: 0.4rem 0;
		font-size: 0.85rem;
	}

	.when {
		color: var(--text-muted);
	}

	/* Un valor que publicó la Superfinanciera, no uno del extracto. */
	.source {
		margin-left: 0.3rem;
		padding: 0 0.3rem;
		border: 1px solid var(--border-strong);
		border-radius: 4px;
		font-size: 0.66rem;
		color: var(--text-dim);
	}

	.price,
	.delta {
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
		text-align: right;
		overflow-wrap: anywhere;
	}

	.remove {
		padding: 0.2rem 0.5rem;
		border: 1px solid var(--border);
		border-radius: 6px;
		background: transparent;
		font: inherit;
		font-size: 0.75rem;
		color: var(--text-dim);
		cursor: pointer;
	}

	.remove:hover {
		border-color: var(--red);
		color: var(--red);
	}

	.danger-zone {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		justify-content: space-between;
		gap: 0.5rem 1rem;
		margin-top: 1.25rem;
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

	.danger-link:hover {
		border-color: var(--red);
		background: rgba(224, 90, 90, 0.08);
	}

	@media (max-width: 480px) {
		.history li {
			grid-template-columns: 1fr auto;
		}

		.delta {
			text-align: left;
		}
	}
</style>
