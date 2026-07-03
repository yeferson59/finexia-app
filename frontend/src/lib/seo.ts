// Central SEO configuration. Keep canonical URL, social metadata and
// structured-data defaults in one place so pages and the sitemap stay in sync.

export const SITE_URL = 'https://finexia.me';
export const SITE_NAME = 'Finexia';

// Contact channel for exercising data-subject rights under Ley 1581 de 2012
// (Habeas Data). Update this if the responsible party's email changes.
export const CONTACT_EMAIL = 'soporte@finexia.me';

export const DEFAULT_TITLE = 'Finexia — Tu patrimonio en un solo mapa';
export const DEFAULT_DESCRIPTION =
	'Registra manualmente dónde tienes tus activos y agrúpalos en los portafolios que tú imaginas, aunque estén en distintas plataformas. Sin conectar cuentas. Lanzamiento 1 oct 2026.';

// 1200×630 social share image served from /static.
export const OG_IMAGE = `${SITE_URL}/og-image.png`;

export const LOCALE = 'es_ES';

/** Build an absolute URL for a path, deduping slashes. */
export function absoluteUrl(path = '/'): string {
	return new URL(path, SITE_URL).href;
}
