import z from 'zod';
import type { Actions } from './$types';
import { env } from '$env/dynamic/private';
import { redirect } from '@sveltejs/kit';

export const actions = {
	default: async ({ request, cookies, fetch }) => {
		const formData = await request.formData();

		const { success, error, data } = await z
			.object({
				name: z.coerce.string().min(2),
				description: z.coerce.string().optional(),
				type: z.coerce.string().min(2)
			})
			.safeParseAsync({
				name: formData.get('name'),
				description: formData.get('description'),
				type: formData.get('type')
			});

		if (!success) {
			return { error: error.message };
		}

		const accessToken = cookies.get('access_token_finexia');

		if (!accessToken) {
			return { error: 'No access token found' };
		}

		const res = await fetch(`${env.BASE_API}/portfolios/sources`, {
			method: 'POST',
			headers: {
				Authorization: `Bearer ${accessToken}`,
				'Content-Type': 'application/json'
			},
			body: JSON.stringify(data)
		});

		if (!res.ok) {
			return { error: 'Failed to add platform' };
		}

		const { success: successResponse } = await res.json();

		if (!successResponse) {
			return { error: 'Failed to add platform' };
		}

		redirect(303, '/dashboard/platforms');
	}
} satisfies Actions;
