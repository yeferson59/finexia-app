/**
 * Feature `landing` — superficie pública.
 *
 * Componentes de la página de marketing (`routes/+page.svelte`) y del layout
 * legal (`routes/(legal)/+layout.svelte`, que reutiliza el `Footer`).
 *
 * La hoja de estilos compartida se importa por su ruta como side-effect:
 * `import '$lib/features/landing/landing.css'`, después de `$lib/ui/public.css`,
 * que pone los tokens y el contenedor `.lp` del que cuelga todo.
 */
export { default as Header } from './components/header.svelte';
export { default as Hero } from './components/hero.svelte';
export { default as ProductTour } from './components/product-tour.svelte';
export { default as HowItWorks } from './components/how-it-works.svelte';
export { default as Trust } from './components/trust.svelte';
export { default as Faq } from './components/faq.svelte';
export { default as FinalCta } from './components/final-cta.svelte';
export { default as Footer } from './components/footer.svelte';
export * from './product-tour';
export * from './schemas';
