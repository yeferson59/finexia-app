<script lang="ts">
	/**
	 * Las cifras de una plataforma: lo invertido, lo que vale hoy y la
	 * diferencia entre las dos.
	 *
	 * Iban en cuatro tarjetas iguales, cada una con un icono ámbar decorativo
	 * que no añadía nada y que empujaba el número a media columna. Aquí las tres
	 * cifras están en línea porque sólo significan algo leídas juntas: la
	 * ganancia *es* la resta de las otras dos, y ponerlas al lado lo enseña.
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
	 * Sobre qué se reparten las posiciones. Diez posiciones son una cuenta
	 * distinta si son diez empresas que si son una empresa en diez portafolios.
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

<dl class="figures">
	<div class="figure">
		<dt>Invertido</dt>
		<dd class="amount">{formatCurrency(platform.totalValue)}</dd>
		{#if platform.percent !== undefined && platform.percent > 0}
			<p class="note">{platform.percent.toFixed(1)}% de la cuenta</p>
		{/if}
	</div>

	{#if platform.marketValue !== undefined}
		<div class="figure">
			<dt>Vale hoy</dt>
			<dd class="amount">{formatCurrency(platform.marketValue)}</dd>
			{#if atCost > 0 && !allAtCost}
				<p class="note">
					{atCost}
					{atCost === 1 ? 'posición valorada' : 'posiciones valoradas'} a coste
				</p>
			{/if}
		</div>
	{/if}

	{#if gain !== null}
		<div class="figure">
			<dt>Diferencia</dt>
			<dd class="amount" class:up={gain > 0} class:down={gain < 0}>
				{formatCurrency(platform.gainLoss ?? '0')}
			</dd>
			{#if platform.gainLossPct !== undefined}
				<p class="note" class:up={gain > 0} class:down={gain < 0}>
					{platform.gainLossPct > 0 ? '+' : ''}{platform.gainLossPct.toFixed(2)}% sobre lo invertido
				</p>
			{/if}
		</div>
	{/if}

	<div class="figure">
		<dt>Posiciones</dt>
		<dd class="amount">{platform.investments}</dd>
		{#if spread}
			<p class="note">{spread}</p>
		{/if}
	</div>
</dl>

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
	/* Las cifras se apilan a la izquierda con un ancho propio en vez de estirarse
	   por el panel: con dos, `1fr` las mandaba a los extremos y la resta que las
	   une —invertido, vale hoy, diferencia— dejaba de leerse de un vistazo. */
	.figures {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(9.5rem, 12rem));
		justify-content: start;
		gap: 2rem 3.5rem;
		margin: 0;
	}

	.figure dt {
		font-size: 0.8rem;
		font-weight: 400;
		color: var(--text-muted);
	}

	.amount {
		margin: 0.4rem 0 0;
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
		font-size: 1.4rem;
		font-weight: 400;
		letter-spacing: -0.01em;
		color: var(--text);
	}

	.note {
		margin: 0.35rem 0 0;
		font-size: 0.75rem;
		color: var(--text-dim);
		font-variant-numeric: tabular-nums;
	}

	.fx-note {
		margin: 1.75rem 0 0;
		padding: 0.5rem 0.7rem;
		border-left: 2px solid rgba(212, 145, 42, 0.45);
		color: rgba(236, 234, 229, 0.7);
		font-size: 0.78rem;
		line-height: 1.45;
	}

	.up {
		color: var(--green);
	}

	.down {
		color: var(--red);
	}
</style>
