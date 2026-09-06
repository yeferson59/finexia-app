<script lang="ts">
	/*
	 * El activo elegido, en dos renglones.
	 *
	 * Era una rejilla de cinco celdas —Ticker, Nombre, Tipo, Exchange, Precio—
	 * con las etiquetas EN VERSALITAS y los cinco valores en ámbar. Cinco
	 * etiquetas para cinco datos que sirven a una sola cosa: confirmar que el
	 * instrumento elegido es el que buscabas. Así se lee de un vistazo, y el
	 * ámbar queda libre para lo único que de verdad manda aquí, que es el total.
	 */
	import { formatAssetType } from '$lib/shared/format/asset-type';
	import type { Asset } from '$lib/api/types';

	let {
		asset,
		formatCurrency
	}: { asset: Asset; formatCurrency: (value: number, code?: string) => string } = $props();

	// El precio de mercado está en la moneda del activo, que no tiene por qué ser
	// aquella en la que se pagará la compra.
	const assetCurrency = $derived(asset.currency?.trim().toUpperCase() || undefined);

	/* La frase se compone aquí y no en el marcado: intercalar `{#if}` entre
	   fragmentos de texto se come los espacios que los separan. */
	const detail = $derived.by(() => {
		const type = formatAssetType(asset.assetType);
		const parts = [asset.exchange ? `${type} en ${asset.exchange}.` : `${type}.`];

		if (asset.currentPrice) {
			const price = formatCurrency(parseFloat(asset.currentPrice.value), assetCurrency);
			parts.push(`Cotiza a ${price}.`);
		}

		return parts.join(' ');
	});
</script>

<p class="chosen">
	<span class="identity">
		<span class="ticker">{asset.ticker}</span>
		<span class="name">{asset.name}</span>
	</span>
	<span class="detail">{detail}</span>
</p>

<style>
	.chosen {
		margin: 0.65rem 0 0;
		font-size: 0.85rem;
		line-height: 1.5;
	}

	.identity {
		display: flex;
		align-items: baseline;
		flex-wrap: wrap;
		gap: 0.15rem 0.6rem;
	}

	/* El ticker en la tipografía de máquina, como en el resto del panel: es una
	   cadena que se compara carácter a carácter con la del bróker. */
	.ticker {
		font-family: var(--font-mono);
		font-size: 0.9em;
		font-weight: 500;
		color: var(--text);
	}

	.name {
		font-weight: 500;
		color: var(--text);
	}

	.detail {
		display: block;
		margin-top: 0.15rem;
		color: var(--text-muted);
	}
</style>
