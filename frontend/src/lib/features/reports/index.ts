/**
 * Feature `reports` — superficie pública.
 *
 * La ficha de resultados de `routes/dashboard/reports`: la cifra de cabecera,
 * la matriz de rentabilidad mes a mes, las notas de movimiento y riesgo, la
 * proyección a cinco años y las descargas.
 *
 * Los cálculos van en dos módulos —`reports.ts` (cabecera, matriz y notas) y
 * `projection.ts` (la proyección y su geometría)—, separados por tamaño. Los
 * consume el loader de la página.
 */
export { default as RecordHeadline } from './components/record-headline.svelte';
export { default as MonthlyReturns } from './components/monthly-returns.svelte';
export { default as KeyStatistics } from './components/key-statistics.svelte';
export { default as GrowthProjection } from './components/growth-projection.svelte';
export { default as ReportDownloads } from './components/report-downloads.svelte';

export * from './reports';
export * from './projection';
