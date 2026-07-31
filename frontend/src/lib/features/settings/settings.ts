/**
 * Helpers y tipos del dominio `settings`.
 *
 * La página de ajustes es un montón de formularios independientes que comparten
 * un solo `form` de SvelteKit: cada acción devuelve `{ action: '<nombre>', … }` y
 * cada sección tiene que decidir si ese resultado es suyo. Ese reparto vivía
 * repetido como `$derived(form?.action === 'x' && (form as {…}).success)` en la
 * página; aquí se centraliza para que las secciones no repitan el cast.
 */

import type { ActiveSession } from '$lib/api/types';

export type { ActiveSession, MarketCredential, TwoFactorStatus } from '$lib/api/types';

/** `form` de la página de ajustes, sin tipar por acción. */
export type SettingsForm = Record<string, unknown> | null | undefined;

/** `true` si `form` es el resultado correcto de esa acción. */
export function actionSucceeded(form: SettingsForm, action: string): boolean {
	return form?.action === action && form?.success === true;
}

/** Mensaje de error de esa acción, o cadena vacía si el `form` no es suyo. */
export function actionError(form: SettingsForm, action: string): string {
	return form?.action === action ? ((form?.error as string) ?? '') : '';
}

/**
 * Campo del resultado de una acción concreta, solo si esa acción tuvo éxito.
 * Sirve para los datos que devuelven las acciones (`imageUrl`, `secret`,
 * `recoveryCodes`…) sin volver a comprobar `action` + `success` en cada sitio.
 */
export function actionData<T>(form: SettingsForm, action: string, field: string): T | undefined {
	return actionSucceeded(form, action) ? (form?.[field] as T | undefined) : undefined;
}

/**
 * Códigos de recuperación recién emitidos.
 *
 * Los devuelven dos acciones distintas —activar 2FA y regenerar los códigos— y
 * solo se muestran una vez: el backend no los vuelve a enviar.
 */
export function issuedRecoveryCodes(form: SettingsForm): string[] {
	const issuedBy = form?.action === 'enable2fa' || form?.action === 'regenerate2faCodes';
	return issuedBy && form?.success === true ? ((form?.recoveryCodes as string[]) ?? []) : [];
}

/**
 * Descripción legible de un dispositivo a partir de su `User-Agent`.
 *
 * Deliberadamente aproximado: solo tiene que ayudar a reconocer una sesión
 * propia frente a una ajena, no identificar el navegador con precisión.
 */
export function describeDevice(userAgent: string | null): string {
	if (!userAgent) return 'Dispositivo desconocido';
	const ua = userAgent.toLowerCase();

	let browser = 'Navegador desconocido';
	if (ua.includes('edg/')) browser = 'Edge';
	else if (ua.includes('opr/') || ua.includes('opera')) browser = 'Opera';
	else if (ua.includes('chrome')) browser = 'Chrome';
	else if (ua.includes('safari')) browser = 'Safari';
	else if (ua.includes('firefox')) browser = 'Firefox';

	let os = '';
	if (ua.includes('windows')) os = 'Windows';
	else if (ua.includes('android')) os = 'Android';
	else if (ua.includes('iphone') || ua.includes('ipad')) os = 'iOS';
	else if (ua.includes('mac os') || ua.includes('macintosh')) os = 'macOS';
	else if (ua.includes('linux')) os = 'Linux';

	return os ? `${browser} · ${os}` : browser;
}

const sessionDateFormatter = new Intl.DateTimeFormat('es', {
	day: '2-digit',
	month: 'short',
	hour: '2-digit',
	minute: '2-digit'
});

/** Fecha de última actividad de una sesión; `—` si el backend manda basura. */
export function formatSessionDate(value: string): string {
	const date = new Date(value);
	return Number.isNaN(date.getTime()) ? '—' : sessionDateFormatter.format(date);
}

/** Sesiones distintas de la actual: las que puede cerrar «cerrar las demás». */
export function countOtherSessions(sessions: ActiveSession[] | undefined): number {
	return (sessions ?? []).filter((s) => !s.current).length;
}
