/**
 * Feature `portfolio` — superficie pública.
 *
 * Componentes del detalle de portafolio (`routes/dashboard/portfolios/[id]`) y
 * del alta (`portfolios/add`). `portfolio.ts` aporta los helpers puros
 * (agrupar holdings, distribución por tipo, segmentos del donut) y los tipos.
 */
export { default as PortfolioEditForm } from './components/portfolio-edit-form.svelte';
export { default as PortfolioSummaryCards } from './components/portfolio-summary-cards.svelte';
export { default as PortfolioStatsCards } from './components/portfolio-stats-cards.svelte';
export { default as AllocationDonut } from './components/allocation-donut.svelte';
export { default as HoldingsTable } from './components/holdings-table.svelte';
export { default as PortfolioAddForm } from './components/portfolio-add-form.svelte';
export { default as PortfolioEntryForm } from './components/portfolio-entry-form.svelte';

// Detalle de un activo dentro del portafolio
// (`portfolios/[id]/assets/[symbol]`). El formulario de alta, el panel de venta
// rápida, la tabla y el modal de edición son internos de
// `asset-transaction-history`; `asset-combobox` lo es de `portfolio-entry-form`.
export { default as AssetPositionHeader } from './components/asset-position-header.svelte';
export { default as AssetPositionSummary } from './components/asset-position-summary.svelte';
export { default as AssetInfoPanel } from './components/asset-info-panel.svelte';
export { default as AssetTransactionHistory } from './components/asset-transaction-history.svelte';

export * from './portfolio';
export * from './asset';
