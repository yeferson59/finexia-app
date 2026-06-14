import type { Actions } from './$types';
import { env } from '$env/dynamic/private';
import z from 'zod';

export const actions = {
	default: async ({ fetch, request }) => {
		const formData = await request.formData();
		const email = await z.email().parseAsync(formData.get('email'));

		const response = await fetch(`${env.BASE_API}/marketing/waitlists`, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({ email })
		});

		if (!response.ok) {
			const data = await response.json();
			return { success: false, error: data.message };
		}

		const { success, message } = await response.json();

		if (!success) {
			return { success: false, error: message };
		}

		return { success: true, message };
	}
} satisfies Actions;
