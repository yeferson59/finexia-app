/**
 * Feature `platforms` — superficie pública.
 *
 * Componentes del área de plataformas de inversión (`routes/dashboard/platforms/**`):
 * la tarjeta del listado, el detalle (ver/editar/eliminar) y el formulario de alta.
 * `platforms.ts` aporta las constantes y reexporta el contrato `Platform` de
 * `$lib/api/types`.
 *
 * `platform-edit-form` y `platform-delete-dialog` son internos de
 * `platform-detail` (import relativo) y no forman parte de la superficie pública.
 */
export { default as PlatformCard } from './components/platform-card.svelte';
export { default as PlatformDetail } from './components/platform-detail.svelte';
export { default as PlatformAddForm } from './components/platform-add-form.svelte';

export * from './platforms';
export * from './schemas';
