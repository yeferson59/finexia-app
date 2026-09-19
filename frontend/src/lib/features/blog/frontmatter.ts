/**
 * El bloque `---` de cabecera de un artículo, separado del cuerpo.
 *
 * Es un subconjunto deliberadamente pequeño de YAML: pares `clave: valor`,
 * cadenas con o sin comillas, booleanos y listas en línea (`[uno, dos]`). Con
 * eso se describe un artículo entero, y a cambio no entra una dependencia más
 * en el proyecto solo para leer seis campos.
 *
 * Aquí no se valida nada: esto devuelve lo que ponga el archivo y es
 * `schemas.ts` quien dice si eso es un artículo válido. Así el error que ve
 * quien escribe habla de campos («falta `description`»), no de sintaxis.
 */

/** Cabecera al principio del archivo, entre dos líneas de tres guiones. */
const FENCE = /^---\r?\n([\s\S]*?)\r?\n---[ \t]*(?:\r?\n|$)/;

export interface Frontmatter {
	/** Los campos declarados en la cabecera, sin interpretar su significado. */
	data: Record<string, unknown>;
	/** El Markdown que viene después. */
	body: string;
}

/** Parte un archivo en cabecera y cuerpo. Sin cabecera, todo es cuerpo. */
export function splitFrontmatter(source: string): Frontmatter {
	const match = FENCE.exec(source);
	if (!match) return { data: {}, body: source };
	return { data: parseBlock(match[1]), body: source.slice(match[0].length) };
}

/** Las líneas de la cabecera, una por campo. Las vacías y los `#` no cuentan. */
function parseBlock(block: string): Record<string, unknown> {
	const data: Record<string, unknown> = {};

	for (const raw of block.split(/\r?\n/)) {
		const line = raw.trim();
		if (!line || line.startsWith('#')) continue;

		const colon = line.indexOf(':');
		if (colon === -1) continue;

		const key = line.slice(0, colon).trim();
		if (key) data[key] = parseValue(line.slice(colon + 1).trim());
	}

	return data;
}

/**
 * El valor de un campo.
 *
 * Entrecomillado se toma literal —ahí un `#` es un `#` y no un comentario—; sin
 * comillas, lo que venga después de ` #` se descarta.
 */
function parseValue(raw: string): unknown {
	if (!raw) return '';

	const quote = raw[0];
	if (quote === '"' || quote === "'") {
		const end = raw.indexOf(quote, 1);
		return end === -1 ? raw.slice(1) : raw.slice(1, end);
	}

	const value = raw.split(' #')[0].trim();

	if (value.startsWith('[') && value.endsWith(']')) {
		return value
			.slice(1, -1)
			.split(',')
			.map((item) => parseValue(item.trim()))
			.filter((item) => item !== '');
	}

	if (value === 'true') return true;
	if (value === 'false') return false;

	return value;
}
