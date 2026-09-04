<script lang="ts">
	/**
	 * El reparto de la cuenta entre plataformas, en una sola barra.
	 *
	 * Es la respuesta a la pregunta de esta pantalla —dónde está el dinero— y la
	 * que el listado de tarjetas no daba: seis tarjetas del mismo tamaño para
	 * seis plataformas que guardan cantidades muy distintas.
	 *
	 * La barra repite datos que la tabla de abajo ya escribe con letra, así que
	 * se oculta al lector de pantalla en vez de duplicarlos.
	 */
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import { shareTint, type PlatformShare } from '../platforms';

	interface Props {
		ranked: PlatformShare[];
		/** Moneda de la cuenta: en ella vienen convertidos los totales. */
		currency: string;
		/** Posiciones que entraron al total sin convertir, por falta de tasa. */
		unconverted: number;
	}

	let { ranked, currency, unconverted }: Props = $props();

	const total = $derived(
		ranked.reduce((sum, { platform }) => sum + (parseFloat(platform.totalValue) || 0), 0)
	);

	const count = $derived(ranked.length);
</script>

<section class="allocation" aria-labelledby="allocation-total">
	<p class="total" id="allocation-total">
		<span class="amount">{privacy.money(formatCurrency(total, currency))}</span>
		<span class="caption">
			repartidos en {count}
			{count === 1 ? 'plataforma' : 'plataformas'}
		</span>
	</p>

	<div class="bar" aria-hidden="true">
		{#each ranked as { platform, share, rank } (platform.id)}
			<span
				class="segment"
				style="flex-grow: {Math.max(share, 0.4)}; --tint: {shareTint(rank, count)}"
			></span>
		{/each}
	</div>

	{#if unconverted > 0}
		<p class="fx-note">
			{unconverted}
			{unconverted === 1 ? 'posición sigue contada' : 'posiciones siguen contadas'} en su propia moneda
			por no haber tasa de cambio guardada: el reparto suma monedas distintas.
		</p>
	{/if}
</section>

<style>
	.allocation {
		margin-bottom: 2.5rem;
	}

	.total {
		margin: 0 0 1rem;
		display: flex;
		align-items: baseline;
		flex-wrap: wrap;
		gap: 0.6rem;
	}

	.amount {
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
		font-size: clamp(1.75rem, 4vw, 2.4rem);
		font-weight: 400;
		letter-spacing: -0.02em;
		color: var(--text);
	}

	.caption {
		font-size: 0.95rem;
		font-weight: 300;
		color: var(--text-muted);
	}

	.bar {
		display: flex;
		gap: 2px;
		height: 10px;
		border-radius: 2px;
		overflow: hidden;
	}

	.segment {
		background: var(--amber);
		opacity: var(--tint);
		min-width: 3px;
	}

	/* El único momento de movimiento de la pantalla: la barra se dibuja una vez
	   al entrar y lleva la vista al reparto antes que a la tabla. */
	@media (prefers-reduced-motion: no-preference) {
		.bar {
			animation: draw 0.55s cubic-bezier(0.22, 1, 0.36, 1) both;
		}
	}

	@keyframes draw {
		from {
			clip-path: inset(0 100% 0 0);
		}
		to {
			clip-path: inset(0 0 0 0);
		}
	}

	.fx-note {
		margin: 1rem 0 0;
		padding: 0.5rem 0.7rem;
		border: 1px solid rgba(212, 145, 42, 0.3);
		border-radius: 8px;
		background: rgba(212, 145, 42, 0.08);
		color: rgba(236, 234, 229, 0.75);
		font-size: 0.78rem;
		line-height: 1.4;
	}
</style>
