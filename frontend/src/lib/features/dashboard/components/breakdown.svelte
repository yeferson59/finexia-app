<script lang="ts">
	/*
	 * Dónde está el dinero: el mismo patrimonio leído de tres formas —por
	 * plataforma, por portafolio y por clase de activo— en un solo sitio.
	 *
	 * Sustituye a tres bloques que se repetían entre sí: una tabla de
	 * portafolios que volvía a sumar el total de arriba y un donut de ocho
	 * porciones cuya paleta no pasaba la comprobación de daltonismo —el morado
	 * de «Bonos» y el azul de «Cripto» quedaban a ΔE 3,4, y el verde de «ETFs»
	 * era exactamente el verde de «+21 %»—. Una barra no necesita ocho colores:
	 * la identidad la lleva el nombre de la fila y la magnitud, el largo.
	 */
	import { resolve } from '$app/paths';
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import { formatPercent, formatSignedPercent } from '$lib/shared/format/percent';
	import { FALLBACK_CURRENCY } from '$lib/shared/currency';
	import { CUTS, breakdownFor, type CutId } from '../breakdown';
	import type { AllocationItem, Platform, PortfolioSummary } from '$lib/api/types';

	interface Props {
		platforms: Platform[];
		summaries: PortfolioSummary[];
		allocation: AllocationItem[];
		currency?: string;
	}

	let {
		platforms = [],
		summaries = [],
		allocation = [],
		currency = FALLBACK_CURRENCY
	}: Props = $props();

	const source = $derived({ platforms, summaries, allocation });

	/*
	 * `null` mientras el usuario no elija: así la pestaña que se abre es la
	 * primera que tenga algo dentro. Arrancar siempre en plataformas enseñaba un
	 * vacío a quien lleva sus portafolios sin registrar dónde están custodiados.
	 */
	let chosen = $state<CutId | null>(null);

	const fallback = $derived(
		CUTS.find((c) => breakdownFor(c.id, source, currency).rows.length > 0)?.id ?? 'portfolio'
	);
	const cut = $derived(chosen ?? fallback);
	const data = $derived(breakdownFor(cut, source, currency));

	const money = (value: number) => privacy.money(formatCurrency(value, currency));

	/*
	 * Las flechas recorren las pestañas, que es lo que espera un `tablist`: solo
	 * la activa entra con el tabulador y dentro del grupo se navega con el
	 * teclado. El manejador va en cada botón y no en el contenedor porque es el
	 * botón quien tiene el foco.
	 */
	function onKey(event: KeyboardEvent & { currentTarget: HTMLButtonElement }) {
		const at = CUTS.findIndex((c) => c.id === cut);
		let next: number;

		if (event.key === 'ArrowRight') next = (at + 1) % CUTS.length;
		else if (event.key === 'ArrowLeft') next = (at - 1 + CUTS.length) % CUTS.length;
		else if (event.key === 'Home') next = 0;
		else if (event.key === 'End') next = CUTS.length - 1;
		else return;

		event.preventDefault();
		chosen = CUTS[next].id;

		const tabs = event.currentTarget.parentElement;
		tabs?.querySelectorAll<HTMLButtonElement>('[role="tab"]')[next]?.focus();
	}

	const EMPTY: Record<CutId, string> = {
		platform: 'Todavía no has registrado dónde tienes tus activos.',
		portfolio: 'Todavía no has creado ningún portafolio.',
		type: 'Todavía no hay posiciones que repartir.'
	};
</script>

