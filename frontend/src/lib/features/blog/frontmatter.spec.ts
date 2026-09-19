import { describe, it, expect } from 'vitest';
import { splitFrontmatter } from './frontmatter';

describe('splitFrontmatter', () => {
	it('separa la cabecera del cuerpo', () => {
		const { data, body } = splitFrontmatter('---\ntitle: Hola\n---\nEl cuerpo.\n');

		expect(data).toEqual({ title: 'Hola' });
		expect(body).toBe('El cuerpo.\n');
	});

	it('deja todo el archivo como cuerpo cuando no hay cabecera', () => {
		expect(splitFrontmatter('# Solo Markdown')).toEqual({ data: {}, body: '# Solo Markdown' });
	});

	it('lee listas en línea', () => {
		const { data } = splitFrontmatter('---\ntags: [guías, producto]\n---\n');

		expect(data.tags).toEqual(['guías', 'producto']);
	});

	it('lee booleanos y deja el resto como texto', () => {
		const { data } = splitFrontmatter('---\ndraft: true\ndate: 2026-09-15\n---\n');

		expect(data).toEqual({ draft: true, date: '2026-09-15' });
	});

	// Un dos puntos dentro del valor es lo normal en un titular; solo el primero
	// separa la clave.
	it('parte por el primer dos puntos y no por los siguientes', () => {
		const { data } = splitFrontmatter('---\ntitle: Finexia: el mapa\n---\n');

		expect(data.title).toBe('Finexia: el mapa');
	});

	it('respeta lo entrecomillado, incluido un # dentro del valor', () => {
		const { data } = splitFrontmatter('---\ntitle: "Todo sobre #finanzas"\n---\n');

		expect(data.title).toBe('Todo sobre #finanzas');
	});

	it('descarta el comentario que va detrás de un valor sin comillas', () => {
		const { data } = splitFrontmatter('---\nauthor: Equipo # firma por defecto\n---\n');

		expect(data.author).toBe('Equipo');
	});

	it('ignora líneas en blanco y comentarios sueltos', () => {
		const { data } = splitFrontmatter('---\n# un comentario\n\ntitle: Hola\n---\n');

		expect(data).toEqual({ title: 'Hola' });
	});

	it('acepta finales de línea de Windows', () => {
		const { data, body } = splitFrontmatter('---\r\ntitle: Hola\r\n---\r\nCuerpo.');

		expect(data).toEqual({ title: 'Hola' });
		expect(body).toBe('Cuerpo.');
	});
});
