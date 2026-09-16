<script lang="ts">
	/**
	 * Lo que la entidad rara vez cambia de una tasa, plegado: la retención, los
	 * tramos por saldo y el conversor de una tasa nominal.
	 *
	 * Los tramos se editan como una tabla corta —desde qué saldo y a qué tasa—
	 * en vez de como pares de campos sueltos: se leen de arriba abajo como los
	 * publica la entidad. Un tramo al 0 % es un tope.
	 *
	 * El conversor no se envía: solo rellena la tasa efectiva anual del
	 * formulario, que es lo que se guarda siempre.
	 */
	import { untrack } from 'svelte';
	import { annualFromNominal, formatAnnualRate, NOMINAL_PERIODS } from '../rates';
	import CashMoneyInput from './cash-money-input.svelte';

	/** Una fila de tramo como la escribe el usuario. */
	export interface TierRow {
		fromBalance: string;
		annualRatePct: string;
	}

	interface Props {
		currency: string;
		withholdingPct: string | number | null;
		tiers: TierRow[];
		/** Pasa al campo de la tasa principal la efectiva que sale del conversor. */
		onConvert: (annualRatePct: string) => void;
	}

	let { currency, withholdingPct = $bindable(), tiers = $bindable(), onConvert }: Props = $props();

	/** Los que acepta el backend en una versión. */
	const MAX_TIERS = 10;

	/*
	 * Las filas se reemplazan enteras en vez de mutarse: el padre las guarda en
	 * un `$derived` reasignable, que avisa cuando se reasigna, no cuando cambia
	 * algo por dentro.
	 */
	function setTier(index: number, field: keyof TierRow, value: string) {
		tiers = tiers.map((tier, i) => (i === index ? { ...tier, [field]: value } : tier));
	}

	function addTier(rate = '') {
		tiers = [...tiers, { fromBalance: '', annualRatePct: rate }];
	}

	function removeTier(index: number) {
		tiers = tiers.filter((_, i) => i !== index);
	}

	let nominalPct = $state<string | number | null>('');
	let nominalPeriods = $state(12);
	const converted = $derived(
		annualFromNominal(parseFloat(String(nominalPct)) || 0, nominalPeriods)
	);

	/*
	 * Abierto de entrada si ya hay algo dentro: plegado escondería una retención o
	 * unos tramos que cambian lo que rinde la cuenta. Solo al abrirse: quitar el
	 * último tramo no puede plegarlo en la mano de quien lo está editando.
	 */
	const startOpen = untrack(
		() => (parseFloat(String(withholdingPct)) || 0) > 0 || tiers.length > 0
	);
</script>

