import { defineConfig } from '@playwright/test';

// Allow pointing the e2e run at an already-installed Chromium (e.g. a sandboxed
// CI image where `playwright install` can't reach the CDN). Mirrors the
// CHROMIUM_EXECUTABLE_PATH override used by the vitest browser tests.
const chromiumExecutablePath = process.env.CHROMIUM_EXECUTABLE_PATH;

export default defineConfig({
	webServer: { command: 'npm run build && npm run preview', port: 4173 },
	testMatch: '**/*.e2e.{ts,js}',
	use: chromiumExecutablePath ? { launchOptions: { executablePath: chromiumExecutablePath } } : {}
});
