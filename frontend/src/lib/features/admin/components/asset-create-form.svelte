<script lang="ts">
	/**
	 * Alta de un activo en el catálogo compartido.
	 *
	 * Crear aquí un ticker que ya aportó un usuario lo cura para todos, así que
	 * este formulario también es la vía para promover activos aportados.
	 *
	 * Los campos son los de cualquier formulario del producto —los de
	 * `routes/layout.css`— y no una segunda familia con etiquetas en versalitas
	 * mono: el mismo activo se da de alta desde una cartera con esos mismos
	 * campos, y no había razón para que aquí se vieran de otra manera.
	 */
	import { enhance } from '$app/forms';
	import Button from '$lib/ui/button.svelte';
	import { SECTOR_OPTIONS, typeHasSector } from '$lib/shared/format/sector';
	import { ASSET_TYPES } from '../admin';

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

	/*
	 * El tipo se sigue desde aquí porque la industria depende de él: detrás de
	 * una cripto, un saldo o un inmueble no hay empresa que clasificar, y el
	 * backend contesta 400 a quien lo intente. Enseñar el campo y que el
	 * servidor lo rechace sería pedir un dato para luego negarlo; mejor no
	 * ofrecerlo.
	 */
	let assetType = $state('');
	const classifiable = $derived(typeHasSector(assetType));
</script>

<form
	class="rail-fields"
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
	<div class="pair">
		<div class="field">
			<label for="ticker">Ticker</label>
			<input
				id="ticker"
				type="text"
				name="ticker"
				placeholder="AAPL"
				autocapitalize="characters"
				required
			/>
		</div>
		<div class="field">
			<label for="currency">Moneda</label>
			<input
				id="currency"
				type="text"
				name="currency"
				placeholder="USD"
				maxlength="3"
				autocapitalize="characters"
				required
			/>
		</div>
	</div>

	<div class="field">
		<label for="name">Nombre</label>
		<input id="name" type="text" name="name" placeholder="Apple Inc." required />
	</div>

	<div class="pair">
		<div class="field">
			<label for="assetType">Tipo</label>
			<select id="assetType" name="assetType" bind:value={assetType} required>
				<option value="" disabled selected>Elige un tipo</option>
				{#each ASSET_TYPES as t (t.value)}
					<option value={t.value}>{t.label}</option>
				{/each}
			</select>
		</div>
		<div class="field">
			<label for="exchange">Mercado <span class="optional">(opcional)</span></label>
			<input id="exchange" type="text" name="exchange" placeholder="NASDAQ" />
		</div>
	</div>

	{#if classifiable}
		<div class="field">
			<label for="sector">Industria <span class="optional">(opcional)</span></label>
			<select id="sector" name="sector">
				<option value="">Sin clasificar</option>
				{#each SECTOR_OPTIONS as s (s.value)}
					<option value={s.value}>{s.label}</option>
				{/each}
			</select>
			<p class="hint">
				Es lo que reparte el panel por industria. Dejarlo sin clasificar no rompe nada: esa parte
				del patrimonio se cuenta aparte, con su propia etiqueta.
			</p>
		</div>
	{/if}

	{#if error}
		<p class="feedback error" role="alert">{error}</p>
	{/if}

	<div class="actions">
		{#if onCancel}
			<Button type="button" variant="ghost" onclick={onCancel}>Cancelar</Button>
		{/if}
		<Button type="submit" loading={creating}>Crear activo</Button>
	</div>
</form>

<style>
	.hint {
		margin: 0.35rem 0 0;
		font-size: 0.78rem;
		line-height: 1.45;
		color: var(--text-dim);
	}

	.actions {
		display: flex;
		justify-content: flex-end;
		gap: 0.75rem;
		margin-top: 0.5rem;
	}
</style>
