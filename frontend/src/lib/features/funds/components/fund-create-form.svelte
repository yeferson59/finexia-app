<script lang="ts">
	/**
	 * Registrar un fondo: qué es, dónde está y cómo se sigue, con su primera
	 * compra.
	 *
	 * La primera pregunta decide el resto: ¿el extracto trae unidades? Si sí, se
	 * escriben las unidades y el valor de unidad tal como aparecen. Si la app solo
	 * enseña el saldo —lo normal en un neobanco o un fondo de pensiones—, se
	 * escribe lo que se metió y lo que dice hoy, y Finexia lleva las unidades por
	 * dentro para separar lo aportado de lo ganado.
	 *
	 * Lo de hoy es opcional pero es lo que se quiere: sin eso el fondo vale lo que
	 * costó hasta la primera actualización. Con eso, un fondo de hace meses entra
	 * con lo que vale, y lo que ya había ganado no se cuenta como rentabilidad del
	 * día en que se registra.
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

	let tracking = $state<'units' | 'balance'>('units');
	let name = $state('');
	let portfolioId = $derived(portfolios.find((p) => p.isDefault)?.id ?? portfolios[0]?.id ?? '');
	let sourceId = $derived(platforms[0]?.id ?? '');
	let fundCurrency = $derived(currency);
	let date = $state(today);
	let units = $state<string | number | null>('');
	let unitValue = $state<string | number | null>('');
	let amount = $state<string | number | null>('');
	let current = $state<string | number | null>('');
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

	const num = (v: string | number | null) => parseFloat(String(v)) || 0;

	/* Lo que costó y lo que vale hoy, en los dos modos: por unidades se multiplica,
	   por saldo se lee tal cual. */
	const cost = $derived(tracking === 'units' ? num(units) * num(unitValue) : num(amount));
	const worth = $derived(tracking === 'units' ? num(units) * num(current) : num(current));
	const gainPct = $derived(cost > 0 && worth > 0 ? (worth / cost - 1) * 100 : null);

	const money = (value: number) => privacy.money(formatCurrency(value, fundCurrency));
</script>

<Modal
	{open}
	hidden={dialog.hidden}
	title="Nuevo fondo"
	description="Un fondo de inversión colectiva, de pensiones voluntarias o un money market: no rinde una tasa fija, sino lo que resulte."
	size="md"
	onClose={close}
>
	{#if open}
		<form method="POST" action="?/createFund" class="rail-fields" use:enhance={handler}>
			<fieldset class="choice">
				<legend>¿Tu extracto muestra unidades?</legend>
				<label class:on={tracking === 'units'}>
					<input type="radio" name="tracking" value="units" bind:group={tracking} />
					<span>Sí, unidades y valor de unidad</span>
				</label>
				<label class:on={tracking === 'balance'}>
					<input type="radio" name="tracking" value="balance" bind:group={tracking} />
					<span>No, solo el saldo</span>
				</label>
			</fieldset>

			<div class="field">
				<label for="fund-name">Nombre</label>
				<input
					id="fund-name"
					name="name"
					type="text"
					maxlength="100"
					placeholder={tracking === 'units' ? 'FIC Renta Fija' : 'Bolsillo de inversión'}
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

			{#if tracking === 'units'}
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
							Si hiciste varios aportes, registra el primero aquí y los demás como compras desde el
							portafolio.
						</p>
					</div>
				</fieldset>
			{:else}
				<fieldset class="group">
					<legend>Lo que metiste</legend>
					<div class="pair">
						<div class="field">
							<label for="fund-amount">Aportaste</label>
							<input
								id="fund-amount"
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
							<span class="field-label">Desde el</span>
							<DatePicker name="date" bind:value={date} required />
						</div>
					</div>
					<p class="hint">
						Si ya lo tenías de antes, pon el total que has metido y el día del primer aporte. Los
						aportes y retiros que vengan los anotas desde la tarjeta del fondo.
					</p>
				</fieldset>
			{/if}

			<fieldset class="group">
				<legend>Lo que vale hoy <span class="optional">(opcional)</span></legend>
				<div class="pair">
					<div class="field">
						<label for="fund-current">{tracking === 'units' ? 'Valor de unidad' : 'Saldo'}</label>
						<input
							id="fund-current"
							name={tracking === 'units' ? 'currentUnitValue' : 'currentBalance'}
							type="number"
							inputmode="decimal"
							step="any"
							min="0"
							bind:value={current}
						/>
					</div>
					<div class="field">
						<span class="field-label">Al día</span>
						<DatePicker name="currentDate" bind:value={currentDate} />
					</div>
				</div>
				<p class="hint">
					Sin este dato, el fondo vale lo que costó hasta que lo actualices. Con él, lo que ya ganó
					antes de registrarlo no se cuenta como rentabilidad de hoy.
				</p>
			</fieldset>

			{#if cost > 0}
				<dl class="preview" aria-live="polite">
					<div>
						<dt>{tracking === 'units' ? 'Costó' : 'Aportado'}</dt>
						<dd>{money(cost)}</dd>
					</div>
					{#if worth > 0}
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
	.choice {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(12rem, 1fr));
		gap: 0.5rem;
		margin: 0;
		padding: 0;
		border: 0;
	}

	.choice legend {
		margin-bottom: 0.5rem;
		padding: 0;
		font-size: 0.87rem;
		font-weight: 500;
		color: var(--text);
	}

	.choice label {
		display: flex;
		align-items: center;
		gap: 0.6rem;
		padding: 0.75rem 0.9rem;
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

	.choice input {
		accent-color: var(--amber);
	}

	.group {
		display: grid;
		gap: 1rem;
		margin: 0;
		padding: 0;
		border: 0;
	}

	.group legend {
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
