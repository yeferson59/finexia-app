<script lang="ts">
	/**
	 * Buscar un fondo en el catálogo de la Superintendencia Financiera: los FIC
	 * cuyo valor de unidad se publica cada día como dato abierto.
	 *
	 * Un mismo fondo sale varias veces, una por tipo de participación, porque cada
	 * una tiene su propio valor de unidad. La lista enseña ese valor y su día al
	 * lado de cada una: comparado con el del extracto, dice cuál es la propia.
	 *
	 * Lo elegido viaja en un campo oculto (`name`), como cualquier otro campo del
	 * formulario que lo envuelve.
	 */
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import { formatCalendarDate } from '$lib/shared/format/date';
	import { publicFundDetail, shortFundName, type PublicFund } from '../funds';
	import { publicFundSearchErrorMessage } from '../schemas';

	interface Props {
		selected?: PublicFund | null;
		/** El campo oculto que lleva el id del fondo elegido. */
		name?: string;
		id?: string;
		onSelect?: (fund: PublicFund | null) => void;
	}

	let {
		selected = $bindable(null),
		name = 'publicFundId',
		id = 'public-fund-search',
		onSelect
	}: Props = $props();

	let search = $state('');
	let results = $state<PublicFund[]>([]);
	let open = $state(false);
	let searching = $state(false);
	let error = $state('');
	let active = $state(-1);
	let timer: ReturnType<typeof setTimeout>;
	/* La última búsqueda enviada: una respuesta que llega tarde no pisa a la nueva. */
	let latest = 0;

	const listId = $derived(`${id}-list`);

	async function fetchResults(q: string) {
		const ticket = ++latest;
		searching = true;
		error = '';

		try {
			const res = await fetch(`/api/funds/catalog?q=${encodeURIComponent(q)}`);
			const body = await res.json().catch(() => null);
			if (ticket !== latest) return;

			if (!res.ok || !body?.success) {
				results = [];
				error = publicFundSearchErrorMessage(res.status);
			} else {
				results = body.data ?? [];
			}
		} catch {
			if (ticket === latest) {
				results = [];
				error = publicFundSearchErrorMessage(0);
			}
		} finally {
			if (ticket === latest) {
				searching = false;
				active = -1;
			}
		}
	}

	function onInput() {
		open = true;
		clearTimeout(timer);

		const q = search.trim();
		if (q.length < 2) {
			latest++;
			results = [];
			error = '';
			searching = false;
			return;
		}

		timer = setTimeout(() => fetchResults(q), 300);
	}

	function choose(fund: PublicFund | null) {
		selected = fund;
		open = false;
		search = '';
		results = [];
		onSelect?.(fund);
	}

	function onKeydown(e: KeyboardEvent) {
		if (!open || results.length === 0) {
			if (e.key === 'ArrowDown' && results.length > 0) open = true;
			return;
		}

		switch (e.key) {
			case 'ArrowDown':
				e.preventDefault();
				active = (active + 1) % results.length;
				break;
			case 'ArrowUp':
				e.preventDefault();
				active = active <= 0 ? results.length - 1 : active - 1;
				break;
			case 'Enter':
				if (active >= 0) {
					e.preventDefault();
					choose(results[active]);
				}
				break;
			case 'Escape':
				open = false;
				break;
		}
	}

	function clickOutside(node: HTMLElement) {
		function listener(e: MouseEvent) {
			if (!node.contains(e.target as Node)) open = false;
		}
		document.addEventListener('mousedown', listener);
		return {
			destroy() {
				document.removeEventListener('mousedown', listener);
			}
		};
	}

	const day = (iso: string) => formatCalendarDate(iso, { day: 'numeric', month: 'short' });

	const unitValue = (value: string) =>
		privacy.money(formatCurrency(parseFloat(value) || 0, 'COP', 2));
</script>

