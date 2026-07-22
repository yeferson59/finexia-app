/**
 * Feature `landing` — superficie pública.
 *
 * Componentes de la página de marketing (`routes/+page.svelte`) y del layout
 * legal (`routes/(legal)/+layout.svelte`, que reutiliza `Brand` y `Footer`).
 *
 * La hoja de estilos compartida se importa por su ruta como side-effect:
 * `import '$lib/features/landing/landing.css'`.
 */
export { default as Ticker } from './components/ticker.svelte';
export { default as Header } from './components/header.svelte';
export { default as Hero } from './components/hero.svelte';
export { default as Benefits } from './components/benefits.svelte';
export { default as HowItWorks } from './components/how-it-works.svelte';
export { default as Metrics } from './components/metrics.svelte';
export { default as Faq } from './components/faq.svelte';
export { default as FinalCta } from './components/final-cta.svelte';
export { default as Footer } from './components/footer.svelte';
export { default as Brand } from './components/brand.svelte';