<section class="where" aria-labelledby="where-title">
	<header class="head">
		<h2 id="where-title">Dónde está</h2>

		<div class="tabs" role="tablist" aria-label="Cómo repartir el patrimonio">
			{#each CUTS as option (option.id)}
				<button
					role="tab"
					type="button"
					id="cut-{option.id}"
					class="tab"
					class:on={cut === option.id}
					aria-selected={cut === option.id}
					aria-controls="cut-panel"
					tabindex={cut === option.id ? 0 : -1}
					onclick={() => (chosen = option.id)}
					onkeydown={onKey}
				>
					{option.label}
				</button>
			{/each}
		</div>
	</header>

	<div id="cut-panel" role="tabpanel" aria-labelledby="cut-{cut}" tabindex="-1">
		{#if data.rows.length === 0}
			<p class="empty">{EMPTY[cut]}</p>
		{:else}
			<table class="rows">
				<caption class="sr-only">
					Reparto del patrimonio por {CUTS.find((c) => c.id === cut)?.label.toLowerCase()}
				</caption>
				<thead>
					<tr>
						<th scope="col">{CUTS.find((c) => c.id === cut)?.label}</th>
						<th scope="col" class="col-bar">Participación</th>
						<th scope="col" class="num">Valor</th>
						<th scope="col" class="num">
							{data.trailing === 'gain' ? 'Rendimiento' : 'Del total'}
						</th>
					</tr>
				</thead>
				<tbody>
					{#each data.rows as row (row.key)}
						<tr>
							<th scope="row" class="who">
								{#if cut === 'platform'}
									<a class="name" href={resolve(`/dashboard/platforms/${row.key}`)}>{row.label}</a>
								{:else if cut === 'portfolio'}
									<a class="name" href={resolve(`/dashboard/portfolios/${row.key}`)}>{row.label}</a>
								{:else}
									<span class="name">{row.label}</span>
								{/if}
								{#if row.detail}<span class="detail">{row.detail}</span>{/if}
							</th>

							<td class="col-bar">
								<span class="track" aria-hidden="true">
									<span class="fill" style="width: {(row.share * 100).toFixed(2)}%"></span>
								</span>
								<span class="sr-only">{formatPercent(row.share * 100)} del total</span>
							</td>

							<td class="num value">{money(row.value)}</td>

							<td
								class="num trail"
								class:up={(row.gainPct ?? 0) >= 0}
								class:has={row.gainPct !== null}
							>
								{#if data.trailing === 'share'}
									{formatPercent(row.share * 100)}
								{:else if row.gainPct === null}
									<span class="none" title="Sin precio de mercado con el que calcularlo">—</span>
								{:else}
									{formatSignedPercent(row.gainPct, 2)}
								{/if}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>

			{#if data.excluded > 0 || data.unconverted > 0}
				<p class="fx">
					{#if data.excluded > 0}
						{data.excluded === 1 ? 'Una fila queda' : `${data.excluded} filas quedan`} fuera de este reparto:
						no hay tasa para pasarla{data.excluded === 1 ? '' : 's'} a {currency}.
					{/if}
					{#if data.unconverted > 0}
						{data.unconverted}
						{data.unconverted === 1 ? 'posición se suma' : 'posiciones se suman'} sin convertir, así que
						el reparto mezcla monedas.
					{/if}
				</p>
			{/if}
		{/if}
	</div>
</section>

<style>
	.where {
		padding: 2rem 0;
		border-bottom: 1px solid var(--border);
	}

	.head {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
		margin-bottom: 1.5rem;
	}

	h2 {
		margin: 0;
		font-family: var(--font-body);
		font-size: 1.05rem;
		font-weight: 500;
		color: var(--text);
	}

	.tabs {
		display: flex;
		gap: 0.15rem;
		padding: 0.2rem;
		border: 1px solid var(--border);
		border-radius: 9px;
	}

	.tab {
		padding: 0.35rem 0.75rem;
		border: none;
		border-radius: 7px;
		background: none;
		color: var(--text-muted);
		font-family: inherit;
		font-size: 0.8rem;
		cursor: pointer;
		transition:
			background 0.15s ease,
			color 0.15s ease;
	}

	.tab:hover {
		color: var(--text);
	}

	.tab.on {
		background: var(--panel-2);
		color: var(--text);
	}

	#cut-panel:focus {
		outline: none;
	}

	.rows {
		width: 100%;
		border-collapse: collapse;
	}

	thead th {
		padding: 0 0 0.6rem;
		border-bottom: 1px solid var(--border);
		font-size: 0.75rem;
		font-weight: 400;
		color: var(--text-dim);
		text-align: left;
	}

	tbody tr {
		transition: background 0.15s ease;
	}

	/* Solo donde hay puntero: en una pantalla táctil el `:hover` se queda pegado
	   a la última fila tocada y parece que estuviera seleccionada. */
	@media (hover: hover) {
		tbody tr:hover {
			background: var(--panel);
		}
	}

	tbody th,
	tbody td {
		padding: 0.7rem 0.75rem 0.7rem 0;
		border-bottom: 1px solid var(--border);
		vertical-align: middle;
	}

	tbody th:first-child,
	tbody td:first-child {
		padding-left: 0.6rem;
	}

	/* La última fila no lleva filete: el de la sección va dos dedos más abajo y
	   las dos líneas juntas se leían como una fila vacía. */
	tbody tr:last-child th,
	tbody tr:last-child td {
		border-bottom: none;
	}

	tbody td:last-child {
		padding-right: 0.6rem;
	}

	.who {
		font-weight: 400;
		text-align: left;
	}

	.name {
		display: block;
		font-size: 0.9rem;
		color: var(--text);
		text-decoration: none;
		overflow-wrap: anywhere;
	}

	a.name:hover {
		text-decoration: underline;
		text-underline-offset: 3px;
	}

	.detail {
		display: block;
		margin-top: 0.15rem;
		font-size: 0.75rem;
		color: var(--text-dim);
	}

	/* El nombre no necesita un tercio de la fila: acotarlo acerca la barra a su
	   etiqueta y quita el vacío que quedaba entre la barra y las cifras. */
	thead th:first-child {
		width: 26%;
	}

	.col-bar {
		width: 38%;
	}

	/*
	 * Una sola serie, así que un solo color y ninguna leyenda: el largo dice la
	 * magnitud y el nombre de la fila, de quién es. El ámbar es el color del
	 * valor de mercado en toda la página.
	 */
	.track {
		display: block;
		height: 6px;
		border-radius: 3px;
		background: var(--panel-2);
		overflow: hidden;
	}

	.fill {
		display: block;
		height: 100%;
		min-width: 2px;
		border-radius: 0 3px 3px 0;
		background: var(--amber);
	}

	.num {
		text-align: right;
	}

	/* Tabulares aquí sí: son columnas y tienen que cuadrar entre filas. */
	.value,
	.trail {
		font-family: var(--font-mono);
		font-size: 0.85rem;
		font-variant-numeric: tabular-nums;
		white-space: nowrap;
		color: var(--text);
	}

	.trail.has {
		color: var(--red);
	}

	.trail.has.up {
		color: var(--green);
	}

	.none {
		color: var(--text-dim);
		cursor: help;
	}

	.empty {
		margin: 0;
		padding: 2.5rem 0;
		font-size: 0.9rem;
		color: var(--text-dim);
	}

	.fx {
		max-width: 62ch;
		margin: 1rem 0 0;
		padding-left: 0.75rem;
		border-left: 2px solid rgba(212, 145, 42, 0.45);
		font-size: 0.8rem;
		line-height: 1.5;
		color: var(--text-muted);
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

	@media (max-width: 700px) {
		.col-bar {
			display: none;
		}

		/* Sin la barra, el nombre se queda con el ancho que necesite: acotado al
		   26 % partía «Broker Demo» en dos líneas. */
		thead th:first-child {
			width: auto;
		}
	}
</style>
