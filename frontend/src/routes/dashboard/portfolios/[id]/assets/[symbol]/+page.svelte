<script lang="ts">
	import type { PageProps } from './$types';
	import { goto, invalidateAll } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { enhance } from '$app/forms';
	import DatePicker from '$components/ui/date-picker.svelte';

	const { params, data, form }: PageProps = $props();

	const entries = $derived(data.entries);
	const transactions = $derived(data.transactions ?? []);
	const portfolioTotalValue = $derived(data.portfolioTotalValue);
	const txnMeta = $derived(data.txnMeta);
	const currentPage = $derived(txnMeta.page);
	const totalPages = $derived(txnMeta.totalPages);

	let showAddForm = $state(false);
	let isSubmitting = $state(false);
	let formError = $derived(form?.success === false);

	$effect(() => {
		if (form?.success === true && !form?.edited) {
			showAddForm = false;
			sellFromTxn = null;
			goto(
				resolve('/dashboard/portfolios/[id]/assets/[symbol]', {
					id: params.id,
					symbol: params.symbol
				})
			);
		}
	});

	const TRANSACTION_TYPES = [
		{ value: 'buy', label: 'Compra' },
		{ value: 'sell', label: 'Venta' },
		{ value: 'dividend', label: 'Dividendo' },
		{ value: 'transfer_in', label: 'Transferencia entrada' },
		{ value: 'transfer_out', label: 'Transferencia salida' },
		{ value: 'fee', label: 'Comisión' },
		{ value: 'interest', label: 'Interés' },
		{ value: 'split', label: 'Split' }
	] as const;

	const TYPE_STYLE: Record<string, string> = {
		buy: 'type-buy',
		sell: 'type-sell',
		dividend: 'type-dividend',
		fee: 'type-fee',
		interest: 'type-interest',
		transfer_in: 'type-transfer',
		transfer_out: 'type-transfer',
		split: 'type-split'
	};

	const TYPE_LABEL: Record<string, string> = {
		buy: 'Compra',
		sell: 'Venta',
		dividend: 'Dividendo',
		fee: 'Comisión',
		interest: 'Interés',
		transfer_in: 'T. Entrada',
		transfer_out: 'T. Salida',
		split: 'Split'
	};

	// Default form state — entryId/currency are populated reactively by the $effect below
	let txnForm = $state({
		entryId: '',
		type: 'buy',
		quantity: '',
		price: '',
		currency: 'USD',
		fees: '',
		transactionDate: new Date().toISOString().split('T')[0],
		notes: ''
	});

	$effect(() => {
		txnForm.entryId = entries[0]?.id ?? '';
		txnForm.currency = entries[0]?.costCurrency ?? 'USD';
	});

	// Edit transaction state
	let editingTxn = $state<(typeof transactions)[0] | null>(null);
	let editForm = $state({
		type: 'buy',
		quantity: '',
		price: '',
		currency: 'USD',
		fees: '',
		transactionDate: new Date().toISOString().split('T')[0],
		notes: ''
	});
	let isEditSubmitting = $state(false);
	let editError = $state(false);
	let editErrorMessage = $state('');

	function startEdit(txn: (typeof transactions)[0]) {
		editingTxn = txn;
		editForm = {
			type: txn.type,
			quantity: txn.quantity,
			price: txn.price,
			currency: txn.currency,
			fees: txn.fees,
			transactionDate: txn.transactionDate.split('T')[0],
			notes: txn.notes
		};
		editError = false;
		editErrorMessage = '';
	}

	// Quick-sell state: sell directly from a buy lot
	let sellFromTxn = $state<(typeof transactions)[0] | null>(null);
	let sellMode = $state<'full' | 'partial'>('full');
	let sellQty = $state('');
	let sellPrice = $state('');
	let sellFees = $state('');
	let sellDate = $state(new Date().toISOString().split('T')[0]);
	let sellNotes = $state('');
	let isSellSubmitting = $state(false);
	let sellPanelEl = $state<HTMLElement | null>(null);

	$effect(() => {
		if (sellFromTxn) {
			sellMode = 'full';
			sellQty = sellFromTxn.quantity;
			sellPrice = position?.marketPrice ? position.marketPrice.toFixed(2) : sellFromTxn.price;
			sellFees = '';
			sellNotes = '';
			sellDate = new Date().toISOString().split('T')[0];
			setTimeout(() => sellPanelEl?.scrollIntoView({ behavior: 'smooth', block: 'start' }), 50);
		}
	});

	$effect(() => {
		if (sellMode === 'full' && sellFromTxn) {
			sellQty = sellFromTxn.quantity;
		}
	});

	// Classify how each type renders the form
	// 'trade'   → cantidad + precio unitario + comisión  (buy, sell, transfer_in, transfer_out)
	// 'amount'  → solo monto (dividendo, comisión, interés) — cantidad = 1 implícita
	// 'split'   → nuevas acciones recibidas, sin precio
	const txnMode = $derived(
		txnForm.type === 'split'
			? 'split'
			: ['fee', 'dividend', 'interest'].includes(txnForm.type)
				? 'amount'
				: 'trade'
	);

	const editTxnMode = $derived(
		editForm.type === 'split'
			? 'split'
			: ['fee', 'dividend', 'interest'].includes(editForm.type)
				? 'amount'
				: 'trade'
	);

	const editPriceLabel = $derived(
		editTxnMode === 'amount'
			? editForm.type === 'dividend'
				? 'Monto del dividendo'
				: editForm.type === 'interest'
					? 'Monto del interés'
					: 'Monto de la comisión'
			: 'Precio unitario'
	);

	const priceLabel = $derived(
		txnMode === 'amount'
			? txnForm.type === 'dividend'
				? 'Monto del dividendo'
				: txnForm.type === 'interest'
					? 'Monto del interés'
					: 'Monto de la comisión'
			: 'Precio unitario'
	);

	// Position metrics from trigger-maintained entry aggregates (accurate regardless of pagination).
	const position = $derived.by(() => {
		if (entries.length === 0) return null;

		const first = entries[0];

		const totalQty = entries.reduce((s, e) => s + (parseFloat(e.quantity) || 0), 0);
		const rawCost = entries.reduce(
			(s, e) => s + (parseFloat(e.quantity) || 0) * (parseFloat(e.price) || 0),
			0
		);
		const averageCost = totalQty > 0 ? rawCost / totalQty : 0;
		const totalCost = rawCost;

		const marketPrice = parseFloat(first.marketPrice) || averageCost;
		const totalValue = totalQty * marketPrice;
		const gainLoss = totalValue - totalCost;
		const gainLossPercent = totalCost > 0 ? (gainLoss / totalCost) * 100 : 0;
		const allocation = portfolioTotalValue > 0 ? (totalValue / portfolioTotalValue) * 100 : 0;

		return {
			ticker: first.ticker,
			name: first.name,
			assetType: first.assetType,
			exchange: first.exchange,
			currency: first.currency,
			costCurrency: first.costCurrency,
			marketPrice,
			totalQty,
			totalCost,
			averageCost,
			totalValue,
			gainLoss,
			gainLossPercent,
			allocation
		};
	});

	function fmt(value: number, decimals = 2): string {
		return new Intl.NumberFormat('es-CO', {
			style: 'currency',
			currency: position?.costCurrency || 'USD',
			currencyDisplay: 'narrowSymbol',
			minimumFractionDigits: decimals,
			maximumFractionDigits: decimals
		}).format(value);
	}

	function fmtPct(value: number): string {
		return `${value >= 0 ? '+' : ''}${value.toFixed(2)}%`;
	}

	function fmtDate(iso: string): string {
		return new Date(iso).toLocaleDateString('es-CO', {
			year: 'numeric',
			month: 'short',
			day: 'numeric'
		});
	}

	function goBack() {
		goto(resolve('/dashboard/portfolios/[id]', { id: params.id }));
	}
