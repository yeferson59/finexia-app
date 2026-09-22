<script lang="ts">
	/*
	 * Dónde está el dinero: el mismo patrimonio leído de cuatro formas —por
	 * plataforma, por portafolio, por clase de activo y por industria— en un
	 * solo sitio.
	 *
	 * Sustituye a tres bloques que se repetían entre sí: una tabla de
	 * portafolios que volvía a sumar el total de arriba y un donut de ocho
	 * porciones cuya paleta no pasaba la comprobación de daltonismo —el morado
	 * de «Bonos» y el azul de «Cripto» quedaban a ΔE 3,4, y el verde de «ETFs»
	 * era exactamente el verde de «+21 %»—. La franja de arriba no necesita
	 * ocho colores: es un solo ámbar a cinco intensidades, en el orden de las
	 * filas, y la identidad la lleva el nombre.
	 */
	import { resolve } from '$app/paths';
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import { formatPercent, formatSignedPercent } from '$lib/shared/format/percent';
	import { FALLBACK_CURRENCY } from '$lib/shared/currency';
	import BreakdownStrip from './breakdown-strip.svelte';
	import { CUTS, breakdownFor, stripSegments, toneOf, type CutId } from '../breakdown';
	import type {
		AllocationItem,
		Platform,
		PortfolioSummary,
		SectorAllocationItem
	} from '$lib/api/types';

	interface Props {
		platforms: Platform[];
		summaries: PortfolioSummary[];
		allocation: AllocationItem[];
		sectors?: SectorAllocationItem[];
		currency?: string;
	}

	let {
		platforms = [],
		summaries = [],
		allocation = [],
		sectors = [],
		currency = FALLBACK_CURRENCY
	}: Props = $props();

	const source = $derived({ platforms, summaries, allocation, sectors });

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
	const segments = $derived(stripSegments(data.rows));
	const cutLabel = $derived(CUTS.find((c) => c.id === cut)?.label ?? '');

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
		type: 'Todavía no hay posiciones que repartir.',
		sector: 'Todavía no hay posiciones que repartir.'
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
			<BreakdownStrip {segments} />

			<table class="rows">
				<caption class="sr-only">Reparto del patrimonio por {cutLabel.toLowerCase()}</caption>
				<thead>
					<tr>
						<th scope="col">{cutLabel}</th>
						<th scope="col" class="num">Del total</th>
						<th scope="col" class="num">Valor</th>
						{#if data.trailing === 'gain'}
							<th scope="col" class="num">Rendimiento</th>
						{/if}
					</tr>
				</thead>
				<tbody>
					{#each data.rows as row, i (row.key)}
						<tr>
							<th scope="row" class="who">
								<span class="who-in">
									<span
										class="swatch"
										style="background: var(--tone-{toneOf(i)})"
										aria-hidden="true"
									></span>
									<span class="names">
										{#if cut === 'platform'}
											<a class="name" href={resolve(`/dashboard/platforms/${row.key}`)}
												>{row.label}</a
											>
										{:else if cut === 'portfolio'}
											<a class="name" href={resolve(`/dashboard/portfolios/${row.key}`)}
												>{row.label}</a
											>
										{:else}
											<span class="name">{row.label}</span>
										{/if}
										{#if row.detail}<span class="detail">{row.detail}</span>{/if}
									</span>
								</span>
							</th>

							<td class="num share">
								<span aria-hidden="true">{formatPercent(row.share * 100)}</span>
								<span class="sr-only">{formatPercent(row.share * 100)} del total</span>
							</td>

							<td class="num value">{money(row.value)}</td>

							{#if data.trailing === 'gain'}
								<td
									class="num trail"
									class:up={(row.gainPct ?? 0) >= 0}
									class:has={row.gainPct !== null}
								>
									{#if row.gainPct === null}
										<span class="none" title="Sin precio de mercado con el que calcularlo">—</span>
									{:else}
										{formatSignedPercent(row.gainPct, 2)}
									{/if}
								</td>
							{/if}
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
	/*
	 * Va pegado a la cifra de arriba, sin filete entre los dos: es esa misma
	 * cifra partida, no una sección aparte.
	 *
	 * Los cinco tonos de la franja, mezclados con el fondo y no con opacidad:
	 * uno translúcido cambiaría según lo que tuviera detrás. El salto entre
	 * escalones se lee también en escala de grises, que es lo que pide una rampa.
	 */
	.where {
		--tone-0: var(--amber);
		--tone-1: color-mix(in oklab, var(--amber) 62%, var(--bg));
		--tone-2: color-mix(in oklab, var(--amber) 40%, var(--bg));
		--tone-3: color-mix(in oklab, var(--amber) 29%, var(--bg));
		--tone-4: color-mix(in oklab, var(--amber) 21%, var(--bg));

		padding: 0 0 2.5rem;
		border-bottom: 1px solid var(--border);
	}

	.head {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
		margin-bottom: 1.1rem;
	}

	h2 {
		margin: 0;
		font-family: var(--font-display);
		font-size: 1.3rem;
		font-weight: 400;
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
		margin-top: 1.75rem;
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

	thead th + th {
		padding-left: 1.5rem;
	}

	tbody th,
	tbody td {
		padding: 0.75rem 0;
		border-bottom: 1px solid var(--border);
		vertical-align: middle;
	}

	tbody td {
		padding-left: 1.5rem;
	}

	/* La última fila no lleva filete: el de la sección va justo debajo y las dos
	   líneas juntas se leían como una fila vacía. */
	tbody tr:last-child th,
	tbody tr:last-child td {
		border-bottom: none;
	}

	.who {
		font-weight: 400;
		text-align: left;
	}

	.who-in {
		display: flex;
		align-items: baseline;
		gap: 0.75rem;
	}

	/* El tono del tramo de la franja: es lo que ata cada fila a su parte, así
	   que no hace falta leyenda. */
	.swatch {
		flex-shrink: 0;
		width: 10px;
		height: 10px;
		border-radius: 2px;
		transform: translateY(1px);
	}

	.names {
		min-width: 0;
	}

	.name {
		display: block;
		font-size: 0.92rem;
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

	.num {
		width: 1%;
		text-align: right;
		white-space: nowrap;
	}

	/* Tabulares: son columnas y tienen que cuadrar entre filas. */
	.share,
	.value,
	.trail {
		font-family: var(--font-figures);
		font-size: 0.92rem;
		font-stretch: 88%;
		font-variant-numeric: tabular-nums;
		color: var(--text);
	}

	.share {
		color: var(--text-muted);
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

	/* En un teléfono la franja pierde los rótulos, así que la participación se
	   queda en su columna; lo que se estrecha es el aire entre ellas. */
	@media (max-width: 560px) {
		thead th + th,
		tbody td {
			padding-left: 0.85rem;
		}
	}
</style>
