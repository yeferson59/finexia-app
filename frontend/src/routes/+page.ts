// The landing page is static marketing content — prerender it to static HTML so
// it is served without an SSR round-trip. Waitlist signup runs through the
// /api/waitlist endpoint (see hero.svelte).
export const prerender = true;
