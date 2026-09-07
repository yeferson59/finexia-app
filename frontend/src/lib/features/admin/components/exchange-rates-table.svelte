<script lang="ts">
	/**
	 * Tasas de cambio compartidas, con ajuste manual fila a fila.
	 *
	 * El origen es la única insignia de la tabla y solo la llevan las filas del
	 * feed: son las que un refresco reescribe, incluida una tasa corregida a
	 * mano, así que es lo que hay que saber antes de escribir encima. Las
	 * manuales no llevan marca porque son la norma.
	 *
	 * Las dos procedencias envejecen a ritmos distintos y por eso la columna de
	 * fecha las mide distinto: el feed corre cada hora —dos días sin moverse es
	 * un feed roto— y una paridad escrita a mano aguanta semanas.
	 */
	import { enhance } from '$app/forms';
	import Badge from '$lib/ui/badge.svelte';
	import DataTable from '$lib/ui/data-table.svelte';
	import EmptyState from '$lib/ui/empty-state.svelte';
	import AdminBlock from './admin-block.svelte';
	import { formatDateTime, formatRate, rateSourceLabel, type ExchangeRate } from '../admin';
	import { STALE_RATE_AFTER_DAYS, describeRates, formatAge, isStale } from '../desk';

	interface Props {
		rates: ExchangeRate[];
		/** `form` de la página, para el resultado del ajuste por fila. */
		form: Record<string, unknown> | null;
	}

	let { rates, form }: Props = $props();

	let updatingId = $state<string | null>(null);
	let rateInputs = $state<Record<string, string>>({});

	/** Un feed parado dos días ya es noticia; una tasa a mano, no. */
	const FEED_STALE_AFTER_DAYS = 2;
	const rateIsStale = (rate: ExchangeRate) =>
		isStale(
			rate.rateDate,
			rate.source === 'manual' ? STALE_RATE_AFTER_DAYS : FEED_STALE_AFTER_DAYS
		);

	$effect(() => {
		for (const rate of rates) {
			if (!(rate.id in rateInputs)) {
				rateInputs[rate.id] = rate.rate ?? '';
			}
		}
	});
</script>

<AdminBlock title="Tasas compartidas" summary={describeRates(rates)}>
	{#if rates.length === 0}
		<EmptyState
			title="No hay ninguna tasa guardada"
			description="Trae la TRM del feed público o escribe la primera paridad a mano."
		/>
	{:else}
		<DataTable caption="Tasas de cambio compartidas, quién las mantiene y de cuándo son">
			<thead>
				<tr>
					<th>Par</th>
					<th class="num">Tasa</th>
					<th>Origen</th>
					<th>Fecha de la tasa</th>
					<th class="num">Nueva tasa</th>
				</tr>
			</thead>
			<tbody>
				{#each rates as rate (rate.id)}
					{@const hasError = form?.updateError && form?.errorId === rate.id}
					{@const saved = form?.updateSuccess && form?.updatedId === rate.id}
					<tr>
						<td class="cell-key">{rate.fromCurrency}/{rate.toCurrency}</td>
						<td class="num">{formatRate(rate.rate)}</td>
						<td>
							{#if rate.source === 'manual'}
								<span class="by-hand">Manual</span>
							{:else}
								<Badge tone="amber">{rateSourceLabel(rate.source)}</Badge>
							{/if}
						</td>
						<td
							class="cell-age"
							class:aged={rateIsStale(rate)}
							title={formatDateTime(rate.rateDate)}
						>
							{formatAge(rate.rateDate)}
						</td>
						<td class="num">
							<form
								class="edit"
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
								<label class="sr-only" for="rate-{rate.id}">
									Nueva tasa de {rate.fromCurrency} a {rate.toCurrency}
								</label>
								<input
									id="rate-{rate.id}"
									type="number"
									name="rate"
									class="edit-input"
									class:invalid={hasError}
									bind:value={rateInputs[rate.id]}
									min="0.00000001"
									step="any"
									placeholder="0.00"
									required
								/>
								<button class="row-action" type="submit" disabled={updatingId === rate.id}>
									{updatingId === rate.id ? 'Guardando…' : 'Guardar'}
								</button>
							</form>
							{#if hasError}
								<p class="row-error">{form.updateError}</p>
							{:else if saved}
								<p class="row-note">Tasa guardada</p>
							{/if}
						</td>
					</tr>
				{/each}
			</tbody>
		</DataTable>
	{/if}
</AdminBlock>

<style>
	/* Sin insignia: lo mantenido a mano es la norma en esta tabla. */
	.by-hand {
		font-size: 0.82rem;
		color: var(--text-dim);
	}
</style>
