import { describe, it, expect } from 'vitest';
import {
	countOtherSessions,
	describeAccess,
	describeConnections,
	describeDevice,
	describeIdentity,
	formatMCPTokenDate,
	formatSessionDate,
	issuedMCPToken,
	issuedRecoveryCodes
} from './settings';
import { mcpTokenExpirySchema, mcpTokenNameSchema, profileSchema } from './schemas';
import type { ActiveSession, MarketCredential, MCPToken, OAuthGrant } from '$lib/api/types';

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

describe('profileSchema', () => {
	const parse = (preferredCurrency: string) =>
		profileSchema.safeParse({ name: 'Jane Doe', preferredCurrency });

	it('normaliza la moneda antes de validarla', () => {
		const parsed = parse('  eur ');
		expect(parsed.success && parsed.data.preferredCurrency).toBe('EUR');
	});

	// La moneda preferida es a la que el panel convierte todos los totales: una
	// sin tasa detrás la rechaza el backend, así que se corta aquí y con un
	// mensaje que dice cuáles valen, en vez de esperar al 400.
	it('rechaza una moneda que la app no puede convertir', () => {
		const parsed = parse('ARS');
		expect(parsed.success).toBe(false);
		expect(parsed.success === false && parsed.error.issues[0].message).toContain('USD');
	});
});

describe('issuedMCPToken', () => {
	const token = {
		id: 't1',
		name: 'Claude Desktop',
		last4: 'a3f9',
		expiresAt: null,
		lastUsedAt: null,
		rotatedAt: null,
		createdAt: '2026-03-01T10:00:00Z',
		expired: false,
		token: 'fnx_mcp_secreto'
	};

	it('lo recoge tanto al crear como al rotar', () => {
		expect(issuedMCPToken({ action: 'createMcpToken', success: true, mcpToken: token })).toEqual(
			token
		);
		expect(issuedMCPToken({ action: 'rotateMcpToken', success: true, mcpToken: token })).toEqual(
			token
		);
	});

	it('no lo muestra en otras acciones ni en un intento fallido', () => {
		expect(issuedMCPToken({ action: 'deleteMcpToken', success: true })).toBeNull();
		expect(issuedMCPToken({ action: 'createMcpToken', error: 'Ya tienes un token' })).toBeNull();
		expect(issuedMCPToken(null)).toBeNull();
	});
});

describe('formatMCPTokenDate', () => {
	it('devuelve — cuando no hay fecha: un token sin usar y uno sin caducidad', () => {
		expect(formatMCPTokenDate(null)).toBe('—');
	});

	it('formatea una fecha real', () => {
		expect(formatMCPTokenDate('2026-03-01T10:00:00Z')).not.toBe('—');
	});
});

describe('schemas de tokens MCP', () => {
	it('exige un nombre y recorta los espacios', () => {
		expect(mcpTokenNameSchema.safeParse('   ').success).toBe(false);
		expect(mcpTokenNameSchema.parse('  Claude Desktop  ')).toBe('Claude Desktop');
		expect(mcpTokenNameSchema.safeParse('x'.repeat(61)).success).toBe(false);
	});

	it('acepta el 0 como «sin caducidad» y rechaza más de un año', () => {
		expect(mcpTokenExpirySchema.parse('0')).toBe(0);
		expect(mcpTokenExpirySchema.parse('90')).toBe(90);
		expect(mcpTokenExpirySchema.safeParse('366').success).toBe(false);
		expect(mcpTokenExpirySchema.safeParse('-1').success).toBe(false);
	});
});

describe('describeIdentity', () => {
	const user = {
		name: 'Usuaria Prueba',
		email: 'user@finexia.test',
		emailVerified: true,
		image: '',
		role: 'user',
		preferredCurrency: 'USD',
		createdAt: '2025-06-14T09:00:00.000Z',
		updatedAt: '2026-03-14T09:00:00.000Z'
	};

	it('dice a nombre de quién está la cuenta y desde cuándo', () => {
		const line = describeIdentity(user);

		expect(line).toContain('Usuaria Prueba');
		// El correo se lee aquí porque no se puede editar: era un campo de
		// formulario desactivado.
		expect(line).toContain('user@finexia.test');
		expect(line).toMatch(/junio de 2025/);
	});

	it('se calla la antigüedad cuando la fecha no sirve', () => {
		const line = describeIdentity({ ...user, createdAt: 'no-es-una-fecha' });

		expect(line).toContain('user@finexia.test');
		expect(line).not.toMatch(/Es tuya desde/);
	});

	it('lo dice cuando no hay usuario que describir', () => {
		expect(describeIdentity(null)).toMatch(/No pudimos cargar/);
	});
});

describe('describeAccess', () => {
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

	it('avisa de que solo hay contraseña cuando la 2FA está apagada', () => {
		expect(
			describeAccess({ enabled: false, pendingSetup: false, recoveryCodesLeft: 0 }, [
				session('a', true)
			])
		).toBe(
			'Entras solo con tu contraseña: la verificación en dos pasos está desactivada. La única sesión abierta es la de este navegador.'
		);
	});

	it('cuenta las sesiones que no son la de este navegador', () => {
		const line = describeAccess({ enabled: true, pendingSetup: false, recoveryCodesLeft: 8 }, [
			session('a', true),
			session('b', false),
			session('c', false)
		]);

		expect(line).toContain('un código de tu autenticador');
		expect(line).toContain('otras 2 sesiones abiertas');
	});

	it('habla en singular con una sola sesión ajena', () => {
		const line = describeAccess(undefined, [session('a', true), session('b', false)]);

		expect(line).toContain('hay otra sesión abierta');
	});

	it('no inventa sesiones cuando no se pudieron cargar', () => {
		const line = describeAccess(
			{ enabled: true, pendingSetup: false, recoveryCodesLeft: 3 },
			undefined
		);

		expect(line).not.toMatch(/sesi[oó]n/i);
	});
});

describe('describeConnections', () => {
	const credential = { provider: 'finnhub' } as MarketCredential;
	const token = { id: 't1' } as MCPToken;
	const grant = { id: 'g1' } as OAuthGrant;

	it('invita a conectar algo cuando no hay nada', () => {
		expect(describeConnections([], [], [])).toBe('Todavía no has conectado nada a tu cuenta.');
		expect(describeConnections(undefined, undefined, undefined)).toBe(
			'Todavía no has conectado nada a tu cuenta.'
		);
	});

	it('enumera lo que hay, con género y número', () => {
		expect(describeConnections([credential], [token], [grant])).toBe(
			'Tienes una clave de precios, un token de asistente y una aplicación conectada.'
		);
	});

	it('enumera solo lo que hay: «tienes ninguna clave» no es una frase', () => {
		expect(describeConnections([], [token, token], [])).toBe('Tienes 2 tokens de asistente.');
		expect(describeConnections([credential], [], [grant])).toBe(
			'Tienes una clave de precios y una aplicación conectada.'
		);
	});
});
