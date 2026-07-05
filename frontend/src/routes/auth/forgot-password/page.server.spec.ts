import { describe, it, expect, vi } from 'vitest';
import { actions } from './+page.server';
import { jsonResponse } from '$lib/server/testing';

type RequestEvent = Parameters<typeof actions.default>[0];

function buildEvent(fields: Record<string, string>, fetch: ReturnType<typeof vi.fn>) {
	const formData = new FormData();
	for (const [key, value] of Object.entries(fields)) formData.append(key, value);

	return {
		request: { formData: async () => formData },
		fetch
	} as unknown;
}

describe('forgot-password action', () => {
	it('fails validation for a bad email without hitting the backend', async () => {
		const fetch = vi.fn();
		const result = (await actions.default(
			buildEvent({ email: 'nope' }, fetch) as RequestEvent
		)) as {
			status: number;
			data: { errors: Record<string, string> };
		};

		expect(result?.status).toBe(400);
		expect(result?.data.errors).toEqual({ email: 'Ingresa un email válido' });
		expect(fetch).not.toHaveBeenCalled();
	});

	it('reports success after calling the backend for a valid email', async () => {
		const fetch = vi.fn().mockResolvedValue(jsonResponse({ success: true }));

		const result = await actions.default(
			buildEvent({ email: 'user@finexia.me' }, fetch) as RequestEvent
		);

		expect(fetch).toHaveBeenCalledTimes(1);
		expect(result).toEqual({ sent: true });
	});

	it('reports success even when the backend errors, to avoid leaking whether the email exists', async () => {
		const fetch = vi.fn().mockResolvedValue(jsonResponse({ success: false }, { status: 500 }));

		const result = await actions.default(
			buildEvent({ email: 'user@finexia.me' }, fetch) as RequestEvent
		);

		expect(result).toEqual({ sent: true });
	});
});
