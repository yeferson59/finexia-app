<script lang="ts">
	/**
	 * Un cajón de una cuenta: la cuenta principal de una moneda, uno de sus
	 * bolsillos o un depósito a plazo. Es la misma fila para los tres, con columnas
	 * que caen en la misma vertical en todas las plataformas: qué es, a qué tasa
	 * rinde, cuánto guarda en su moneda —la cifra del extracto— y qué se le hace.
	 *
	 * Un bolsillo cuelga de su cuenta por un riel en la primera columna: su dinero
	 * suma en la plataforma y lo suyo es la tasa, así que se anida en lugar de
	 * listarse aparte. Un depósito lleva además la línea de su plazo, que es lo
	 * que se le pregunta.
	 */
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import type { CashPocket } from '$lib/api/types';
	import type { CashAccount } from '../cash';
	import { cashAccountHistory } from '../pockets';
	import { cashAccountRate, formatAnnualRate, type CashRate } from '../rates';
	import { cashCurrencyName, cashRateLines } from '../yield';
	import CashTermTrack from './cash-term-track.svelte';

	interface Props {
		account: CashAccount;
		/** El bolsillo de la fila; `null` en la cuenta principal. */
		pocket: CashPocket | null;
		rates: CashRate[];
		platformName: string;
		showPortfolio: boolean;
		today: string;
		/** Un bolsillo o un depósito: cuelga del riel de su cuenta. */
		nested?: boolean;
		/** El último cajón del riel, donde el riel dobla y termina. */
		last?: boolean;
		/** Una cuenta principal con cajones debajo: el riel empieza en ella. */
		rail?: boolean;
		onRecord: () => void;
		onRate: () => void;
		onPocket: () => void;
		onDeposit: () => void;
	}

	let {
		account,
		pocket,
		rates,
		platformName,
		showPortfolio,
		today,
		nested = false,
		last = false,
		rail = false,
		onRecord,
		onRate,
		onPocket,
		onDeposit
	}: Props = $props();

	const money = (amount: number, currency = account.currency) =>
		privacy.money(formatCurrency(amount, currency));

	const status = $derived(
		cashAccountRate(rates, account.sourceId, account.currency, today, account.pocketId)
	);

	const fixed = $derived(account.pocketKind === 'fixed');
	const lines = $derived(cashRateLines(status, (amount) => money(amount)));

	/* Un depósito conserva la tasa del día en que se abrió: rija o no hoy, es esa. */
	const fixedRate = $derived(status.current ?? status.latest);

	const earned = $derived(
		account.balances.reduce((sum, b) => sum + (parseFloat(b.interestThisMonth) || 0), 0)
	);
	const pending = $derived(
		account.balances.reduce((sum, b) => sum + (parseFloat(b.pendingInterest) || 0), 0)
	);

	/*
	 * El nombre de la fila. Con cajones debajo, la de la moneda es «Cuenta
	 * principal», que es como la llama el formulario al elegir dónde cae el
	 * dinero; sola, basta con decir de qué moneda es.
	 */
	const title = $derived(
		nested ? account.pocketName : rail ? 'Cuenta principal' : cashCurrencyName(account.currency)
	);

	/* Lo que distingue esta fila de las demás para quien no la ve entera. */
	const where = $derived(
		nested ? `${platformName}, ${account.pocketName}` : `${platformName}, ${account.currency}`
	);
</script>

<li
	class="drawer"
	data-account={account.key}
	class:nested
	class:last
	class:rail
	class:emptied={account.balance === 0}
	class:has-term={fixed && !!pocket?.maturesOn}
