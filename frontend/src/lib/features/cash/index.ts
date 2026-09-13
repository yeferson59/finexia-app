/**
 * Feature `cash` — superficie pública.
 *
 * El efectivo de la cuenta (`routes/dashboard/cash`): el total y su reparto por
 * moneda, los saldos de cada plataforma, los movimientos y el formulario para
 * registrarlos o editarlos. `cash.ts` aporta los helpers puros y reexporta los
 * contratos `CashBalance` y `CashMovement` de `$lib/api/types`.
 *
 * `cash-delete-confirm` es interno de `cash-movements-table` (import relativo)
 * y no forma parte de la superficie pública.
 */
export { default as CashSummary } from './components/cash-summary.svelte';
export { default as CashBalancesTable } from './components/cash-balances-table.svelte';
export { default as CashMovementsTable } from './components/cash-movements-table.svelte';
export {
	default as CashMovementForm,
	type CashFormTarget
} from './components/cash-movement-form.svelte';

export * from './cash';
export * from './schemas';
