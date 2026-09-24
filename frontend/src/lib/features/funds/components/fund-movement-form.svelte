<script lang="ts">
	/**
	 * Aportar a un fondo que se sigue por saldo, o retirar de él, y la lista de lo
	 * que ya se anotó.
	 *
	 * Solo se habla de dinero. Las unidades las calcula Finexia al valor del último
	 * saldo anotado antes del día del movimiento, que es lo que separa lo aportado
	 * de lo ganado. Si ese saldo es viejo, «Saldo justo antes» lo hace exacto: es
	 * lo que tenía el fondo al cierre del día anterior.
	 *
	 * Borrar un movimiento recalcula todo lo que viene después, como corregirlo.
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

	let kind = $state<'contribution' | 'withdrawal'>('contribution');
	let date = $derived.by(() => {
		void fund;
		return today;
	});
	let amount = $state<string | number | null>('');
	let fees = $state<string | number | null>('');
	let balanceBefore = $state<string | number | null>('');
	let all = $state(false);

	/* Por defecto, donde ya está el fondo: casi siempre se aporta al mismo sitio. */
	let portfolioId = $derived(fund?.positions[0]?.portfolioId ?? portfolios[0]?.id ?? '');
	let sourceId = $derived(fund?.positions[0]?.sourceId ?? platforms[0]?.id ?? '');
	let entryId = $derived(fund?.positions[0]?.entryId ?? '');

	const canWithdraw = $derived((fund?.positions.length ?? 0) > 0);

	const dialog = new OptimisticDialog(() => fund);

	function close() {
		dialog.reset();
		amount = '';
		fees = '';
		balanceBefore = '';
		all = false;
		kind = 'contribution';
		onClose();
	}

	const handler = dialog.submit({
		fallbackError: 'No pudimos guardar el movimiento.',
		onDone: close
	});

	const money = (value: string) =>
		fund ? privacy.money(formatCurrency(parseFloat(value) || 0, fund.currency)) : value;

	const day = (iso: string) =>
		formatCalendarDate(iso, { day: 'numeric', month: 'short', year: 'numeric' });

	/* Retirar todo propone lo que vale la posición: casi siempre es lo que llegó. */
	function takeAll() {
		const position = fund?.positions.find((p) => p.entryId === entryId);
		if (all && position && fund?.unitValue && !amount) {
			amount = (parseFloat(position.units) * parseFloat(fund.unitValue)).toFixed(2);
		}
	}
</script>

<Modal
	open={fund !== null}
	hidden={dialog.hidden}
	title={fund ? `Aportes y retiros de ${fund.name}` : 'Aportes y retiros'}
	description="Escribe el dinero que entró o salió; Finexia calcula cuánto del saldo es aporte y cuánto es rendimiento."
	size="md"
	onClose={close}
>
	{#if fund}
		<form
			method="POST"
			action={kind === 'contribution' ? '?/contribute' : '?/withdraw'}
			class="rail-fields"
			use:enhance={handler}
		>
			<input type="hidden" name="id" value={fund.assetId} />

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

			<div class="pair">
				<div class="field">
					<label for="fund-mv-amount">{kind === 'contribution' ? 'Aportaste' : 'Retiraste'}</label>
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

			{#if kind === 'contribution'}
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
							Sin él, el aporte entra al valor del último saldo que anotaste. Con él, entra al valor
							exacto de ese día.
						</p>
					</div>
				</details>
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
					<label class="check">
						<input type="checkbox" name="all" bind:checked={all} onchange={takeAll} />
						<span>Retirar todo</span>
					</label>
				</div>
				<p class="hint">
					Lo que se queda la entidad —una penalidad, un impuesto— cuenta como pérdida. «Retirar
					todo» vacía la posición aunque el importe no cuadre al centavo con el último saldo.
				</p>
			{/if}

			{#if dialog.error}
				<p class="feedback error" role="alert">{dialog.error}</p>
			{/if}

			<div class="modal-actions">
				<Button type="button" variant="ghost" onclick={close}>Cerrar</Button>
				<Button type="submit" loading={dialog.submitting}>
					{kind === 'contribution' ? 'Anotar aporte' : 'Anotar retiro'}
				</Button>
			</div>
		</form>

		<section class="history" aria-labelledby="fund-movements-title">
			<h3 id="fund-movements-title">Movimientos</h3>
			{#if movements.length === 0}
				<p class="hint">Todavía no hay ninguno.</p>
			{:else}
				<ul>
					{#each movements as m (m.txnId)}
						<li>
							<span class="when">{day(m.date)}</span>
							<span class="what" class:out={m.kind === 'withdrawal'}>
								{m.kind === 'contribution' ? 'Aporte' : m.all ? 'Retiro total' : 'Retiro'}
							</span>
							<span class="amount">
								{m.kind === 'withdrawal' ? '−' : '+'}{money(m.amount)}
							</span>
							<form method="POST" action="?/deleteMovement" use:enhance>
								<input type="hidden" name="txnId" value={m.txnId} />
								<button
									type="submit"
									class="remove"
									aria-label="Borrar el movimiento del {day(m.date)}"
								>
									Borrar
								</button>
							</form>
						</li>
					{/each}
				</ul>
			{/if}
		</section>
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
		grid-template-columns: minmax(6rem, 1fr) minmax(5rem, 1fr) minmax(6rem, 1.2fr) auto;
		align-items: center;
		gap: 0.75rem;
		padding: 0.4rem 0;
		font-size: 0.85rem;
	}

	.when {
		color: var(--text-muted);
	}

	/* Sin verde: un aporte es dinero que metiste, no ganancia. */
	.what {
		color: var(--text);
	}

	.what.out {
		color: var(--text-muted);
	}

	.amount {
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

	@media (max-width: 480px) {
		.history li {
			grid-template-columns: 1fr auto;
		}
	}
</style>
