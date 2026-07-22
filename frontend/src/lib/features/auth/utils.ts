/**
 * Normaliza los errores de una form action a un mapa `campo -> mensaje`.
 *
 * Las actions devuelven bien un array de issues de Zod
 * (`[{ path, message }]`), bien un objeto ya mapeado (`{ email: '...' }`); esta
 * función unifica ambos casos para que los formularios lean siempre igual.
 */
export function parseErrors(errors: unknown): Record<string, string> {
	if (!errors) return {};
	if (Array.isArray(errors)) {
		const result: Record<string, string> = {};
		for (const issue of errors) {
			const key = issue?.path?.[0] ? String(issue.path[0]) : 'server';
			if (!result[key]) result[key] = issue.message ?? 'Campo inválido';
		}
		return result;
	}
	if (typeof errors === 'object') return errors as Record<string, string>;
	return {};
}
