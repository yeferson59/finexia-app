/**
 * Feature `platforms` — superficie pública.
 *
 * Componentes del área de plataformas de inversión (`routes/dashboard/platforms/**`):
 * la tarjeta del listado, el detalle (ver/editar/eliminar) y el formulario de alta.
 * `platforms.ts` aporta las constantes/tipos compartidos.
 */
export { default as PlatformCard } from './components/platform-card.svelte';
export { default as PlatformDetail } from './components/platform-detail.svelte';
export { default as PlatformAddForm } from './components/platform-add-form.svelte';

export * from './platforms';
