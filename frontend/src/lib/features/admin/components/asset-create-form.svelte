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
	import { optimisticSubmit } from '$lib/shared/optimistic.svelte';
	import Button from '$lib/ui/button.svelte';
	import { typeHasSector } from '$lib/shared/format/sector';
	import { ASSET_TYPES } from '../admin';
	import SectorClassification from './sector-classification.svelte';

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
	 * Lo valida el servidor, así que se espera su respuesta; pero el diálogo se
	 * cierra en cuanto llega y la página se refresca de fondo, en vez de esperar
	 * a que se recargue entera con el botón girando.
	 */
	const submit = optimisticSubmit({
		fallbackError: 'No se pudo crear el activo.',
		syncForm: true,
		apply: () => {
			creating = true;
		},
		onSuccess: () => {
			creating = false;
			onSuccess?.();
		},
		onError: () => (creating = false)
	});

	/*
	 * El tipo se sigue desde aquí porque la industria depende de él: detrás de
	 * una cripto, un saldo o un inmueble no hay empresa que clasificar, y el
	 * backend contesta 400 a quien lo intente. Enseñar el campo y que el
	 * servidor lo rechace sería pedir un dato para luego negarlo; mejor no
	 * ofrecerlo.
	 */
	let assetType = $state('');
	const classifiable = $derived(typeHasSector(assetType));

	// La clasificación vive aquí y no dentro del componente porque los dos
	// campos se envían juntos y el componente solo los edita.
	let sector = $state('');
	let weights = $state<Record<string, number | null>>({});
</script>

<form class="rail-fields" method="POST" action="?/createAsset" use:enhance={submit}>
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
		<SectorClassification bind:sector bind:weights />
	{/if}

	{#if error}
		<p class="feedback error" role="alert">{error}</p>
	{/if}

	<div class="modal-actions">
		{#if onCancel}
			<Button type="button" variant="ghost" onclick={onCancel}>Cancelar</Button>
		{/if}
		<Button type="submit" loading={creating}>Crear activo</Button>
	</div>
</form>
