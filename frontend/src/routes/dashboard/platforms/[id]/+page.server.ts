import { error, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import * as platforms from '$lib/api/platforms';
import { platformUpdateSchema } from '$lib/features/platforms';

export const load: PageServerLoad = async ({ params, parent }) => {
	const { platforms } = await parent();

	const platform = platforms.find((p) => p.id === params.id);

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
		} = await platformUpdateSchema.safeParseAsync({
			name: formData.get('name'),
			description: formData.get('description'),
			type: formData.get('type'),
			isActive: formData.get('isActive')
		});

		if (!success) {
			return { success: false, error: zodError.message };
		}

		const res = await platforms.updateSource({ cookies, fetch }, params.id, data);

		if (!res.ok) return { success: false, error: 'Error al actualizar la plataforma' };

		return { success: res.success };
	},

	delete: async ({ cookies, fetch, params }) => {
		const res = await platforms.deleteSource({ cookies, fetch }, params.id);

		if (!res.ok) return { success: false, error: 'Error al eliminar la plataforma' };

		redirect(303, '/dashboard/platforms');
	}
};