</script>

<svelte:head>
	<title>{params.symbol} - Portfolio - FINEXIA</title>
	<meta name="description" content="Detalles de la posición {params.symbol}" />
</svelte:head>

<div class="container">
	<button class="btn-back" onclick={goBack}>
		<svg
			width="20"
			height="20"
			viewBox="0 0 24 24"
			fill="none"
			stroke="currentColor"
			stroke-width="2"
		>
			<path d="M15 19l-7-7 7-7" />
		</svg>
		Volver
	</button>

	{#if !position}
		<div class="empty-state">
			<p>No se encontraron entradas para <strong>{params.symbol}</strong> en este portafolio.</p>
		</div>
	{:else}
		<!-- Header -->
		<div class="header-section">
			<div class="header-content">
				<div class="symbol-info">
					<h1>{position.ticker}</h1>
					<p class="asset-name">{position.name}</p>
					<span class="asset-badge">{position.assetType} · {position.exchange}</span>
				</div>

				<div class="price-display">
					<p class="current-price">{fmt(position.marketPrice)}</p>
					<p class="price-label">Precio de mercado</p>
				</div>
			</div>
		</div>

		<!-- Position Summary -->
		<section class="panel">
			<header class="panel-header">
				<h2>Resumen de Posición</h2>
			</header>

			<div class="metrics-grid">
				<article class="metric-card">
					<p class="metric-label">Cantidad total</p>
					<p class="metric-value">
						{position.totalQty.toLocaleString('es-CO', { maximumFractionDigits: 8 })}
						<span class="metric-unit">{position.ticker}</span>
					</p>
				</article>

				<article class="metric-card">
					<p class="metric-label">Precio promedio</p>
					<p class="metric-value">{fmt(position.averageCost)}</p>
				</article>

				<article class="metric-card">
					<p class="metric-label">Precio actual</p>
					<p class="metric-value">{fmt(position.marketPrice)}</p>
				</article>

				<article class="metric-card">
					<p class="metric-label">Costo total</p>
					<p class="metric-value">{fmt(position.totalCost, 0)}</p>
				</article>

				<article class="metric-card">
					<p class="metric-label">Valor de mercado</p>
					<p class="metric-value">{fmt(position.totalValue, 0)}</p>
				</article>

				<article class="metric-card gain">
					<p class="metric-label">Ganancia / Pérdida</p>
					<p class="metric-value {position.gainLoss >= 0 ? 'positive' : 'negative'}">
						{fmt(position.gainLoss, 0)}
					</p>
					<p class="metric-pct {position.gainLoss >= 0 ? 'positive' : 'negative'}">
						{fmtPct(position.gainLossPercent)}
					</p>
				</article>
			</div>
		</section>

		<!-- Asset Info -->
		<section class="panel">
			<header class="panel-header">
				<h2>Información del Activo</h2>
			</header>

			<div class="perf-grid">
				<article class="perf-card">
					<h3>Tipo</h3>
					<p class="perf-value">{position.assetType}</p>
				</article>

				<article class="perf-card">
					<h3>Exchange</h3>
					<p class="perf-value">{position.exchange || '—'}</p>
				</article>

				<article class="perf-card">
					<h3>Moneda</h3>
					<p class="perf-value">{position.currency}</p>
				</article>

				<article class="perf-card">
					<h3>Asignación</h3>
					<p class="perf-value">{position.allocation.toFixed(1)}%</p>
					<div class="bar-wrap">
						<div class="bar-fill" style="width: {Math.min(position.allocation, 100)}%"></div>
					</div>
				</article>

				<article class="perf-card">
					<h3>Transacciones</h3>
					<p class="perf-value">{txnMeta.total}</p>
				</article>

				<article class="perf-card">
					<h3>ROI</h3>
					<p class="perf-value {position.gainLossPercent >= 0 ? 'positive' : 'negative'}">
						{fmtPct(position.gainLossPercent)}
					</p>
				</article>
			</div>
		</section>

		<!-- Transaction History -->
		<section class="panel">
			<header class="panel-header">
				<h2>Historial de Transacciones</h2>
				<div class="header-actions">
					<span>
						{txnMeta.total}
						{txnMeta.total === 1 ? 'transacción' : 'transacciones'}
						{#if totalPages > 1}
							· página {currentPage} de {totalPages}
						{/if}
					</span>
					<button class="btn-add" onclick={() => (showAddForm = !showAddForm)}>
						{#if showAddForm}
							Cancelar
						{:else}
							+ Agregar
						{/if}
					</button>
				</div>
			</header>

			<!-- Add transaction inline form -->
			{#if showAddForm}
				<form
					method="POST"
					class="add-txn-form"
					action="?/createTransaction"
					use:enhance={() => {
						isSubmitting = true;
						return async ({ update }) => {
							await update({ reset: false });
							isSubmitting = false;
						};
					}}
				>
					<input type="hidden" name="entryId" value={txnForm.entryId} />
					<input type="hidden" name="currency" value={txnForm.currency} />

					{#if entries.length > 1}
						<div class="form-row">
							<div class="form-group">
								<label class="form-label" for="txn-platform">Plataforma</label>
								<select
									id="txn-platform"
									class="form-select"
									name="entryId"
									bind:value={txnForm.entryId}
								>
									{#each entries as entry (entry.id)}
										<option value={entry.id}>
											{entry.costCurrency} · {new Date(entry.entryDate).toLocaleDateString(
												'es-CO',
												{ year: 'numeric', month: 'short', day: 'numeric' }
											)}
										</option>
									{/each}
								</select>
							</div>
						</div>
					{/if}

					<div class="form-row">
						<div class="form-group">
							<label class="form-label" for="txn-type">Tipo <span class="required">*</span></label>
							<select
								id="txn-type"
								class="form-select"
								name="type"
								bind:value={txnForm.type}
								required
							>
								{#each TRANSACTION_TYPES as t (t.value)}
									<option value={t.value}>{t.label}</option>
								{/each}
							</select>
						</div>
						<div class="form-group">
							<label class="form-label" for="txn-date">Fecha <span class="required">*</span></label>
							<DatePicker name="transactionDate" bind:value={txnForm.transactionDate} required />
						</div>
					</div>

					<!-- trade: cantidad + precio unitario + comisión -->
					{#if txnMode === 'trade'}
						<div class="form-row">
							<div class="form-group">
								<label class="form-label" for="txn-qty"
									>Cantidad <span class="required">*</span></label
								>
								<input
									id="txn-qty"
									type="number"
									class="form-input"
									name="quantity"
									bind:value={txnForm.quantity}
									placeholder="100"
									min="0"
									step="any"
									required
								/>
							</div>
							<div class="form-group">
								<label class="form-label" for="txn-price"
									>Precio unitario <span class="required">*</span></label
								>
								<input
									id="txn-price"
									type="number"
									class="form-input"
									name="price"
									bind:value={txnForm.price}
									placeholder="150.50"
									min="0"
									step="0.01"
									required
								/>
							</div>
							<div class="form-group">
								<label class="form-label" for="txn-fees">Comisión</label>
								<input
									id="txn-fees"
									type="number"
									class="form-input"
									name="fees"
									bind:value={txnForm.fees}
									placeholder="0"
									min="0"
									step="0.01"
								/>
							</div>
						</div>

						<!-- amount: solo monto (dividendo / comisión / interés) -->
					{:else if txnMode === 'amount'}
						<input type="hidden" name="quantity" value="1" />
						<input type="hidden" name="fees" value="0" />
						<div class="form-row">
							<div class="form-group">
								<label class="form-label" for="txn-amount"
									>{priceLabel} <span class="required">*</span></label
								>
								<input
									id="txn-amount"
									type="number"
									class="form-input"
									name="price"
									bind:value={txnForm.price}
									placeholder="0.00"
									min="0"
									step="0.01"
									required
								/>
							</div>
						</div>

						<!-- split: nuevas acciones, sin precio -->
					{:else}
						<input type="hidden" name="price" value="0" />
						<input type="hidden" name="fees" value="0" />
						<div class="form-row">
							<div class="form-group">
								<label class="form-label" for="txn-split-qty"
									>Nuevas acciones recibidas <span class="required">*</span></label
								>
								<input
									id="txn-split-qty"
									type="number"
									class="form-input"
									name="quantity"
									bind:value={txnForm.quantity}
									placeholder="100"
									min="0"
									step="0.00000001"
									required
								/>
							</div>
						</div>
					{/if}

					<div class="form-group">
						<label class="form-label" for="txn-notes">Notas</label>
						<input
							id="txn-notes"
							type="text"
							class="form-input"
							name="notes"
							bind:value={txnForm.notes}
							placeholder="Observaciones opcionales..."
						/>
					</div>

					{#if formError}
						<p class="form-error-msg">No se pudo registrar la transacción. Verifica los datos.</p>
					{/if}

					<div class="form-actions">
						<button type="button" class="btn-cancel" onclick={() => (showAddForm = false)}>
							Cancelar
						</button>
						<button type="submit" class="btn-submit" disabled={isSubmitting}>
							{isSubmitting ? 'Guardando…' : 'Registrar transacción'}
						</button>
					</div>
				</form>
			{/if}

			<!-- Quick-sell panel: appears when a buy lot is selected -->
			{#if sellFromTxn}
				<div class="sell-panel" bind:this={sellPanelEl}>
					<div class="sell-panel-header">
						<div class="sell-panel-info">
							<span class="sell-panel-title">Vender desde compra</span>
							<span class="sell-panel-lot">
								Lote: {parseFloat(sellFromTxn.quantity).toLocaleString('es-CO', {
									maximumFractionDigits: 8
								})}
								unidades @ {fmt(parseFloat(sellFromTxn.price))} ·
								{fmtDate(sellFromTxn.transactionDate)}
							</span>
						</div>
						<button class="sell-panel-close" type="button" onclick={() => (sellFromTxn = null)}
							>✕</button
						>
					</div>

					<div class="sell-mode-toggle">
						<button
							type="button"
							class="sell-mode-btn"
							class:active={sellMode === 'full'}
							onclick={() => (sellMode = 'full')}
						>
							Venta Completa
						</button>
						<button
							type="button"
							class="sell-mode-btn"
							class:active={sellMode === 'partial'}
							onclick={() => (sellMode = 'partial')}
						>
							Venta Parcial
						</button>
					</div>

					<form
						method="POST"
						class="sell-form"
						use:enhance={() => {
							isSellSubmitting = true;
							return async ({ update }) => {
								await update({ reset: false });
								isSellSubmitting = false;
							};
						}}
					>
						<input type="hidden" name="entryId" value={sellFromTxn.entryId} />
						<input type="hidden" name="type" value="sell" />
						<input
							type="hidden"
							name="currency"
							value={entries.find((e) => e.id === sellFromTxn?.entryId)?.costCurrency ??
								txnForm.currency}
						/>

						<div class="form-row">
							<div class="form-group">
								<label class="form-label" for="sell-qty">
									Cantidad <span class="required">*</span>
									{#if sellMode === 'full'}
										<span class="sell-label-hint">(lote completo)</span>
									{/if}
								</label>
								<input
									id="sell-qty"
									type="number"
									class="form-input"
									name="quantity"
									bind:value={sellQty}
									disabled={sellMode === 'full'}
									min="0.00000001"
									max={parseFloat(sellFromTxn.quantity)}
									step="0.00000001"
									required
								/>
							</div>
							<div class="form-group">
								<label class="form-label" for="sell-price"
									>Precio unitario <span class="required">*</span></label
								>
								<input
									id="sell-price"
									type="number"
									class="form-input"
									name="price"
									bind:value={sellPrice}
									min="0"
									step="0.01"
									required
								/>
							</div>
							<div class="form-group">
								<label class="form-label" for="sell-fees">Comisión</label>
								<input
									id="sell-fees"
									type="number"
									class="form-input"
									name="fees"
									bind:value={sellFees}
									placeholder="0"
									min="0"
									step="0.01"
								/>
							</div>
							<div class="form-group">
								<label class="form-label" for="sell-date"
									>Fecha <span class="required">*</span></label
								>
								<DatePicker name="transactionDate" bind:value={sellDate} required />
							</div>
						</div>

						<div class="form-group">
							<label class="form-label" for="sell-notes">Notas</label>
							<input
								id="sell-notes"
								type="text"
								class="form-input"
								name="notes"
								bind:value={sellNotes}
								placeholder="Observaciones opcionales..."
							/>
						</div>

						{#if formError && !showAddForm}
							<p class="form-error-msg">No se pudo registrar la venta. Verifica los datos.</p>
						{/if}

						<div class="form-actions">
							<button type="button" class="btn-cancel" onclick={() => (sellFromTxn = null)}>
								Cancelar
							</button>
							<button type="submit" class="btn-sell-submit" disabled={isSellSubmitting}>
								{isSellSubmitting
									? 'Guardando…'
									: sellMode === 'full'
										? 'Confirmar Venta Total'
										: 'Registrar Venta Parcial'}
							</button>
						</div>
					</form>
				</div>
			{/if}

			{#if transactions.length === 0}
				<p class="empty-txn">No hay transacciones registradas aún.</p>
			{:else}
				<div class="transactions-table">
					<div class="table-header">
						<p>Tipo</p>
						<p>Fecha</p>
						<p>Cantidad</p>
						<p>Precio</p>
						<p>Comisión</p>
						<p>Total</p>
						<p>Notas</p>
						<p>Acciones</p>
					</div>

					{#each transactions as txn (txn.id)}
						{@const qty = parseFloat(txn.quantity) || 0}
						{@const price = parseFloat(txn.price) || 0}
						{@const fees = parseFloat(txn.fees) || 0}
						{@const total = qty * price}
						{@const isBuyLot = txn.type === 'buy' || txn.type === 'transfer_in'}
						{@const isActiveSell = sellFromTxn?.id === txn.id}
						<div class="table-row" class:row-selling={isActiveSell}>
							<p>
								<span class="type-badge {TYPE_STYLE[txn.type] ?? ''}">
									{TYPE_LABEL[txn.type] ?? txn.type}
								</span>
							</p>
							<p class="date">{fmtDate(txn.transactionDate)}</p>
							<p class="qty">{qty.toLocaleString('es-CO', { maximumFractionDigits: 8 })}</p>
							<p class="price">{fmt(price)}</p>
							<p class="fees">{fees > 0 ? fmt(fees) : '—'}</p>
							<p class="total">{fmt(total, 0)}</p>
							<p class="notes">{txn.notes || '—'}</p>
							<p class="cell-action">
								<button
									type="button"
									class="btn-edit-row"
									onclick={() => startEdit(txn)}
									aria-label="Editar transacción"
								>
									<svg
										width="13"
										height="13"
										viewBox="0 0 24 24"
										fill="none"
										stroke="currentColor"
										stroke-width="2.5"
									>
										<path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" />
										<path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
									</svg>
								</button>
								{#if isBuyLot}
									<button
										type="button"
										class="btn-sell-row"
										class:active={isActiveSell}
										onclick={() => (sellFromTxn = isActiveSell ? null : txn)}
									>
										{isActiveSell ? 'Cancelar' : 'Vender'}
									</button>
								{/if}
							</p>
						</div>
					{/each}
				</div>

				{#if totalPages > 1}
					<form class="pagination" method="GET">
						<input type="hidden" name="limit" value={txnMeta.limit} />
						<button
							type="submit"
							name="page"
							value={currentPage - 1}
							class="pg-btn"
							disabled={currentPage === 1}
						>
							‹ Anterior
						</button>
						{#each Array.from({ length: totalPages }, (_, i) => i + 1) as p (p)}
							<button
								type="submit"
								name="page"
								value={p}
								class="pg-btn pg-num"
								class:pg-active={p === currentPage}>{p}</button
							>
						{/each}
						<button
							type="submit"
							name="page"
							value={currentPage + 1}
							class="pg-btn"
							disabled={currentPage === totalPages}
						>
							Siguiente ›
						</button>
					</form>
				{/if}
			{/if}
		</section>
	{/if}
</div>

<!-- Edit transaction modal -->
{#if editingTxn}
	<div
		class="modal-backdrop"
		role="button"
		tabindex="0"
		aria-label="Cerrar modal"
		onclick={() => (editingTxn = null)}
		onkeydown={(e) => e.key === 'Enter' && (editingTxn = null)}
	></div>
	<div class="modal" role="dialog" aria-modal="true" aria-label="Editar transacción">
		<header class="modal-header">
			<span>Editar transacción</span>
			<button class="modal-close" type="button" onclick={() => (editingTxn = null)}>✕</button>
		</header>

		<form
			method="POST"
			action="?/editTransaction"
			class="modal-form"
			use:enhance={() => {
				isEditSubmitting = true;
				editError = false;
				editErrorMessage = '';
				return async ({ result, update }) => {
					await update({ reset: false });
					isEditSubmitting = false;
					const data =
						result.type === 'success'
							? (result.data as { success?: boolean; error?: string } | undefined)
							: undefined;
					if (data?.success) {
						editingTxn = null;
						await invalidateAll();
					} else {
						editError = true;
						editErrorMessage = data?.error ?? '';
					}
				};
			}}
		>
			<input type="hidden" name="txnId" value={editingTxn.id} />
			<input type="hidden" name="currency" value={editForm.currency} />

			<div class="form-row">
				<div class="form-group">
					<label class="form-label" for="edit-type">Tipo <span class="required">*</span></label>
					<select
						id="edit-type"
						class="form-select"
						name="type"
						bind:value={editForm.type}
						required
					>
						{#each TRANSACTION_TYPES as t (t.value)}
							<option value={t.value}>{t.label}</option>
						{/each}
					</select>
				</div>
				<div class="form-group">
					<span class="form-label">Fecha <span class="required">*</span></span>
					<DatePicker name="transactionDate" bind:value={editForm.transactionDate} required />
				</div>
			</div>

			{#if editTxnMode === 'trade'}
				<div class="form-row">
					<div class="form-group">
						<label class="form-label" for="edit-qty">Cantidad <span class="required">*</span></label
						>
						<input
							id="edit-qty"
							type="number"
							class="form-input"
							name="quantity"
							bind:value={editForm.quantity}
							placeholder="100"
							min="0"
							step="any"
							required
						/>
					</div>
					<div class="form-group">
						<label class="form-label" for="edit-price"
							>{editPriceLabel} <span class="required">*</span></label
						>
						<input
							id="edit-price"
							type="number"
							class="form-input"
							name="price"
							bind:value={editForm.price}
							placeholder="150.50"
							min="0"
							step="0.01"
							required
						/>
					</div>
					<div class="form-group">
						<label class="form-label" for="edit-fees">Comisión</label>
						<input
							id="edit-fees"
							type="number"
							class="form-input"
							name="fees"
							bind:value={editForm.fees}
							placeholder="0"
							min="0"
							step="0.01"
						/>
					</div>
				</div>
			{:else if editTxnMode === 'amount'}
				<input type="hidden" name="quantity" value="1" />
				<input type="hidden" name="fees" value="0" />
				<div class="form-row">
					<div class="form-group">
						<label class="form-label" for="edit-amount"
							>{editPriceLabel} <span class="required">*</span></label
						>
						<input
							id="edit-amount"
							type="number"
							class="form-input"
							name="price"
							bind:value={editForm.price}
							placeholder="0.00"
							min="0"
							step="0.01"
							required
						/>
					</div>
				</div>
			{:else}
				<input type="hidden" name="price" value="0" />
				<input type="hidden" name="fees" value="0" />
				<div class="form-row">
					<div class="form-group">
						<label class="form-label" for="edit-split-qty"
							>Nuevas acciones recibidas <span class="required">*</span></label
						>
						<input
							id="edit-split-qty"
							type="number"
							class="form-input"
							name="quantity"
							bind:value={editForm.quantity}
							placeholder="100"
							min="0"
							step="0.00000001"
							required
						/>
					</div>
				</div>
			{/if}

			<div class="form-group">
				<label class="form-label" for="edit-notes">Notas</label>
				<input
					id="edit-notes"
					type="text"
					class="form-input"
					name="notes"
					bind:value={editForm.notes}
					placeholder="Observaciones opcionales..."
				/>
			</div>

			{#if editError}
				<p class="form-error-msg">
					{editErrorMessage || 'No se pudo actualizar la transacción. Verifica los datos.'}
				</p>
			{/if}

			<div class="form-actions">
				<button type="button" class="btn-cancel" onclick={() => (editingTxn = null)}
					>Cancelar</button
				>
				<button type="submit" class="btn-submit" disabled={isEditSubmitting}>
					{isEditSubmitting ? 'Guardando…' : 'Guardar cambios'}
				</button>
			</div>
		</form>
	</div>
{/if}

<style>
	.container {
		max-width: 1400px;
		margin: 0 auto;
	}

	.btn-back {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.65rem 1rem;
		margin-bottom: 1.5rem;
		background: transparent;
		border: 1.5px solid rgba(212, 145, 42, 0.25);
		border-radius: 8px;
		color: var(--amber);
		font-weight: 600;
		font-size: 0.9rem;
		cursor: pointer;
		transition: all 0.3s ease;
		font-family: var(--font-body);
	}

	.btn-back:hover {
		background: var(--border);
		border-color: var(--amber);
	}

	.empty-state {
		padding: 3rem;
		text-align: center;
		color: rgba(236, 234, 229, 0.5);
		border: 1px dashed var(--border-strong);
		border-radius: 12px;
	}

	/* Header */
	.header-section {
		margin-bottom: 1.5rem;
		padding-bottom: 1.5rem;
		border-bottom: 1px solid var(--border-strong);
	}

	.header-content {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 2rem;
		flex-wrap: wrap;
	}

	.symbol-info h1 {
		margin: 0 0 0.25rem;
		font-size: 2.25rem;
		font-weight: 700;
		color: var(--amber);
		letter-spacing: -0.5px;
		font-family: var(--font-display);
	}

	.asset-name {
		margin: 0 0 0.5rem;
		color: rgba(236, 234, 229, 0.7);
		font-size: 1rem;
	}

	.asset-badge {
		display: inline-block;
		padding: 0.25rem 0.75rem;
		border-radius: 20px;
		background: rgba(212, 145, 42, 0.1);
		border: 1px solid rgba(212, 145, 42, 0.25);
		color: var(--amber);
		font-size: 0.8rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.price-display {
		text-align: right;
	}

	.current-price {
		margin: 0 0 0.25rem;
		font-size: 2rem;
		font-weight: 700;
		color: var(--text);
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
	}

	.price-label {
		margin: 0;
		font-size: 0.8rem;
		color: rgba(236, 234, 229, 0.4);
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	/* Panels */
	.panel {
		background: var(--surface);
		border: 1px solid var(--border-strong);
		border-radius: 16px;
		padding: 1.75rem;
		margin-bottom: 1.5rem;
		box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
	}

	.panel-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
		margin-bottom: 1.5rem;
		padding-bottom: 1rem;
		border-bottom: 1px solid var(--border);
	}

	.panel-header h2 {
		margin: 0;
		font-size: 1.1rem;
		font-weight: 400;
		color: var(--text);
		font-family: var(--font-display);
	}

	.panel-header span {
		color: rgba(236, 234, 229, 0.4);
		font-size: 0.85rem;
	}

	/* Metrics grid */
	.metrics-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
		gap: 1rem;
	}

	.metric-card {
		background: rgba(212, 145, 42, 0.06);
		border: 1px solid rgba(212, 145, 42, 0.2);
		border-radius: 10px;
		padding: 1.25rem;
		text-align: center;
	}

	.metric-label {
		margin: 0 0 0.6rem;
		font-size: 0.75rem;
		color: rgba(236, 234, 229, 0.5);
		text-transform: uppercase;
		letter-spacing: 0.5px;
		font-weight: 600;
	}

	.metric-value {
		margin: 0;
		font-size: 1.25rem;
		font-weight: 700;
		color: var(--amber);
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
	}

	.metric-value.positive {
		color: var(--green);
	}
	.metric-value.negative {
		color: var(--red);
	}

	.metric-unit {
		font-size: 0.7rem;
		color: rgba(236, 234, 229, 0.4);
		margin-left: 0.25rem;
	}

	.metric-pct {
		margin: 0.25rem 0 0;
		font-size: 0.85rem;
		font-weight: 600;
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
	}

	.metric-pct.positive {
		color: var(--green);
	}
	.metric-pct.negative {
		color: var(--red);
	}

	/* Performance grid */
	.perf-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
		gap: 1rem;
	}

	.perf-card {
		background: rgba(255, 255, 255, 0.022);
		border: 1px solid var(--border);
		border-radius: 10px;
		padding: 1.25rem;
		text-align: center;
	}

	.perf-card h3 {
		margin: 0 0 0.6rem;
		font-size: 0.75rem;
		color: rgba(236, 234, 229, 0.5);
		text-transform: uppercase;
		letter-spacing: 0.5px;
		font-weight: 600;
	}

	.perf-value {
		margin: 0;
		font-size: 1.1rem;
		font-weight: 700;
		color: var(--amber);
	}

	.perf-value.positive {
		color: var(--green);
	}
	.perf-value.negative {
		color: var(--red);
	}

	.bar-wrap {
		width: 100%;
		height: 4px;
		background: var(--border);
		border-radius: 2px;
		margin-top: 0.6rem;
		overflow: hidden;
	}

	.bar-fill {
		height: 100%;
		background: linear-gradient(90deg, var(--amber) 0%, var(--amber-light, var(--amber)) 100%);
		border-radius: 2px;
	}

	/* Panel header actions */
	.header-actions {
		display: flex;
		align-items: center;
		gap: 1rem;
	}

	.btn-add {
		padding: 0.4rem 0.9rem;
		border: 1.5px solid rgba(212, 145, 42, 0.4);
		border-radius: 6px;
		background: transparent;
		color: var(--amber);
		font-size: 0.82rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s ease;
		font-family: var(--font-body);
	}

	.btn-add:hover {
		background: rgba(212, 145, 42, 0.12);
		border-color: var(--amber);
	}

	/* Add transaction form */
	.add-txn-form {
		margin-bottom: 1.5rem;
		padding: 1.25rem;
		border: 1px solid rgba(212, 145, 42, 0.2);
		border-radius: 10px;
		background: rgba(212, 145, 42, 0.04);
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.form-row {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
		gap: 1rem;
	}

	.form-group {
		display: flex;
		flex-direction: column;
		gap: 0.4rem;
	}

	.form-label {
		font-size: 0.8rem;
		font-weight: 600;
		color: rgba(236, 234, 229, 0.6);
		text-transform: uppercase;
		letter-spacing: 0.3px;
	}

	.required {
		color: var(--red);
	}

	.form-input {
		padding: 0.6rem 0.85rem;
		border: 1.5px solid rgba(212, 145, 42, 0.2);
		border-radius: 8px;
		background: rgba(255, 255, 255, 0.04);
		color: var(--text);
		font-size: 0.9rem;
		font-family: var(--font-body);
		transition: border-color 0.2s ease;
	}

	.form-input:focus {
		outline: none;
		border-color: var(--amber);
	}

	.form-select {
		padding: 0.6rem 2.2rem 0.6rem 0.85rem;
		border: 1.5px solid rgba(212, 145, 42, 0.35);
		border-radius: 8px;
		background-color: rgba(212, 145, 42, 0.06);
		background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='16' height='16' viewBox='0 0 24 24' fill='none' stroke='%23d4912a' stroke-width='2.5' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='M6 9l6 6 6-6'/%3E%3C/svg%3E");
		background-repeat: no-repeat;
		background-position: right 0.65rem center;
		background-size: 1rem;
		appearance: none;
		-webkit-appearance: none;
		color: var(--text);
		font-size: 0.9rem;
		font-family: var(--font-body);
		cursor: pointer;
		transition:
			border-color 0.2s ease,
			background-color 0.2s ease;
	}

	.form-select:focus {
		outline: none;
		border-color: var(--amber);
		background-color: rgba(212, 145, 42, 0.1);
	}

	.form-error-msg {
		margin: 0;
		padding: 0.6rem 0.9rem;
		border-radius: 6px;
		background: rgba(224, 90, 90, 0.1);
		border: 1px solid rgba(224, 90, 90, 0.3);
		color: rgba(224, 90, 90, 0.9);
		font-size: 0.85rem;
	}

	.form-actions {
		display: flex;
		gap: 0.75rem;
		justify-content: flex-end;
	}

	.btn-cancel {
		padding: 0.55rem 1.1rem;
		border: 1.5px solid rgba(212, 145, 42, 0.2);
		border-radius: 7px;
		background: transparent;
		color: rgba(236, 234, 229, 0.6);
		font-size: 0.88rem;
		font-weight: 600;
		cursor: pointer;
		font-family: var(--font-body);
		transition: all 0.2s ease;
	}

	.btn-cancel:hover {
		border-color: rgba(212, 145, 42, 0.4);
		color: var(--text);
	}

	.btn-submit {
		padding: 0.55rem 1.25rem;
		border: none;
		border-radius: 7px;
		background: var(--amber);
		color: #0d0800;
		font-size: 0.88rem;
		font-weight: 700;
		cursor: pointer;
		font-family: var(--font-body);
		transition: all 0.2s ease;
	}

	.btn-submit:hover:not(:disabled) {
		transform: translateY(-1px);
		box-shadow: 0 6px 16px rgba(212, 145, 42, 0.25);
	}

	.btn-submit:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.empty-txn {
		margin: 0;
		padding: 1.5rem;
		text-align: center;
		color: rgba(236, 234, 229, 0.4);
		font-size: 0.9rem;
	}

	/* Transactions table */
	.transactions-table {
		overflow-x: auto;
	}

	.table-header {
		display: grid;
		grid-template-columns: 110px 100px 1fr 1fr 1fr 1fr 1fr 120px;
		gap: 1rem;
		padding: 0.75rem 1rem;
		background: rgba(0, 0, 0, 0.2);
		border-radius: 8px 8px 0 0;
		border-bottom: 1px solid var(--border-strong);
		font-weight: 600;
		font-size: 0.8rem;
		color: rgba(236, 234, 229, 0.5);
		text-transform: uppercase;
		letter-spacing: 0.3px;
	}

	.table-row {
		display: grid;
		grid-template-columns: 110px 100px 1fr 1fr 1fr 1fr 1fr 120px;
		gap: 1rem;
		padding: 1rem;
		border-bottom: 1px solid var(--border);
		align-items: center;
		transition: background 0.2s ease;
	}

	.row-selling {
		background: rgba(224, 90, 90, 0.05) !important;
		border-left: 2px solid rgba(224, 90, 90, 0.4);
	}

	.cell-action {
		display: flex;
		justify-content: flex-end;
	}

	.btn-sell-row {
		padding: 0.3rem 0.65rem;
		border: 1.5px solid rgba(224, 90, 90, 0.4);
		border-radius: 6px;
		background: transparent;
		color: var(--red);
		font-size: 0.78rem;
		font-weight: 600;
		cursor: pointer;
		font-family: var(--font-body);
		transition: all 0.2s ease;
		white-space: nowrap;
	}

	.btn-sell-row:hover,
	.btn-sell-row.active {
		background: rgba(224, 90, 90, 0.12);
		border-color: var(--red);
	}

	/* Sell panel */
	.sell-panel {
		margin-bottom: 1.5rem;
		padding: 1.25rem;
		border: 1px solid rgba(224, 90, 90, 0.3);
		border-radius: 10px;
		background: rgba(224, 90, 90, 0.05);
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.sell-panel-header {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 1rem;
	}

	.sell-panel-info {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.sell-panel-title {
		font-size: 0.9rem;
		font-weight: 700;
		color: var(--red);
		text-transform: uppercase;
		letter-spacing: 0.3px;
	}

	.sell-panel-lot {
		font-size: 0.82rem;
		color: rgba(236, 234, 229, 0.6);
		font-family: var(--font-mono);
	}

	.sell-panel-close {
		padding: 0.2rem 0.5rem;
		border: none;
		background: transparent;
		color: rgba(236, 234, 229, 0.4);
		font-size: 1rem;
		cursor: pointer;
		border-radius: 4px;
		transition: color 0.2s ease;
		flex-shrink: 0;
	}

	.sell-panel-close:hover {
		color: var(--text);
	}

	.sell-mode-toggle {
		display: flex;
		gap: 0.5rem;
	}

	.sell-mode-btn {
		padding: 0.45rem 1rem;
		border: 1.5px solid rgba(224, 90, 90, 0.3);
		border-radius: 6px;
		background: transparent;
		color: rgba(236, 234, 229, 0.6);
		font-size: 0.85rem;
		font-weight: 600;
		cursor: pointer;
		font-family: var(--font-body);
		transition: all 0.2s ease;
	}

	.sell-mode-btn:hover {
		border-color: rgba(224, 90, 90, 0.5);
		color: var(--text);
	}

	.sell-mode-btn.active {
		background: rgba(224, 90, 90, 0.15);
		border-color: var(--red);
		color: var(--red);
	}

	.sell-form {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.sell-label-hint {
		font-weight: 400;
		color: rgba(236, 234, 229, 0.4);
		text-transform: none;
		letter-spacing: 0;
		font-size: 0.75rem;
	}

	.btn-sell-submit {
		padding: 0.55rem 1.25rem;
		border: none;
		border-radius: 7px;
		background: var(--red);
		color: #fff;
		font-size: 0.88rem;
		font-weight: 700;
		cursor: pointer;
		font-family: var(--font-body);
		transition: all 0.2s ease;
	}

	.btn-sell-submit:hover:not(:disabled) {
		transform: translateY(-1px);
		box-shadow: 0 6px 16px rgba(224, 90, 90, 0.25);
	}

	.btn-sell-submit:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.table-row:hover {
		background: rgba(212, 145, 42, 0.03);
	}

	.table-row:last-child {
		border-bottom: none;
	}

	.table-row p {
		margin: 0;
		font-size: 0.9rem;
	}

	/* Transaction type badges */
	.type-badge {
		display: inline-block;
		padding: 0.2rem 0.6rem;
		border-radius: 20px;
		font-size: 0.75rem;
		font-weight: 700;
		letter-spacing: 0.2px;
	}

	.type-buy {
		background: rgba(80, 200, 120, 0.15);
		color: var(--green);
		border: 1px solid rgba(80, 200, 120, 0.3);
	}

	.type-sell {
		background: rgba(224, 90, 90, 0.15);
		color: var(--red);
		border: 1px solid rgba(224, 90, 90, 0.3);
	}

	.type-dividend,
	.type-interest {
		background: rgba(212, 145, 42, 0.15);
		color: var(--amber);
		border: 1px solid rgba(212, 145, 42, 0.3);
	}

	.type-fee {
		background: rgba(150, 150, 150, 0.15);
		color: rgba(236, 234, 229, 0.5);
		border: 1px solid rgba(150, 150, 150, 0.2);
	}

	.type-transfer,
	.type-split {
		background: rgba(100, 160, 230, 0.15);
		color: #7ab4f0;
		border: 1px solid rgba(100, 160, 230, 0.3);
	}

	.date {
		color: rgba(236, 234, 229, 0.5);
	}

	.qty {
		color: var(--text);
		font-weight: 500;
	}

	.price {
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
		color: var(--amber);
		font-weight: 500;
	}

	.fees {
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
		color: rgba(236, 234, 229, 0.4);
		font-size: 0.85rem;
	}

	.total {
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
		color: var(--text);
		font-weight: 600;
	}

	.notes {
		color: rgba(236, 234, 229, 0.4);
		font-size: 0.85rem;
		font-style: italic;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	@media (max-width: 768px) {
		.header-content {
			flex-direction: column;
			align-items: flex-start;
		}

		.price-display {
			text-align: left;
		}

		.metrics-grid {
			grid-template-columns: repeat(2, 1fr);
		}

		.perf-grid {
			grid-template-columns: repeat(2, 1fr);
		}

		.table-header,
		.table-row {
			grid-template-columns: 90px 90px 1fr 1fr 1fr 70px;
		}

		.table-header p:nth-child(5),
		.table-header p:nth-child(6),
		.table-row .fees,
		.table-row .notes {
			display: none;
		}
	}

	@media (max-width: 480px) {
		.metrics-grid {
			grid-template-columns: 1fr 1fr;
		}

		.perf-grid {
			grid-template-columns: 1fr 1fr;
		}

		.table-header {
			display: none;
		}

		.table-row {
			grid-template-columns: 1fr 1fr;
			gap: 0.5rem;
			background: rgba(255, 255, 255, 0.022);
			border: 1px solid var(--border-strong);
			border-radius: 8px;
			margin-bottom: 0.75rem;
		}

		.table-row:last-child {
			border-bottom: 1px solid var(--border-strong);
		}

		.fees,
		.notes {
			display: none;
		}
	}

	.btn-edit-row {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		padding: 0.28rem 0.5rem;
		border: 1.5px solid rgba(212, 145, 42, 0.3);
		border-radius: 6px;
		background: transparent;
		color: rgba(212, 145, 42, 0.6);
		cursor: pointer;
		transition: all 0.2s ease;
		flex-shrink: 0;
	}

	.btn-edit-row:hover {
		background: rgba(212, 145, 42, 0.1);
		border-color: var(--amber);
		color: var(--amber);
	}

	/* Modal */
	.modal-backdrop {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.65);
		z-index: 100;
	}

	.modal {
		position: fixed;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
		z-index: 101;
		width: min(540px, 92vw);
		background: var(--surface);
		border: 1.5px solid rgba(212, 145, 42, 0.35);
		border-radius: 16px;
		box-shadow: 0 24px 64px rgba(0, 0, 0, 0.5);
		backdrop-filter: blur(16px);
		overflow: hidden;
	}

	.modal-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 1rem 1.25rem;
		border-bottom: 1px solid var(--border);
		font-size: 0.95rem;
		font-weight: 700;
		color: var(--amber);
	}

	.modal-close {
		background: transparent;
		border: none;
		color: rgba(236, 234, 229, 0.4);
		font-size: 1rem;
		cursor: pointer;
		padding: 0.2rem 0.4rem;
		border-radius: 4px;
		transition: color 0.2s ease;
		line-height: 1;
	}

	.modal-close:hover {
		color: var(--text);
	}

	.modal-form {
		padding: 1.25rem;
		display: flex;
		flex-direction: column;
		gap: 1rem;
		max-height: 80vh;
		overflow-y: auto;
	}

	/* Pagination */
	.pagination {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.35rem;
		padding: 1rem 0 0.25rem;
		flex-wrap: wrap;
	}

	.pg-btn {
		padding: 0.35rem 0.75rem;
		border: 1.5px solid rgba(212, 145, 42, 0.25);
		border-radius: 6px;
		background: transparent;
		color: rgba(236, 234, 229, 0.6);
		font-size: 0.82rem;
		font-weight: 600;
		cursor: pointer;
		font-family: var(--font-body);
		transition: all 0.2s ease;
	}

	.pg-btn:hover:not(:disabled) {
		border-color: var(--amber);
		color: var(--amber);
		background: rgba(212, 145, 42, 0.08);
	}

	.pg-btn:disabled {
		opacity: 0.3;
		cursor: default;
	}

	.pg-num {
		min-width: 2rem;
	}

	.pg-active {
		background: rgba(212, 145, 42, 0.15);
		border-color: var(--amber);
		color: var(--amber);
	}
</style>
