<script lang="ts">
	import { resolve } from '$app/paths';
	import PageHeader from '$lib/ui/page-header.svelte';
	import Button from '$lib/ui/button.svelte';
	import EmptyState from '$lib/ui/empty-state.svelte';
	import CurrencySelect from '$lib/ui/currency-select.svelte';
	import { todayLocalDateString } from '$lib/shared/format/date';
	import { OptimisticList } from '$lib/shared/optimistic.svelte';
	import {
		CashAccounts,
		CashDepositForm,
		CashMoveForm,
		CashMovementForm,
		CashMovements,
		CashPocketForm,
		CashRateForm,
		CashRecalcNotice,
		CashSummary,
		cashAccountRate,
		cashInterestTotal,
		cashRecalcChange,
		cashYield,
		cashYieldMap,
		groupCashAccounts,
		suggestCashPortfolio,
		summarizeCash,
		type CashAccount,
		type CashDepositTarget,
		type CashFormTarget,
		type CashMoveTarget,
		type CashPocketTarget,
		type CashRateTarget,
		type CashRecalcDone,
		type CashYieldBlock
	} from '$lib/features/cash';
	import type { CashMovement, CashPocket } from '$lib/api/types';
	import type { PageProps } from './$types';

	const { data }: PageProps = $props();

	const today = todayLocalDateString();

	const summary = $derived(summarizeCash(data.balances, data.currency));
	/* Cada cajón por su lado: la tasa media y el mapa miden lo que suma el total. */
	const accounts = $derived(groupCashAccounts(data.balances));
	const yielding = $derived(cashYield(accounts, data.rates, today));
	const yieldMap = $derived(cashYieldMap(accounts, data.rates, today));

	/* Sin portafolio o sin plataforma no hay saldo sobre el que anotar nada. */
	const canRecord = $derived(data.portfolios.length > 0 && data.platforms.length > 0);

	/* Con un solo portafolio no hay otro sitio donde el efectivo pueda contar: no se nombra. */
	const showPortfolio = $derived(data.portfolios.length > 1);

	let target = $state<CashFormTarget | null>(null);
	const pendingMovements = new OptimisticList<CashMovement>();

	/* La cuenta cuya rentabilidad está abierta. */
	let rateTarget = $state<CashRateTarget | null>(null);

	/* El bolsillo que se está abriendo o editando, y el traslado en curso. */
	let pocketTarget = $state<CashPocketTarget | null>(null);
	let moveTarget = $state<CashMoveTarget | null>(null);

	/* El depósito a tasa fija que se abre, o el que se está mirando. */
	let depositTarget = $state<CashDepositTarget | null>(null);

	/*
	 * El último recálculo de intereses: se queda arriba hasta que se cierra, y
	 * su fila destella un momento. Lo de después se lee de los saldos ya
	 * refrescados, así que es lo mismo que enseña la fila.
	 */
	let recalc = $state<CashRecalcDone | null>(null);
	let highlight = $state<string | null>(null);
	let highlightTimer: ReturnType<typeof setTimeout> | undefined;

	const recalcChange = $derived.by(() => {
		if (!recalc) return null;
		const account = accounts.find((a) => a.key === recalc?.key);
		return cashRecalcChange(recalc, account ? cashInterestTotal(account.balances) : 0);
	});

	function recalculated(done: CashRecalcDone) {
		recalc = done;
		clearTimeout(highlightTimer);
		highlight = done.key;
		highlightTimer = setTimeout(() => (highlight = null), 2500);
	}

	function record() {
		target = { mode: 'create' };
	}

	function openPocket(account: CashAccount, pocket: CashPocket | null) {
		pocketTarget = pocket
			? { mode: 'edit', pocket }
			: {
					mode: 'create',
					sourceId: account.sourceId,
					sourceName: account.sourceName,
					currency: account.currency
				};
	}

	/* El portafolio que se propone para una cuenta, y cómo se llama: aquel en
	   que ya guarda más, que es el que casi siempre se quiere. Lo comparten el
	   depósito y el traslado, que lo proponen igual. */
	function suggestPortfolio(account: CashAccount) {
		const portfolioId = suggestCashPortfolio(
			data.balances,
			account.sourceId,
			account.currency,
			data.portfolios[0]?.id ?? ''
		);

		return {
			portfolioId,
			portfolioName: data.portfolios.find((p) => p.id === portfolioId)?.name ?? ''
		};
	}

	/*
	 * Un depósito es un lote de dinero comprado una vez, así que es de un solo
	 * portafolio: se propone aquel en que la cuenta ya guarda más, como al mover.
	 * Uno ya abierto no se edita —su tasa es la del día en que se abrió—, así que
	 * lo que se abre es la ficha desde la que se cancela o se borra.
	 */
	function openDeposit(account: CashAccount, pocket: CashPocket | null) {
		if (pocket) {
			depositTarget = {
				mode: 'manage',
				pocket,
				rate: cashAccountRate(data.rates, pocket.sourceId, pocket.currency, today, pocket.id)
					.latest,
				balance: parseFloat(pocket.balance) || 0
			};

			return;
		}

		depositTarget = {
			mode: 'create',
			sourceId: account.sourceId,
			sourceName: account.sourceName,
			currency: account.currency,
			...suggestPortfolio(account)
		};
	}

	/* Mover es dentro de un portafolio: cambia de cajón o de plataforma, no de dónde cuenta. */
	function openMove(account: CashAccount) {
		moveTarget = { account, ...suggestPortfolio(account) };
	}

	/* Un bloque del mapa abre lo que rinde su cajón: la ficha si es un depósito,
	   su tasa si no. Un bloque parado lleva así directo a darle una. */
	function selectBlock(block: CashYieldBlock) {
		const account = accounts.find((a) => a.key === block.key);
		if (!account) return;

		if (block.fixed) {
			openDeposit(account, data.pockets.find((p) => p.id === account.pocketId) ?? null);
		} else {
			rateTarget = { account };
		}
	}
