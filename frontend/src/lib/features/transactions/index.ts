/**
 * Feature `transactions` — superficie pública.
 *
 * `ImportWizard` orquesta el flujo de importación (upload → mapeo/preview →
 * resultado) que consume `routes/dashboard/transactions/import`. Los pasos
 * (`import-upload-step`, `import-mapping-step`, `import-result-step`) son
 * internos de la feature.
 */
export { default as ImportWizard } from './components/import-wizard.svelte';

export * from './types';
