#!/usr/bin/env node
/**
 * Genera el PDF del manual de usuario a partir de `docs/MANUAL_DE_USUARIO.md`.
 *
 * Uso: `pnpm manual:build` (desde `frontend/`).
 *
 * Escribe tres cosas:
 *   - `docs/MANUAL_DE_USUARIO.pdf` — el documento del repositorio.
 *   - `frontend/static/manual-usuario.pdf` — la copia que descarga el usuario.
 *   - `.../features/guide/manual-meta.ts` — versión, fecha, tamaño, índice y la
 *     huella de las fuentes, para que la aplicación pueda decir de cuándo es la
 *     guía que está sirviendo y `check.mjs` pueda detectar que se quedó vieja.
 *
 * Imprime con el Chromium de Playwright, que ya está instalado para los tests,
 * así que no añade ninguna herramienta nueva ni en local ni en CI.
 */
import { chromium } from 'playwright';
import { copyFileSync, readFileSync, rmSync, statSync, writeFileSync, mkdirSync } from 'node:fs';
import { dirname, join } from 'node:path';
import {
	DOCS,
	MARKDOWN,
	META,
	PDF_DOCS,
	PDF_STATIC,
	readFrontMatter,
	readSections,
	renderHtml,
	sourceHash
} from './manual.mjs';

const markdown = readFileSync(MARKDOWN, 'utf8');
const { version, date } = readFrontMatter(markdown);
const sections = readSections(markdown);

// El HTML se escribe dentro de `docs/` para que las capturas, referenciadas
// como `img/manual/...`, resuelvan solas al abrirlo como file://.
const tempHtml = join(DOCS, '.manual-build.html');
writeFileSync(tempHtml, renderHtml(markdown));

const footer = `
	<div style="width:100%;padding:0 16mm;font-family:Helvetica,Arial,sans-serif;font-size:7pt;color:#8a8d93;display:flex;justify-content:space-between;">
		<span>Finexia — Manual de Usuario · v${version}</span>
		<span class="pageNumber"></span>
	</div>`;

const browser = await chromium.launch({
	...(process.env.CHROMIUM_EXECUTABLE_PATH
		? { executablePath: process.env.CHROMIUM_EXECUTABLE_PATH }
		: {})
});

try {
	const page = await browser.newPage();
	await page.goto(`file://${tempHtml}`, { waitUntil: 'load' });
	// Las capturas son grandes: sin esto la primera página puede imprimirse
	// antes de que el navegador las haya decodificado.
	await page.evaluate(() =>
		Promise.all(
			[...document.images].filter((img) => !img.complete).map((img) => img.decode().catch(() => {}))
		)
	);

	await page.pdf({
		path: PDF_DOCS,
		format: 'A4',
		printBackground: true,
		displayHeaderFooter: true,
		headerTemplate: '<span></span>',
		footerTemplate: footer,
		margin: { top: '18mm', bottom: '20mm', left: '16mm', right: '16mm' }
	});
} finally {
	await browser.close();
	rmSync(tempHtml, { force: true });
}

copyFileSync(PDF_DOCS, PDF_STATIC);

const bytes = statSync(PDF_DOCS).size;
const hash = sourceHash();

mkdirSync(dirname(META), { recursive: true });
writeFileSync(
	META,
	`// GENERADO por \`pnpm manual:build\` — no editar a mano.
//
// Describe el PDF que hay en \`static/manual-usuario.pdf\`. \`sourceHash\` es la
// huella del manual y sus capturas: \`pnpm check:manual\` la recalcula y falla
// si el PDF publicado ya no corresponde al manual del repositorio.

export interface ManualMeta {
	/** Versión declarada en la portada del manual. */
	version: string;
	/** Fecha declarada en la portada. */
	date: string;
	/** Tamaño del PDF en bytes. */
	bytes: number;
	/** Fecha de generación del PDF, en ISO. */
	generatedAt: string;
	/** Huella de \`docs/MANUAL_DE_USUARIO.md\` y sus capturas. */
	sourceHash: string;
	/** Secciones de primer nivel, para el índice de la página de la guía. */
	sections: { number: number; title: string }[];
}

export const manual: ManualMeta = ${JSON.stringify(
		{
			version,
			date,
			bytes,
			generatedAt: new Date().toISOString(),
			sourceHash: hash,
			sections
		},
		null,
		'\t'
	).replace(/\n/g, '\n')};
`
);

const mb = (bytes / 1024 / 1024).toFixed(1);
console.log(`manual:build — v${version} (${date}), ${sections.length} secciones, ${mb} MB`);
console.log(`  ${PDF_DOCS}`);
console.log(`  ${PDF_STATIC}`);
console.log(`  ${META}`);
