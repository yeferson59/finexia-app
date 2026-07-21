import { defineConfig } from '@playwright/test';

// Allow pointing the e2e run at an already-installed Chromium (e.g. a sandboxed
// CI image where `playwright install` can't reach the CDN). Mirrors the
// CHROMIUM_EXECUTABLE_PATH override used by the vitest browser tests.
const chromiumExecutablePath = process.env.CHROMIUM_EXECUTABLE_PATH;

// The smoke tests run the real SvelteKit server against a stub backend
// (e2e/mocks/mock-api.mjs) that serves the docs/API.md contract with fixed
// fixtures, so authenticated flows can be exercised without a Go backend.
const MOCK_API_PORT = 4174;
const APP_PORT = 4173;

export default defineConfig({
	webServer: [
		{
			command: 'node e2e/mocks/mock-api.mjs',
			port: MOCK_API_PORT,
			env: { MOCK_API_PORT: String(MOCK_API_PORT) }
		},
		{
			command: 'npm run build && npm run preview',
			port: APP_PORT,
			timeout: 180_000,
			env: { BASE_API: `http://127.0.0.1:${MOCK_API_PORT}/api/v1` }
		}
	],
	testMatch: '**/*.e2e.{ts,js}',
	use: {
		baseURL: `http://localhost:${APP_PORT}`,
		...(chromiumExecutablePath ? { launchOptions: { executablePath: chromiumExecutablePath } } : {})
	}
});
