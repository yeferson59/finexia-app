<script lang="ts">
	/*
	 * Quién es este activo y dónde está.
	 *
	 * Ya no trae el precio de mercado en cuerpo 32: era la cifra más grande de
	 * la página y no es la que se viene a ver. Un precio por acción es un dato
	 * del mercado; lo que el usuario tiene es la posición, y esa manda ahora en
	 * el titular de debajo. El precio sigue en la página, en la frase que lo
	 * compara con lo que pagó, que es donde significa algo.
	 *
	 * La vuelta nombra el portafolio del que se salió, y la insignia
	 * «STOCK · NASDAQ» —versalitas, punto medio y el enum crudo del backend— es
	 * ahora la línea que describe el activo en castellano.
	 */
	import { resolve } from '$app/paths';
	import { formatAssetType } from '$lib/shared/format/asset-type';
	import type { AssetPosition } from '../asset';

	let {
		position,
		portfolioId,
		portfolioName,
		onAddTransaction
	}: {
		position: AssetPosition;
		portfolioId: string;
		/** Nombre del portafolio; vacío si la carga no lo trajo. */
		portfolioName: string;
		onAddTransaction: () => void;
	} = $props();

	/**
	 * Qué es y dónde cotiza: «acciones en NASDAQ», «cripto».
	 *
	 * La clase va en minúscula porque entra a mitad de frase, y el mercado
	 * conserva su caja, que es como se escribe su nombre. Un activo cuyo nombre
	 * ya dice su clase se la salta: «Efectivo en dólares» no necesita
	 * completarse con «, efectivo».
	 */
	const classLine = $derived.by(() => {
		const label = formatAssetType(position.assetType).toLowerCase();
		const parts = position.name.toLowerCase().includes(label) ? [] : [label];
		if (position.exchange) parts.push(`en ${position.exchange}`);

		return parts.join(' ');
	});
</script>

<header class="head">
	<a class="back" href={resolve('/dashboard/portfolios/[id]', { id: portfolioId })}>
		{portfolioName ? `Volver a ${portfolioName}` : 'Volver al portafolio'}
	</a>

	<div class="row">
		<div class="who">
			<h1>{position.ticker}</h1>
			<p class="name">{position.name}{classLine ? `, ${classLine}` : ''}</p>
		</div>

		<button type="button" class="add" onclick={onAddTransaction}>Registrar movimiento</button>
	</div>
</header>

<style>
	.head {
		margin-bottom: 2rem;
		padding-bottom: 1.75rem;
		border-bottom: 1px solid var(--border);
	}

	.back {
		display: inline-block;
		margin-bottom: 1.1rem;
		font-size: 0.82rem;
		color: var(--text-muted);
		text-decoration: none;
	}

	.back::before {
		content: '←';
		margin-right: 0.4rem;
	}

	.back:hover {
		color: var(--text);
		text-decoration: underline;
		text-underline-offset: 3px;
	}

	.row {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 2rem;
		flex-wrap: wrap;
	}

	.who {
		min-width: 0;
	}

	h1 {
		margin: 0;
		font-family: var(--font-display);
		font-size: clamp(2rem, 4vw, 2.75rem);
		font-weight: 300;
		line-height: 1.05;
		letter-spacing: -0.02em;
		color: var(--text);
		overflow-wrap: anywhere;
	}

	.name {
		max-width: 58ch;
		margin: 0.5rem 0 0;
		font-size: 0.95rem;
		font-weight: 300;
		line-height: 1.45;
		color: var(--text-muted);
	}

	.add {
		flex-shrink: 0;
		padding: 0.6rem 1.15rem;
		border: 1px solid var(--amber);
		border-radius: 9px;
		background: var(--amber);
		color: #0d0800;
		font-family: var(--font-body);
		font-size: 0.88rem;
		font-weight: 600;
		white-space: nowrap;
		cursor: pointer;
		transition:
			background 0.2s ease,
			border-color 0.2s ease;
	}

	.add:hover {
		border-color: var(--amber-light);
		background: var(--amber-light);
	}

	@media (max-width: 620px) {
		.row {
			flex-direction: column;
			align-items: flex-start;
			gap: 1.25rem;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.add {
			transition: none;
		}
	}
</style>
