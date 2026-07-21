<script lang="ts">
	interface Props {
		/** Current page, 1-based. Bindable so the parent can slice its list. */
		page?: number;
		/** Total number of items across all pages. */
		total: number;
		/** Items shown per page. */
		perPage: number;
		/** Noun used in the "N–M de T elementos" summary. */
		label?: string;
	}

	let { page = $bindable(1), total, perPage, label = 'elementos' }: Props = $props();

	const totalPages = $derived(Math.max(1, Math.ceil(total / perPage)));

	// Keep the page in range when the underlying list grows or shrinks
	// (e.g. after a filter or a delete removes the last row of a page).
	$effect(() => {
		if (page > totalPages) page = totalPages;
		else if (page < 1) page = 1;
	});

	const rangeStart = $derived(total === 0 ? 0 : (page - 1) * perPage + 1);
	const rangeEnd = $derived(Math.min(page * perPage, total));

	// Compact, windowed page list with `…` gaps around the current page.
	const items = $derived.by<(number | '…')[]>(() => {
		if (totalPages <= 7) {
			return Array.from({ length: totalPages }, (_, i) => i + 1);
		}
		const pages: (number | '…')[] = [1];
		const from = Math.max(2, page - 1);
		const to = Math.min(totalPages - 1, page + 1);
		if (from > 2) pages.push('…');
		for (let p = from; p <= to; p++) pages.push(p);
		if (to < totalPages - 1) pages.push('…');
		pages.push(totalPages);
		return pages;
	});

	function go(target: number) {
		page = Math.min(totalPages, Math.max(1, target));
	}
</script>

{#if totalPages > 1}
	<nav class="pagination" aria-label="Paginación">
		<p class="summary">
			{rangeStart}–{rangeEnd} de {total}
			{label}
		</p>

		<div class="controls">
			<button
				type="button"
				class="page-btn nav"
				onclick={() => go(page - 1)}
				disabled={page <= 1}
				aria-label="Página anterior"
			>
				←
			</button>

			{#each items as item, i (typeof item === 'number' ? `p${item}` : `gap${i}`)}
				{#if item === '…'}
					<span class="gap" aria-hidden="true">…</span>
				{:else}
					<button
						type="button"
						class="page-btn"
						class:active={item === page}
						aria-current={item === page ? 'page' : undefined}
						aria-label={`Página ${item}`}
						onclick={() => go(item)}
					>
						{item}
					</button>
				{/if}
			{/each}

			<button
				type="button"
				class="page-btn nav"
				onclick={() => go(page + 1)}
				disabled={page >= totalPages}
				aria-label="Página siguiente"
			>
				→
			</button>
		</div>
	</nav>
{/if}

<style>
	.pagination {
		display: flex;
		align-items: center;
		justify-content: space-between;
		flex-wrap: wrap;
		gap: 0.75rem 1rem;
		padding: 1rem 0.25rem 0.25rem;
	}

	.summary {
		margin: 0;
		font-family: var(--font-mono);
		font-size: 0.75rem;
		color: var(--text-dim);
		font-variant-numeric: tabular-nums;
	}

	.controls {
		display: flex;
		align-items: center;
		gap: 0.35rem;
		margin-left: auto;
	}

	.page-btn {
		min-width: 2rem;
		height: 2rem;
		padding: 0 0.55rem;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		border: 1px solid var(--border);
		border-radius: 8px;
		background: transparent;
		color: var(--text-muted);
		font-family: var(--font-mono);
		font-size: 0.8rem;
		font-variant-numeric: tabular-nums;
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.page-btn:hover:not(:disabled):not(.active) {
		border-color: var(--amber);
		color: var(--amber);
	}

	.page-btn.active {
		background: var(--amber);
		border-color: var(--amber);
		color: #0d0800;
		font-weight: 700;
	}

	.page-btn:disabled {
		opacity: 0.4;
		cursor: not-allowed;
	}

	.gap {
		color: var(--text-dim);
		padding: 0 0.15rem;
		font-size: 0.8rem;
		user-select: none;
	}

	@media (max-width: 560px) {
		.pagination {
			justify-content: center;
		}

		.controls {
			margin-left: 0;
		}
	}
</style>
