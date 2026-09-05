<script lang="ts">
	/**
	 * Moneda en la que se miran los importes de la pantalla.
	 *
	 * Era un par de pestañas cuando solo había USD y COP. Con la lista completa
	 * de monedas convertibles no caben en la cabecera, así que es un desplegable:
	 * mismo comportamiento —cambia `?currency=` y recarga— en el ancho de una.
	 *
	 * Solo cambia la vista: la preferencia de la cuenta se toca en ajustes, y es
	 * la que manda cuando no hay parámetro en la URL.
	 *
	 * Vivía dentro del panel. Bajó aquí cuando la vista consolidada de activos
	 * pasó a necesitarlo: allí los pesos de cada activo solo significan algo si
	 * todas las filas llegan convertidas a la misma moneda, y la página ya leía
	 * `?currency=` sin que nada permitiera ponerlo.
	 */
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { SUPPORTED_CURRENCIES, FALLBACK_CURRENCY } from '$lib/shared/currency';

	const { currency = FALLBACK_CURRENCY }: { currency?: string } = $props();

	function select(target: string) {
		if (target === currency) return;
		const url = new URL(page.url);
		url.searchParams.set('currency', target);
		// Only the query string changes; the pathname is already the current,
		// router-validated route, so there's no route id for resolve() to check.
		// eslint-disable-next-line svelte/no-navigation-without-resolve
		goto(url, { noScroll: true, keepFocus: true, invalidateAll: true });
	}
</script>

<select
	class="currency-select"
	aria-label="Moneda de visualización"
	value={currency}
	onchange={(e) => select(e.currentTarget.value)}
>
	{#each SUPPORTED_CURRENCIES as c (c)}
		<option value={c}>{c}</option>
	{/each}
</select>

<style>
	.currency-select {
		background: rgba(255, 255, 255, 0.03);
		border: 1px solid var(--border);
		border-radius: 8px;
		padding: 0.4rem 1.75rem 0.4rem 0.65rem;
		color: var(--amber-light);
		font-size: 0.72rem;
		font-weight: 600;
		font-family: var(--font-mono);
		cursor: pointer;
		flex-shrink: 0;
		outline: none;
		appearance: none;
		background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 24 24' fill='none' stroke='%23888' stroke-width='2'%3E%3Cpolyline points='6 9 12 15 18 9'%3E%3C/polyline%3E%3C/svg%3E");
		background-repeat: no-repeat;
		background-position: right 0.5rem center;
		transition:
			background-color 0.15s ease,
			border-color 0.15s ease;
	}

	.currency-select:hover {
		background-color: rgba(255, 255, 255, 0.05);
	}

	.currency-select:focus-visible {
		border-color: var(--amber);
	}

	.currency-select option {
		background: var(--bg);
		color: var(--text);
	}
</style>
