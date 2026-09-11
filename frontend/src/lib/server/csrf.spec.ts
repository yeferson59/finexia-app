import { describe, it, expect } from 'vitest';
import { crossSiteFormResponse, isCrossSiteFormSubmission } from './csrf';

const APP = new URL('https://finexia.example/auth?/login');
const FORM = 'application/x-www-form-urlencoded';

function submit(headers: Record<string, string>, method = 'POST') {
	return new Request(APP, { method, headers, body: 'email=a' });
}

describe('isCrossSiteFormSubmission', () => {
	it('rejects a form posted from another origin', () => {
		const request = submit({ 'content-type': FORM, origin: 'https://evil.example' });

		expect(isCrossSiteFormSubmission(request, APP)).toBe(true);
	});

	it('rejects a form that carries no Origin at all, as SvelteKit does', () => {
		expect(isCrossSiteFormSubmission(submit({ 'content-type': FORM }), APP)).toBe(true);
	});

	it('accepts a form posted from the app itself', () => {
		const request = submit({ 'content-type': FORM, origin: 'https://finexia.example' });

		expect(isCrossSiteFormSubmission(request, APP)).toBe(false);
	});

	it.each([
		'multipart/form-data; boundary=----x',
		'text/plain;charset=UTF-8',
		'Application/X-WWW-Form-Urlencoded',
		'application/x-sveltekit-formdata'
	])('treats %s as a form', (contentType) => {
		const request = submit({ 'content-type': contentType, origin: 'https://evil.example' });

		expect(isCrossSiteFormSubmission(request, APP)).toBe(true);
	});

	it.each(['PUT', 'PATCH', 'DELETE'])('applies to %s as well as POST', (method) => {
		const request = submit({ 'content-type': FORM, origin: 'https://evil.example' }, method);

		expect(isCrossSiteFormSubmission(request, APP)).toBe(true);
	});

	it('leaves JSON alone: a cross-site page cannot send it without a CORS preflight', () => {
		const request = submit({ 'content-type': 'application/json', origin: 'https://evil.example' });

		expect(isCrossSiteFormSubmission(request, APP)).toBe(false);
	});

	it('leaves safe methods alone', () => {
		const request = new Request(APP, {
			headers: { 'content-type': FORM, origin: 'https://evil.example' }
		});

		expect(isCrossSiteFormSubmission(request, APP)).toBe(false);
	});
});

describe('crossSiteFormResponse', () => {
	it('answers 403 in the same words as SvelteKit', async () => {
		const res = crossSiteFormResponse(submit({}));

		expect(res.status).toBe(403);
		expect(await res.text()).toBe('Cross-site POST form submissions are forbidden');
	});

	it('answers in JSON to a client that asks for it', async () => {
		const res = crossSiteFormResponse(submit({ accept: 'application/json' }, 'DELETE'));

		expect(res.status).toBe(403);
		expect(await res.json()).toEqual({
			message: 'Cross-site DELETE form submissions are forbidden'
		});
	});
});
