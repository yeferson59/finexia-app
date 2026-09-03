/**
 * Feature `reports` — superficie pública.
 *
 * Los cuatro bloques de `routes/dashboard/reports`: calendario de rentabilidad,
 * estadísticas de riesgo, proyección de crecimiento y descargas. `reports.ts`
 * aporta los cálculos que alimentan la página —los consume el loader— y las
 * constantes de presentación.
 *
 * `report-panel` es el marco compartido por los cuatro bloques y no forma parte
 * de la superficie pública.
 *
 * Los cálculos van en dos módulos: `reports.ts` (calendario y estadísticas) y
 * `projection.ts` (la proyección a cinco años), separados por tamaño.
 */
export { default as PerformanceCalendars } from './components/performance-calendars.svelte';
export { default as KeyStatistics } from './components/key-statistics.svelte';
export { default as GrowthProjection } from './components/growth-projection.svelte';
export { default as ReportDownloads } from './components/report-downloads.svelte';

export * from './reports';
export * from './projection';
