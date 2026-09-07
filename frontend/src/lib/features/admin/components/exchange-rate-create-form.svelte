<script lang="ts">
	/**
	 * Alta manual de una tasa de cambio del catálogo compartido.
	 *
	 * El aviso de que el feed puede pisarla no es una nota al pie: es la razón
	 * por la que una tasa escrita aquí puede desaparecer sola, y se dice antes de
	 * escribirla y no después.
	 */
	import { enhance } from '$app/forms';
	import Button from '$lib/ui/button.svelte';

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

<form
	class="rail-fields"
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
	<div class="pair">
		<div class="field">
			<label for="fromCurrency">Moneda origen</label>
			<input
				id="fromCurrency"
				type="text"
				name="fromCurrency"
				placeholder="USD"
				maxlength="3"
				autocapitalize="characters"
				required
			/>
		</div>
		<div class="field">
			<label for="toCurrency">Moneda destino</label>
			<input
				id="toCurrency"
				type="text"
				name="toCurrency"
				placeholder="COP"
				maxlength="3"
				autocapitalize="characters"
				required
			/>
		</div>
	</div>

	<div class="field">
		<label for="rate">Cuánto vale una unidad de la moneda origen</label>
		<input
			id="rate"
			name="rate"
			type="number"
			placeholder="4000.00"
			min="0.00000001"
			step="any"
			required
		/>
		<p class="hint">
			Si el par es USD/COP, escribe cuántos pesos cuesta un dólar. El feed público reescribe la TRM
			cada hora, así que una tasa manual de ese par no dura.
		</p>
	</div>

	{#if error}
		<p class="feedback error" role="alert">{error}</p>
	{/if}

	<div class="actions">
		{#if onCancel}
			<Button type="button" variant="ghost" onclick={onCancel}>Cancelar</Button>
		{/if}
		<Button type="submit" loading={creating}>Crear tasa</Button>
	</div>
</form>

<style>
	.actions {
		display: flex;
		justify-content: flex-end;
		gap: 0.75rem;
		margin-top: 0.5rem;
	}
</style>
