<script lang="ts">
	/**
	 * Lo que hizo un recálculo de intereses, dicho en la página cuando ya trae
	 * los números nuevos: cuánto llevaba generado la cuenta antes y cuánto
	 * después, y cuántos días se volvieron a calcular.
	 *
	 * Un recálculo sobre el mismo saldo y la misma tasa da lo mismo, y sin este
	 * aviso no se distingue de uno que no se hizo. Por eso, cuando no cambia
	 * nada, dice por qué.
	 */
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import { formatCalendarDate } from '$lib/shared/format/date';
	import type { CashRecalcChange, CashRecalcDone } from '../interest';

	interface Props {
		done: CashRecalcDone;
		change: CashRecalcChange;
		onDismiss: () => void;
	}

	let { done, change, onDismiss }: Props = $props();

	const money = (amount: number) => privacy.money(formatCurrency(amount, done.currency));

	const longDate = (iso: string) =>
		formatCalendarDate(iso.slice(0, 10), { day: 'numeric', month: 'long' });

	const days = $derived(done.result.recomputed);
	const signed = $derived(`${change.delta > 0 ? '+' : '−'}${money(Math.abs(change.delta))}`);
</script>

<div class="notice" class:changed={change.delta !== 0} role="status">
	<div class="body">
		<p class="head">Intereses recalculados · {done.where}</p>

		{#if change.delta !== 0}
			<p class="figures">
				<span class="before">{money(done.before)}</span>
				<span class="arrow" aria-label="pasa a">→</span>
				<span class="after">{money(change.after)}</span>
				<span class="delta" class:down={change.delta < 0}>{signed}</span>
			</p>
		{:else}
			<p class="figures">
				Sin cambios: <span class="after">{money(change.after)}</span>
			</p>
		{/if}

		<p class="detail">
			{#if days > 0}
				Se volvieron a calcular {days}
				{days === 1 ? 'día' : 'días'}, hasta el {longDate(done.result.through)}.
			{:else}
				No había días con tasa que calcular desde el {longDate(done.from)}.
			{/if}
			{#if change.delta === 0 && days > 0}
				El saldo y la tasa de esos días son los mismos que ya se habían usado; el resultado cambia
				cuando anotas un depósito o un retiro con fecha pasada, o cambias la tasa.
			{/if}
			{#if change.beforeRate && done.rateFrom}
				La tasa empieza el {longDate(done.rateFrom)}: los días anteriores no generan intereses.
			{/if}
		</p>
	</div>

	<button type="button" class="dismiss" onclick={onDismiss} aria-label="Cerrar aviso">×</button>
</div>

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

	.before {
		color: var(--text-muted);
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
