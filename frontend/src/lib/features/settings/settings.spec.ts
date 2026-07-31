import { describe, it, expect } from 'vitest';
import {
	actionData,
	actionError,
	actionSucceeded,
	countOtherSessions,
	describeDevice,
	formatSessionDate,
	issuedRecoveryCodes
} from './settings';
import type { ActiveSession } from '$lib/api/types';

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

describe('issuedRecoveryCodes', () => {
	it('los recoge tanto al activar 2FA como al regenerarlos', () => {
		expect(
			issuedRecoveryCodes({ action: 'enable2fa', success: true, recoveryCodes: ['a'] })
		).toEqual(['a']);
		expect(
			issuedRecoveryCodes({
				action: 'regenerate2faCodes',
				success: true,
				recoveryCodes: ['b', 'c']
			})
		).toEqual(['b', 'c']);
	});

	it('no muestra códigos de otras acciones ni de un intento fallido', () => {
		expect(issuedRecoveryCodes({ action: 'disable2fa', success: true })).toEqual([]);
		expect(issuedRecoveryCodes({ action: 'enable2fa', error: 'Código incorrecto' })).toEqual([]);
	});
});

describe('describeDevice', () => {
	it('resuelve navegador y sistema operativo', () => {
		expect(
			describeDevice(
				'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0 Safari/537.36'
			)
		).toBe('Chrome · macOS');
	});

	it('prefiere Edge y Opera sobre el Chrome de su user agent', () => {
		expect(
			describeDevice('Mozilla/5.0 (Windows NT 10.0) Chrome/120.0 Safari/537.36 Edg/120.0')
		).toBe('Edge · Windows');
		expect(
			describeDevice('Mozilla/5.0 (X11; Linux x86_64) Chrome/120.0 Safari/537.36 OPR/106.0')
		).toBe('Opera · Linux');
	});

	it('cae a un texto genérico sin user agent', () => {
		expect(describeDevice(null)).toBe('Dispositivo desconocido');
		expect(describeDevice('curl/8.4.0')).toBe('Navegador desconocido');
	});
});

describe('formatSessionDate', () => {
	it('formatea una fecha ISO', () => {
		expect(formatSessionDate('2026-03-14T09:05:00.000Z')).not.toBe('—');
	});

	it('no rompe con una fecha inválida', () => {
		expect(formatSessionDate('no-es-una-fecha')).toBe('—');
	});
});

describe('countOtherSessions', () => {
	const session = (id: string, current: boolean): ActiveSession => ({
		id,
		ipAddress: null,
		userAgent: null,
		location: null,
		createdAt: '2026-03-14T09:00:00.000Z',
		lastActiveAt: '2026-03-14T09:00:00.000Z',
		expiresAt: '2026-04-14T09:00:00.000Z',
		current
	});

	it('cuenta todas menos la actual', () => {
		expect(countOtherSessions([session('a', true), session('b', false), session('c', false)])).toBe(
			2
		);
	});

	it('vale 0 sin sesiones cargadas', () => {
		expect(countOtherSessions(undefined)).toBe(0);
	});
});
