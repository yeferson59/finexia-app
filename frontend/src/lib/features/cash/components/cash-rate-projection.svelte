<script lang="ts">
	/**
	 * Lo que rendiría la cuenta con la tasa que se está escribiendo: al día, en
	 * 30 días y en un año, sobre el saldo de hoy y neto de retención.
	 *
	 * Sirve para compararlo con lo que abona la entidad. Con tramos, lo primero
	 * que hay que decir es a cuánto rinde el saldo en conjunto, porque ya no es la
	 * tasa de arriba.
	 */
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import { formatAnnualRate, type InterestProjection } from '../rates';

	interface Props {
		balance: number;
		currency: string;
		rate: number;
		projection: InterestProjection;
		/** La tasa a la que rinde todo el saldo con tramos; `null` sin ellos. */
		blended: number | null;
		posting: 'daily' | 'monthly';
	}

	let { balance, currency, rate, projection, blended, posting }: Props = $props();

	const money = (amount: number) => privacy.money(formatCurrency(amount, currency));
</script>

<div class="projection" aria-live="polite">
	{#if rate <= 0}
		<p class="lead">Escribe la tasa para ver cuánto rendiría la cuenta.</p>
	{:else if balance > 0}
		<p class="lead">
			{#if blended !== null}
				Con los tramos, los {money(balance)} de hoy rinden en conjunto un
				<strong>{formatAnnualRate(Number(blended.toFixed(2)))}</strong>. Rendirían
			{:else}
				Con los {money(balance)} de hoy, rendiría
			{/if}
		</p>
		<dl>
			<div>
				<dt>Al día</dt>
				<dd>+{money(projection.day)}</dd>
			</div>
			<div>
				<dt>En 30 días</dt>
				<dd>+{money(projection.month)}</dd>
			</div>
			<div class="year">
				<dt>En un año</dt>
				<dd>+{money(projection.year)}</dd>
			</div>
		</dl>
	{:else}
		<p class="lead">Cuando la cuenta tenga saldo, aquí verás cuánto rinde.</p>
	{/if}
	<p class="note">
		Se calcula cada día sobre el saldo al cierre.
		{posting === 'monthly'
			? 'Con abono mensual se guarda hasta el último día del mes y entra todo junto en un movimiento de intereses.'
			: 'Se abona solo a la mañana siguiente, como un movimiento de intereses.'}
	</p>
</div>

<style>
	.projection {
		display: grid;
		gap: 0.8rem;
		padding: 1rem;
		border: 1px solid var(--border);
		border-radius: 10px;
		background: rgba(255, 255, 255, 0.02);
	}

	.lead {
		margin: 0;
		font-size: 0.84rem;
		line-height: 1.5;
		color: var(--text-muted);
	}

	.lead strong {
		font-family: var(--font-mono);
		font-weight: 400;
		color: var(--text);
	}

	dl {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 0.75rem;
		margin: 0;
	}

	dl div {
		display: grid;
		gap: 0.15rem;
		min-width: 0;
	}

	dt {
		font-size: 0.74rem;
		color: var(--text-dim);
	}

	/* Lo que rendiría, en el verde de los intereses del extracto; el año, que es
	   la cifra que se compara con la entidad, un punto más grande. */
	dd {
		margin: 0;
		font-family: var(--font-mono);
		font-size: 0.92rem;
		font-variant-numeric: tabular-nums;
		color: var(--green);
		overflow-wrap: anywhere;
	}

	.year dd {
		font-size: 1.05rem;
	}

	.note {
		margin: 0;
		font-size: 0.76rem;
		line-height: 1.45;
		color: var(--text-dim);
	}

	@media (max-width: 520px) {
		dl {
			grid-template-columns: minmax(0, 1fr);
		}
	}
</style>
