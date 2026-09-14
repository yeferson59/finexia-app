/**
 * Feature `cash` — superficie pública.
 *
 * El efectivo de la cuenta (`routes/dashboard/cash`): el total y su reparto por
 * moneda, las cuentas de cada plataforma, los movimientos por meses y el
 * formulario para registrarlos o editarlos. `cash.ts` aporta los helpers puros
 * y reexporta los contratos `CashBalance` y `CashMovement` de `$lib/api/types`.
 *
 * `cash-delete-confirm` es interno de `cash-movements` (import relativo) y no
 * forma parte de la superficie pública.
 */
export { default as CashSummary } from './components/cash-summary.svelte';
export { default as CashAccounts } from './components/cash-accounts.svelte';
export { default as CashMovements } from './components/cash-movements.svelte';
export {
	default as CashMovementForm,
	type CashFormTarget
} from './components/cash-movement-form.svelte';

export * from './cash';
export * from './schemas';
