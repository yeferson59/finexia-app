/**
 * Feature `auth` — superficie pública.
 *
 * `routes/auth/**` compone estos componentes y sus form actions validan con los
 * schemas de `schemas.ts`. Los sub-formularios de `login-register`
 * (login-form, register-form, two-factor-challenge, invite-only-notice,
 * password-input, auth-field) son internos: solo se importan dentro de la feature.
 * El armazón de las pantallas vive en `$lib/ui/public-shell.svelte`: lo comparten
 * con la autorización OAuth y la página de error.
 */
export { default as LoginRegister } from './components/login-register.svelte';
export { default as ForgotPasswordForm } from './components/forgot-password-form.svelte';
export { default as ResetPasswordForm } from './components/reset-password-form.svelte';
export { default as VerifyEmailPanel } from './components/verify-email-panel.svelte';
export { default as AcceptInviteForm } from './components/accept-invite-form.svelte';

export * from './schemas';
