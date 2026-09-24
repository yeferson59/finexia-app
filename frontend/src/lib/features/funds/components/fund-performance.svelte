<script lang="ts">
	/**
	 * Cómo le fue a un fondo: su rentabilidad por periodo, como la publica la
	 * entidad, la gráfica de su valor de unidad y el dinero que entró y salió.
	 *
	 * El porcentaje y el dinero responden cosas distintas y los dos se enseñan: el
	 * porcentaje es cómo le fue al fondo —ponderado por tiempo, sin que un aporte
	 * grande cuente como rendimiento—; el dinero es cuánto ganó el dueño con sus
	 * propias fechas de entrada y salida.
	 *
	 * En un fondo que se sigue por saldo, la serie es un índice que empieza en 100
	 * el día del primer aporte: no es un precio que exista fuera de Finexia.
	 */
	import Modal from '$lib/ui/modal.svelte';
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import { formatSignedPercent } from '$lib/shared/format/percent';
	import { formatCalendarDate } from '$lib/shared/format/date';
	import { fundPeriodRows, type FundPerformance } from '../performance';
	import type { Fund } from '../funds';
	import FundChart from './fund-chart.svelte';

	interface Props {
		fund: Fund | null;
		performance: FundPerformance | null;
		onClose: () => void;
	}

	let { fund, performance, onClose }: Props = $props();

	const byBalance = $derived(fund?.tracking === 'balance');
	const currency = $derived(fund?.currency ?? 'USD');

	const money = (value: string | number) =>
		privacy.money(formatCurrency(typeof value === 'number' ? value : Number(value) || 0, currency));

	const signedMoney = (value: number) => (value >= 0 ? '+' : '−') + money(Math.abs(value));

	const day = (iso: string) =>
		formatCalendarDate(iso, { day: 'numeric', month: 'short', year: 'numeric' });

	const rows = $derived(performance ? fundPeriodRows(performance.periods) : []);
	const points = $derived(
		(performance?.series ?? []).map((p) => ({
			date: p.date.slice(0, 10),
			value: Number(p.unitValue)
		}))
	);

	const realized = $derived(Number(performance?.realizedGain ?? 0));
	const unrealized = $derived(
		performance?.unrealizedGain == null ? null : Number(performance.unrealizedGain)
	);
	const total = $derived(unrealized === null ? null : realized + unrealized);

	const formatValue = (v: number) =>
		byBalance
			? new Intl.NumberFormat('es-CO', { maximumFractionDigits: 4 }).format(v)
			: privacy.money(formatCurrency(v, currency, 6));
	const formatAxis = (v: number) =>
		new Intl.NumberFormat('es-CO', { maximumFractionDigits: 2 }).format(v);
</script>

<Modal
	open={fund !== null}
	title={fund ? fund.name : 'Rentabilidad'}
	description={fund?.valuedOn
		? `Rentabilidad al ${day(fund.valuedOn)}, con los valores que has anotado.`
		: 'Todavía no hay valores anotados: el fondo vale lo que costó.'}
	size="lg"
	{onClose}
