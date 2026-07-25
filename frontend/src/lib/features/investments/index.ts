/**
 * Feature `investments` — superficie pública.
 *
 * Catálogo de productos de inversión (`routes/dashboard/investments/**`): el
 * detalle de un producto y el formulario de alta. `state/investments.svelte.ts`
 * mantiene el store compartido por el listado, el detalle y el alta;
 * `investments.ts` aporta el catálogo mock, los helpers y los tipos.
 */
export { default as InvestmentDetail } from './components/investment-detail.svelte';
export { default as InvestmentAddForm } from './components/investment-add-form.svelte';

export { investmentStore } from './state/investments.svelte';
export type { Investment, NewInvestment } from './state/investments.svelte';

export * from './investments';
