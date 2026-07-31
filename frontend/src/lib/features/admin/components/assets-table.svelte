<script lang="ts">
	/**
	 * Catálogo de activos con edición del precio manual fila a fila.
	 *
	 * El precio de mercado no se toca desde aquí: lo sincroniza cada usuario con
	 * su propia clave. Este campo es el respaldo manual del catálogo.
	 */
	import { enhance } from '$app/forms';
	import Badge from '$lib/ui/badge.svelte';
	import Button from '$lib/ui/button.svelte';
	import DataTable from '$lib/ui/data-table.svelte';
	import Pagination from '$lib/ui/pagination.svelte';
	import AdminTableCard from './admin-table-card.svelte';
	import { formatDateTime, formatPrice, type Asset } from '../admin';

	interface Props {
		assets: Asset[];
		/** `form` de la página, para el resultado del ajuste por fila. */
		form: Record<string, unknown> | null;
	}

	let { assets, form }: Props = $props();

	const PER_PAGE = 20;
	let page = $state(1);
	const pagedAssets = $derived(assets.slice((page - 1) * PER_PAGE, page * PER_PAGE));

	let updatingId = $state<string | null>(null);
	let priceInputs = $state<Record<string, string>>({});

	$effect(() => {
		for (const asset of assets) {
			if (!(asset.id in priceInputs)) {
				priceInputs[asset.id] = asset.currentPrice?.value ?? '';
			}
		}
	});
</script>

<AdminTableCard>
	{#if assets.length === 0}
		<p class="empty-state">No hay activos en el sistema.</p>
	{:else}
		<DataTable>
			<thead>
				<tr>
					<th>Ticker</th>
					<th>Nombre</th>
					<th>Tipo</th>
					<th>Origen</th>
					<th>Precio manual</th>
					<th>Actualizado</th>
					<th>Nuevo precio</th>
				</tr>
			</thead>
			<tbody>
				{#each pagedAssets as asset (asset.id)}
					{@const isUpdating = updatingId === asset.id}
					{@const hasUpdateError = form?.updateError && form?.errorId === asset.id}
					{@const hasUpdateSuccess = form?.updateSuccess && form?.updatedId === asset.id}
					<tr class:row-success={hasUpdateSuccess}>
						<td class="cell-ticker">{asset.ticker}</td>
						<td class="cell-name">{asset.name}</td>
						<td>
							<Badge tone="neutral">{asset.assetType}</Badge>
						</td>
						<td>
							<!-- Un activo aportado por un usuario solo lo ve quien lo aportó;
							     crearlo aquí con el mismo ticker lo cura para todos. -->
							<Badge tone={asset.isCurated ? 'info' : 'warning'}>
								{asset.isCurated ? 'Catálogo' : 'Aportado'}
							</Badge>
						</td>
						<td class="cell-price">{formatPrice(asset.currentPrice, asset.currency)}</td>
						<td class="cell-date">{formatDateTime(asset.priceUpdatedAt)}</td>
						<td class="cell-update">
							<form
								method="POST"
								action="?/updatePrice"
								use:enhance={() => {
									updatingId = asset.id;
									return async ({ update }) => {
										updatingId = null;
										await update({ reset: false });
									};
								}}
							>
								<input type="hidden" name="id" value={asset.id} />
								<input
									type="hidden"
									name="currency"
									value={asset.currentPrice?.currency ?? asset.currency ?? 'USD'}
								/>
								<div class="update-row">
									<input
										type="number"
										name="price"
										class="price-input"
										class:input-error={hasUpdateError}
										bind:value={priceInputs[asset.id]}
										min="0.0001"
										step="any"
										placeholder="0.00"
										required
									/>
									<Button type="submit" size="sm" variant="secondary" loading={isUpdating}>
										OK
									</Button>
								</div>
								{#if hasUpdateError}
									<p class="row-error">{form.updateError}</p>
								{/if}
							</form>
						</td>
					</tr>
				{/each}
			</tbody>
		</DataTable>
		<div class="pagination-wrap">
			<Pagination bind:page total={assets.length} perPage={PER_PAGE} label="activos" />
		</div>
	{/if}
</AdminTableCard>

<style>
	.pagination-wrap {
		padding: 0 1.25rem 0.5rem;
	}
</style>
