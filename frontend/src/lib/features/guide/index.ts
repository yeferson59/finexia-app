/**
 * Feature `guide` — superficie pública.
 *
 * La guía de usuario dentro de la aplicación: el visor y la descarga del PDF
 * que se genera desde `docs/MANUAL_DE_USUARIO.md`.
 *
 * `manual-meta.ts` lo escribe `pnpm manual:build`; no se edita a mano.
 */
export { default as GuideViewer } from './components/guide-viewer.svelte';
export { default as GuideContents } from './components/guide-contents.svelte';
export * from './manual-meta';
export * from './guide';
