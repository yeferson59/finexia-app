#!/usr/bin/env node
/**
 * Regenera las capturas del manual de usuario (`docs/img/manual/`).
 *
 * Uso: `pnpm manual:shots` (desde `frontend/`), y después `pnpm manual:build`.
 *
 * Las capturas se sacan de la aplicación real compilada, hablando con el stub
 * de la API que ya usan los e2e (`e2e/mocks/mock-api.mjs`): así el manual
 * enseña la interfaz que hay hoy en el repositorio y no la de hace tres meses.
 * Sin esto, cualquier cambio en el menú lateral deja las dieciséis capturas
 * desfasadas a la vez y nadie se entera hasta que un usuario lo nota.
 */
import { spawn } from 'node:child_process';
import { mkdirSync } from 'node:fs';
import { join } from 'node:path';
import { chromium } from 'playwright';
import { IMAGES, ROOT } from './manual.mjs';

const FRONTEND = join(ROOT, 'frontend');
const MOCK_PORT = 4188;
const APP_PORT = 4189;
const BASE = `http://127.0.0.1:${APP_PORT}`;

/** Fixtures del stub, iguales que en `e2e/helpers.ts`. */
const USER = { email: 'user@finexia.test', password: 'Password123!' };
const PORTFOLIO_ID = '11111111-1111-4111-8111-111111111111';
const PLATFORM_ID = '33333333-3333-4333-8333-333333333333';

const DESKTOP = { width: 1440, height: 900, deviceScaleFactor: 2 };
const MOBILE = { width: 390, height: 844, deviceScaleFactor: 3 };

const children = [];

/*
 * `detached: true` pone cada servidor en su propio grupo de procesos. `pnpm
 * preview` arranca a su vez un `vite preview`, y matando solo al padre el nieto
 * sobrevivía: la siguiente ejecución encontraba el puerto ocupado y se quedaba
 * esperando a un servidor que nunca era el suyo.
 */
function run(command, args, env = {}) {
	const child = spawn(command, args, {
		cwd: FRONTEND,
		env: { ...process.env, ...env },
		stdio: ['ignore', 'pipe', 'pipe'],
		detached: true
	});
	children.push(child);
	return child;
}

async function waitFor(url, timeoutMs = 120_000) {
	const deadline = Date.now() + timeoutMs;
	while (Date.now() < deadline) {
		try {
			const res = await fetch(url);
			if (res.status < 500) return;
		} catch {
			// El servidor todavía no escucha; se reintenta.
		}
		await new Promise((resolve) => setTimeout(resolve, 500));
	}
	throw new Error(`El servidor no respondió a tiempo: ${url}`);
}

let stopped = false;
function shutdown() {
	if (stopped) return;
	stopped = true;
	for (const child of children) {
		try {
			process.kill(-child.pid, 'SIGTERM');
		} catch {
			// El proceso ya había terminado por su cuenta.
		}
	}
}

process.on('exit', shutdown);
process.on('SIGINT', () => {
	shutdown();
	process.exit(130);
});

// --- Servidores -------------------------------------------------------------

for (const port of [MOCK_PORT, APP_PORT]) {
	const free = await fetch(`http://127.0.0.1:${port}/`)
		.then(() => false)
		.catch(() => true);
	if (!free) {
		console.error(
			`El puerto ${port} está ocupado. Cierra el proceso que lo usa y vuelve a intentarlo.`
		);
		process.exit(1);
	}
}

console.log('· levantando el stub de la API…');
run('node', ['e2e/mocks/mock-api.mjs'], { MOCK_API_PORT: String(MOCK_PORT) });

console.log('· compilando la aplicación…');
await new Promise((resolve, reject) => {
	const build = run('pnpm', ['build']);
	build.on('exit', (code) =>
		code === 0 ? resolve() : reject(new Error(`build salió con ${code}`))
	);
});

console.log('· arrancando el servidor de vista previa…');
// `--host 127.0.0.1` explícito: por defecto `vite preview` escucha solo en
// `::1`, y el sondeo de abajo —y el `fetch` del SSR contra el stub— van por
// IPv4. Sin esto el script espera dos minutos a un servidor que sí arrancó.
run('pnpm', ['preview', '--port', String(APP_PORT), '--host', '127.0.0.1'], {
	BASE_API: `http://127.0.0.1:${MOCK_PORT}/api/v1`
});
await waitFor(`${BASE}/`);