<div class="picker" use:clickOutside>
	<input type="hidden" {name} value={selected?.id ?? ''} />

	{#if selected}
		<div class="chosen">
			<div class="chosen-text">
				<span class="chosen-name">{shortFundName(selected.fundName)}</span>
				<span class="chosen-meta"
					>{selected.entityName} · Participación {selected.participation}</span
				>
			</div>
			<button type="button" class="change" onclick={() => choose(null)}>Cambiar</button>
		</div>
	{:else}
		<div class="input-wrap">
			<input
				{id}
				type="text"
				role="combobox"
				autocomplete="off"
				placeholder="Fiducuenta, Renta Liquidez, Bancolombia…"
				aria-expanded={open && results.length > 0}
				aria-controls={listId}
				aria-autocomplete="list"
				aria-activedescendant={active >= 0 ? `${listId}-${active}` : undefined}
				bind:value={search}
				oninput={onInput}
				onfocus={() => (open = true)}
				onkeydown={onKeydown}
			/>
			{#if searching}
				<span class="spinner" aria-hidden="true"></span>
			{/if}
		</div>

		{#if error}
			<p class="feedback error" role="alert">{error}</p>
		{:else if open && !searching && search.trim().length >= 2 && results.length === 0}
			<p class="hint">
				No encontramos ese fondo en la Superfinanciera. Prueba con otra palabra del nombre o con la
				entidad. Si no está, escribe tú el valor de unidad.
			</p>
		{/if}

		{#if open && results.length > 0}
			<ul class="panel" id={listId} role="listbox">
				{#each results as fund, i (fund.id)}
					<li
						id={`${listId}-${i}`}
						role="option"
						aria-selected={i === active}
						class:active={i === active}
						onmousedown={() => choose(fund)}
					>
						<div class="option-text">
							<span class="option-name">{shortFundName(fund.fundName)}</span>
							<span class="option-meta">{fund.entityName}</span>
							<span class="option-meta">{publicFundDetail(fund)}</span>
						</div>
						<div class="option-value">
							<span>{unitValue(fund.unitValue)}</span>
							<span class="option-meta">al {day(fund.valueDate)}</span>
						</div>
					</li>
				{/each}
			</ul>
		{/if}
	{/if}
</div>

<style>
	.picker {
		position: relative;
		display: grid;
		gap: 0.4rem;
	}

	.input-wrap {
		position: relative;
		display: flex;
		align-items: center;
	}

	.input-wrap input {
		width: 100%;
		padding-right: 2.5rem;
	}

	.spinner {
		position: absolute;
		right: 0.9rem;
		width: 14px;
		height: 14px;
		border: 2px solid var(--border-strong);
		border-top-color: var(--text-muted);
		border-radius: 50%;
		animation: spin 0.6s linear infinite;
	}

	.panel {
		position: absolute;
		top: calc(100% + 4px);
		left: 0;
		right: 0;
		z-index: 50;
		max-height: 300px;
		margin: 0;
		padding: 0.35rem 0;
		overflow-y: auto;
		list-style: none;
		background: #101114;
		border: 1px solid var(--border-strong);
		border-radius: 10px;
		box-shadow: 0 12px 32px rgba(0, 0, 0, 0.4);
	}

	.panel li {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 0.75rem;
		padding: 0.6rem 1rem;
		cursor: pointer;
	}

	.panel li:hover,
	.panel li.active {
		background: var(--surface-2);
	}

	.option-text,
	.chosen-text {
		display: grid;
		gap: 0.1rem;
		min-width: 0;
	}

	.option-name,
	.chosen-name {
		font-size: 0.88rem;
		color: var(--text);
		overflow-wrap: anywhere;
	}

	.option-meta,
	.chosen-meta {
		font-size: 0.74rem;
		color: var(--text-dim);
		overflow-wrap: anywhere;
	}

	.option-value {
		display: grid;
		justify-items: end;
		gap: 0.1rem;
		flex-shrink: 0;
		font-family: var(--font-mono);
		font-size: 0.82rem;
		font-variant-numeric: tabular-nums;
		color: var(--text-muted);
	}

	.chosen {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.75rem;
		padding: 0.65rem 0.9rem;
		border: 1px solid var(--amber);
		border-radius: 9px;
		background: rgba(212, 145, 42, 0.06);
	}

	.change {
		flex-shrink: 0;
		padding: 0.3rem 0.65rem;
		border: 1px solid var(--border-strong);
		border-radius: 7px;
		background: none;
		font: inherit;
		font-size: 0.8rem;
		color: var(--text);
		cursor: pointer;
	}

	.change:hover {
		border-color: var(--text-dim);
	}

	.feedback,
	.hint {
		margin: 0;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.spinner {
			animation-duration: 1.5s;
		}
	}
</style>
