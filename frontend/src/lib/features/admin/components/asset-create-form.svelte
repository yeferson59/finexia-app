<script lang="ts">
	/**
	 * Alta de un activo en el catálogo compartido.
	 *
	 * Crear aquí un ticker que ya aportó un usuario lo cura para todos, así que
	 * este formulario también es la vía para promover activos aportados.
	 */
	import { enhance } from '$app/forms';
	import Button from '$lib/ui/button.svelte';
	import AdminFormCard from './admin-form-card.svelte';
	import { ASSET_TYPES } from '../admin';

	interface Props {
		error?: string;
		/**
		 * Se llama cuando el envío sale bien. La página cierra el panel desde aquí
		 * y no desde el `form` común, que también cambia con el resto de actions.
		 */
		onSuccess?: () => void;
	}

	let { error = '', onSuccess }: Props = $props();

	let creating = $state(false);
</script>

<AdminFormCard title="Nuevo activo">
	<form
		method="POST"
		action="?/createAsset"
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
				<label class="field-label" for="ticker">Ticker <span class="required">*</span></label>
				<input id="ticker" name="ticker" class="field-input" placeholder="AAPL" required />
			</div>
			<div class="form-field">
				<label class="field-label" for="name">Nombre <span class="required">*</span></label>
				<input id="name" name="name" class="field-input" placeholder="Apple Inc." required />
			</div>
			<div class="form-field">
				<label class="field-label" for="assetType">Tipo <span class="required">*</span></label>
				<select id="assetType" name="assetType" class="field-input field-select" required>
					<option value="" disabled selected>Seleccionar tipo</option>
					{#each ASSET_TYPES as t (t.value)}
						<option value={t.value}>{t.label}</option>
					{/each}
				</select>
			</div>
			<div class="form-field">
				<label class="field-label" for="currency">Moneda <span class="required">*</span></label>
				<input
					id="currency"
					name="currency"
					class="field-input"
					placeholder="USD"
					maxlength="3"
					required
				/>
			</div>
			<div class="form-field">
				<label class="field-label" for="exchange"
					>Exchange <span class="optional">(opcional)</span></label
				>
				<input id="exchange" name="exchange" class="field-input" placeholder="NASDAQ" />
			</div>
		</div>
		{#if error}
			<p class="form-error">{error}</p>
		{/if}
		<div class="form-actions">
			<Button type="submit" loading={creating}>Crear activo</Button>
		</div>
	</form>
</AdminFormCard>
