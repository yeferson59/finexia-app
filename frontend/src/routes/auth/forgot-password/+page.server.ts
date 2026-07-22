import z from 'zod';
import type { Actions } from './$types';
import * as auth from '$lib/api/auth';
import { fail } from '@sveltejs/kit';

export const actions = {
	default: async ({ request, fetch }) => {
		const formData = await request.formData();

		const parsed = await z
			.object({ email: z.email().min(2) })
			.safeParseAsync({ email: formData.get('email') });

		if (!parsed.success) {
			return fail(400, { errors: { email: 'Ingresa un email válido' } });
		}

		// The backend response never reveals whether the email exists, so the
		// UI reports the same generic success regardless of the outcome.
		await auth.requestPasswordReset(fetch, parsed.data.email);

		return { sent: true as const };
	}
} satisfies Actions;
