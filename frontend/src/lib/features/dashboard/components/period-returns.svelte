<script lang="ts">
	/*
	 * Cuánto se ganó el último día, la última semana, el último mes… debajo del
	 * patrimonio.
	 *
	 * La cifra de arriba dice cuánto se ha ganado sobre lo invertido desde
	 * siempre; esta franja parte ese mismo resultado en ventanas. Cada ventana
	 * tiene dos lecturas: el porcentaje, ponderado por tiempo —el mismo que
	 * dibuja la vista `%` de la gráfica—, y lo ganado en dinero. Las dos
	 * descuentan lo que se metió o sacó en la ventana, así que un depósito no
	 * sale aquí como ganancia de la semana.
	 *
	 * Es una regla: las ventanas van de la más corta a la más larga, cada una
	 * es una marca sobre el mismo filete y lleva solo su porcentaje, que es lo
	 * que se compara de un vistazo. La elegida se lee debajo en una frase —desde
	 * cuándo, cuánto dinero—. Antes las siete llevaban también el importe y la
	 * fecha iba en un `title`: catorce cifras en verde y rojo, y la fecha
	 * invisible en un móvil.
	 *
	 * Una ventana a la que el historial no llega se queda en raya y su frase
	 * dice desde cuándo hay historial: enseñar la historia entera bajo «1 año»
	 * sería inventarse el año.
	 *
	 * Sin aire propio encima ni debajo: la usan la cabecera del panel y el
	 * detalle de un portafolio, y cada uno la separa de lo que tiene al lado a
	 * su manera. El filete de arriba sí es suyo: es la regla.
	 */
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import { formatSignedPercent } from '$lib/shared/format/percent';
	import { FALLBACK_CURRENCY } from '$lib/shared/currency';
	import { defaultTrailingPeriod, toTrailingCells, trailingReading } from '../trailing';
	import type { TrailingPeriod, TrailingReturn } from '$lib/api/types';

	interface Props {
		returns?: TrailingReturn[];
		currency?: string;
	}

	let { returns = [], currency = FALLBACK_CURRENCY }: Props = $props();

	const uid = $props.id();

	const cells = $derived(toTrailingCells(returns));

	// Lo que eligió el usuario sobrevive a un cambio de moneda; si esa ventana
	// ya no viene, se vuelve a la de por defecto.
	let chosen = $state<TrailingPeriod | null>(null);
	const current = $derived(
		cells.find((cell) => cell.period === chosen) ??
			cells.find((cell) => cell.period === defaultTrailingPeriod(cells))
	);
	const reading = $derived(current ? trailingReading(current) : null);

	const tabs: HTMLButtonElement[] = $state([]);

	// Flechas, Inicio y Fin, como cualquier grupo de pestañas: el tabulador
	// entra a la elegida y sale de la regla en un solo paso.
	function move(event: KeyboardEvent, index: number) {
		const last = cells.length - 1;
		const targets: Record<string, number> = {
			ArrowRight: index === last ? 0 : index + 1,
			ArrowLeft: index === 0 ? last : index - 1,
			Home: 0,
			End: last
		};
		const next = targets[event.key];
		if (next === undefined) return;

		event.preventDefault();
		chosen = cells[next].period;
		tabs[next]?.focus();
	}
</script>

