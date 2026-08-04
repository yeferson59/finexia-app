/**
 * Helpers puros de la feature `guide`.
 *
 * Los datos del manual vienen de `manual-meta.ts`, que genera el build; aquí
 * solo está la presentación de esos datos.
 */

/** Tamaño del PDF en la unidad que toque, para poder decidir si descargarlo. */
export function formatBytes(bytes: number): string {
	if (!Number.isFinite(bytes) || bytes <= 0) return '—';
	if (bytes < 1024) return `${bytes} B`;
	if (bytes < 1024 * 1024) return `${Math.round(bytes / 1024)} KB`;
	return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

/** Fecha de generación en formato largo; el ISO crudo no dice nada al usuario. */
export function formatGeneratedAt(iso: string): string {
	const date = new Date(iso);
	if (Number.isNaN(date.getTime())) return '—';
	return date.toLocaleDateString('es-CO', { day: '2-digit', month: 'long', year: 'numeric' });
}
