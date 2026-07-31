#!/usr/bin/env node
/**
 * Comprueba los criterios de arquitectura que ESLint no puede expresar.
 *
 * Las reglas de dependencia entre capas viven en `eslint.config.js`; aquí van
 * las que hablan del *tamaño* de los archivos y de qué puede aparecer en
 * `routes/`, más los restos del legacy que la Fase 7 borró. Eran una foto en
 * `docs/FRONTEND_MIGRATION_BASELINE.md`; ejecutándose en CI dejan de poder
 * volver sin que nadie se entere.
 *
 * Uso: `pnpm check:arch` (o `node scripts/check-architecture.mjs`).
 */
import { readdirSync, readFileSync, statSync, existsSync } from 'node:fs';
import { join, relative } from 'node:path';

const ROOT = new URL('..', import.meta.url).pathname;
const SRC = join(ROOT, 'src');

/** Presupuestos de la sección 3 de docs/FRONTEND_ARCHITECTURE.md. */
const MAX_LINES = 500;
const MAX_PAGE_LINES = 300;

const problems = [];
const fail = (rule, detail) => problems.push({ rule, detail });

function walk(dir, out = []) {
	for (const entry of readdirSync(dir)) {
		const full = join(dir, entry);
		if (statSync(full).isDirectory()) walk(full, out);
		else out.push(full);
	}
	return out;
}

const files = walk(SRC);
const isSpec = (f) => /\.(spec|test)\.(ts|js)$/.test(f);
const isSource = (f) => /\.(svelte|ts|js)$/.test(f) && !isSpec(f);
const rel = (f) => relative(ROOT, f);
const lines = (f) => readFileSync(f, 'utf8').split('\n').length;

/**
 * Módulos que importa un archivo, sin comentarios: un `import` citado dentro de
 * un bloque de documentación no es una dependencia.
 */
function importsOf(file) {
	const code = readFileSync(file, 'utf8')
		.replace(/\/\*[\s\S]*?\*\//g, '')
		.replace(/(^|\s)\/\/.*$/gm, '$1');
	return [...code.matchAll(/(?:from|import)\s+['"]([^'"]+)['"]/g)].map((m) => m[1]);
}

// --- Presupuesto de tamaño --------------------------------------------------

for (const file of files.filter(isSource)) {
	const n = lines(file);
	if (n > MAX_LINES) fail('tamaño', `${rel(file)} tiene ${n} líneas (máximo ${MAX_LINES})`);
	if (file.endsWith('+page.svelte') && n > MAX_PAGE_LINES) {
		fail('tamaño de página', `${rel(file)} tiene ${n} líneas (máximo ${MAX_PAGE_LINES})`);
	}
}

// --- `routes/` solo orquesta ------------------------------------------------

const ROUTE_BANS = [
	[/^zod$/, 'declara un schema Zod; los schemas viven en features/<x>/schemas.ts'],
	[/lib\/api\/client$/, 'usa el cliente HTTP directo; llama a un módulo de dominio de lib/api'],
	[/lib\/server\/api$/, 'usa el cliente HTTP heredado; llama a un módulo de dominio de lib/api']
];

for (const file of files.filter((f) => f.startsWith(join(SRC, 'routes')) && isSource(f))) {
	for (const spec of importsOf(file)) {
		for (const [pattern, why] of ROUTE_BANS) {
			if (pattern.test(spec)) fail('routes/ solo orquesta', `${rel(file)} ${why}`);
		}
	}
	if (/\bauthedFetch\b/.test(readFileSync(file, 'utf8'))) {
		fail('routes/ solo orquesta', `${rel(file)} llama a authedFetch`);
	}
}

// --- Restos del legacy que borró la Fase 7 ----------------------------------

const LEGACY = [
	[/^\$components\//, 'el alias $components ya no existe'],
	[/^\$lib\/utils$/, 'lib/utils.ts se repartió en lib/shared'],
	[/^\$lib\/stores\//, 'lib/stores se movió a lib/shared']
];

for (const file of files.filter(isSource)) {
	for (const spec of importsOf(file)) {
		for (const [pattern, why] of LEGACY) {
			if (pattern.test(spec)) fail('legacy', `${rel(file)} importa ${spec}: ${why}`);
		}
	}
}

for (const dir of ['components', 'lib/stores']) {
	if (existsSync(join(SRC, dir))) fail('legacy', `src/${dir} debería haberse borrado en la Fase 7`);
}
if (existsSync(join(SRC, 'lib/utils.ts')))
	fail('legacy', 'src/lib/utils.ts debería haberse borrado');

// --- Anatomía de una feature ------------------------------------------------

const FEATURES = join(SRC, 'lib/features');
for (const feature of readdirSync(FEATURES)) {
	const dir = join(FEATURES, feature);
	if (!statSync(dir).isDirectory()) continue;

	if (!existsSync(join(dir, 'index.ts'))) {
		fail('feature sin superficie pública', `${feature} no tiene index.ts`);
	}
	for (const file of readdirSync(dir)) {
		if (file.endsWith('.svelte')) {
			fail('componente fuera de components/', `lib/features/${feature}/${file}`);
		}
	}
}

// --- Informe ----------------------------------------------------------------

if (problems.length === 0) {
	console.log('check:arch — sin incidencias');
	process.exit(0);
}

console.error(`check:arch — ${problems.length} incidencia(s):\n`);
for (const { rule, detail } of problems) console.error(`  [${rule}] ${detail}`);
console.error('\nVer docs/FRONTEND_ARCHITECTURE.md');
process.exit(1);
