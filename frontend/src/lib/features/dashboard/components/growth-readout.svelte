<script lang="ts">
	/*
	 * El detalle del punto señalado en la gráfica: la fecha y las dos series.
	 *
	 * Salió de `portfolio-growth` cuando la tarjeta pasó del presupuesto de 500
	 * líneas. Es además la parte que menos tiene que ver con el resto: aquella
	 * decide qué se mide y esta solo lo escribe.
	 *
	 * Sin `aria-live`: la gráfica es un `slider` y su `aria-valuetext` ya anuncia
	 * el punto; repetirlo aquí haría que el lector de pantalla lo dijera dos
	 * veces.
	 */
	import type { GrowthPoint } from '../dashboard';

	interface Props {
		/** Punto bajo el cursor, o `null` cuando no hay ninguno. */
		point: GrowthPoint | null;
		primaryLabel: string;
		secondaryLabel: string;
		/** La gráfica está en rentabilidad; cambia qué es la diferencia y su color. */
		isPercent: boolean;
		formatValue: (value: number) => string;
		formatDate: (iso: string) => string;
	}

	let { point, primaryLabel, secondaryLabel, isPercent, formatValue, formatDate }: Props = $props();

	const gap = $derived(point ? point.mv - point.cb : 0);
</script>

<p class="readout">
	{#if point}
		<span class="date">{formatDate(point.date)}</span>
		<span class="item">
			<i class="swatch value"></i>{primaryLabel}
			<b>{formatValue(point.mv)}</b>
		</span>
		<span class="item">
			<i class="swatch cost"></i>{secondaryLabel}
			<b>{formatValue(point.cb)}</b>
		</span>
		<!--
			La diferencia entre los dos trazos: en dinero es la ganancia de ese día;
			en porcentaje, cuánto se separan las dos lecturas, que es justo lo que
			aportaron los movimientos de dinero. Aquella se tiñe de verde o rojo
			porque es ganar o perder; esta va en gris: que la rentabilidad vaya por
			encima de la ganancia sobre coste no es bueno ni malo, solo dice cuándo
			entró el dinero.
		-->
		<span class="item" class:up={!isPercent && gap >= 0} class:down={!isPercent && gap < 0}>
			{isPercent ? 'Diferencia' : 'Ganancia'}
			<b>
				{#if isPercent}
					{formatValue(gap)}
				{:else}
					{gap >= 0 ? '+' : '−'}{formatValue(Math.abs(gap))}
				{/if}
			</b>
		</span>
	{:else}
		<span class="hint">
			Pasa el cursor por la gráfica —o enfócala y usa las flechas— para ver el detalle de cada día.
		</span>
	{/if}
</p>

<style>
	/* Alto fijo: el detalle aparece y desaparece sin mover la gráfica. */
	.readout {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 0.5rem 1.25rem;
		min-height: 2.4rem;
		margin: 0 0 0.5rem;
		font-size: 0.78rem;
		color: var(--text-muted);
	}

	.date {
		font-family: var(--font-mono);
		font-size: 0.7rem;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		color: var(--text-dim);
	}

	.item {
		display: inline-flex;
		align-items: center;
		gap: 0.4rem;
	}

	.item b {
		font-family: var(--font-mono);
		font-weight: 600;
		font-variant-numeric: tabular-nums;
		color: var(--text);
	}

	.item.up b {
		color: var(--green);
	}

	.item.down b {
		color: var(--red);
	}

	.swatch {
		width: 10px;
		height: 2px;
		border-radius: 1px;
	}

	.swatch.value {
		background: var(--amber);
	}

	.swatch.cost {
		background: var(--cost);
	}

	.hint {
		color: var(--text-dim);
	}

	@media (max-width: 600px) {
		.readout {
			min-height: 3.4rem;
		}
	}
</style>
