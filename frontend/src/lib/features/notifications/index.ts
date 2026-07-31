/**
 * Feature `notifications` — superficie pública.
 *
 * Los canales de notificación de `routes/dashboard/notifications`: las
 * preferencias de correo, que es el único con formulario, y las alertas en la
 * app, todavía por llegar.
 *
 * `notification-section` es el marco compartido por los dos y no forma parte de
 * la superficie pública.
 */
export { default as EmailPreferences } from './components/email-preferences.svelte';
export { default as InAppAlerts } from './components/in-app-alerts.svelte';

export * from './schemas';
