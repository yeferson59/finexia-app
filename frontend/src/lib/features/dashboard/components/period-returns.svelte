<script lang="ts">
	/*
	 * Cuánto se ganó el último día, la última semana, el último mes… debajo del
	 * patrimonio.
	 *
	 * La cifra de arriba dice cuánto se ha ganado sobre lo invertido desde
	 * siempre; esta franja parte ese mismo resultado en ventanas. Cada una lleva
	 * dos lecturas: el porcentaje, ponderado por tiempo —el mismo que dibuja la
	 * vista `%` de la gráfica—, y lo ganado en dinero. Las dos descuentan lo que
	 * se metió o sacó en la ventana, así que un depósito no sale aquí como
	 * ganancia de la semana.
	 *
	 * Una ventana a la que el historial no llega se queda en raya, con la fecha
	 * en que empieza: enseñar la historia entera bajo «1 año» sería inventarse el
	 * año.
	 *
	 * Sin filete ni aire propios: la usan la cabecera del panel y el detalle de
	 * un portafolio, y cada uno la separa de lo que tiene al lado a su manera.
	 */
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import { formatSignedPercent } from '$lib/shared/format/percent';
	import { FALLBACK_CURRENCY } from '$lib/shared/currency';
	import { toTrailingCells } from '../trailing';
	import type { TrailingReturn } from '$lib/api/types';

	interface Props {
		returns?: TrailingReturn[];
		currency?: string;
	}

	let { returns = [], currency = FALLBACK_CURRENCY }: Props = $props();

	const cells = $derived(toTrailingCells(returns));

	function signedMoney(value: number): string {
		const sign = value > 0 ? '+' : value < 0 ? '−' : '';
		return `${sign}${privacy.money(formatCurrency(Math.abs(value), currency))}`;
	}
</script>

{#if cells.length > 0}
	<section class="trailing" aria-labelledby="trailing-title">
		<h2 class="title" id="trailing-title">Rentabilidad por periodo</h2>

		<dl class="grid">
			{#each cells as cell (cell.period)}
				<div class="cell" title={cell.note}>
					<dt class="label">{cell.label}</dt>
					<dd class="pct tone-{cell.pctTone}">
						{cell.pct === null ? '—' : formatSignedPercent(cell.pct, 2)}
					</dd>
					<dd class="gain tone-{cell.gainTone}">
						{#if cell.gain !== null}{signedMoney(cell.gain)}{/if}
					</dd>
					<!-- La nota va en `title` para el ratón y aquí para el lector de
					     pantalla, que no lee `title` en un `div`. -->
					<dd class="sr-only">{cell.note}</dd>
				</div>
			{/each}
		</dl>

		<p class="hint">Sin contar lo que metiste o sacaste en cada periodo.</p>
	</section>
{/if}

<style>
	.trailing {
		min-width: 0;
	}

	.title {
		margin: 0 0 0.9rem;
		font-family: var(--font-body);
		font-size: 0.8rem;
		font-weight: 400;
		color: var(--text-muted);
	}

	.grid {
		display: grid;
		grid-template-columns: repeat(7, minmax(0, 1fr));
		gap: 1.1rem 1.5rem;
		margin: 0;
	}

	.cell {
		min-width: 0;
	}

	dd {
		margin: 0;
	}

	.label {
		font-size: 0.78rem;
		color: var(--text-muted);
		white-space: nowrap;
	}

	/* Cifras tabulares: las siete columnas se comparan de un vistazo. */
	.pct,
	.gain {
		font-family: var(--font-figures);
		font-variant-numeric: tabular-nums;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.pct {
		margin-top: 0.3rem;
		font-size: 1.15rem;
		color: var(--text);
	}

	.gain {
		/* Alto fijo: una ventana sin cifra en dinero no descuadra la fila. */
		min-height: 1.2em;
		margin-top: 0.15rem;
		font-size: 0.8rem;
		color: var(--text-muted);
	}

	.tone-up {
		color: var(--green);
	}

	.tone-down {
		color: var(--red);
	}

	.sr-only {
		position: absolute;
		width: 1px;
		height: 1px;
		padding: 0;
		margin: -1px;
		overflow: hidden;
		clip-path: inset(50%);
		white-space: nowrap;
	}

	.hint {
		margin: 1rem 0 0;
		font-size: 0.75rem;
		color: var(--text-dim);
	}

	@media (max-width: 1080px) {
		.grid {
			grid-template-columns: repeat(4, minmax(0, 1fr));
		}
	}

	@media (max-width: 560px) {
		.grid {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}
	}
</style>
