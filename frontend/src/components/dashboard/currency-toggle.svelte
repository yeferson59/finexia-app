<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';

	const SUPPORTED_CURRENCIES = ['USD', 'COP'] as const;

	const { currency = 'USD' }: { currency?: string } = $props();

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

<div class="currency-tabs" role="tablist" aria-label="Moneda de visualización">
	{#each SUPPORTED_CURRENCIES as c (c)}
		<button
			type="button"
			role="tab"
			aria-selected={currency === c}
			class="currency-btn"
			class:active={currency === c}
			onclick={() => select(c)}
		>
			{c}
		</button>
	{/each}
</div>

<style>
	.currency-tabs {
		display: flex;
		gap: 0.2rem;
		background: rgba(255, 255, 255, 0.03);
		border: 1px solid var(--border);
		border-radius: 8px;
		padding: 0.2rem;
		flex-shrink: 0;
	}

	.currency-btn {
		padding: 0.3rem 0.65rem;
		border: none;
		background: transparent;
		color: var(--text-dim);
		border-radius: 6px;
		font-size: 0.72rem;
		font-weight: 600;
		font-family: var(--font-mono);
		cursor: pointer;
		transition:
			background 0.15s ease,
			color 0.15s ease;
	}

	.currency-btn.active {
		background: rgba(212, 145, 42, 0.18);
		color: var(--amber-light);
	}

	.currency-btn:hover:not(.active) {
		background: rgba(255, 255, 255, 0.05);
		color: var(--text);
	}
</style>
