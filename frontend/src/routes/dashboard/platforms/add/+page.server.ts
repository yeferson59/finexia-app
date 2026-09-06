import type { Actions } from './$types';
import { fail, redirect } from '@sveltejs/kit';
import * as platforms from '$lib/api/platforms';
import { platformCreateSchema } from '$lib/features/platforms';

export const actions = {
	default: async ({ request, cookies, fetch }) => {
		const formData = await request.formData();

		const { success, error, data } = await platformCreateSchema.safeParseAsync({
			name: formData.get('name'),
			description: formData.get('description'),
			type: formData.get('type')
		});

		/*
		 * `fail`, no un objeto suelto: devolver `{ error }` responde 200, así que
		 * el navegador da el alta por buena. Y el mensaje es el del schema, no el
		 * `error.message` de Zod, que es el JSON de las incidencias en crudo.
		 */
		if (!success) {
			return fail(400, { error: error.issues[0].message });
		}

		const res = await platforms.createSource({ cookies, fetch }, data);

		if (!res.ok || !res.success) {
			return fail(res.status >= 400 ? res.status : 400, {
				error: 'No pudimos guardar la plataforma. Vuelve a intentarlo en un momento.'
			});
		}

		redirect(303, '/dashboard/platforms');
	}
} satisfies Actions;
