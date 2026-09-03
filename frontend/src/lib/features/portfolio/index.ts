/**
 * Feature `portfolio` — superficie pública.
 *
 * Componentes del detalle de portafolio (`routes/dashboard/portfolios/[id]`),
 * del alta (`portfolios/add`) y de la vista consolidada de activos
 * (`dashboard/assets`). `portfolio.ts` aporta los helpers puros (agrupar
 * holdings, distribución por tipo, segmentos del donut) y `asset-holdings.ts`
 * los de la vista consolidada.
 */
export { default as PortfolioEditForm } from './components/portfolio-edit-form.svelte';
export { default as PortfolioSummaryCards } from './components/portfolio-summary-cards.svelte';
export { default as PortfolioStatsCards } from './components/portfolio-stats-cards.svelte';
export { default as AllocationDonut } from './components/allocation-donut.svelte';
export { default as HoldingsTable } from './components/holdings-table.svelte';
export { default as PortfolioCard } from './components/portfolio-card.svelte';
export { default as PortfolioDetailHeader } from './components/portfolio-detail-header.svelte';
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
export { default as AssetDeletePosition } from './components/asset-delete-position.svelte';

// Vista consolidada de activos (`dashboard/assets`): lo que el usuario tiene
// de cada activo sumando todos sus portafolios.
export { default as AssetConcentration } from './components/asset-concentration.svelte';
export { default as AssetHoldingsTable } from './components/asset-holdings-table.svelte';

export * from './portfolio';
export * from './asset-holdings';
export * from './asset';
export * from './schemas';
