<script lang="ts">
	/**
	 * Las métricas de una plataforma: lo invertido, lo que vale, la diferencia
	 * entre las dos y qué parte de la cuenta representa.
	 *
	 * Sale del detalle porque las cuatro tarjetas y sus estilos eran la mitad de
	 * aquel archivo, y porque son una unidad: cada una solo significa algo
	 * medida contra las otras.
	 */
	import type { Platform } from '$lib/api/types';

	let {
		platform,
		unconverted,
		gain,
		formatCurrency
	}: {
		platform: Platform;
		/** Posiciones sumadas a valor nominal por falta de tasa. */
		unconverted: number;
		/** La ganancia ya parseada, o `null` si el backend no la manda. */
		gain: number | null;
		formatCurrency: (value: string) => string;
	} = $props();

	/**
	 * Sobre qué se reparten las posiciones. Va de nota bajo el contador en vez
	 * de en tarjetas propias: son dos números que solo significan algo pegados
	 * al que cuentan, y cuatro tarjetas ya son las que caben.
	 */
	const spread = $derived.by(() => {
		const parts: string[] = [];
		if (platform.assets !== undefined) {
			parts.push(`${platform.assets} ${platform.assets === 1 ? 'activo' : 'activos'}`);
		}
		if (platform.portfolios !== undefined && platform.portfolios > 0) {
			parts.push(
				`${platform.portfolios} ${platform.portfolios === 1 ? 'portafolio' : 'portafolios'}`
			);
		}
		return parts.join(' · ');
	});

	/**
	 * Cuántas posiciones siguen valoradas a su propio coste.
	 *
	 * Es lo que hace legible la ganancia: una posición sin precio de mercado se
	 * valora al coste contra el que se la compara, así que aporta exactamente
	 * cero. Sin este número, una ganancia de cero puede ser una plataforma que
	 * no se movió o una que nadie ha valorado, y son cosas distintas.
	 */
	const atCost = $derived(platform.positionsAtCost ?? 0);

	/**
	 * Cuando *todas* lo están, la ganancia no es una cifra pequeña: es cero por
	 * construcción, y decirlo entero merece un aviso y no una nota al pie.
	 */
	const allAtCost = $derived(platform.investments > 0 && atCost === platform.investments);
</script>

