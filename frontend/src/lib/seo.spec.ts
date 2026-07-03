import { describe, it, expect } from 'vitest';
import { absoluteUrl, SITE_URL } from './seo';

describe('absoluteUrl', () => {
	it('defaults to the site root when no path is given', () => {
		expect(absoluteUrl()).toBe(`${SITE_URL}/`);
	});

	it('resolves a relative path against the site URL', () => {
		expect(absoluteUrl('/auth')).toBe(`${SITE_URL}/auth`);
	});

	it('dedupes a leading slash instead of doubling it', () => {
		expect(absoluteUrl('/dashboard/settings')).toBe(`${SITE_URL}/dashboard/settings`);
	});

	it('preserves query strings and fragments', () => {
		expect(absoluteUrl('/sitemap.xml?page=1')).toBe(`${SITE_URL}/sitemap.xml?page=1`);
	});

	it('returns an absolute URL unchanged when it is already on the site host', () => {
		const url = `${SITE_URL}/og-image.png`;
		expect(absoluteUrl(url)).toBe(url);
	});
});
