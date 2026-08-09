<script lang="ts">
	/** Tasas de cambio compartidas, con ajuste manual fila a fila. */
	import { enhance } from '$app/forms';
	import Badge from '$lib/ui/badge.svelte';
	import Button from '$lib/ui/button.svelte';
	import DataTable from '$lib/ui/data-table.svelte';
	import AdminTableCard from './admin-table-card.svelte';
	import {
		formatDateTime,
		formatRate,
		rateSourceLabel,
		rateSourceTone,
		type ExchangeRate
	} from '../admin';

	interface Props {
		rates: ExchangeRate[];
		/** `form` de la página, para el resultado del ajuste por fila. */
		form: Record<string, unknown> | null;
	}

	let { rates, form }: Props = $props();

	let updatingId = $state<string | null>(null);
	let rateInputs = $state<Record<string, string>>({});

	$effect(() => {
		for (const rate of rates) {
			if (!(rate.id in rateInputs)) {
				rateInputs[rate.id] = rate.rate ?? '';
			}
		}
	});
</script>

<AdminTableCard>
	{#if rates.length === 0}
		<p class="empty-state">No hay tasas de cambio en el sistema.</p>
	{:else}
		<DataTable>
			<thead>
				<tr>
					<th>Par</th>
					<th>Tasa actual</th>
					<th>Origen</th>
					<th>Fecha de tasa</th>
					<th>Actualizado</th>
					<th>Nueva tasa</th>
				</tr>
			</thead>
			<tbody>
				{#each rates as rate (rate.id)}
					{@const isUpdating = updatingId === rate.id}
					{@const hasUpdateError = form?.updateError && form?.errorId === rate.id}
					{@const hasUpdateSuccess = form?.updateSuccess && form?.updatedId === rate.id}
					<tr class:row-success={hasUpdateSuccess}>
						<td class="cell-ticker">
							{rate.fromCurrency}/{rate.toCurrency}
						</td>
						<td class="cell-price">{formatRate(rate.rate)}</td>
						<td>
							<Badge tone={rateSourceTone(rate.source)}>{rateSourceLabel(rate.source)}</Badge>
						</td>
						<td class="cell-date">{formatDateTime(rate.rateDate)}</td>
						<td class="cell-date">{formatDateTime(rate.createdAt)}</td>
						<td class="cell-update">
							<form
								method="POST"
								action="?/updateRate"
								use:enhance={() => {
									updatingId = rate.id;
									return async ({ update }) => {
										updatingId = null;
										await update({ reset: false });
									};
								}}
							>
								<input type="hidden" name="id" value={rate.id} />
								<div class="update-row">
									<input
										type="number"
										name="rate"
										class="price-input"
										class:input-error={hasUpdateError}
										bind:value={rateInputs[rate.id]}
										min="0.00000001"
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
	{/if}
</AdminTableCard>

<style>
	/* El par va en una celda con dos textos, a diferencia del ticker suelto de
	   la tabla de activos. */
	.cell-ticker {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}
</style>
