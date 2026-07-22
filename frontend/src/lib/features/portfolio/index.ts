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

export * from './portfolio';