</script>

<svelte:head>
	<title>Efectivo - FINEXIA</title>
	<meta
		name="description"
		content="El efectivo de tus plataformas, por moneda, y sus movimientos"
	/>
</svelte:head>

<PageHeader
	title="Efectivo"
	subtitle="El dinero que tienes sin invertir, en cada plataforma y en cada moneda."
>
	{#snippet actions()}
		<CurrencySelect currency={data.currency} />
		<Button type="button" onclick={record} disabled={!canRecord}>Registrar movimiento</Button>
	{/snippet}
</PageHeader>

{#if !canRecord}
	<!-- Lo que falta, con el enlace a donde se arregla: un botón deshabilitado
	     sin explicación no dice por dónde se sigue. -->
	<p class="feedback prereq">
		Para anotar efectivo necesitas
		{#if data.portfolios.length === 0}
			un <a href={resolve('/dashboard/portfolios/add')}>portafolio</a>{data.platforms.length === 0
				? ' y '
				: ''}
		{/if}
		{#if data.platforms.length === 0}
			una <a href={resolve('/dashboard/platforms/add')}>plataforma activa</a>
		{/if}
		donde guardarlo.
	</p>
{/if}

{#if data.loadFailed}
	<p class="feedback error">No pudimos cargar tus saldos. Vuelve a intentarlo en un momento.</p>
{:else if data.balances.length === 0}
	<EmptyState
		bordered
		title="Todavía no hay efectivo registrado"
		description="Anota lo que tienes en tu banco, tu bróker o tu billetera. Suma a tu patrimonio, y los intereses que te abonen cuentan como rendimiento."
	>
		{#snippet action()}
			{#if canRecord}
				<Button type="button" onclick={record}>Registrar el primer depósito</Button>
			{/if}
		{/snippet}
	</EmptyState>
{:else}
	<CashSummary {summary} {yielding} map={yieldMap} onSelect={selectBlock} />

	{#if recalc && recalcChange}
		<CashRecalcNotice done={recalc} change={recalcChange} onDismiss={() => (recalc = null)} />
	{/if}

	<section class="block" aria-labelledby="cash-accounts-title">
		<h2 id="cash-accounts-title">Cuentas</h2>
		<CashAccounts
			balances={data.balances}
			rates={data.rates}
			pockets={data.pockets}
			{showPortfolio}
			{today}
			canTransfer={data.platforms.length > 1}
			onRecord={(balance) => (target = { mode: 'create', balance })}
			onRate={(account) => (rateTarget = { account })}
			onPocket={openPocket}
			onMove={openMove}
			onDeposit={openDeposit}
			{highlight}
		/>
	</section>
{/if}

{#if data.movements.length > 0 || pendingMovements.delta > 0}
	<section class="block" aria-labelledby="cash-movements-title">
		<header class="block-head">
			<h2 id="cash-movements-title">Movimientos</h2>
			<!-- La regla que distingue esta pantalla de un simple saldo, junto a la
			     lista donde se ve: el verde de los intereses es su leyenda. -->
			<p class="rule">
				Los depósitos y retiros no cuentan como ganancia ni como pérdida. Los
				<span class="income">intereses</span> sí: suman a tu rentabilidad.
			</p>
		</header>
		<CashMovements
			movements={data.movements}
			total={data.movementsTotal}
			{showPortfolio}
			pending={pendingMovements}
			onEdit={(movement) => (target = { mode: 'edit', movement })}
		/>
	</section>
{/if}

<CashMovementForm
	{target}
	portfolios={data.portfolios}
	platforms={data.platforms}
	balances={data.balances}
	pockets={data.pockets}
	currency={data.currency}
	pending={pendingMovements}
	onClose={() => (target = null)}
/>

<CashRateForm
	target={rateTarget}
	rates={data.rates}
	onClose={() => (rateTarget = null)}
	onRecalculated={recalculated}
/>

<CashPocketForm target={pocketTarget} onClose={() => (pocketTarget = null)} />

<CashMoveForm
	target={moveTarget}
	pockets={data.pockets}
	platforms={data.platforms}
	onClose={() => (moveTarget = null)}
/>

<CashDepositForm target={depositTarget} onClose={() => (depositTarget = null)} />

<style>
	.prereq {
		margin: 0 0 1.5rem;
	}

	.prereq a {
		color: var(--amber);
	}

	.block + .block {
		margin-top: 3.5rem;
	}

	.block h2 {
		margin: 0 0 1rem;
		font-family: var(--font-display);
		font-size: 1.5rem;
		font-weight: 300;
		letter-spacing: -0.01em;
		color: var(--text);
	}

	.block-head {
		margin-bottom: 1rem;
	}

	.block-head h2 {
		margin-bottom: 0.3rem;
	}

	.rule {
		max-width: 62ch;
		margin: 0;
		font-size: 0.84rem;
		line-height: 1.5;
		color: var(--text-muted);
	}

	.income {
		color: var(--green);
	}
</style>
