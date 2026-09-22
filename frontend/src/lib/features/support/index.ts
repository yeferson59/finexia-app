/**
 * Feature `support` — superficie pública.
 *
 * La página `/apoyar`: el aporte con el botón de pagos de Bold. La firma de
 * la orden y la consulta del resultado viven en `$lib/server/bold`.
 */
export { default as BoldPayment } from './components/bold-payment.svelte';
export { default as PaymentResult } from './components/payment-result.svelte';
export { default as SupportNotes } from './components/support-notes.svelte';
export * from './support';
export * from './schemas';
