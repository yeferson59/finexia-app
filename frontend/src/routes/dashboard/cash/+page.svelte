<script lang="ts">
	import { resolve } from '$app/paths';
	import PageHeader from '$lib/ui/page-header.svelte';
	import Button from '$lib/ui/button.svelte';
	import Card from '$lib/ui/card.svelte';
	import EmptyState from '$lib/ui/empty-state.svelte';
	import CurrencySelect from '$lib/ui/currency-select.svelte';
	import {
		CashAccounts,
		CashMovementForm,
		CashMovements,
		CashSummary,
		summarizeCash,
		type CashFormTarget
	} from '$lib/features/cash';
	import type { PageProps } from './$types';

	const { data }: PageProps = $props();

	const summary = $derived(summarizeCash(data.balances, data.currency));

	/* Sin portafolio o sin plataforma no hay saldo sobre el que anotar nada. */
	const canRecord = $derived(data.portfolios.length > 0 && data.platforms.length > 0);

	/* Con un solo portafolio no hay otro sitio donde el efectivo pueda contar: no se nombra. */
	const showPortfolio = $derived(data.portfolios.length > 1);

	let target = $state<CashFormTarget | null>(null);

	function record() {
		target = { mode: 'create' };
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
	<Card variant="elevated" padding="md">
		<EmptyState
			title="Todavía no hay efectivo registrado"
			description="Anota lo que tienes en tu banco, tu bróker o tu billetera. Suma a tu patrimonio, y los intereses que te abonen cuentan como rendimiento."
		>
			{#snippet action()}
				{#if canRecord}
					<Button type="button" onclick={record}>Registrar el primer depósito</Button>
				{/if}
			{/snippet}
		</EmptyState>
	</Card>
{:else}
	<CashSummary {summary} />

	<section class="block" aria-labelledby="cash-accounts-title">
		<h2 id="cash-accounts-title">Cuentas</h2>
		<Card variant="elevated" padding="none">
			<CashAccounts
				balances={data.balances}
				{showPortfolio}
				onRecord={(balance) => (target = { mode: 'create', balance })}
			/>
		</Card>
	</section>
{/if}

{#if data.movements.length > 0}
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
		<Card variant="elevated" padding="none">
			<CashMovements
				movements={data.movements}
				total={data.movementsTotal}
				{showPortfolio}
				onEdit={(movement) => (target = { mode: 'edit', movement })}
			/>
		</Card>
	</section>
{/if}

<CashMovementForm
	{target}
	portfolios={data.portfolios}
	platforms={data.platforms}
	balances={data.balances}
	currency={data.currency}
	onClose={() => (target = null)}
/>

<style>
	.prereq {
		margin: 0 0 1.5rem;
	}

	.prereq a {
		color: var(--amber);
	}

	.block + .block {
		margin-top: 3rem;
	}

	.block h2 {
		margin: 0 0 0.9rem;
		font-family: var(--font-display);
		font-size: 1.35rem;
		font-weight: 500;
		color: var(--text);
	}

	.block-head {
		display: flex;
		flex-wrap: wrap;
		align-items: baseline;
		justify-content: space-between;
		gap: 0 2rem;
	}

	.rule {
		max-width: 52ch;
		margin: 0 0 0.9rem;
		font-size: 0.82rem;
		line-height: 1.5;
		color: var(--text-muted);
	}

	.income {
		color: var(--green);
	}
</style>
