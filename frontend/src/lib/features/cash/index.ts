/**
 * Feature `cash` — superficie pública.
 *
 * El efectivo de la cuenta (`routes/dashboard/cash`): el total y su reparto por
 * moneda, las cuentas de cada plataforma, los movimientos por meses y el
 * formulario para registrarlos o editarlos. `cash.ts` aporta los helpers puros
 * y reexporta los contratos `CashBalance` y `CashMovement` de `$lib/api/types`.
 *
 * Los bolsillos (000047) son subcuentas de una cuenta: su dinero sigue contando
 * en la plataforma y lo propio de cada uno es su tasa. `cash-pocket-form` los
 * abre y renombra, y `cash-move-form` mueve dinero entre los cajones de una
 * cuenta en una sola transacción.
 *
 * Un depósito a tasa fija (000048) es un bolsillo que conserva la tasa del día
 * en que se abrió: `cash-deposit-form` lo abre —dinero, plazo y tasa a la vez—
 * y luego lo cancela o lo borra, porque no hay nada más que hacerle.
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
export { default as CashRateForm, type CashRateTarget } from './components/cash-rate-form.svelte';
export {
	default as CashPocketForm,
	type CashPocketTarget
} from './components/cash-pocket-form.svelte';
export { default as CashMoveForm, type CashMoveTarget } from './components/cash-move-form.svelte';
export {
	default as CashDepositForm,
	type CashDepositTarget
} from './components/cash-deposit-form.svelte';

export * from './cash';
export * from './pockets';
export * from './rates';
export * from './deposits';
export * from './interest';
export * from './schemas';
