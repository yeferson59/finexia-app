<script lang="ts">
	/**
	 * El valor automático de un fondo: enlazarlo al fondo que publica la
	 * Superintendencia Financiera, o deshacer el enlace.
	 *
	 * Enlazado, el valor de unidad llega solo cada día —con dos días de retraso,
	 * que es lo que tarda en publicarse— y la gráfica de crecimiento se corrige
	 * desde el día de cada valor. Lo que escriba el dueño manda: un valor
	 * publicado nunca reemplaza uno suyo.
	 *
	 * Deshacer el enlace borra los valores que trajo, y deja los del dueño. Así un
	 * enlace al tipo de participación equivocado no deja nada detrás.
	 */
	import { enhance } from '$app/forms';
	import Button from '$lib/ui/button.svelte';
	import { submitError } from '$lib/shared/optimistic.svelte';
	import { formatCalendarDate } from '$lib/shared/format/date';
	import { shortFundName, type Fund, type PublicFund } from '../funds';
	import PublicFundPicker from './public-fund-picker.svelte';
	import type { SubmitFunction } from '@sveltejs/kit';

	interface Props {
		fund: Fund;
	}

	let { fund }: Props = $props();

	let chosen = $state<PublicFund | null>(null);
	let busy = $state(false);
	let error = $state('');

	const submit =
		(fallback: string): SubmitFunction =>
		() => {
			busy = true;
			error = '';

			return async ({ result, update }) => {
				const failed = submitError(result, fallback);
				if (failed) {
					error = failed;
				} else {
					chosen = null;
					await update({ reset: false });
				}
				busy = false;
			};
		};
</script>

<section class="link" aria-labelledby="fund-link-title">
	<h3 id="fund-link-title">Valor automático</h3>

	{#if fund.publicFund}
		<p class="linked">
			<span class="badge">Superfinanciera</span>
			<span>
				{shortFundName(fund.publicFund.fundName)} · {fund.publicFund.entityName} · Participación
				{fund.publicFund.participation}
			</span>
		</p>
		<p class="hint">
			El valor de unidad llega solo cada día, con los dos días que tarda en publicarse. Si escribes
			uno tú, el tuyo manda.
		</p>
		<form
			method="POST"
			action="?/unlinkFund"
			class="actions"
			use:enhance={submit('No pudimos desenlazar el fondo.')}
		>
			<input type="hidden" name="id" value={fund.assetId} />
			<p class="hint">Desenlazar borra los valores que trajo; los que escribiste tú se quedan.</p>
			<Button type="submit" variant="ghost" size="sm" loading={busy}>Desenlazar</Button>
		</form>
	{:else}
		<p class="hint">
			Si tu fondo es un FIC, la Superintendencia Financiera publica su valor de unidad cada día.
			Búscalo y enlázalo: el valor se actualiza solo, desde tu primera compra.
		</p>
		<form
			method="POST"
			action="?/linkFund"
			class="rail-fields"
			use:enhance={submit('No pudimos enlazar el fondo.')}
		>
			<input type="hidden" name="id" value={fund.assetId} />
			<div class="field">
				<label for="fund-link-search">Fondo en la Superfinanciera</label>
				<PublicFundPicker id="fund-link-search" bind:selected={chosen} />
			</div>
			{#if chosen}
				<p class="hint">
					Revisa que el valor de unidad del {formatCalendarDate(chosen.valueDate, {
						day: 'numeric',
						month: 'short'
					})} se parezca al de tu extracto: cada tipo de participación tiene el suyo.
				</p>
				<div class="actions">
					<Button type="submit" size="sm" loading={busy}>Enlazar</Button>
				</div>
			{/if}
		</form>
	{/if}

	{#if error}
		<p class="feedback error" role="alert">{error}</p>
	{/if}
</section>

<style>
	.link {
		display: grid;
		gap: 0.6rem;
		margin-top: 1.5rem;
		padding-top: 1.25rem;
		border-top: 1px solid var(--border);
	}

	h3 {
		margin: 0;
		font-size: 0.78rem;
		font-weight: 500;
		letter-spacing: 0.04em;
		text-transform: uppercase;
		color: var(--text-dim);
	}

	.linked {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 0.4rem 0.6rem;
		margin: 0;
		font-size: 0.86rem;
		color: var(--text);
	}

	.badge {
		padding: 0.1rem 0.45rem;
		border: 1px solid var(--amber);
		border-radius: 999px;
		font-size: 0.7rem;
		color: var(--amber);
	}

	.actions {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		justify-content: space-between;
		gap: 0.5rem 1rem;
	}

	.actions .hint {
		flex: 1 1 14rem;
	}

	.hint,
	.feedback {
		margin: 0;
	}
</style>
