<script lang="ts">
	import { resolve } from '$app/paths';
	import PageHeader from '$lib/ui/page-header.svelte';
	import Button from '$lib/ui/button.svelte';
	import Card from '$lib/ui/card.svelte';
	import EmptyState from '$lib/ui/empty-state.svelte';
	import CurrencySelect from '$lib/ui/currency-select.svelte';
	import {
		CashBalancesTable,
		CashMovementForm,
		CashMovementsTable,
		CashSummary,
		summarizeCash,
		type CashFormTarget
	} from '$lib/features/cash';
	import type { PageProps } from './$types';

	const { data }: PageProps = $props();

	const summary = $derived(summarizeCash(data.balances, data.currency));

	/* Sin portafolio o sin plataforma no hay saldo sobre el que anotar nada. */
	const canRecord = $derived(data.portfolios.length > 0 && data.platforms.length > 0);

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

	<Card variant="elevated" padding="none">
		<CashBalancesTable
			balances={data.balances}
			onRecord={(balance) => (target = { mode: 'create', balance })}
		/>
	</Card>
{/if}

{#if data.movements.length > 0}
	<section class="movements" aria-labelledby="cash-movements-title">
		<h2 id="cash-movements-title">Movimientos</h2>
		<Card variant="elevated" padding="none">
			<CashMovementsTable
				movements={data.movements}
				total={data.movementsTotal}
				onEdit={(movement) => (target = { mode: 'edit', movement })}
			/>
		</Card>
	</section>
{/if}

<CashMovementForm
	{target}
	portfolios={data.portfolios}
	platforms={data.platforms}
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

	.movements {
		margin-top: 2.5rem;
	}

	.movements h2 {
		margin: 0 0 0.9rem;
		font-family: var(--font-display);
		font-size: 1.25rem;
		font-weight: 500;
		color: var(--text);
	}
</style>
