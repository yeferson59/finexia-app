<script lang="ts">
	/**
	 * Los aportes de `/apoyar`, del más nuevo al más viejo.
	 *
	 * Solo se leen: el estado lo escriben Bold y la conciliación, y cualquier
	 * cosa que se haga con un pago —una devolución, una anulación— se hace en el
	 * panel de Bold. Por eso la fila lleva el id del pago allí y no acciones.
	 *
	 * El filtro y la paginación son del servidor, así que van en formularios GET
	 * como la tabla de usuarios: funcionan sin JS y la URL se puede compartir.
	 */
	import Badge from '$lib/ui/badge.svelte';
	import DataTable from '$lib/ui/data-table.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import AdminBlock from './admin-block.svelte';
	import { formatDateTime, type PageMeta } from '../admin';
	import {
		CONTRIBUTION_STATUSES,
		contributionStatusLabel,
		contributionStatusTone,
		paymentMethodLabel,
		type AdminContribution,
		type SupportStatus
	} from '../contributions';

	interface Props {
		contributions: AdminContribution[];
		meta: PageMeta;
		/** El filtro activo; sin él, todos. */
		status?: SupportStatus;
		/** La frase de estado del bloque. */
		summary: string;
	}

	let { contributions, meta, status, summary }: Props = $props();

	const cop = (value: number) => formatCurrency(value, 'COP');
</script>

<AdminBlock title="Historial" {summary}>
	{#snippet actions()}
		<form class="filter" method="GET" aria-label="Filtrar por estado">
			<button type="submit" class="chip" aria-pressed={!status}>Todos</button>
			{#each CONTRIBUTION_STATUSES as option (option.value)}
				<button
					type="submit"
					name="status"
					value={option.value}
					class="chip"
					aria-pressed={status === option.value}
				>
					{option.label}
				</button>
			{/each}
		</form>
	{/snippet}

	{#if contributions.length === 0}
		<p class="empty">
			{status ? 'No hay aportes en este estado.' : 'Todavía no ha llegado ningún aporte.'}
		</p>
	{:else}
		<DataTable caption="Aportes recibidos por Bold, del más reciente al más antiguo">
			<thead>
				<tr>
					<th>Fecha</th>
					<th class="num">Monto</th>
					<th>Estado</th>
					<th>Medio</th>
					<th>Orden</th>
				</tr>
			</thead>
			<tbody>
				{#each contributions as c (c.orderId)}
					<tr class:row-muted={c.status === 'created'}>
						<td class="cell-age">{formatDateTime(c.createdAt)}</td>
						<td class="num">
							{cop(c.amount)}
							<!-- Lo cobrado solo se enseña cuando no es lo pedido: con
							     pesos enteros casi nunca difiere, y repetirlo en cada fila
							     sería ruido. -->
							{#if c.totalCharged !== null && c.totalCharged !== c.amount}
								<span class="cell-sector">cobrado {cop(c.totalCharged)}</span>
							{/if}
						</td>
						<td>
							<Badge tone={contributionStatusTone(c.status)}>
								{contributionStatusLabel(c.status)}
							</Badge>
						</td>
						<td>{paymentMethodLabel(c.paymentMethod)}</td>
						<td>
							<span class="cell-key order">{c.orderId}</span>
							{#if c.paymentId}
								<span class="cell-sector" title="Id del pago en el panel de Bold">
									Bold {c.paymentId}
								</span>
							{/if}
						</td>
					</tr>
				{/each}
			</tbody>
		</DataTable>
	{/if}

	{#snippet footer()}
		{#if meta.totalPages > 1}
			<form class="pager" method="GET">
				{#if status}
					<input type="hidden" name="status" value={status} />
				{/if}
				<p class="pager-info">Página {meta.currentPage} de {meta.totalPages}</p>
				<div class="pager-controls">
					<button
						type="submit"
						name="page"
						value={meta.currentPage - 1}
						class="pager-btn"
						disabled={!meta.previous}
					>
						Anterior
					</button>
					<button
						type="submit"
						name="page"
						value={meta.currentPage + 1}
						class="pager-btn"
						disabled={!meta.next}
					>
						Siguiente
					</button>
				</div>
			</form>
		{/if}
	{/snippet}
</AdminBlock>

<style>
	.filter {
		display: flex;
		flex-wrap: wrap;
		gap: 0.35rem;
	}

	/* Texto, no botones con borde, como las acciones de fila: el filtro activo
	   se marca con el fondo de la sección abierta del menú, no con ámbar. */
	.chip {
		padding: 0.3rem 0.65rem;
		border: none;
		border-radius: 6px;
		background: none;
		font-family: var(--font-body);
		font-size: 0.8rem;
		color: var(--text-muted);
		cursor: pointer;
		transition:
			background 0.15s ease,
			color 0.15s ease;
	}

	.chip:hover {
		color: var(--text);
	}

	.chip[aria-pressed='true'] {
		background: var(--panel-2);
		color: var(--text);
		font-weight: 500;
	}

	.num {
		text-align: right;
		font-variant-numeric: tabular-nums;
		white-space: nowrap;
	}

	/* El id de la orden es largo y casi nunca se lee entero: se busca por él. */
	.order {
		font-size: 0.72rem;
		font-weight: 500;
	}

	.empty {
		margin: 0;
		padding: 2rem 0;
		font-size: 0.875rem;
		color: var(--text-muted);
	}

	.pager {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1.5rem;
	}

	.pager-info {
		margin: 0;
		font-family: var(--font-mono);
		font-size: 0.72rem;
		color: var(--text-dim);
	}

	.pager-controls {
		display: flex;
		gap: 1.25rem;
	}

	.pager-btn {
		padding: 0;
		border: none;
		background: none;
		font-family: var(--font-body);
		font-size: 0.82rem;
		color: var(--text-muted);
		cursor: pointer;
	}

	.pager-btn:hover:not(:disabled) {
		color: var(--text);
		text-decoration: underline;
		text-underline-offset: 3px;
	}

	.pager-btn:disabled {
		color: var(--text-dim);
		cursor: default;
	}

	@media (prefers-reduced-motion: reduce) {
		.chip {
			transition: none;
		}
	}
</style>