<details class="advanced" open={startOpen}>
	<summary>
		<svg viewBox="0 0 16 16" aria-hidden="true"><path d="M6 4l4 4-4 4" /></svg>
		Retención, tramos y tasa nominal
	</summary>

	<div class="body">
		<div class="field withholding">
			<label for="cash-rate-withholding">Retención <span class="optional">(opcional)</span></label>
			<CashMoneyInput
				id="cash-rate-withholding"
				name="withholdingPct"
				unit="%"
				max="99.99"
				bind:value={withholdingPct}
				aria-describedby="cash-rate-withholding-hint"
			/>
			<p class="hint" id="cash-rate-withholding-hint">
				Si la entidad te la descuenta, lo que rinde es neto.
			</p>
		</div>

		<fieldset class="tiers">
			<legend class="field-label">Tramos <span class="optional">(opcional)</span></legend>
			<p class="hint">
				Si la entidad paga distinto según el saldo. La tasa de arriba rige desde cero y cada tramo
				desde el saldo que escribas. Un tramo al 0 % es un tope: desde ahí la cuenta no rinde.
			</p>

			{#if tiers.length > 0}
				<div class="tier-table" role="group" aria-label="Tramos">
					<span class="tier-head" aria-hidden="true">Desde el saldo</span>
					<span class="tier-head" aria-hidden="true">Rinde</span>
					<span aria-hidden="true"></span>
					{#each tiers as tier, i (i)}
						<div class="tier-cell">
							<label class="sr-only" for="cash-rate-tier-from-{i}">
								Desde qué saldo rige el tramo {i + 1}
							</label>
							<CashMoneyInput
								id="cash-rate-tier-from-{i}"
								name="tierFromBalance"
								unit={currency}
								value={tier.fromBalance}
								oninput={(event) => setTier(i, 'fromBalance', event.currentTarget.value)}
							/>
						</div>
						<div class="tier-cell">
							<label class="sr-only" for="cash-rate-tier-rate-{i}">
								Tasa efectiva anual del tramo {i + 1}
							</label>
							<CashMoneyInput
								id="cash-rate-tier-rate-{i}"
								name="tierAnnualRatePct"
								unit="% E.A."
								max="100"
								value={tier.annualRatePct}
								oninput={(event) => setTier(i, 'annualRatePct', event.currentTarget.value)}
							/>
						</div>
						<button type="button" class="remove" onclick={() => removeTier(i)}>
							<svg viewBox="0 0 16 16" aria-hidden="true"><path d="M4 4l8 8M12 4l-8 8" /></svg>
							<span class="sr-only">Quitar el tramo {i + 1}</span>
						</button>
					{/each}
				</div>
			{/if}

			<div class="tier-actions">
				<button
					type="button"
					class="tool"
					disabled={tiers.length >= MAX_TIERS}
					onclick={() => addTier()}
				>
					Agregar tramo
				</button>
				<button
					type="button"
					class="tool"
					disabled={tiers.length >= MAX_TIERS}
					onclick={() => addTier('0')}
				>
					Agregar tope
				</button>
			</div>
		</fieldset>

		<div class="converter">
			<p class="hint">
				¿La entidad publica una tasa nominal? Escríbela y la pasamos a efectiva anual.
			</p>
			<div class="converter-row">
				<div>
					<label class="sr-only" for="cash-rate-nominal">Tasa nominal</label>
					<CashMoneyInput
						id="cash-rate-nominal"
						unit="% N.A."
						placeholder="12"
						bind:value={nominalPct}
					/>
				</div>
				<div>
					<label class="sr-only" for="cash-rate-periods">Capitaliza</label>
					<select id="cash-rate-periods" bind:value={nominalPeriods}>
						{#each NOMINAL_PERIODS as period (period.value)}
							<option value={period.value}>{period.label}</option>
						{/each}
					</select>
				</div>
				<button
					type="button"
					class="tool use"
					disabled={converted <= 0}
					onclick={() => onConvert(String(Number(converted.toFixed(4))))}
				>
					{converted > 0 ? `Usar ${formatAnnualRate(converted.toFixed(4))}` : 'Convertir'}
				</button>
			</div>
		</div>
	</div>
</details>

<style>
	.advanced {
		border-top: 1px solid var(--border);
		border-bottom: 1px solid var(--border);
	}

	summary {
		display: flex;
		align-items: center;
		gap: 0.45rem;
		padding: 0.8rem 0;
		font-size: 0.86rem;
		color: var(--text-muted);
		cursor: pointer;
		list-style: none;
	}

	summary::-webkit-details-marker {
		display: none;
	}

	summary:hover {
		color: var(--text);
	}

	summary svg {
		width: 0.8rem;
		height: 0.8rem;
		fill: none;
		stroke: currentColor;
		stroke-width: 1.75;
		stroke-linecap: round;
		stroke-linejoin: round;
		transition: transform 0.2s ease;
	}

	.advanced[open] summary svg {
		transform: rotate(90deg);
	}

	.body {
		display: grid;
		gap: 1.35rem;
		padding: 0.25rem 0 1.25rem;
	}

	.withholding {
		max-width: 14rem;
	}

	.tiers {
		display: grid;
		gap: 0.6rem;
		min-width: 0;
		margin: 0;
		padding: 0;
		border: none;
	}

	.tiers legend {
		margin-bottom: 0.45rem;
		padding: 0;
	}

	.tier-table {
		display: grid;
		grid-template-columns: minmax(0, 1.4fr) minmax(0, 1fr) 2.25rem;
		align-items: center;
		gap: 0.45rem;
	}

	.tier-head {
		font-size: 0.74rem;
		color: var(--text-dim);
	}

	.tier-cell {
		min-width: 0;
	}

	.remove {
		display: grid;
		place-items: center;
		width: 2.25rem;
		height: 2.25rem;
		padding: 0;
		border: none;
		border-radius: 7px;
		background: transparent;
		color: var(--text-dim);
		cursor: pointer;
	}

	.remove:hover {
		background: var(--surface-2);
		color: var(--red);
	}

	.remove svg {
		width: 0.85rem;
		height: 0.85rem;
		fill: none;
		stroke: currentColor;
		stroke-width: 1.6;
		stroke-linecap: round;
	}

	.tier-actions {
		display: flex;
		flex-wrap: wrap;
		gap: 0.4rem;
	}

	.tool {
		padding: 0.45rem 0.8rem;
		border: 1px solid var(--border-strong);
		border-radius: 7px;
		background: transparent;
		font: inherit;
		font-size: 0.82rem;
		white-space: nowrap;
		color: var(--text-muted);
		cursor: pointer;
	}

	.tool:hover:not(:disabled) {
		border-color: var(--amber);
		color: var(--amber-light);
	}

	.tool:disabled {
		cursor: default;
		opacity: 0.45;
	}

	.converter {
		display: grid;
		gap: 0.55rem;
	}

	.converter-row {
		display: grid;
		grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) auto;
		align-items: center;
		gap: 0.45rem;
	}

	.use:not(:disabled) {
		border-color: rgba(212, 145, 42, 0.45);
		color: var(--amber);
	}

	@media (max-width: 640px) {
		.converter-row {
			grid-template-columns: minmax(0, 1fr);
		}

		.withholding {
			max-width: none;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		summary svg {
			transition: none;
		}
	}
</style>
