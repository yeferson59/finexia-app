import { describe, it, expect } from 'vitest';
import { errorActions, errorCopy } from './http-error';

describe('errorCopy', () => {
	it('no confunde un error de la petición con uno del servidor', () => {
		expect(errorCopy(400).title).toBe('No pudimos abrir esta página');
		expect(errorCopy(403).title).toBe('No tienes acceso a esta página');
		expect(errorCopy(502).title).toBe('Algo falló de nuestro lado');
	});

	it('enseña el mensaje que escribió quien lanzó el error, con su punto final', () => {
		expect(errorCopy(404, 'Portafolio no encontrado').detail).toBe('Portafolio no encontrado.');
		expect(
			errorCopy(404, 'Esta petición de autorización ya no es válida. Vuelve a conectar.').detail
		).toBe('Esta petición de autorización ya no es válida. Vuelve a conectar.');
	});

	it('calla los mensajes en inglés que pone SvelteKit por defecto', () => {
		expect(errorCopy(404, 'Not Found').detail).toBeUndefined();
		expect(errorCopy(500, 'Internal Error').detail).toBeUndefined();
		expect(errorCopy(500, '   ').detail).toBeUndefined();
	});

	it('ofrece soporte solo cuando la persona no puede arreglarlo sola', () => {
		expect(errorCopy(404).contact).toBe(false);
		expect(errorCopy(403).contact).toBe(true);
		expect(errorCopy(503).contact).toBe(true);
	});
});

describe('errorActions', () => {
	it('fuera del panel vuelve a la portada', () => {
		expect(errorActions(404, '/no-existe')).toEqual([{ label: 'Volver al inicio', route: '/' }]);
	});

	it('dentro del panel vuelve al panel', () => {
		expect(errorActions(404, '/dashboard/portfolios/abc')).toEqual([
			{ label: 'Ir al panel', route: '/dashboard' }
		]);
		expect(errorActions(404, '/dashboardx')).toEqual([{ label: 'Volver al inicio', route: '/' }]);
	});

	it('ante un fallo del servidor propone reintentar primero', () => {
		expect(errorActions(500, '/dashboard')).toEqual([
			{ label: 'Volver a intentarlo', reload: true },
			{ label: 'Ir al panel', route: '/dashboard' }
		]);
	});

	it('ante un 401 manda a iniciar sesión', () => {
		expect(errorActions(401, '/dashboard/settings')[0]).toEqual({
			label: 'Iniciar sesión',
			route: '/auth'
		});
	});
});
