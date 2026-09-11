import adapter from '@sveltejs/adapter-node';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	compilerOptions: {
		// Force runes mode for the project, except for libraries. Can be removed in svelte 6.
		runes: ({ filename }) => (filename.split(/[/\\]/).includes('node_modules') ? undefined : true)
	},
	preprocess: vitePreprocess(),
	kit: {
		// adapter-auto only supports some environments, see https://svelte.dev/docs/kit/adapter-auto for a list.
		// If your environment is not supported, or you settled on a specific environment, switch out the adapter.
		// See https://svelte.dev/docs/kit/adapters for more information about adapters.
		// Sin aliases propios: todo el código compartido vive bajo `$lib`, el
		// alias estándar de SvelteKit (ver docs/FRONTEND_ARCHITECTURE.md).
		adapter: adapter({ bodySize: 5 * 1024 * 1024 }),
		// El origen de los formularios lo comprueba `hooks.server.ts`
		// (`$lib/server/csrf`), no SvelteKit: su check vale para todas las rutas o
		// para ninguna, y `/oauth/token`, que la app reenvía al backend
		// (`$lib/api/proxy`), recibe por definición formularios de otros
		// servidores. `'*'` es la forma no deprecada de `checkOrigin: false`.
		csrf: { trustedOrigins: ['*'] }
	}
};

export default config;