{#if current && reading}
	<section class="trailing" aria-labelledby="{uid}-title">
		<h2 class="sr-only" id="{uid}-title">Rentabilidad por periodo</h2>

		<div class="rule" role="tablist" aria-labelledby="{uid}-title">
			{#each cells as cell, i (cell.period)}
				{@const active = cell.period === current.period}
				<button
					type="button"
					role="tab"
					id="{uid}-{cell.period}"
					class="stop"
					class:active
					aria-selected={active}
					aria-controls="{uid}-reading"
					tabindex={active ? 0 : -1}
					bind:this={tabs[i]}
					onclick={() => (chosen = cell.period)}
					onkeydown={(event) => move(event, i)}
				>
					<span class="label">{cell.label}</span>
					<span class="pct tone-{cell.pctTone}">
						{cell.pct === null ? '—' : formatSignedPercent(cell.pct, 2)}
					</span>
				</button>
			{/each}
		</div>

		<div
			class="reading"
			role="tabpanel"
			id="{uid}-reading"
			aria-labelledby="{uid}-{current.period}"
			aria-live="polite"
		>
			<p>
				{reading.before}{#if reading.amount !== null}<span class="amount tone-{reading.tone}"
						>{privacy.money(formatCurrency(reading.amount, currency))}</span
					>{/if}{reading.after}
			</p>
		</div>
	</section>
{/if}

<style>
	.trailing {
		min-width: 0;
	}

	/*
	 * El filete es continuo porque las paradas van pegadas: cada una pone su
	 * tramo de borde. Si no caben las siete —un móvil—, la regla se desliza de
	 * lado en vez de partirse en filas, que rompería el orden de corta a larga.
	 */
	.rule {
		display: grid;
		grid-auto-columns: minmax(5.75rem, 1fr);
		grid-auto-flow: column;
		overflow-x: auto;
		scroll-snap-type: x proximity;
		scrollbar-width: none;
	}

	.rule::-webkit-scrollbar {
		display: none;
	}

	.stop {
		position: relative;
		display: flex;
		flex-direction: column;
		align-items: flex-start;
		gap: 0.4rem;
		min-width: 0;
		padding: 1rem 1rem 0.25rem 0;
		border: 0;
		border-top: 1px solid var(--border-strong);
		background: none;
		color: inherit;
		font: inherit;
		text-align: left;
		cursor: pointer;
		scroll-snap-align: start;
	}

	/* La marca de la regla: dónde empieza cada ventana. */
	.stop::after {
		content: '';
		position: absolute;
		top: -4px;
		left: 0;
		width: 1px;
		height: 7px;
		background: var(--border-strong);
	}

	/*
	 * El tramo de la ventana elegida, sobre el filete. Crece desde su marca al
	 * elegirla: lo único que se mueve en la cabecera, y solo porque alguien lo
	 * pidió.
	 */
	.stop::before {
		content: '';
		position: absolute;
		top: -1px;
		left: 0;
		right: 1rem;
		height: 2px;
		background: var(--amber);
		transform: scaleX(0);
		transform-origin: left;
		transition: transform 240ms cubic-bezier(0.2, 0.7, 0.2, 1);
	}

	.stop.active::before {
		transform: scaleX(1);
	}

	.stop.active::after {
		background: var(--amber);
	}

	.stop:focus-visible {
		outline: 2px solid var(--amber);
		outline-offset: 4px;
		border-radius: 2px;
	}

	.label {
		font-size: 0.78rem;
		color: var(--text-muted);
		white-space: nowrap;
		transition: color 160ms ease;
	}

	.stop:hover .label,
	.stop.active .label {
		color: var(--text);
	}

	/* Cifras tabulares: las siete se comparan de un vistazo. */
	.pct,
	.amount {
		font-family: var(--font-figures);
		font-variant-numeric: tabular-nums;
		white-space: nowrap;
	}

	.pct {
		font-size: 1.2rem;
		font-stretch: 94%;
		letter-spacing: -0.01em;
		color: var(--text);
	}

	.pct:not(.tone-up):not(.tone-down) {
		color: var(--text-dim);
	}

	.tone-up {
		color: var(--green);
	}

	.tone-down {
		color: var(--red);
	}

	/* Una línea en escritorio; en un móvil puede ir en dos sin que salte. */
	.reading {
		max-width: 62ch;
		min-height: 1.5em;
		margin: 1.1rem 0 0;
		font-size: 0.9rem;
		line-height: 1.5;
		color: var(--text-muted);
	}

	.reading p {
		margin: 0;
	}

	.amount {
		font-weight: 500;
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

	/* Algo más estrechas en un teléfono, para que asome la siguiente parada y
	   se note que la regla sigue. */
	@media (max-width: 560px) {
		.rule {
			grid-auto-columns: minmax(5.25rem, 1fr);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.stop::before,
		.label {
			transition: none;
		}
	}
</style>
