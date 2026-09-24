<script lang="ts">
	/**
	 * Registrar un fondo: qué es, dónde está y la primera compra de unidades, como
	 * las trae el extracto.
	 *
	 * Lo que vale hoy una unidad es opcional pero es lo que se quiere: sin eso el
	 * fondo vale lo que costó hasta la primera marca. Con eso, un fondo comprado
	 * hace meses entra con lo que vale, y la ganancia que ya traía no se cuenta
	 * como rentabilidad del día en que se registra.
	 */
	import { enhance } from '$app/forms';
	import Button from '$lib/ui/button.svelte';
	import DatePicker from '$lib/ui/date-picker.svelte';
	import Modal from '$lib/ui/modal.svelte';
	import { OptimisticDialog } from '$lib/shared/optimistic.svelte';
	import { SUPPORTED_CURRENCIES } from '$lib/shared/currency';
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import { formatSignedPercent } from '$lib/shared/format/percent';
	import { todayLocalDateString } from '$lib/shared/format/date';

	interface Option {
		id: string;
		name: string;
	}

	interface Props {
		open: boolean;
		portfolios: (Option & { isDefault: boolean })[];
		platforms: Option[];
		/** La moneda que se propone: la de la cuenta del usuario. */
		currency: string;
		onClose: () => void;
	}

	let { open, portfolios, platforms, currency, onClose }: Props = $props();

	const today = todayLocalDateString();

	let name = $state('');
	let portfolioId = $derived(portfolios.find((p) => p.isDefault)?.id ?? portfolios[0]?.id ?? '');
	let sourceId = $derived(platforms[0]?.id ?? '');
	let fundCurrency = $derived(currency);
	let date = $state(today);
	let units = $state<string | number | null>('');
	let unitValue = $state<string | number | null>('');
	let currentUnitValue = $state<string | number | null>('');
	let currentDate = $state(today);

	const dialog = new OptimisticDialog(() => open);

	function close() {
		dialog.reset();
		onClose();
	}

	const handler = dialog.submit({
		fallbackError: 'No pudimos guardar el fondo.',
		onDone: close
	});

	const unitsN = $derived(parseFloat(String(units)) || 0);
	const cost = $derived(unitsN * (parseFloat(String(unitValue)) || 0));
	const current = $derived(parseFloat(String(currentUnitValue)) || 0);
	const worth = $derived(unitsN * current);
	const gainPct = $derived(cost > 0 && current > 0 ? (worth / cost - 1) * 100 : null);

	const money = (value: number) => privacy.money(formatCurrency(value, fundCurrency));
</script>

<Modal
	{open}
	hidden={dialog.hidden}
	title="Nuevo fondo"
	description="Un fondo de inversión colectiva, de pensiones voluntarias o un money market: unidades cuyo valor publica la entidad."
	size="md"
	onClose={close}
>
	{#if open}
		<form method="POST" action="?/createFund" class="rail-fields" use:enhance={handler}>
			<div class="field">
				<label for="fund-name">Nombre</label>
				<input
					id="fund-name"
					name="name"
					type="text"
					maxlength="100"
					placeholder="FIC Renta Fija"
					bind:value={name}
					required
				/>
			</div>

			<div class="pair">
				<div class="field">
					<label for="fund-source">Plataforma</label>
					<select id="fund-source" name="sourceId" bind:value={sourceId} required>
						{#each platforms as platform (platform.id)}
							<option value={platform.id}>{platform.name}</option>
						{/each}
					</select>
				</div>
				<div class="field">
					<label for="fund-currency">Moneda</label>
					<select id="fund-currency" name="currency" bind:value={fundCurrency} required>
						{#each SUPPORTED_CURRENCIES as code (code)}
							<option value={code}>{code}</option>
						{/each}
					</select>
				</div>
			</div>

			{#if portfolios.length > 1}
				<div class="field">
					<label for="fund-portfolio">Portafolio</label>
					<select id="fund-portfolio" name="portfolioId" bind:value={portfolioId} required>
						{#each portfolios as p (p.id)}
							<option value={p.id}>{p.name}</option>
						{/each}
					</select>
				</div>
			{:else}
				<input type="hidden" name="portfolioId" value={portfolioId} />
			{/if}

			<fieldset class="group">
				<legend>La compra</legend>
				<div class="pair">
					<div class="field">
						<label for="fund-units">Unidades</label>
						<input
							id="fund-units"
							name="units"
							type="number"
							inputmode="decimal"
							step="any"
							min="0"
							bind:value={units}
							required
						/>
					</div>
					<div class="field">
						<label for="fund-unit-value">Valor de unidad</label>
						<input
							id="fund-unit-value"
							name="unitValue"
							type="number"
							inputmode="decimal"
							step="any"
							min="0"
							bind:value={unitValue}
							required
						/>
					</div>
				</div>
				<div class="field">
					<span class="field-label">Día de la compra</span>
					<DatePicker name="date" bind:value={date} required />
					<p class="hint">
						Tu extracto trae las unidades y el valor de unidad de cada aporte. Si hiciste varios,
						registra el primero aquí y los demás como compras desde el portafolio.
					</p>
				</div>
			</fieldset>

			<fieldset class="group">
				<legend>Lo que vale hoy <span class="optional">(opcional)</span></legend>
				<div class="pair">
					<div class="field">
						<label for="fund-current">Valor de unidad</label>
						<input
							id="fund-current"
							name="currentUnitValue"
							type="number"
							inputmode="decimal"
							step="any"
							min="0"
							bind:value={currentUnitValue}
						/>
					</div>
					<div class="field">
						<span class="field-label">Al día</span>
						<DatePicker name="currentDate" bind:value={currentDate} />
					</div>
				</div>
				<p class="hint">
					Sin este valor, el fondo vale lo que costó hasta que lo actualices. Con él, lo que ya ganó
					antes de registrarlo no se cuenta como rentabilidad de hoy.
				</p>
			</fieldset>

			{#if cost > 0}
				<dl class="preview" aria-live="polite">
					<div>
						<dt>Costó</dt>
						<dd>{money(cost)}</dd>
					</div>
					{#if current > 0}
						<div>
							<dt>Vale hoy</dt>
							<dd>{money(worth)}</dd>
						</div>
						{#if gainPct !== null}
							<div>
								<dt>Rentabilidad</dt>
								<dd class:gain={gainPct >= 0} class:loss={gainPct < 0}>
									{formatSignedPercent(gainPct, 2)}
								</dd>
							</div>
						{/if}
					{/if}
				</dl>
			{/if}

			{#if dialog.error}
				<p class="feedback error" role="alert">{dialog.error}</p>
			{/if}

			<div class="modal-actions">
				<Button type="button" variant="ghost" onclick={close}>Cancelar</Button>
				<Button type="submit" loading={dialog.submitting}>Registrar fondo</Button>
			</div>
		</form>
	{/if}
</Modal>

<style>
	.group {
		display: grid;
		gap: 1rem;
		margin: 0;
		padding: 0;
		border: 0;
	}

	legend {
		margin-bottom: 0.25rem;
		padding: 0;
		font-size: 0.78rem;
		font-weight: 500;
		letter-spacing: 0.04em;
		text-transform: uppercase;
		color: var(--text-dim);
	}

	.preview {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(8rem, 1fr));
		gap: 0.75rem 1rem;
		margin: 0;
		padding: 1rem;
		border: 1px solid var(--border);
		border-radius: 10px;
		background: rgba(255, 255, 255, 0.02);
	}

	.preview div {
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

	dd.loss {
		color: var(--red);
	}

	.feedback {
		margin: 0;
	}
</style>
