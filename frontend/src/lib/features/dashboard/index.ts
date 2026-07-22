/**
 * Feature `dashboard` — superficie pública.
 *
 * Componentes del área de inversión (`routes/dashboard/**`): el shell
 * (header, sidebar) y los widgets del dashboard principal, más `PortfolioGrowth`
 * que reutiliza el detalle de portafolio. `CurrencyToggle` es interno de
 * `net-worth-card` y no se exporta.
 *
 * `state/investments.svelte.ts` mantiene el store de productos de inversión
 * compartido por las páginas de `routes/dashboard/investments/**`.
 */
export { default as DashboardHeader } from './components/header.svelte';
export { default as Sidebar } from './components/sidebar.svelte';
export { default as NetWorthCard } from './components/net-worth-card.svelte';
export { default as PortfolioOverview } from './components/portfolio-overview.svelte';
export { default as AssetAllocation } from './components/asset-allocation.svelte';
export { default as RecentActivity } from './components/recent-activity.svelte';
export { default as PortfolioGrowth } from './components/portfolio-growth.svelte';

export { investmentStore } from './state/investments.svelte';
export type { Investment, NewInvestment } from './state/investments.svelte';
