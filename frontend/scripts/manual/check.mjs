#!/usr/bin/env node
/**
 * Comprueba que el PDF que descarga el usuario corresponde al manual vigente.
 *
 * Uso: `pnpm check:manual` (desde `frontend/`). Se ejecuta en CI.
 *
 * El manual y su PDF son dos archivos distintos, y nada obliga a tocarlos a la
 * vez: es fácil corregir una sección y publicar durante meses un PDF que ya no
 * dice eso. Este check compara la huella de las fuentes con la que guardó el
 * último `pnpm manual:build` y falla si no coinciden.
 */
import { existsSync, readFileSync, statSync } from 'node:fs';
import { relative } from 'node:path';
import { META, PDF_DOCS, PDF_STATIC, ROOT, sourceHash } from './manual.mjs';

const problems = [];
const rel = (path) => relative(ROOT, path);

if (!existsSync(META)) {
	problems.push(`falta ${rel(META)}: nunca se ejecutó \`pnpm manual:build\``);
} else {
	// El archivo lo escribe el build y luego pasa por prettier, que puede quitar
	// las comillas de la clave o cambiarlas de dobles a simples: la expresión
	// admite las dos formas para no romperse por el formato.
	const recorded = /sourceHash["']?\s*:\s*["']([a-f0-9]+)["']/.exec(
		readFileSync(META, 'utf8')
	)?.[1];
	const current = sourceHash();

	if (!recorded) {
		problems.push(`${rel(META)} no declara sourceHash`);
	} else if (recorded !== current) {
		problems.push(
			`el manual cambió después de generar el PDF\n` +
				`      esperado ${recorded.slice(0, 16)}…\n` +
				`      actual   ${current.slice(0, 16)}…`
		);
	}
}

for (const pdf of [PDF_DOCS, PDF_STATIC]) {
	if (!existsSync(pdf)) problems.push(`falta ${rel(pdf)}`);
	else if (statSync(pdf).size === 0) problems.push(`${rel(pdf)} está vacío`);
}

// Las dos copias tienen que ser el mismo documento: `static/` es lo que se
// sirve, `docs/` lo que se revisa en el repositorio.
if (existsSync(PDF_DOCS) && existsSync(PDF_STATIC)) {
	const docsBytes = readFileSync(PDF_DOCS);
	const staticBytes = readFileSync(PDF_STATIC);
	if (!docsBytes.equals(staticBytes)) {
		problems.push(`${rel(PDF_DOCS)} y ${rel(PDF_STATIC)} no son el mismo PDF`);
	}
}

if (problems.length === 0) {
	console.log('check:manual — el PDF publicado corresponde al manual vigente');
	process.exit(0);
}

console.error(`check:manual — ${problems.length} incidencia(s):\n`);
for (const problem of problems) console.error(`  - ${problem}`);
console.error('\nRegenera la guía con `pnpm manual:build` y sube el PDF resultante.');
process.exit(1);
