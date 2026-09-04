/**
 * Pantalla de consentimiento OAuth.
 *
 * Es el único punto del flujo donde interviene una persona, y lo único que
 * aporta es precisamente eso: la identidad. Todo lo demás —qué cliente pregunta,
 * con qué ámbitos, a dónde vuelve— lo aparcó el backend en `/oauth/authorize` y
 * se lee de ahí; por el navegador solo viaja el id de la petición.
 *
 * Vive en el frontend y no en la API porque la sesión vive aquí: la API está en
 * otro origen y no ve la cookie, así que no puede saber quién está mirando la
 * pantalla.
 */
import type { Actions, PageServerLoad } from './$types';
import { error, redirect } from '@sveltejs/kit';
import * as user from '$lib/api/user';
import { setReturnTo } from '$lib/server/session';

export const load: PageServerLoad = async (event) => {
	const requestId = event.url.searchParams.get('request');

	if (!requestId) {
		error(400, 'Falta el identificador de la petición de autorización.');
	}

	// Llegar sin sesión es lo normal aquí: se entra desde un enlace que abre el
	// cliente MCP. Se deja anotado a dónde volver para que el login no pierda la
	// petición que se estaba autorizando.
	if (!event.locals.session || !event.locals.user) {
		setReturnTo(event.cookies, event.url.pathname + event.url.search);
		redirect(303, '/auth');
	}

	const result = await user.getOAuthConsent(event, requestId);

	if (!result.ok || !result.data) {
		error(
			result.status === 404 ? 404 : 502,
			result.status === 404
				? 'Esta petición de autorización ya no es válida. Vuelve a conectar desde la aplicación.'
				: 'No se pudo cargar la petición de autorización.'
		);
	}

	return { consent: result.data, requestId };
};

export const actions = {
	/**
	 * Aprueba o deniega. El destino lo calcula el backend a partir de la URI
	 * registrada del cliente, nunca de nada que venga del formulario: es lo que
	 * separa un flujo OAuth de un redirect abierto.
	 */
	decide: async (event) => {
		const requestId = event.url.searchParams.get('request');

		if (!requestId) {
			error(400, 'Falta el identificador de la petición de autorización.');
		}

		const formData = await event.request.formData();
		const approved = formData.get('decision') === 'approve';

		const result = await user.decideOAuthConsent(event, requestId, approved);

		if (!result.ok || !result.data?.redirectTo) {
			error(
				result.status === 404 ? 404 : 502,
				result.status === 404
					? 'Esta petición de autorización ya no es válida. Vuelve a conectar desde la aplicación.'
					: 'No se pudo completar la autorización.'
			);
		}

		redirect(303, result.data.redirectTo);
	}
} satisfies Actions;
