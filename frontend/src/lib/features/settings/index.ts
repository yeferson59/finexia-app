/**
 * Feature `settings` — superficie pública.
 *
 * Las secciones de `routes/dashboard/settings`, una por tarjeta: perfil,
 * apariencia, seguridad (contraseña), 2FA, datos de mercado, tokens MCP,
 * aplicaciones conectadas por OAuth y sesiones activas.
 * Todas leen el mismo `form` de la página y se quedan solo con el resultado de
 * sus propias acciones (ver `settings.ts`).
 *
 * `settings-section` (el chrome de tarjeta), `avatar-uploader`,
 * `two-factor-setup` y `two-factor-manage` son internos de sus secciones y no
 * forman parte de la superficie pública.
 */
export { default as ProfileSection } from './components/profile-section.svelte';
export { default as AppearanceSection } from './components/appearance-section.svelte';
export { default as PasswordSection } from './components/password-section.svelte';
export { default as TwoFactorSection } from './components/two-factor-section.svelte';
export { default as MarketCredentials } from './components/market-credentials.svelte';
export { default as MCPTokensSection } from './components/mcp-tokens-section.svelte';
export { default as OAuthGrantsSection } from './components/oauth-grants-section.svelte';
export { default as SessionsSection } from './components/sessions-section.svelte';

export * from './settings';
export * from './schemas';
