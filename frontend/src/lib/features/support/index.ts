/**
 * Feature `support` — superficie pública.
 *
 * La página `/apoyar`: el aporte con el botón de pagos de Bold. La firma de
 * la orden, su guardado y el webhook viven en el backend (`$lib/api/support`).
 */
export { default as BoldPayment } from './components/bold-payment.svelte';
export { default as PaymentResult } from './components/payment-result.svelte';
export { default as SupportNotes } from './components/support-notes.svelte';
export * from './support';
export * from './schemas';
