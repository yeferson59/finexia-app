<script lang="ts">
	/**
	 * Lo que hizo un recálculo de intereses: lo que rendían los días que ya
	 * estaban calculados, lo que rinden ahora, y aparte los días que se
	 * calcularon por primera vez.
	 *
	 * Un recálculo sobre el mismo saldo y la misma tasa da lo mismo, y sin este
	 * aviso no se distingue de uno que no se hizo. Por eso, cuando no cambia
	 * nada, dice por qué; y un día nuevo no se cuenta como cambio.
	 */
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import { formatCalendarDate } from '$lib/shared/format/date';
	import type { CashRecalcChange, CashRecalcDone } from '../interest';
	import type { CashRecalcFeedback } from '../recalc.svelte';

	interface Props {
		/** El último recálculo; sin ninguno no se pinta nada. */
		feedback: CashRecalcFeedback;
	}

	let { feedback }: Props = $props();

	const money = (amount: number, currency: string) =>
		privacy.money(formatCurrency(amount, currency));

	const shortDate = (iso: string) =>
		formatCalendarDate(iso.slice(0, 10), { day: 'numeric', month: 'short' });

	/** «16 – 22 sep.», o un solo día si empieza y acaba en el mismo. */
	function span(from: string | null, through: string | null): string {
		if (!from || !through) return '';
		return from.slice(0, 10) === through.slice(0, 10)
			? shortDate(from)
			: `${shortDate(from)} – ${shortDate(through)}`;
	}

	const plural = (n: number) => (n === 1 ? 'día' : 'días');
</script>

{#if feedback.done && feedback.change}
	{@render notice(feedback.done, feedback.change)}
{/if}

{#snippet notice(done: CashRecalcDone, change: CashRecalcChange)}
	{@const redone = done.result.before}
	{@const fresh = done.result.new}
	{@const signed = `${change.delta > 0 ? '+' : '−'}${money(Math.abs(change.delta), done.currency)}`}
	<div class="notice" class:changed={change.delta !== 0} role="status">
		<div class="body">
			<p class="head">Intereses recalculados · {done.where}</p>

			{#if redone.days > 0}
				<p class="figures">
					<span class="label"
						>{redone.days} {plural(redone.days)} ({span(redone.from, redone.through)})</span
					>
					{#if change.delta !== 0}
						<span class="before">{money(change.before, done.currency)}</span>
						<span class="arrow" aria-label="pasa a">→</span>
						<span class="after">{money(change.after, done.currency)}</span>
						<span class="delta" class:down={change.delta < 0}>{signed}</span>
					{:else}
						<span class="after">{money(change.after, done.currency)}</span>
						<span class="same">sin cambios</span>
					{/if}
				</p>
			{:else}
				<p class="figures">No había días calculados que rehacer.</p>
			{/if}

			{#if fresh.days > 0}
				<p class="figures fresh">
					<span class="label">
						{fresh.days}
						{fresh.days === 1 ? 'día nuevo' : 'días nuevos'} ({span(fresh.from, fresh.through)})
					</span>
					<span class="delta">+{money(change.fresh, done.currency)}</span>
				</p>
			{/if}

			<p class="detail">
				{#if change.delta === 0 && redone.days > 0}
					El saldo y la tasa de esos días son los mismos que ya se habían usado. El resultado cambia
					si anotas un depósito o un retiro con fecha pasada.
				{/if}
				{#if fresh.days > 0}
					Los días nuevos no estaban calculados todavía: suman aunque nada haya cambiado.
				{/if}
				{#if change.beforeRate && done.rateFrom}
					La tasa empieza el {shortDate(done.rateFrom)} y los días anteriores no generan intereses. Si
					la cuenta rendía desde antes, usa «Cambiar fecha» en su tasa.
				{/if}
			</p>
		</div>

		<button
			type="button"
			class="dismiss"
			onclick={() => feedback.dismiss()}
			aria-label="Cerrar aviso">×</button
		>
	</div>
{/snippet}

<style>
	.notice {
		display: flex;
		align-items: flex-start;
		gap: 1rem;
		margin: 0 0 1.5rem;
		padding: 0.9rem 1rem;
		border: 1px solid var(--border);
		border-left: 2px solid var(--text-dim);
		border-radius: 10px;
		background: rgba(255, 255, 255, 0.02);
	}

	.notice.changed {
		border-left-color: var(--green);
	}

	.body {
		flex: 1;
		min-width: 0;
	}

	.head {
		margin: 0;
		font-size: 0.78rem;
		letter-spacing: 0.02em;
		color: var(--text-muted);
	}

	.figures {
		display: flex;
		flex-wrap: wrap;
		align-items: baseline;
		gap: 0.5rem;
		margin: 0.3rem 0 0;
		font-size: 1.05rem;
		color: var(--text);
		font-variant-numeric: tabular-nums;
	}

	.label {
		font-size: 0.83rem;
		color: var(--text-muted);
	}

	.before,
	.same {
		color: var(--text-muted);
	}

	.same {
		font-size: 0.83rem;
	}

	.arrow {
		color: var(--text-dim);
	}

	.delta {
		font-size: 0.88rem;
		color: var(--green);
	}

	.delta.down {
		color: var(--red);
	}

	.detail {
		max-width: 62ch;
		margin: 0.35rem 0 0;
		font-size: 0.83rem;
		line-height: 1.5;
		color: var(--text-muted);
	}

	.dismiss {
		flex-shrink: 0;
		padding: 0 0.25rem;
		border: 0;
		background: none;
		font-size: 1.2rem;
		line-height: 1;
		color: var(--text-dim);
		cursor: pointer;
	}

	.dismiss:hover,
	.dismiss:focus-visible {
		color: var(--text);
	}
</style>
