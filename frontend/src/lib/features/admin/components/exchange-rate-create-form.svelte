<script lang="ts">
	/** Alta manual de una tasa de cambio del catálogo compartido. */
	import { enhance } from '$app/forms';
	import Button from '$lib/ui/button.svelte';
	import AdminFormFields from './admin-form-fields.svelte';

	interface Props {
		error?: string;
		/**
		 * Se llama cuando el envío sale bien. La página cierra el panel desde aquí
		 * y no desde el `form` común, que también cambia con el resto de actions.
		 */
		onSuccess?: () => void;
		/** Cierra el modal sin enviar. */
		onCancel?: () => void;
	}

	let { error = '', onSuccess, onCancel }: Props = $props();

	let creating = $state(false);
</script>

<AdminFormFields>
	<form
		method="POST"
		action="?/createRate"
		use:enhance={() => {
			creating = true;
			return async ({ result, update }) => {
				creating = false;
				await update();
				if (result.type === 'success') onSuccess?.();
			};
		}}
	>
		<div class="form-grid">
			<div class="form-field">
				<label class="field-label" for="fromCurrency"
					>Moneda origen <span class="required">*</span></label
				>
				<input
					id="fromCurrency"
					name="fromCurrency"
					class="field-input"
					placeholder="USD"
					maxlength="3"
					required
				/>
			</div>
			<div class="form-field">
				<label class="field-label" for="toCurrency"
					>Moneda destino <span class="required">*</span></label
				>
				<input
					id="toCurrency"
					name="toCurrency"
					class="field-input"
					placeholder="COP"
					maxlength="3"
					required
				/>
			</div>
			<div class="form-field">
				<label class="field-label" for="rate">Tasa <span class="required">*</span></label>
				<input
					id="rate"
					name="rate"
					type="number"
					class="field-input"
					placeholder="4000.00"
					min="0.00000001"
					step="any"
					required
				/>
			</div>
		</div>
		{#if error}
			<p class="form-error">{error}</p>
		{/if}
		<div class="form-actions">
			{#if onCancel}
				<Button type="button" variant="ghost" onclick={onCancel}>Cancelar</Button>
			{/if}
			<Button type="submit" loading={creating}>Crear tasa</Button>
		</div>
	</form>
</AdminFormFields>
