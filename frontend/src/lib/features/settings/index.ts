/**
 * Feature `settings` — superficie pública.
 *
 * Las secciones de `routes/dashboard/settings`, repartidas en dos grupos:
 * cómo entras a la cuenta (contraseña, 2FA, sesiones) y qué tienes conectado
 * (datos de mercado, asistentes MCP, aplicaciones OAuth), con el perfil
 * abriendo la página. Todas leen el mismo `form` y se quedan solo con el
 * resultado de sus propias acciones (ver `settings.ts`).
 *
 * `settings-section` (el carril y los controles que comparten todas),
 * `avatar-uploader`, `two-factor-setup` y `two-factor-manage` son internos y no
 * forman parte de la superficie pública.
 *
 * La sección «Apariencia» ya no existe: su único contenido era decir que no hay
 * más de un tema, así que era una sección sobre su propia ausencia.
 */
export { default as SettingsGroup } from './components/settings-group.svelte';
export { default as ProfileSection } from './components/profile-section.svelte';
export { default as PasswordSection } from './components/password-section.svelte';
export { default as TwoFactorSection } from './components/two-factor-section.svelte';
export { default as MarketCredentials } from './components/market-credentials.svelte';
export { default as MCPTokensSection } from './components/mcp-tokens-section.svelte';
export { default as OAuthGrantsSection } from './components/oauth-grants-section.svelte';
export { default as SessionsSection } from './components/sessions-section.svelte';

export * from './settings';
export * from './schemas';
