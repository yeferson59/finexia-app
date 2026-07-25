<script lang="ts">
	import type { Risk } from '$lib/api/types';

	let {
		risks,
		selected = $bindable(''),
		disabled = false
	}: { risks: Risk[]; selected?: string; disabled?: boolean } = $props();
</script>

<div class="form-group">
	<label class="label" for="risk">Nivel de Riesgo *</label>
	<fieldset class="risk-options">
		{#each risks as risk (risk.id)}
			<label class="radio-label">
				<input
					type="radio"
					name="riskId"
					value={risk.id}
					bind:group={selected}
					{disabled}
					required
				/>
				<span class="radio-content">
					<span class="radio-title">{risk.name}</span>
					<span class="radio-description">{risk.description}</span>
				</span>
			</label>
		{/each}
	</fieldset>
</div>

<style>
	.form-group {
		display: grid;
		gap: 0.6rem;
	}

	.label {
		font-size: 0.95rem;
		font-weight: 600;
		color: var(--text);
	}

	.risk-options {
		display: grid;
		gap: 1rem;
		border: none;
		padding: 0;
		margin: 0;
	}

	.radio-label {
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

	.radio-label:hover {
		background: var(--border);
		border-color: rgba(212, 145, 42, 0.3);
	}

	.radio-label input[type='radio'] {
		margin-top: 0.25rem;
		cursor: pointer;
		accent-color: var(--amber);
		width: 18px;
		height: 18px;
	}

	.radio-label input[type='radio']:disabled {
		cursor: not-allowed;
		opacity: 0.6;
	}

	.radio-content {
		display: flex;
		flex-direction: column;
		gap: 0.3rem;
	}

	.radio-title {
		font-weight: 600;
		color: var(--text);
	}

	.radio-description {
		font-size: 0.85rem;
		color: rgba(236, 234, 229, 0.5);
	}
</style>