<div class="stats-grid">
	<div class="stat-card">
		<div class="stat-icon">
			<svg width="24" height="24" viewBox="0 0 24 24" fill="currentColor">
				<path
					d="M19 3H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zM9 17H7v-7h2v7zm4 0h-2V7h2v10zm4 0h-2v-4h2v4z"
				></path>
			</svg>
		</div>
		<div class="stat-content">
			<span class="stat-label">Posiciones</span>
			<span class="stat-value">{platform.investments}</span>
			{#if spread}
				<span class="stat-note">{spread}</span>
			{/if}
		</div>
	</div>
	<div class="stat-card">
		<div class="stat-icon">
			<svg width="24" height="24" viewBox="0 0 24 24" fill="currentColor">
				<path
					d="M11.8 10.9c-2.27-.59-3-1.2-3-2.15 0-1.09 1.01-1.85 2.7-1.85 1.78 0 2.44.85 2.5 2.1h2.21c-.07-1.72-1.12-3.3-3.21-3.81V3h-3v2.16c-1.94.42-3.5 1.68-3.5 3.61 0 2.31 1.91 3.46 4.7 4.13 2.5.6 3 1.48 3 2.41 0 .69-.49 1.79-2.7 1.79-2.06 0-2.87-.92-2.98-2.1h-2.2c.12 2.19 1.76 3.42 3.68 3.83V21h3v-2.15c1.95-.37 3.5-1.5 3.5-3.55 0-2.84-2.43-3.81-4.7-4.4z"
				></path>
			</svg>
		</div>
		<div class="stat-content">
			<span class="stat-label">Total Invertido</span>
			<span class="stat-value">{formatCurrency(platform.totalValue)}</span>
			{#if platform.percent !== undefined && platform.percent > 0}
				<span class="stat-note">{platform.percent.toFixed(1)}% de la cuenta</span>
			{/if}
		</div>
	</div>
	{#if platform.marketValue !== undefined}
		<div class="stat-card">
			<div class="stat-icon">
				<svg width="24" height="24" viewBox="0 0 24 24" fill="currentColor">
					<path d="M16 6l2.29 2.29-4.88 4.88-4-4L2 16.59 3.41 18l6-6 4 4 6.3-6.29L22 12V6z"></path>
				</svg>
			</div>
			<div class="stat-content">
				<span class="stat-label">Valor de Mercado</span>
				<span class="stat-value">{formatCurrency(platform.marketValue)}</span>
				<!-- Cuando lo están todas, el aviso de abajo lo dice entero y esta
				     nota sería la mitad del mismo mensaje. -->
				{#if atCost > 0 && !allAtCost}
					<span class="stat-note">
						{atCost}
						{atCost === 1 ? 'posición valorada' : 'posiciones valoradas'} a coste
					</span>
				{/if}
			</div>
		</div>
	{/if}
	{#if gain !== null}
		<div class="stat-card">
			<div class="stat-icon">
				<svg width="24" height="24" viewBox="0 0 24 24" fill="currentColor">
					<path
						d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"
					></path>
				</svg>
			</div>
			<div class="stat-content">
				<span class="stat-label">Ganancia</span>
				<span class="stat-value" class:up={gain > 0} class:down={gain < 0}>
					{formatCurrency(platform.gainLoss ?? '0')}
				</span>
				{#if platform.gainLossPct !== undefined}
					<span class="stat-note" class:up={gain > 0} class:down={gain < 0}>
						{platform.gainLossPct > 0 ? '+' : ''}{platform.gainLossPct.toFixed(2)}% sobre lo
						invertido
					</span>
				{/if}
			</div>
		</div>
	{/if}
</div>

{#if allAtCost}
	<p class="fx-note">
		Ninguna posición de esta plataforma tiene precio de mercado guardado, así que se valoran a su
		propio coste: el valor de mercado repite lo invertido y la ganancia sale en cero porque no hay
		con qué compararla, no porque no se haya movido.
	</p>
{/if}

{#if unconverted > 0}
	<p class="fx-note">
		{unconverted}
		{unconverted === 1 ? 'posición sigue' : 'posiciones siguen'} contadas en su propia moneda porque no
		hay tasa de cambio guardada: el total suma monedas distintas.
	</p>
{/if}

<style>
	.stats-grid {
		display: grid;
		/* Hasta cuatro tarjetas, y en pantalla estrecha bajan de línea antes que
		   estrecharse hasta ser ilegibles. */
		grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
		gap: 1.5rem;
	}

	/* Mismo aviso que en la tarjeta de portafolio: el total incluye importes
	   sin convertir, así que se marca en vez de pasar por comparable. */
	.fx-note {
		margin: 1.25rem 0 0;
		padding: 0.5rem 0.7rem;
		border: 1px solid rgba(212, 145, 42, 0.3);
		border-radius: 8px;
		background: rgba(212, 145, 42, 0.08);
		color: rgba(236, 234, 229, 0.75);
		font-size: 0.78rem;
		line-height: 1.4;
	}

	.stat-card {
		display: flex;
		gap: 1rem;
		padding: 1.25rem;
		border-radius: 12px;
		background: var(--border);
		border: 1px solid var(--border-strong);
		align-items: center;
	}

	.stat-icon {
		width: 44px;
		height: 44px;
		border-radius: 10px;
		background: rgba(212, 145, 42, 0.12);
		border: 1px solid rgba(212, 145, 42, 0.2);
		color: var(--amber);
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	.stat-content {
		display: flex;
		flex-direction: column;
		gap: 0.35rem;
	}

	.stat-label {
		font-size: 0.8rem;
		color: rgba(236, 234, 229, 0.6);
	}

	.stat-value {
		font-size: 1.25rem;
		font-weight: 700;
		color: var(--text);
		font-variant-numeric: tabular-nums;
	}

	.stat-note {
		font-size: 0.75rem;
		color: rgba(236, 234, 229, 0.45);
		font-variant-numeric: tabular-nums;
	}

	.up {
		color: var(--green);
	}

	.down {
		color: var(--red);
	}

	@media (max-width: 768px) {
		.stats-grid {
			grid-template-columns: 1fr;
		}
	}
</style>
