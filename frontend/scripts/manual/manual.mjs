/**
 * Piezas compartidas por `build.mjs` (genera el PDF) y `check.mjs` (comprueba
 * que el PDF publicado sigue el manual vigente).
 *
 * La fuente única es `docs/MANUAL_DE_USUARIO.md` con sus capturas. De ahí sale
 * el PDF que se descarga el usuario, y de ahí sale también la huella que
 * permite detectar en CI que alguien tocó el manual y olvidó regenerarlo.
 */
import { createHash } from 'node:crypto';
import { readFileSync, readdirSync } from 'node:fs';
import { join } from 'node:path';
import { marked } from 'marked';

/** Raíz del repositorio, subiendo desde `frontend/scripts/manual/`. */
export const ROOT = new URL('../../../', import.meta.url).pathname;
export const DOCS = join(ROOT, 'docs');
export const MARKDOWN = join(DOCS, 'MANUAL_DE_USUARIO.md');
export const IMAGES = join(DOCS, 'img', 'manual');

/** Salidas del build: el PDF de `docs/` y su copia servida por la aplicación. */
export const PDF_DOCS = join(DOCS, 'MANUAL_DE_USUARIO.pdf');
export const PDF_STATIC = join(ROOT, 'frontend', 'static', 'manual-usuario.pdf');
export const META = join(ROOT, 'frontend', 'src', 'lib', 'features', 'guide', 'manual-meta.ts');

/**
 * Versión de la plantilla. Súbela al cambiar el HTML o el CSS del PDF: entra en
 * la huella, así que el check pedirá regenerar aunque el manual no haya
 * cambiado ni una coma.
 */
export const TEMPLATE_VERSION = 2;

/**
 * Huella de todo lo que acaba dentro del PDF: el texto del manual, cada
 * captura y la plantilla. Si cambia cualquiera, el PDF publicado está viejo.
 */
export function sourceHash() {
	const hash = createHash('sha256');
	hash.update(`template:${TEMPLATE_VERSION}\n`);
	hash.update(readFileSync(MARKDOWN));
	for (const name of readdirSync(IMAGES).sort()) {
		hash.update(name);
		hash.update(readFileSync(join(IMAGES, name)));
	}
	return hash.digest('hex');
}

/** Portada del manual: versión y fecha declaradas en las primeras líneas. */
export function readFrontMatter(markdown) {
	const version = markdown.match(/\*\*Versión del documento:\*\*\s*(.+)/)?.[1]?.trim() ?? '';
	const date = markdown.match(/\*\*Fecha:\*\*\s*(.+)/)?.[1]?.trim() ?? '';
	return { version, date };
}

/**
 * Ancla de un encabezado al estilo GitHub, que es el que ya usan los enlaces de
 * la tabla de contenido del manual: minúsculas, sin puntuación y con guiones.
 */
export function slugify(text) {
	return text
		.toLowerCase()
		.replace(/[^\p{L}\p{N}\s-]/gu, '')
		.trim()
		.replace(/\s+/g, '-');
}

/** Secciones de primer nivel (`## N. Título`), para el índice de la aplicación. */
export function readSections(markdown) {
	const sections = [];
	for (const line of markdown.split('\n')) {
		const match = /^##\s+(\d+)\.\s+(.+)$/.exec(line);
		if (match) sections.push({ number: Number(match[1]), title: match[2].trim() });
	}
	return sections;
}

/**
 * HTML del manual listo para imprimir.
 *
 * Sin fuentes remotas a propósito: el PDF se genera igual en un portátil sin
 * red que en CI, y no depende de que Google Fonts responda ese día.
 */
