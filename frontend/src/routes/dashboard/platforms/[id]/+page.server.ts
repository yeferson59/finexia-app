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

		// Un 409 no es un fallo del servidor: es la plataforma diciendo que
		// todavía la apuntan posiciones, y el backend se niega en vez de
		// borrarlas. Devolvía «Error al eliminar la plataforma» para eso y para
		// una caída de red por igual, así que el usuario no tenía forma de saber
		// que le bastaba con quitar las posiciones primero.
		//
		// El motivo se escribe aquí y no se copia del `details` del backend: ese
		// texto está en inglés y nombra tablas, y esta es la pantalla donde hay
		// que decir qué hacer a continuación.
		if (res.status === 409) {
			return {
				success: false,
				error:
					'No se puede eliminar: la plataforma todavía tiene posiciones registradas, ' +
					'incluidas las que ya vendiste. Elimina esas posiciones primero.'
			};
		}

		if (!res.ok) return { success: false, error: 'Error al eliminar la plataforma' };

		redirect(303, '/dashboard/platforms');
	}
};
