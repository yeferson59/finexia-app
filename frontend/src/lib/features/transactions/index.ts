/**
 * Feature `transactions` — superficie pública.
 *
 * `ImportWizard` orquesta el flujo de importación (upload → mapeo/preview →
 * resultado) que consume `routes/dashboard/transactions/import`. Los pasos
 * (`import-upload-step`, `import-mapping-step`, `import-result-step`) son
 * internos de la feature.
 *
 * `TransactionLedger` es la tabla de movimientos de la cuenta entera que pinta
 * `routes/dashboard/transactions`.
 */
export { default as ImportWizard } from './components/import-wizard.svelte';
export { default as TransactionLedger } from './components/transaction-ledger.svelte';

export * from './transactions';
export * from './types';
