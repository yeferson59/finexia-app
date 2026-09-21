<script lang="ts">
	/**
	 * La fila de conversión del alta de una transacción: en qué moneda se
	 * ejecutó, a qué tasa liquidó la cuenta y cuánto costó al final.
	 *
	 * Vive aparte del formulario porque solo aparece cuando el activo cotiza en
	 * una moneda y la cuenta liquida en otra —la minoría de las operaciones—, y
	 * porque el coste en la moneda de la cuenta es lo que se contrasta contra la
	 * confirmación del bróker antes de guardar.
	 */
	import { formatSettled } from '../asset';

	let {
		fxRate = $bindable(''),
		tradeCurrency,
		assetCurrency,
		costCurrency,
		settledTotal
	}: {
		fxRate?: string;
		/** La moneda en la que va el precio de arriba. */
		tradeCurrency: string;
		assetCurrency: string;
		costCurrency: string;
		settledTotal: number;
	} = $props();
</script>

<div class="form-row fx-row">
	<div class="form-group">
		<span class="form-label">Moneda de la operación</span>
		<p class="fx-static">{tradeCurrency}</p>
		<p class="hint">
			{assetCurrency} es la moneda en la que cotiza el activo; el precio y la comisión de arriba van en
			ella.
		</p>
	</div>
	<div class="form-group">
		<label class="form-label" for="txn-fx"
			>Tasa a {costCurrency} <span class="required">*</span></label
		>
		<input
			id="txn-fx"
			type="number"
			class="form-input"
			name="fxRate"
			bind:value={fxRate}
			placeholder="1.0638"
			min="0"
			step="any"
			required
		/>
		<p class="hint">
			Cuántos {costCurrency} costaba 1 {assetCurrency} ese día, según la confirmación del bróker.
		</p>
	</div>
	<div class="form-group">
		<span class="form-label">Coste en {costCurrency}</span>
		<p class="fx-static">{formatSettled(settledTotal, costCurrency)}</p>
		<p class="hint">Contrástalo con el importe que te debitaron.</p>
	</div>
</div>

<style>
	.fx-row {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
		gap: 1rem;
		padding: 0.9rem;
		border: 1px dashed rgba(212, 145, 42, 0.3);
		border-radius: 8px;
		background: rgba(212, 145, 42, 0.04);
	}

	.form-group {
		display: flex;
		flex-direction: column;
		gap: 0.4rem;
	}

	.form-label {
		font-size: 0.87rem;
		font-weight: 500;
		color: var(--text);
	}

	.required {
		color: var(--red);
	}

	.fx-static {
		margin: 0;
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
		font-size: 0.95rem;
		font-weight: 700;
		color: var(--amber);
	}

	.form-input {
		padding: 0.6rem 0.85rem;
		border: 1.5px solid rgba(212, 145, 42, 0.2);
		border-radius: 8px;
		background: rgba(255, 255, 255, 0.04);
		color: var(--text);
		font-size: 0.9rem;
		font-family: var(--font-body);
		transition: border-color 0.2s ease;
	}

	.form-input:focus {
		outline: none;
		border-color: var(--amber);
	}
</style>
