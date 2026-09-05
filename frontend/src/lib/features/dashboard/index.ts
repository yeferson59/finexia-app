/**
 * Feature `dashboard` — superficie pública.
 *
 * Componentes del área de inversión (`routes/dashboard/**`): el chrome (header,
 * sidebar) y los bloques de la portada del panel, más `PortfolioGrowth`, que
 * reutiliza el detalle de portafolio. Lo interno de un bloque —la curva en
 * miniatura, el detalle de la gráfica— no sale de aquí; el selector de moneda
 * dejó de ser interno y vive en `lib/ui`, porque la vista de activos lo comparte.
 */
export { default as DashboardHeader } from './components/header.svelte';
export { default as Sidebar } from './components/sidebar.svelte';
export { default as WealthHeadline } from './components/wealth-headline.svelte';
export { default as Breakdown } from './components/breakdown.svelte';
export { default as RecentActivity } from './components/recent-activity.svelte';
export { default as PortfolioGrowth } from './components/portfolio-growth.svelte';
export { default as MarketKeyNotice } from './components/market-key-notice.svelte';
