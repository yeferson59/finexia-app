/*
 * Devuelve a los `define` de SvelteKit su valor real dentro del navegador.
 *
 * Vite 8 no sustituye los `define` en el entorno cliente: los inyecta como
 * globales cuando arranca la página. Vitest hace esa inyección por su cuenta y
 * desde la 5.0 pasa la cadena de código tal cual está en la config de Vite en
 * vez del valor ya evaluado, así que `__SVELTEKIT_PATHS_BASE__` llega como la
 * cadena `'""'` —dos comillas— en lugar de `''`, y `__SVELTEKIT_HASH_ROUTING__`
 * como `'false'`, que es truthy. Con eso `resolve('/x')` de `$app/paths`
 * devolvía `""#/x` y cualquier `href` de la app quedaba roto en los tests.
 *
 * Evaluamos aquí esas cadenas, que es lo que el navegador ve en `vite dev`.
 */
const globals = globalThis as unknown as Record<string, unknown>;

for (const key of Object.keys(globals)) {
	if (!key.startsWith('__SVELTEKIT_')) continue;
	const value = globals[key];
	if (typeof value !== 'string') continue;
	try {
		globals[key] = JSON.parse(value);
	} catch {
		// `__SVELTEKIT_PAYLOAD__` es una referencia a código, no un literal JSON.
	}
}
