/**
 * Qué dice la página de error (`routes/+error.svelte`) según el código y la
 * ruta en la que se produjo.
 *
 * Solo sabía distinguir el 404 de «todo lo demás», y todo lo demás era «error
 * interno del servidor»: un 400 de la autorización OAuth o un 403 salían como
 * si se hubiera caído el backend, y el mensaje que cada `error()` escribe para
 * la persona —«Portafolio no encontrado», «Esta petición de autorización ya no
 * es válida…»— no llegaba nunca a la pantalla.
 */

/** Rutas a las que puede mandar la página de error. */
export type ErrorRoute = '/' | '/auth' | '/dashboard';

export type ErrorAction = { label: string; route: ErrorRoute } | { label: string; reload: true };

export interface ErrorCopy {
	title: string;
	/** Qué ha pasado y qué hacer, en general, para ese código. */
	hint: string;
	/** El mensaje propio del `error()` que se lanzó, si lo escribió alguien. */
	detail?: string;
	/** Si merece la pena ofrecer el correo de soporte. */
	contact: boolean;
}

/*
 * Los mensajes que pone SvelteKit cuando nadie escribió uno: el 404 de una ruta
 * que no existe y el 500 de una excepción sin capturar. Están en inglés y no
 * explican nada, así que no se enseñan.
 */
const FRAMEWORK_MESSAGES = new Set([
	'Not Found',
	'Internal Error',
	'Bad Request',
	'Unauthorized',
	'Forbidden',
	'Method Not Allowed',
	'Service Unavailable'
]);

function ownMessage(message: string | undefined): string | undefined {
	const text = message?.trim();
	if (!text || FRAMEWORK_MESSAGES.has(text)) return undefined;
	return /[.!?…]$/.test(text) ? text : `${text}.`;
}

export function errorCopy(status: number, message?: string): ErrorCopy {
	const detail = ownMessage(message);

	if (status === 404) {
		return {
			title: 'No encontramos esta página',
			hint: 'Puede que el enlace esté mal escrito o que lo que buscas ya no exista.',
			detail,
			contact: false
		};
	}

	if (status === 401) {
		return {
			title: 'Tienes que iniciar sesión',
			hint: 'Esta página es de tu cuenta. Entra y vuelve a abrirla.',
			detail,
			contact: false
		};
	}

	if (status === 403) {
		return {
			title: 'No tienes acceso a esta página',
			hint: 'Tu cuenta no tiene permiso para verla. Si crees que debería tenerlo, escríbenos.',
			detail,
			contact: true
		};
	}

	if (status >= 500) {
		return {
			title: 'Algo falló de nuestro lado',
			hint: 'No es nada que hayas hecho. Vuelve a intentarlo en unos minutos.',
			detail,
			contact: true
		};
	}

	return {
		title: 'No pudimos abrir esta página',
		hint: 'El enlace está incompleto o ya no es válido. Vuelve a abrirlo desde donde lo encontraste.',
		detail,
		contact: false
	};
}

/**
 * La salida principal y, si la hay, una segunda. Dentro del panel se vuelve al
 * panel, no a la portada: quien estaba trabajando no quiere empezar de cero.
 */
export function errorActions(status: number, pathname: string): ErrorAction[] {
	const inDashboard = pathname === '/dashboard' || pathname.startsWith('/dashboard/');
	const home: ErrorAction = inDashboard
		? { label: 'Ir al panel', route: '/dashboard' }
		: { label: 'Volver al inicio', route: '/' };

	if (status >= 500) {
		return [{ label: 'Volver a intentarlo', reload: true }, home];
	}

	if (status === 401) {
		return [
			{ label: 'Iniciar sesión', route: '/auth' },
			{ label: 'Volver al inicio', route: '/' }
		];
	}

	return [home];
}