>
	<span class="mark" aria-hidden={nested}>
		{#if !nested}
			<span class="code">{account.currency}</span>
		{/if}
	</span>

	<div class="name">
		<p class="title">
			{title}
			{#if fixed}<span class="kind">a plazo fijo</span>{/if}
		</p>
		{#if account.balances.length > 1}
			<ul class="portions" aria-label="Cómo se reparte entre portafolios">
				{#each account.balances as balance (balance.entryId)}
					<li>
						{balance.portfolioName}
						<span class="portion-amount">{money(parseFloat(balance.balance) || 0)}</span>
					</li>
				{/each}
			</ul>
		{:else if showPortfolio && account.balances.length === 1}
			<p class="meta">Suma en {account.balances[0].portfolioName}</p>
		{/if}
		{#if !fixed}
			<p class="meta">{cashAccountHistory(account)}</p>
		{/if}
	</div>

	<div class="yield">
		{#if fixed}
			<p class="rate earning">
				<span class="key" aria-hidden="true"></span>
				{fixedRate ? formatAnnualRate(fixedRate.annualRatePct) : 'Sin tasa'}
				<span class="rate-note">fija</span>
			</p>
			{#if fixedRate?.posting === 'at_maturity'}
				<p class="meta">abono al vencer</p>
			{/if}
		{:else if lines}
			<p class="rate" class:earning={lines.earning}>
				<span class="key" aria-hidden="true"></span>
				{lines.rate}
			</p>
			{#if lines.detail}
				<p class="meta">{lines.detail}</p>
			{/if}
		{:else}
			<p class="rate">
				<span class="key" aria-hidden="true"></span>
				Sin tasa
			</p>
		{/if}
		{#if earned > 0}
			<p class="meta"><span class="interest">+{money(earned)}</span> este mes</p>
		{/if}
		<!-- Calculado pero todavía fuera del saldo: con abono mensual espera al
		     cierre del mes, y en un depósito que abona al vencer, a su último día. -->
		{#if pending > 0}
			<p class="meta">
				+{money(pending)} por abonar {fixed ? 'al vencer' : 'al cerrar el mes'}
			</p>
		{/if}
	</div>

	<p class="figures">
		<span class="native">{money(account.balance)}</span>
		<!-- Vacía no hay nada que convertir: «≈ $0,00» solo añade ruido. -->
		{#if account.currency !== account.displayCurrency && account.balance !== 0}
			{#if account.fxConverted}
				<span class="converted">≈ {money(account.value, account.displayCurrency)}</span>
			{:else}
				<span class="converted">sin tasa a {account.displayCurrency}</span>
			{/if}
		{/if}
	</p>

	<div class="actions">
		{#if fixed}
			<button type="button" class="act" onclick={onDeposit}>
				Ver depósito<span class="sr-only">: {where}</span>
			</button>
		{:else}
			{#if !nested}
				<button type="button" class="act strong" onclick={onRecord}>
					Registrar<span class="sr-only"> un movimiento en {where}</span>
				</button>
			{/if}
			<button type="button" class="act" onclick={onRate}>
				{lines ? 'Tasa' : 'Darle tasa'}<span class="sr-only"> de {where}</span>
			</button>
			{#if nested}
				<button type="button" class="act" onclick={onPocket}>
					Editar<span class="sr-only"> el bolsillo {account.pocketName}</span>
				</button>
			{/if}
		{/if}
	</div>

	{#if fixed && pocket?.maturesOn}
		<div class="term">
			<CashTermTrack openedOn={pocket.openedOn} maturesOn={pocket.maturesOn} {today} />
		</div>
	{/if}
</li>

<style>
	/*
	 * Las columnas tienen ancho fijo salvo las dos de texto: así los importes y
	 * las acciones caen en la misma vertical en todas las plataformas.
	 */
	.drawer {
		--mark: 3.5rem;
		--elbow: 1.25rem;
		--pad-y: 1rem;

		display: grid;
		grid-template-columns: var(--mark) minmax(0, 1.2fr) minmax(0, 1fr) 11rem 10.5rem;
		/* La fila de abajo es la línea del plazo; sin ella no ocupa nada. */
		grid-template-areas:
			'mark name yield figures actions'
			'mark term term figures actions';
		align-items: start;
		column-gap: 1.25rem;
		padding: var(--pad-y) 1.5rem var(--pad-y) 1.25rem;
	}

	.drawer.has-term {
		row-gap: 0.7rem;
	}

	.nested {
		--pad-y: 0.8rem;
	}

	p {
		margin: 0;
	}

	.mark {
		grid-area: mark;
		position: relative;
		align-self: stretch;
		/* Hasta los bordes de la fila, por encima de su relleno: el riel de un
		   cajón tiene que tocar el del siguiente. */
		margin: calc(-1 * var(--pad-y)) 0;
		padding: var(--pad-y) 0;
	}

	.code {
		display: inline-block;
		min-width: 2.6rem;
		padding: 0.2rem 0.4rem;
		border: 1px solid var(--border-strong);
		border-radius: 5px;
		font-size: 0.72rem;
		font-weight: 600;
		letter-spacing: 0.06em;
		text-align: center;
		color: var(--text-muted);
	}

	/*
	 * El riel: baja desde el código de la moneda por todos sus cajones y dobla
	 * hacia cada uno. Es una línea de estructura, del gris de los filetes.
	 */
	.rail .mark::after,
	.nested .mark::before,
	.nested .mark::after {
		content: '';
		position: absolute;
		background: var(--border-strong);
	}

	.rail .mark::after {
		top: 2.55rem;
		bottom: 0;
		left: var(--elbow);
		width: 1px;
	}

	.nested .mark::before {
		top: 0;
		bottom: 0;
		left: var(--elbow);
		width: 1px;
	}

	.nested.last .mark::before {
		bottom: auto;
		height: 1.55rem;
	}

	.nested .mark::after {
		top: 1.55rem;
		left: var(--elbow);
		width: calc(var(--mark) - var(--elbow) + 0.5rem);
		height: 1px;
	}

	.name {
		grid-area: name;
		min-width: 0;
	}

	.title {
		display: flex;
		flex-wrap: wrap;
		align-items: baseline;
		gap: 0 0.5rem;
		font-size: 0.95rem;
		font-weight: 500;
		line-height: 1.4;
		color: var(--text);
	}

	.kind {
		font-size: 0.74rem;
		font-weight: 400;
		color: var(--text-dim);
	}

	.meta {
		margin-top: 0.15rem;
		font-size: 0.79rem;
		line-height: 1.45;
		color: var(--text-dim);
	}

	.portions {
		display: flex;
		flex-wrap: wrap;
		gap: 0.1rem 1rem;
		margin: 0.2rem 0 0;
		padding: 0;
		list-style: none;
		font-size: 0.79rem;
		color: var(--text-muted);
	}

	.portion-amount {
		margin-left: 0.3rem;
		font-family: var(--font-mono);
		font-size: 0.76rem;
		font-variant-numeric: tabular-nums;
		white-space: nowrap;
		color: var(--text);
	}

	.yield {
		grid-area: yield;
		min-width: 0;
	}

	/* La tasa en la letra de las cifras, con su marca: verde llena mientras
	   rinde, un aro gris cuando no. El texto no lleva el color. */
	.rate {
		display: flex;
		align-items: baseline;
		gap: 0.5rem;
		font-family: var(--font-mono);
		font-size: 0.88rem;
		line-height: 1.4;
		color: var(--text-muted);
	}

	.rate.earning {
		color: var(--text);
	}

	.key {
		flex-shrink: 0;
		align-self: center;
		width: 7px;
		height: 7px;
		border: 1.5px solid var(--text-dim);
		border-radius: 50%;
	}

	.earning .key {
		border-color: var(--green);
		background: var(--green);
	}

	.rate-note {
		font-family: var(--font-body);
		font-size: 0.76rem;
		color: var(--text-dim);
	}

	.interest {
		font-family: var(--font-mono);
		font-size: 0.78rem;
		font-variant-numeric: tabular-nums;
		color: var(--green);
	}

	.figures {
		grid-area: figures;
		display: flex;
		flex-direction: column;
		align-items: flex-end;
		gap: 0.2rem;
	}

	.native,
	.converted {
		font-family: var(--font-mono);
		font-size: 1.05rem;
		font-variant-numeric: tabular-nums;
		white-space: nowrap;
		color: var(--text);
	}

	.converted {
		font-size: 0.76rem;
		color: var(--text-dim);
	}

	.emptied .native,
	.emptied .code {
		color: var(--text-dim);
	}

	.actions {
		grid-area: actions;
		display: flex;
		flex-wrap: wrap;
		justify-content: flex-end;
		gap: 0.25rem;
		margin-top: -0.3rem;
	}

	.act {
		padding: 0.35rem 0.65rem;
		border: 1px solid transparent;
		border-radius: 7px;
		background: transparent;
		font: inherit;
		font-size: 0.8rem;
		white-space: nowrap;
		color: var(--text-muted);
		cursor: pointer;
		transition:
			background-color 0.15s ease,
			border-color 0.15s ease,
			color 0.15s ease;
	}

	.act:hover {
		background: var(--surface-2);
		color: var(--text);
	}

	/* «Registrar» es lo de todos los días: el único con borde de la fila. */
	.act.strong {
		border-color: var(--border-strong);
		color: var(--text);
	}

	.act.strong:hover {
		border-color: var(--amber);
		background: rgba(212, 145, 42, 0.08);
		color: var(--amber-light);
	}

	.term {
		grid-area: term;
		min-width: 0;
	}

	@media (max-width: 980px) {
		.drawer {
			--mark: 2.75rem;
			--elbow: 1rem;

			grid-template-columns: var(--mark) minmax(0, 1fr) auto;
			grid-template-areas:
				'mark name figures'
				'mark yield yield'
				'mark term term'
				'mark actions actions';
			column-gap: 0.85rem;
			row-gap: 0.55rem;
			padding: var(--pad-y) 1rem var(--pad-y) 0.75rem;
		}

		.code {
			min-width: 2.3rem;
			padding: 0.15rem 0.3rem;
			font-size: 0.68rem;
		}

		.actions {
			justify-content: flex-start;
			margin: 0 0 0 -0.65rem;
		}
	}
</style>