>
	{#if fund && performance}
		<dl class="figures">
			<div>
				<dt>Vale</dt>
				<dd>{money(performance.value)}</dd>
			</div>
			<div>
				<dt>Aportado</dt>
				<dd>{money(performance.invested)}</dd>
			</div>
			<div>
				<dt>Retirado</dt>
				<dd>{money(performance.withdrawn)}</dd>
			</div>
			<div>
				<dt>Ganancia total</dt>
				<dd class:gain={total !== null && total >= 0} class:loss={total !== null && total < 0}>
					{total === null ? '—' : signedMoney(total)}
				</dd>
			</div>
			<div>
				<dt>Ya realizada, en retiros</dt>
				<dd>{signedMoney(realized)}</dd>
			</div>
			<div>
				<dt>Sin realizar</dt>
				<dd>{unrealized === null ? '—' : signedMoney(unrealized)}</dd>
			</div>
		</dl>

		<section class="block" aria-labelledby="fund-returns-title">
			<h3 id="fund-returns-title">Rentabilidad por periodo</h3>
			{#if rows.every((r) => r.pct === null)}
				<p class="hint">
					Hace falta más de un valor anotado para medir la rentabilidad. Anota el de tu próximo
					extracto, o pega varios de una vez.
				</p>
			{:else}
				<table class="returns">
					<thead>
						<tr>
							<th scope="col">Periodo</th>
							<th scope="col">Rentabilidad</th>
							<th scope="col">E.A.</th>
						</tr>
					</thead>
					<tbody>
						{#each rows as row (row.key)}
							<tr>
								<th scope="row">
									{row.label}
									{#if row.from && row.days !== null}
										<span class="since">desde el {day(row.from)} · {row.days} días</span>
									{/if}
								</th>
								<td
									class:gain={row.pct !== null && row.pct >= 0}
									class:loss={row.pct !== null && row.pct < 0}
								>
									{row.pct === null ? '—' : formatSignedPercent(row.pct, 2)}
								</td>
								<td>
									{row.ea === null ? '—' : `${formatSignedPercent(row.ea, 2)}`}
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
				<p class="hint">
					Como la publica la entidad: cómo se movió el valor de unidad, sin contar aportes ni
					retiros. «—» es un periodo para el que aún no hay historia; la E.A. solo se da en periodos
					de 28 días o más.
				</p>
			{/if}
		</section>

		<section class="block" aria-labelledby="fund-chart-title">
			<h3 id="fund-chart-title">
				{byBalance ? 'Índice de rentabilidad (base 100)' : 'Valor de unidad'}
			</h3>
			{#if points.length >= 2}
				<FundChart
					{points}
					caption={byBalance ? 'Índice de rentabilidad del fondo' : 'Valor de unidad del fondo'}
					{formatValue}
					{formatAxis}
					formatDate={day}
				/>
				{#if byBalance}
					<p class="hint">
						Empieza en 100 el día del primer aporte. Sube y baja con lo que gana o pierde el fondo,
						no con lo que metes o sacas.
					</p>
				{/if}
			{:else}
				<p class="hint">La gráfica aparece en cuanto haya dos valores.</p>
			{/if}
		</section>
	{/if}
</Modal>

<style>
	.figures {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(9rem, 1fr));
		gap: 0.9rem 1.25rem;
		margin: 0;
	}

	.figures div {
		display: grid;
		gap: 0.2rem;
		min-width: 0;
	}

	dt {
		font-size: 0.74rem;
		color: var(--text-dim);
	}

	dd {
		margin: 0;
		font-family: var(--font-mono);
		font-size: 0.98rem;
		font-variant-numeric: tabular-nums;
		color: var(--text);
		overflow-wrap: anywhere;
	}

	.gain {
		color: var(--green);
	}

	.loss {
		color: var(--red);
	}

	.block {
		margin-top: 1.75rem;
		padding-top: 1.25rem;
		border-top: 1px solid var(--border);
	}

	.block h3 {
		margin: 0 0 0.75rem;
		font-size: 0.78rem;
		font-weight: 500;
		letter-spacing: 0.04em;
		text-transform: uppercase;
		color: var(--text-dim);
	}

	.returns {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.86rem;
	}

	.returns th,
	.returns td {
		padding: 0.5rem 0;
		border-bottom: 1px solid var(--border);
		text-align: right;
	}

	.returns thead th {
		font-size: 0.72rem;
		font-weight: 500;
		color: var(--text-dim);
	}

	.returns th[scope='row'],
	.returns thead th:first-child {
		text-align: left;
		font-weight: 400;
		color: var(--text);
	}

	.returns td {
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
		color: var(--text-muted);
	}

	.returns td.gain {
		color: var(--green);
	}

	.returns td.loss {
		color: var(--red);
	}

	.since {
		display: block;
		font-size: 0.72rem;
		color: var(--text-dim);
	}

	.hint {
		margin-top: 0.75rem;
	}
</style>
