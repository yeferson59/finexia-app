<script lang="ts">
	let {
		currency,
		targetAmount = $bindable(''),
		isDefault = $bindable(false),
		error,
		disabled = false
	}: {
		currency: string;
		targetAmount?: string;
		isDefault?: boolean;
		error?: string;
		disabled?: boolean;
	} = $props();
</script>

<fieldset class="form-section">
	<legend class="section-title">Objetivo Financiero</legend>

	<div class="form-group">
		<label for="targetAmount" class="label">Monto Objetivo (opcional)</label>
		<div class="input-with-prefix">
			<span class="prefix">{currency}</span>
			<input
				type="number"
				id="targetAmount"
				bind:value={targetAmount}
				placeholder="0.00"
				class="input"
				name="priceValue"
				class:error={!!error}
				{disabled}
				step="0.01"
				min="0"
			/>
		</div>
		{#if error}
			<span class="error-message">{error}</span>
		{/if}
		<p class="help-text">Define el monto que deseas alcanzar en este portafolio</p>
	</div>

	<div class="form-group">
		<label class="checkbox-label" for="isDefault">
			<input type="checkbox" id="isDefault" name="isDefault" bind:checked={isDefault} {disabled} />
			<span class="checkbox-content">
				<span class="checkbox-title">Marcar como portafolio por defecto</span>
				<span class="checkbox-description">
					Este portafolio se usará como selección predeterminada
				</span>
			</span>
		</label>
	</div>
</fieldset>

<style>
	.form-section {
		display: grid;
		gap: 1.5rem;
		padding: 1.5rem;
		border: 1px solid var(--border-strong);
		border-radius: 16px;
		background: var(--surface);
		backdrop-filter: blur(16px);
	}

	.section-title {
		margin: 0 0 0.5rem;
		font-size: 1.15rem;
		font-weight: 400;
		color: var(--amber-light);
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.form-group {
		display: grid;
		gap: 0.6rem;
	}

	.label {
		font-size: 0.95rem;
		font-weight: 600;
		color: var(--text);
	}

	.error-message {
		font-size: 0.8rem;
		color: var(--red);
	}

	.help-text {
		margin: 0;
		font-size: 0.8rem;
		color: rgba(236, 234, 229, 0.5);
	}

	.input-with-prefix {
		display: flex;
		align-items: center;
		border: 1px solid rgba(212, 145, 42, 0.2);
		border-radius: 10px;
		background: rgba(255, 255, 255, 0.022);
		overflow: hidden;
		transition: all 0.3s ease;
	}

	.input-with-prefix:focus-within {
		border-color: var(--amber);
		background: rgba(255, 255, 255, 0.022);
		box-shadow: 0 0 0 3px var(--border);
	}

	.prefix {
		padding: 0.85rem;
		color: var(--amber);
		font-weight: 600;
		border-right: 1px solid rgba(212, 145, 42, 0.2);
		background: var(--surface);
	}

	.input-with-prefix .input {
		flex: 1;
		padding: 0.85rem;
		border: none;
		background: transparent;
	}

	.input-with-prefix .input:focus {
		box-shadow: none;
	}

	.checkbox-label {
		display: flex;
		align-items: flex-start;
		gap: 1rem;
		padding: 1rem;
		border: 1px solid var(--border-strong);
		border-radius: 10px;
		background: rgba(255, 255, 255, 0.022);
		cursor: pointer;
		transition: all 0.3s ease;
	}

	.checkbox-label:hover {
		background: var(--border);
		border-color: rgba(212, 145, 42, 0.3);
	}

	.checkbox-label input[type='checkbox'] {
		margin-top: 0.2rem;
		cursor: pointer;
		accent-color: var(--amber);
		width: 18px;
		height: 18px;
	}

	.checkbox-label input[type='checkbox']:disabled {
		cursor: not-allowed;
		opacity: 0.6;
	}

	.checkbox-content {
		display: flex;
		flex-direction: column;
		gap: 0.3rem;
	}

	.checkbox-title {
		font-weight: 600;
		color: var(--text);
	}

	.checkbox-description {
		font-size: 0.85rem;
		color: rgba(236, 234, 229, 0.5);
	}
</style>
