import { describe, it, expect } from 'vitest';
import { actionData, actionError, actionSucceeded } from './form';

describe('reparto del `form` entre secciones', () => {
	it('solo reconoce el resultado de su propia acción', () => {
		const form = { action: 'updateProfile', success: true };
		expect(actionSucceeded(form, 'updateProfile')).toBe(true);
		expect(actionSucceeded(form, 'changePassword')).toBe(false);
	});

	it('no da por buena una acción sin `success`', () => {
		expect(actionSucceeded({ action: 'updateProfile' }, 'updateProfile')).toBe(false);
		expect(actionSucceeded(null, 'updateProfile')).toBe(false);
	});

	it('devuelve el error de su acción y nada de las demás', () => {
		const form = { action: 'changePassword', error: 'Contraseña actual incorrecta' };
		expect(actionError(form, 'changePassword')).toBe('Contraseña actual incorrecta');
		expect(actionError(form, 'updateProfile')).toBe('');
	});

	it('lee los datos que devuelve una acción con éxito', () => {
		const form = { action: 'uploadAvatar', success: true, imageUrl: '/avatars/1.webp' };
		expect(actionData<string>(form, 'uploadAvatar', 'imageUrl')).toBe('/avatars/1.webp');
		// Un fallo de la misma acción no expone sus campos.
		expect(
			actionData<string>({ action: 'uploadAvatar', error: 'x' }, 'uploadAvatar', 'imageUrl')
		).toBeUndefined();
	});
});
