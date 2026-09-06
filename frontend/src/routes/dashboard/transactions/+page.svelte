<script lang="ts">
	import { resolve } from '$app/paths';
	import PageHeader from '$lib/ui/page-header.svelte';
	import Pagination from '$lib/ui/pagination.svelte';
	import { TransactionLedger, sortByDateDesc } from '$lib/features/transactions';
	import type { PageProps } from './$types';

	const { data }: PageProps = $props();

	const ordered = $derived(sortByDateDesc(data.transactions));

	const PER_PAGE = 15;
	let page = $state(1);
	const paged = $derived(ordered.slice((page - 1) * PER_PAGE, page * PER_PAGE));
</script>

<svelte:head>
	<title>Transacciones - FINEXIA</title>
	<meta name="description" content="Historial de movimientos y estados de transacciones" />
</svelte:head>

<!--
	«Monitorea en tiempo real» prometía algo que no pasa: en Finexia los
	movimientos los anotas tú, y esta pantalla es el libro donde quedan.
-->
<PageHeader
	title="Transacciones"
	subtitle="Todo lo que has registrado, del movimiento más reciente al más antiguo."
>
	{#snippet actions()}
		<a class="import" href={resolve('/dashboard/transactions/import')}>Importar desde Excel</a>
	{/snippet}
</PageHeader>

{#if ordered.length === 0}
	<!-- Una invitación, no un aviso: quien llega el primer día no ha hecho nada
	     mal, y «No hay transacciones registradas» no dice por dónde se empieza. -->
	<p class="empty">
		Aquí queda anotada cada compra, venta, dividendo o cargo. Los registras desde la posición del
		activo, dentro de su portafolio, o subes el extracto de tu bróker de una vez con
		<a href={resolve('/dashboard/transactions/import')}>Importar desde Excel</a>.
	</p>
{:else}
	<TransactionLedger transactions={paged} />
	<Pagination bind:page total={ordered.length} perPage={PER_PAGE} label="transacciones" />
{/if}

<style>
	/* La acción secundaria de la cabecera, en el tono de la marca pero sin el
	   salto ni el halo que tenía: es un enlace a otra pantalla, no un
	   acontecimiento. Igual que «Crear portafolio» en su listado. */
	.import {
		display: inline-flex;
		align-items: center;
		padding: 0.6rem 1.15rem;
		border-radius: 9px;
		background: var(--amber);
		color: #0d0800;
		font-size: 0.88rem;
		font-weight: 600;
		text-decoration: none;
		white-space: nowrap;
		transition: background 0.2s ease;
	}

	.import:hover {
		background: var(--amber-light);
	}

	.empty {
		max-width: 62ch;
		margin: 0;
		padding: 2.5rem 0;
		font-size: 0.9rem;
		line-height: 1.6;
		color: var(--text-muted);
	}

	.empty a {
		color: var(--amber);
	}

	@media (prefers-reduced-motion: reduce) {
		.import {
			transition: none;
		}
	}
</style>
