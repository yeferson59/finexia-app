import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vitest/config';
import { playwright } from '@vitest/browser-playwright';
import { sveltekit } from '@sveltejs/kit/vite';

// Allow pointing the browser tests at an already-installed Chromium (e.g. a
// sandboxed CI image where `playwright install` can't reach the CDN). Set
// CHROMIUM_EXECUTABLE_PATH to the binary; otherwise Playwright resolves its
// own managed download as usual.
const chromiumExecutablePath = process.env.CHROMIUM_EXECUTABLE_PATH;

export default defineConfig({
	plugins: [sveltekit(), tailwindcss()],
	build: {
		cssCodeSplit: false
	},
	test: {
		expect: { requireAssertions: true },
		projects: [
			{
				extends: './vite.config.ts',
				test: {
					name: 'client',
					// Los `define` de SvelteKit llegan al navegador sin evaluar; ver
					// el comentario de `vitest-setup-client.ts`.
					setupFiles: ['./vitest-setup-client.ts'],
					browser: {
						enabled: true,
						provider: playwright(
							chromiumExecutablePath
								? { launchOptions: { executablePath: chromiumExecutablePath } }
								: {}
						),
						instances: [{ browser: 'chromium', headless: true }],
						// Vitest 5 pasó `exact` a `true`. Las aserciones de este repo
						// buscan fragmentos —una frase dentro de un párrafo, el símbolo
						// dentro de una celda que también lleva el nombre del activo—,
						// que es lo que describe el comportamiento y no la maquetación.
						locators: { exact: false }
					},
					include: ['src/**/*.svelte.{test,spec}.{js,ts}'],
					exclude: ['src/lib/server/**']
				}
			},

			{
				extends: './vite.config.ts',
				test: {
					name: 'server',
					environment: 'node',
					// `e2e/**` entra por el spec que valida las fixtures del stub
					// contra los schemas de la API; los `.e2e.ts` son de Playwright.
					include: ['src/**/*.{test,spec}.{js,ts}', 'e2e/**/*.{test,spec}.{js,ts}'],
					exclude: ['src/**/*.svelte.{test,spec}.{js,ts}']
				}
			}
		]
	}
});