// --- Capturas ---------------------------------------------------------------

mkdirSync(IMAGES, { recursive: true });

const browser = await chromium.launch({
	...(process.env.CHROMIUM_EXECUTABLE_PATH
		? { executablePath: process.env.CHROMIUM_EXECUTABLE_PATH }
		: {})
});

/** Cierra el aviso de cookies y espera a que la página se asiente. */
async function settle(page) {
	await page
		.getByRole('button', { name: 'Entendido' })
		.click({ timeout: 1500 })
		.catch(() => {});
	// Las animaciones de entrada del dashboard duran medio segundo.
	await page.waitForTimeout(900);
}

async function shot(page, name, { fullPage = false } = {}) {
	await page.screenshot({ path: join(IMAGES, `${name}.png`), fullPage });
	console.log(`  ✓ ${name}.png`);
}

async function login(page) {
	await page.goto(`${BASE}/auth`);
	await page.fill('#login-email', USER.email);
	await page.fill('#login-password', USER.password);
	await page.getByRole('button', { name: 'Iniciar sesión', exact: true }).click();
	await page.waitForURL('**/dashboard');
	await settle(page);
}

try {
	// --- Público ---
	const anon = await browser.newPage({ viewport: DESKTOP });
	await anon.goto(`${BASE}/`);
	await settle(anon);
	await shot(anon, '01-landing');

	await anon.goto(`${BASE}/auth`);
	await settle(anon);
	await shot(anon, '02-login');
	await anon.close();

	// --- Aplicación ---
	const page = await browser.newPage({ viewport: DESKTOP });
	await login(page);
	await shot(page, '03-dashboard', { fullPage: true });

	const pages = [
		['04-portafolios', '/dashboard/portfolios', false],
		['05-portafolio-detalle', `/dashboard/portfolios/${PORTFOLIO_ID}`, true],
		['06-crear-portafolio', '/dashboard/portfolios/add', true],
		['07-plataformas', '/dashboard/platforms', false],
		['08-transacciones', '/dashboard/transactions', false],
		['09-importar-transacciones', '/dashboard/transactions/import', false],
		['10-reportes', '/dashboard/reports', true],
		['11-notificaciones', '/dashboard/notifications', false],
		['12-configuracion', '/dashboard/settings', true],
		['13-activo-detalle', `/dashboard/portfolios/${PORTFOLIO_ID}/assets/AAPL`, true],
		['17-guia', '/dashboard/guia', false],
		['18-mis-activos', '/dashboard/assets', true],
		['19-plataforma-detalle', `/dashboard/platforms/${PLATFORM_ID}`, false]
	];

	for (const [name, path, fullPage] of pages) {
		await page.goto(`${BASE}${path}`);
		await settle(page);
		await shot(page, name, { fullPage });
	}

	// Paso 2 del import: hay que subir un archivo para que exista la pantalla.
	await page.goto(`${BASE}/dashboard/transactions/import`);
	await settle(page);
	await page.setInputFiles(
		'input[type="file"]',
		{
			name: 'movimientos.csv',
			mimeType: 'text/csv',
			buffer: Buffer.from(
				'Fecha,Ticker,Nombre,Tipo,Cantidad,Precio,Comisiones,Moneda\n' +
					'2026-03-12,AAPL,Apple Inc.,buy,10,182.40,1.20,USD\n' +
					'2026-03-08,MSFT,Microsoft,buy,4,410.00,1.20,USD\n' +
					'2026-02-27,TSLA,Tesla,sell,3,242.10,1.20,USD\n'
			)
		},
		{ timeout: 5000 }
	);
	await page.waitForTimeout(2500);
	await shot(page, '14-importar-columnas', { fullPage: true });
	await page.close();

	// --- Móvil ---
	const mobile = await browser.newPage({ viewport: MOBILE });
	await login(mobile);
	await shot(mobile, '15-movil-dashboard');

	await mobile
		.getByRole('button', { name: /men[úu]/i })
		.first()
		.click();
	await mobile.waitForTimeout(600);
	await shot(mobile, '16-movil-menu');
	await mobile.close();
} finally {
	await browser.close();
	shutdown();
}

console.log('\nCapturas regeneradas. Ejecuta `pnpm manual:build` para rehacer el PDF.');
process.exit(0);
