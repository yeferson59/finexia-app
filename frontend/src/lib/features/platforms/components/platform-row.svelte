<script lang="ts">
	/**
	 * Una plataforma en el listado, como fila de un libro y no como tarjeta.
	 *
	 * Eran tarjetas de igual tamaño en una rejilla, que es justo lo que esta
	 * pantalla no puede decir: las plataformas no pesan lo mismo. En columnas se
	 * comparan las cifras de un vistazo, y la regla bajo el nombre repite el
	 * tramo que esta plataforma ocupa en la barra de arriba, con su mismo tono.
	 */
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import { FALLBACK_CURRENCY } from '$lib/shared/currency';
	import { formatSourceType, shareTint, type PlatformShare } from '../platforms';

	interface Props {
		entry: PlatformShare;
		/** Cuántas plataformas hay en total: fija el tono de la regla. */
		count: number;
		/**
		 * Si el listado enseña la columna de ganancia. Cuando ninguna plataforma
		 * la informa, la columna entera sobra: eran tres rayas en fila donde
		 * parecía que faltaba algo.
		 */
		showGain?: boolean;
		onView: (id: string) => void;
	}

	let { entry, count, showGain = true, onView }: Props = $props();

	const platform = $derived(entry.platform);

	// El importe lleva el símbolo de su moneda, no un "$" fijo: el total sale
	// convertido a la moneda de la cuenta y ponerle dólares a un total en pesos
	// es la misma cifra con el significado equivocado.
	const currency = $derived(platform.displayCurrency || FALLBACK_CURRENCY);

	// Posiciones que el backend no pudo convertir: siguen sumadas a valor
	// nominal, así que el total mezcla monedas y hay que decirlo.
	const unconverted = $derived(platform.positionsUnconverted ?? 0);

	function fmtMoney(value: string): string {
		return privacy.money(formatCurrency(parseFloat(value) || 0, currency));
	}

	// La ganancia se muestra solo si el backend la manda. Un `?? 0` la
	// convertiría en «esta plataforma no ganó ni perdió nada», que es una
	// afirmación, no la ausencia de un dato.
	const gain = $derived(platform.gainLoss !== undefined ? parseFloat(platform.gainLoss) : null);
	const gainPct = $derived(platform.gainLossPct);
</script>

<tr class="row">
	<td class="cell-name">
		<button
			type="button"
			class="name-btn"
			onclick={() => onView(platform.id)}
			aria-label={`Ver detalles de ${platform.name}`}
		>
			{platform.name}
		</button>
		<span
			class="share-rule"
			style="width: {Math.max(entry.share, 1.5)}%; --tint: {shareTint(entry.rank, count)}"
			aria-hidden="true"
		></span>
		<p class="meta">
			{formatSourceType(platform.sourceType)}
			<!-- En estrecho no cabe la columna de posiciones y lo primero que se
			     corta es el importe, que es lo que se viene a mirar: el contador
			     baja aquí y la columna se esconde. -->
			<span class="dim narrow-only">
				· {platform.investments}
				{platform.investments === 1 ? 'posición' : 'posiciones'}
			</span>
			<span class="dim">· {entry.share.toFixed(1)}% de la cuenta</span>
			{#if !platform.isActive}
				<span class="inactive">Inactiva</span>
			{/if}
		</p>
		{#if unconverted > 0}
			<p class="fx-note">
				{unconverted}
				{unconverted === 1 ? 'posición sin tasa' : 'posiciones sin tasa'} de cambio: el total suma monedas
				distintas.
			</p>
		{/if}
	</td>

	<td class="num cell-positions">{platform.investments}</td>

	<td class="num">{fmtMoney(platform.totalValue)}</td>

	{#if showGain}
		<td class="num">
			{#if gain !== null}
				<span class:up={gain > 0} class:down={gain < 0}>
					{fmtMoney(platform.gainLoss ?? '0')}
				</span>
				{#if gainPct !== undefined}
					<span class="pct" class:up={gain > 0} class:down={gain < 0}>
						{gainPct > 0 ? '+' : ''}{gainPct.toFixed(2)}%
					</span>
				{/if}
			{:else}
				<span class="absent" title="El backend no informa la ganancia de esta plataforma">—</span>
			{/if}
		</td>
	{/if}
</tr>

<style>
	/* El nombre se queda con todo el hueco sobrante para que las tres cifras
	   viajen juntas al margen derecho en vez de repartirse por el ancho. */
	.cell-name {
		width: 100%;
		min-width: 13rem;
	}

	/* La celda del nombre apila tres cosas, así que sin recortar el aire propio
	   de la tabla cada fila medía lo que tres. */
	.row :global(td) {
		padding-top: 0.9rem;
		padding-bottom: 0.9rem;
	}

	/* El nombre es el enlace: la fila entera no puede serlo dentro de una tabla
	   sin romper la semántica, y un botón «Ver detalles» por fila era una
	   columna entera repitiendo la misma palabra. */
	.name-btn {
		display: block;
		margin: 0;
		padding: 0;
		border: none;
		background: none;
		font-family: var(--font-display);
		font-size: 1.05rem;
		font-weight: 500;
		letter-spacing: -0.01em;
		color: var(--text);
		text-align: left;
		cursor: pointer;
		transition: color 0.2s ease;
	}

	.name-btn:hover {
		color: var(--amber-light);
	}

	.name-btn:focus-visible {
		outline: 2px solid var(--amber);
		outline-offset: 3px;
		border-radius: 2px;
	}

	/* Mismo tono que su tramo en la barra de arriba: es el mismo dato a dos
	   escalas, y verlo dos veces es lo que lo hace legible. */
	.share-rule {
		display: block;
		height: 2px;
		margin: 0.35rem 0 0.3rem;
		border-radius: 1px;
		background: var(--amber);
		opacity: var(--tint);
	}

	.meta {
		margin: 0;
		font-size: 0.8rem;
		color: var(--text-muted);
	}

	.dim {
		color: var(--text-dim);
		font-variant-numeric: tabular-nums;
	}

	/* Activa es el caso normal y no lleva marca. Inactiva sí, en gris: cerrar
	   una cuenta no es perder dinero, y el rojo aquí sólo significa eso. */
	.inactive {
		margin-left: 0.35rem;
		padding: 0.05rem 0.4rem;
		border: 1px solid var(--border-strong);
		border-radius: 4px;
		font-size: 0.7rem;
		color: var(--text-muted);
	}

	.fx-note {
		margin: 0.5rem 0 0;
		padding: 0.35rem 0.55rem;
		border-left: 2px solid rgba(212, 145, 42, 0.45);
		color: rgba(236, 234, 229, 0.7);
		font-size: 0.75rem;
		line-height: 1.4;
	}

	.pct {
		display: block;
		font-size: 0.75rem;
		color: var(--text-dim);
	}

	.absent {
		color: var(--text-dim);
	}

	.up {
		color: var(--green);
	}

	.down {
		color: var(--red);
	}

	.narrow-only {
		display: none;
	}

	@media (max-width: 720px) {
		.cell-name {
			min-width: 0;
		}

		.cell-positions {
			display: none;
		}

		.narrow-only {
			display: inline;
		}

		.row :global(td) {
			padding-left: 0.9rem;
			padding-right: 0.9rem;
		}
	}
</style>
