import z from 'zod';
import { error, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { authedFetch } from '$lib/server/api';

export const load: PageServerLoad = async ({ params, parent }) => {
	const { platforms } = await parent();

	const platform = (platforms as Array<{ id: string }>).find((p) => p.id === params.id);

	if (!platform) {
		error(404, 'Plataforma no encontrada');
	}

	return { platform };
};

export const actions: Actions = {
	update: async ({ request, cookies, fetch, params }) => {
		const formData = await request.formData();

		const {
			success,
			error: zodError,
			data
		} = await z
			.object({
				name: z.string().min(2),
				description: z.string().optional().default(''),
				type: z.string().min(2),
				// El <select> envía "true"/"false" como string; z.coerce.boolean()
				// convertiría "false" en true, así que se compara explícitamente.
				isActive: z.enum(['true', 'false']).transform((v) => v === 'true')
			})
			.safeParseAsync({
				name: formData.get('name'),
				description: formData.get('description'),
				type: formData.get('type'),
				isActive: formData.get('isActive')
			});

		if (!success) {
			return { success: false, error: zodError.message };
		}

		const res = await authedFetch({ cookies, fetch }, `/portfolios/sources/${params.id}`, {
			method: 'PATCH',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(data)
		});

		if (!res.ok) return { success: false, error: 'Error al actualizar la plataforma' };

		const json = await res.json();
		return { success: json.success ?? false };
	},

	delete: async ({ cookies, fetch, params }) => {
		const res = await authedFetch({ cookies, fetch }, `/portfolios/sources/${params.id}`, {
			method: 'DELETE'
		});

		if (!res.ok) return { success: false, error: 'Error al eliminar la plataforma' };

		redirect(303, '/dashboard/platforms');
	}
};
