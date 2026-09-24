<script lang="ts">
	/**
	 * Los fondos que sigue el usuario, uno por tarjeta: lo que vale, a qué valor
	 * de unidad y de qué día, y lo que gana sobre lo que costó.
	 *
	 * Un fondo sin marca vale lo que costó, y la tarjeta lo dice en vez de
	 * enseñar una ganancia de cero: no es que no se haya movido, es que nadie ha
	 * dicho cuánto vale. Una marca vieja se avisa por lo mismo.
	 *
	 * Un fondo que se sigue por saldo no enseña unidades: son de Finexia, no del
	 * extracto, y no significan nada fuera de aquí. Lo que se anota en él son
	 * saldos, aportes y retiros.
	 */
	import Button from '$lib/ui/button.svelte';
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import { formatSignedPercent } from '$lib/shared/format/percent';
	import { formatCalendarDate } from '$lib/shared/format/date';
	import { daysSinceMark, fundGain, fundPlatforms, isStale, type Fund } from '../funds';

	interface Props {
		funds: Fund[];
		today: string;
		/** Con un solo portafolio no hay otro sitio donde pueda estar: no se nombra. */
		showPortfolio: boolean;
		onMark: (fund: Fund) => void;
		/** Aportes y retiros, solo en un fondo que se sigue por saldo. */
		onMove: (fund: Fund) => void;
	}

	let { funds, today, showPortfolio, onMark, onMove }: Props = $props();

	const money = (value: string | number, currency: string) =>
		privacy.money(
			formatCurrency(typeof value === 'number' ? value : parseFloat(value) || 0, currency)
		);

	const unitValue = (value: string, currency: string) =>
		privacy.money(formatCurrency(parseFloat(value) || 0, currency, 6));

	const day = (iso: string) =>
		formatCalendarDate(iso, { day: 'numeric', month: 'short', year: 'numeric' });

	const units = (value: string) =>
		new Intl.NumberFormat('es-CO', { maximumFractionDigits: 8 }).format(parseFloat(value) || 0);
</script>

<ul class="funds">
	{#each funds as fund (fund.assetId)}
		{@const gain = fundGain(fund)}
		{@const stale = isStale(fund, today)}
		{@const age = daysSinceMark(fund, today)}
		{@const byBalance = fund.tracking === 'balance'}
		<li class="fund">
			<header>
				<div class="title">
					<h3>{fund.name}</h3>
					<p class="where">
						{fundPlatforms(fund) || 'Sin posiciones'}
						{#if showPortfolio && fund.positions.length > 0}
							· {[...new Set(fund.positions.map((p) => p.portfolioName))].join(', ')}
						{/if}
					</p>
				</div>
				<div class="actions">
					{#if byBalance}
						<Button type="button" variant="ghost" size="sm" onclick={() => onMove(fund)}>
							Aportar o retirar
						</Button>
					{/if}
					<Button type="button" variant="ghost" size="sm" onclick={() => onMark(fund)}>
						{byBalance ? 'Actualizar saldo' : 'Actualizar valor'}
					</Button>
				</div>
			</header>

			<p class="value">{money(fund.value, fund.currency)}</p>

			{#if gain}
				<p class="gain" class:loss={gain.amount < 0}>
					{gain.amount >= 0 ? '+' : '−'}{money(Math.abs(gain.amount), fund.currency)}
					{#if gain.pct !== null}
						<span>({formatSignedPercent(gain.pct, 2)})</span>
					{/if}
					<span class="dim">sobre lo que costó</span>
				</p>
			{:else}
				<p class="at-cost">
					Sin valor actualizado: vale lo que costó. Escribe {byBalance
						? 'el saldo que muestra tu app'
						: 'el valor de unidad de tu extracto'} para ver lo que gana.
				</p>
			{/if}

			<dl class="figures">
				{#if !byBalance}
					<div>
						<dt>Unidades</dt>
						<dd>{units(fund.units)}</dd>
					</div>
				{/if}
				<div>
					<dt>Costo</dt>
					<dd>{money(fund.cost, fund.currency)}</dd>
				</div>
				{#if byBalance && fund.valuedOn}
					<div>
						<dt>Último saldo anotado</dt>
						<dd>{day(fund.valuedOn)}</dd>
					</div>
				{:else if fund.unitValue && fund.valuedOn}
					<div>
						<dt>Valor de unidad al {day(fund.valuedOn)}</dt>
						<dd>{unitValue(fund.unitValue, fund.currency)}</dd>
					</div>
				{/if}
			</dl>

			{#if stale && age !== null}
				<p class="feedback warning stale">
					El último valor es de hace {age} días. Actualízalo con tu extracto: Finexia no estima lo que
					pasó desde entonces.
				</p>
			{/if}
		</li>
	{/each}
</ul>

<style>
	.funds {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(min(100%, 20rem), 1fr));
		gap: 1rem;
		margin: 0;
		padding: 0;
		list-style: none;
	}

	.fund {
		display: grid;
		gap: 0.75rem;
		align-content: start;
		padding: 1.25rem;
		border: 1px solid var(--border);
		border-radius: 12px;
		background: var(--surface);
	}

	header {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 0.75rem;
	}

	.title {
		min-width: 0;
	}

	.actions {
		display: flex;
		flex-wrap: wrap;
		justify-content: flex-end;
		gap: 0.25rem;
	}

	h3 {
		margin: 0;
		font-size: 1rem;
		font-weight: 500;
		color: var(--text);
		overflow-wrap: anywhere;
	}

	.where {
		margin: 0.2rem 0 0;
		font-size: 0.8rem;
		color: var(--text-dim);
	}

	.value {
		margin: 0;
		font-family: var(--font-figures);
		font-size: 1.6rem;
		font-variant-numeric: tabular-nums;
		color: var(--text);
		overflow-wrap: anywhere;
	}

	.gain {
		margin: 0;
		font-size: 0.88rem;
		font-variant-numeric: tabular-nums;
		color: var(--green);
	}

	.gain.loss {
		color: var(--red);
	}

	.dim {
		color: var(--text-dim);
	}

	.at-cost {
		margin: 0;
		font-size: 0.83rem;
		line-height: 1.5;
		color: var(--text-muted);
	}

	.figures {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(7rem, 1fr));
		gap: 0.6rem 1rem;
		margin: 0;
		padding-top: 0.75rem;
		border-top: 1px solid var(--border);
	}

	.figures div {
		display: grid;
		gap: 0.15rem;
		min-width: 0;
	}

	dt {
		font-size: 0.72rem;
		color: var(--text-dim);
	}

	dd {
		margin: 0;
		font-family: var(--font-mono);
		font-size: 0.86rem;
		font-variant-numeric: tabular-nums;
		color: var(--text);
		overflow-wrap: anywhere;
	}

	.stale {
		margin: 0;
	}
</style>
