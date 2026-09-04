/**
 * Helpers y tipos del dominio `settings`.
 *
 * La página de ajustes es un montón de formularios independientes que comparten
 * un solo `form` de SvelteKit. El reparto de ese `form` entre secciones no tiene
 * nada de ajustes —lo hace igual la página de notificaciones—, así que vive en
 * `lib/shared/form`; aquí se reexporta con el nombre que usan las secciones.
 */

import type { ActiveSession, MCPTokenSecret } from '$lib/api/types';
import type { ActionForm } from '$lib/shared/form';

export type {
	ActiveSession,
	MarketCredential,
	MCPToken,
	MCPTokenSecret,
	OAuthGrant,
	TwoFactorStatus
} from '$lib/api/types';
export { actionSucceeded, actionError, actionData } from '$lib/shared/form';

/** `form` de la página de ajustes, sin tipar por acción. */
export type SettingsForm = ActionForm;

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

/**
 * Token MCP recién emitido.
 *
 * Como los códigos de recuperación, lo devuelven dos acciones distintas —crear
 * y rotar— y solo existe en la respuesta que lo emitió: el backend guarda su
 * hash, así que ni el listado ni una recarga lo pueden volver a mostrar.
 */
export function issuedMCPToken(form: SettingsForm): MCPTokenSecret | null {
	const issuedBy = form?.action === 'createMcpToken' || form?.action === 'rotateMcpToken';
	if (!issuedBy || form?.success !== true) return null;

	return (form?.mcpToken as MCPTokenSecret | undefined) ?? null;
}

/** Opciones de caducidad del desplegable, en días (0 = sin caducidad). */
export const MCP_TOKEN_EXPIRY_OPTIONS = [
	{ days: 30, label: '30 días' },
	{ days: 90, label: '90 días (recomendado)' },
	{ days: 365, label: '1 año' },
	{ days: 0, label: 'Sin caducidad' }
] as const;

/**
 * Fecha de un token, o `—` cuando el campo viene vacío: `lastUsedAt` es null
 * mientras nadie lo haya usado y `expiresAt` cuando no caduca.
 */
export function formatMCPTokenDate(value: string | null): string {
	if (!value) return '—';

	return formatSessionDate(value);
}

/**
 * Mensaje para un fallo del backend en las acciones de tokens MCP.
 *
 * Vive aquí y no en la action por lo mismo que los schemas: es copia de
 * dominio, y tenerla junta evita que «ese token ya no existe» se escriba de dos
 * maneras en las dos acciones que pueden encontrárselo. Sigue siendo una
 * función pura —no toca `lib/api`— para que el barril de la feature no arrastre
 * código de servidor al navegador.
 */
export function mcpTokenError(action: string, status: number, details?: string): string {
	// El 409 lo produce el nombre repetido o el límite de tokens, y el backend
	// dice cuál de los dos: su detalle es más útil que un genérico.
	if (status === 409) return details ?? 'Ya tienes un token con ese nombre.';
	if (status === 404) return 'Ese token ya no existe. Actualiza la página.';

	const verb =
		action === 'createMcpToken' ? 'crear' : action === 'rotateMcpToken' ? 'rotar' : 'eliminar';

	return `No se pudo ${verb} el token. Inténtalo de nuevo.`;
}