export function renderHtml(markdown) {
	const { version, date } = readFrontMatter(markdown);

	// El encabezado de portada ya se pinta aparte; fuera del cuerpo, incluida la
	// regla horizontal que lo cierra (si no, deja una hoja en blanco detrás).
	const body = markdown.replace(/^#\s.+\n+(?:\*\*.+\n)+\s*(?:---\n)?/, '');

	const renderer = new marked.Renderer();
	renderer.heading = ({ text, depth }) => {
		const inline = marked.parseInline(text);
		return `<h${depth} id="${slugify(text)}">${inline}</h${depth}>\n`;
	};

	const content = marked.parse(body, { renderer, gfm: true });

	return `<!doctype html>
<html lang="es">
<head>
<meta charset="utf-8" />
<title>Manual de Usuario — FINEXIA</title>
<style>
	@page {
		size: A4;
		margin: 18mm 16mm 20mm;
	}

	:root {
		--ink: #16181c;
		--muted: #5c6068;
		--rule: #dfe1e5;
		--amber: #a86e14;
		--amber-soft: #fdf6e9;
	}

	* { box-sizing: border-box; }

	body {
		margin: 0;
		font-family: 'Helvetica Neue', Helvetica, Arial, sans-serif;
		font-size: 10.5pt;
		line-height: 1.62;
		color: var(--ink);
		-webkit-print-color-adjust: exact;
		print-color-adjust: exact;
	}

	/* Portada: ocupa su propia hoja. */
	.cover {
		height: 245mm;
		display: flex;
		flex-direction: column;
		justify-content: center;
		page-break-after: always;
	}

	.cover-mark {
		display: flex;
		align-items: center;
		gap: 10px;
		margin-bottom: 34px;
	}

	.cover-mark span {
		font-size: 15pt;
		font-weight: 700;
		letter-spacing: 0.16em;
	}

	.cover h1 {
		margin: 0;
		font-family: Georgia, 'Times New Roman', serif;
		font-size: 34pt;
		font-weight: 400;
		line-height: 1.1;
		letter-spacing: -0.02em;
	}

	.cover p {
		margin: 18px 0 0;
		max-width: 108mm;
		color: var(--muted);
		font-size: 11pt;
	}

	.cover-meta {
		margin-top: 40px;
		padding-top: 18px;
		border-top: 2px solid var(--amber);
		display: flex;
		gap: 34px;
		font-size: 9.5pt;
		color: var(--muted);
	}

	.cover-meta b {
		display: block;
		font-size: 7.5pt;
		letter-spacing: 0.16em;
		text-transform: uppercase;
		color: var(--ink);
		margin-bottom: 3px;
	}

	h2 {
		font-family: Georgia, 'Times New Roman', serif;
		font-size: 18pt;
		font-weight: 400;
		margin: 26px 0 12px;
		padding-bottom: 7px;
		border-bottom: 1px solid var(--rule);
		page-break-after: avoid;
	}

	h3 {
		font-size: 12pt;
		margin: 20px 0 8px;
		page-break-after: avoid;
	}

	h4 {
		font-size: 10.5pt;
		margin: 16px 0 6px;
		page-break-after: avoid;
	}

	p { margin: 0 0 10px; orphans: 3; widows: 3; }

	a { color: var(--amber); text-decoration: none; }

	ul, ol { margin: 0 0 10px; padding-left: 20px; }
	li { margin-bottom: 4px; }

	code {
		font-family: 'Courier New', monospace;
		font-size: 9pt;
		background: #f2f3f5;
		padding: 1px 4px;
		border-radius: 3px;
	}

	blockquote {
		margin: 12px 0;
		padding: 10px 14px;
		background: var(--amber-soft);
		border-left: 3px solid var(--amber);
		page-break-inside: avoid;
	}

	blockquote p:last-child { margin-bottom: 0; }

	table {
		width: 100%;
		border-collapse: collapse;
		margin: 12px 0 16px;
		font-size: 9.5pt;
		page-break-inside: avoid;
	}

	th, td {
		border: 1px solid var(--rule);
		padding: 6px 9px;
		text-align: left;
		vertical-align: top;
	}

	th { background: #f5f6f8; font-size: 8.5pt; text-transform: uppercase; letter-spacing: 0.06em; }

	/*
	 * Las capturas son altas: sin techo, cada una se lleva su hoja y deja media
	 * página en blanco detrás. Acotando el alto caben con el texto que explican.
	 */
	img {
		display: block;
		width: auto;
		height: auto;
		max-width: 100%;
		max-height: 132mm;
		margin: 14px auto;
		border: 1px solid var(--rule);
		border-radius: 5px;
		page-break-inside: avoid;
	}

	hr {
		border: 0;
		border-top: 1px solid var(--rule);
		margin: 22px 0;
		page-break-after: always;
	}

	/* La tabla de contenido no necesita viñetas: ya va numerada. */
	h2#tabla-de-contenido + ol { list-style: none; padding-left: 0; }
	h2#tabla-de-contenido + ol li { padding: 3px 0; border-bottom: 1px dotted var(--rule); }
</style>
</head>
<body>
	<section class="cover">
		<div class="cover-mark">
			<svg width="34" height="34" viewBox="0 0 30 30" fill="none">
				<rect width="30" height="30" rx="7" fill="#d4912a" />
				<path d="M7 22L12.5 14.5L16.5 18.5L23 9" stroke="#0c0a06" stroke-width="2.6"
					stroke-linecap="round" stroke-linejoin="round" />
			</svg>
			<span>FINEXIA</span>
		</div>
		<h1>Manual<br />de Usuario</h1>
		<p>
			Guía completa de Finexia: cómo registrar tus plataformas, construir tus portafolios,
			mantener tus transacciones al día y leer tus reportes.
		</p>
		<div class="cover-meta">
			<div><b>Versión</b>${version}</div>
			<div><b>Fecha</b>${date}</div>
			<div><b>Aplicación</b>Finexia</div>
		</div>
	</section>

	${content}
</body>
</html>`;
}
