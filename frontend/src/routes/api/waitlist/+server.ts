import { env } from '$env/dynamic/private';
import { json, redirect } from '@sveltejs/kit';
import z from 'zod';
import type { RequestHandler } from './$types';

const emailSchema = z.email();

// Waitlist signup lives in its own endpoint (instead of a page action) so the
// landing page can be prerendered as static HTML. The JS path (use:enhance in
// hero.svelte) sends `accept: application/json` and gets JSON back; a plain
// no-JS form submit is redirected back to the landing page.
export const POST: RequestHandler = async ({ request, fetch }) => {
	const wantsJson = (request.headers.get('accept') ?? '').includes('application/json');
	const formData = await request.formData();
	const parsed = await emailSchema.safeParseAsync(formData.get('email'));

	if (!parsed.success) {
		const message = 'Correo electrónico inválido';
		if (wantsJson) return json({ success: false, error: message }, { status: 400 });
		redirect(303, '/?waitlist=invalid');
	}

	const response = await fetch(`${env.BASE_API}/marketing/waitlists`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ email: parsed.data })
	});

	const result = await response.json().catch(() => ({ success: false, message: 'Error' }));

	if (!response.ok || !result.success) {
		if (wantsJson) {
			return json({ success: false, error: result.message ?? 'Error' }, { status: 400 });
		}
		redirect(303, '/?waitlist=error');
	}

	if (wantsJson) return json({ success: true, message: result.message });
	redirect(303, '/?waitlist=success');
};
