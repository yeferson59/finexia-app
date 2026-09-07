import { describe, it, expect } from 'vitest';
import {
	buildWorklist,
	describeAssets,
	describeDesk,
	describeInvitations,
	describeRates,
	describeUsers,
	describeWaitlist,
	formatAge,
	formatDeadline,
	isStale,
	summarizeDesk
} from './desk';
import type { Asset, ExchangeRate, InvitationItem, UserItem, WaitlistItem } from '$lib/api/types';

/** Un martes cualquiera, para que la aritmética de días no dependa de hoy. */
const NOW = new Date('2026-09-15T12:00:00.000Z');

/** `days` días antes de `NOW`, en ISO. */
const ago = (days: number) => new Date(NOW.getTime() - days * 86_400_000).toISOString();

describe('formatAge', () => {
	it('cuenta en la unidad que hace falta en cada tramo', () => {
		expect(formatAge(ago(0), NOW)).toBe('hoy');
		expect(formatAge(ago(1), NOW)).toBe('ayer');
		expect(formatAge(ago(4), NOW)).toBe('hace 4 días');
		expect(formatAge(ago(8), NOW)).toBe('hace una semana');
		expect(formatAge(ago(21), NOW)).toBe('hace 3 semanas');
		expect(formatAge(ago(70), NOW)).toBe('hace 2 meses');
		expect(formatAge(ago(500), NOW)).toBe('hace más de un año');
	});

	// Un precio que nunca se puso no es un precio de hace mucho: es que no hay.
	it('distingue lo que nunca se tocó de lo que se tocó hace tiempo', () => {
		expect(formatAge(null, NOW)).toBe('nunca');
	});
});

describe('isStale', () => {
	it('se mide contra el plazo que se le pase', () => {
		expect(isStale(ago(3), 7, NOW)).toBe(false);
		expect(isStale(ago(9), 7, NOW)).toBe(true);
		expect(isStale(ago(9), 30, NOW)).toBe(false);
	});

	it('cuenta como viejo lo que no tiene fecha', () => {
		expect(isStale(null, 7, NOW)).toBe(true);
	});
});

describe('formatDeadline', () => {
	it('mira hacia delante y hacia atrás', () => {
		expect(formatDeadline(ago(-5), NOW)).toBe('caduca en 5 días');
		expect(formatDeadline(ago(-0.5), NOW)).toBe('caduca mañana');
		expect(formatDeadline(ago(3), NOW)).toBe('caducó hace 3 días');
	});

	// Nadie lee «caduca en 116 días»: a partir de la semana cambia la unidad.
	it('deja de contar días cuando la fecha está lejos', () => {
		expect(formatDeadline(ago(-16), NOW)).toBe('caduca dentro de 2 semanas');
		expect(formatDeadline(ago(-116), NOW)).toBe('caduca dentro de 3 meses');
	});
});

describe('cómo está cada bloque', () => {
	it('dice cuánto lleva esperando la primera solicitud', () => {
		const waitlist = [
			{ id: 'a', email: 'a@x.test', status: 'pending', createdAt: ago(11) },
			{ id: 'b', email: 'b@x.test', status: 'pending', createdAt: ago(2) }
		] as WaitlistItem[];

		expect(describeWaitlist(waitlist, NOW)).toBe(
			'2 personas esperan invitación. La primera pidió acceso hace una semana.'
		);
		expect(describeWaitlist([], NOW)).toBe('Nadie espera invitación ahora mismo.');
	});

	it('avisa de la invitación que se cae antes', () => {
		const invitations = [
			{ id: 'i1', status: 'pending', expiresAt: ago(-2) },
			{ id: 'i2', status: 'accepted', expiresAt: ago(-9) }
		] as InvitationItem[];

		expect(describeInvitations(invitations, NOW)).toBe(
			'Una invitación sin aceptar: caduca en 2 días.'
		);
	});

	// La lista de usuarios llega paginada: las cuentas pequeñas son de la página
	// que se está viendo y la frase no puede darlas por del sistema entero.
	it('acota las cuentas de una página al hablar de ellas', () => {
		const users = [
			{ id: '1', emailVerified: true, bannedAt: null },
			{ id: '2', emailVerified: false, bannedAt: null }
		] as UserItem[];

		expect(describeUsers(users, 2, false)).toBe('2 cuentas: 1 sin verificar el correo.');
		expect(describeUsers(users, 140, true)).toBe(
			'140 cuentas. En esta página, 1 sin verificar el correo.'
		);
	});

	it('separa lo que envejeció de lo que aportó un usuario', () => {
		const assets = [
			{ id: '1', priceUpdatedAt: ago(1), isCurated: true },
			{ id: '2', priceUpdatedAt: ago(40), isCurated: true },
			{ id: '3', priceUpdatedAt: null, isCurated: false }
		] as Asset[];

		expect(describeAssets(assets, NOW)).toBe(
			'3 activos: 2 llevan más de una semana con el mismo precio; uno lo aportó un usuario y solo lo ve quien lo aportó.'
		);

		// Cuando son todos, se dice que son todos: «10 activos: 10 llevan…» es una
		// cuenta que el lector ya tenía.
		expect(describeAssets(assets.slice(1), NOW)).toBe(
			'2 activos: los 2 llevan más de una semana con el mismo precio; uno lo aportó un usuario y solo lo ve quien lo aportó.'
		);
	});

	it('cuenta las tasas por quién las mantiene', () => {
		const rates = [
			{ id: '1', source: 'dolarapi' },
			{ id: '2', source: 'manual' },
			{ id: '3', source: 'manual' }
		] as ExchangeRate[];

		expect(describeRates(rates)).toBe(
			'3 tasas compartidas. Una la reescribe el feed cada hora; las otras 2 las mantienes tú.'
		);
	});
});

describe('lo que hay pendiente', () => {
	const state = (over: Partial<Parameters<typeof buildWorklist>[0]> = {}) => ({
		waiting: 0,
		oldestWaitAt: null,
		expiringInvites: 0,
		nextExpiryAt: null,
		stalePrices: 0,
		oldestPriceAt: null,
		staleRates: 0,
		oldestRateAt: null,
		...over
	});

	it('solo saca una fila cuando esa fila tiene trabajo', () => {
		expect(buildWorklist(state(), NOW)).toEqual([]);

		const tasks = buildWorklist(state({ waiting: 3, oldestWaitAt: ago(11) }), NOW);
		expect(tasks).toHaveLength(1);
		expect(tasks[0].key).toBe('waitlist');
		expect(tasks[0].detail).toBe('La solicitud más antigua entró hace una semana.');
	});

	// Un activo sin precio no lleva «sin cambiar desde nunca».
	it('dice de otra forma lo que no tiene fecha', () => {
		const [task] = buildWorklist(state({ stalePrices: 4, oldestPriceAt: null }), NOW);
		expect(task.detail).toBe('Al menos uno no tiene precio todavía.');
	});

	it('encadena lo abierto en una frase, y dice cuándo no hay nada', () => {
		const tasks = buildWorklist(
			state({ waiting: 3, oldestWaitAt: ago(11), stalePrices: 12, oldestPriceAt: ago(24) }),
			NOW
		);

		expect(describeDesk(tasks)).toBe(
			'Hay 3 personas esperando invitación y 12 precios sin actualizar.'
		);
		expect(describeDesk([])).toBe('Nada pendiente: el acceso y el catálogo están al día.');
	});
});

describe('summarizeDesk', () => {
	const input = {
		assets: [
			{ id: 'a1', priceUpdatedAt: ago(1) },
			{ id: 'a2', priceUpdatedAt: ago(30) }
		] as Asset[],
		rates: [
			{ id: 'r1', source: 'dolarapi', rateDate: ago(0) },
			{ id: 'r2', source: 'manual', rateDate: ago(45) }
		] as ExchangeRate[],
		invitations: [
			{ id: 'i1', status: 'pending', expiresAt: ago(-1) },
			{ id: 'i2', status: 'pending', expiresAt: ago(-20) }
		] as InvitationItem[],
		waitlist: [
			{ id: 'w1', status: 'pending', createdAt: ago(6) },
			{ id: 'w2', status: 'invited', createdAt: ago(30) }
		] as WaitlistItem[]
	};

	it('cuenta solo lo que de verdad está pendiente', () => {
		const desk = summarizeDesk(input, NOW);

		// La entrada ya invitada no espera a nadie.
		expect(desk.waiting).toBe(1);
		// Una invitación con veinte días por delante no es trabajo de hoy.
		expect(desk.expiringInvites).toBe(1);
		expect(desk.stalePrices).toBe(1);
		// La del feed se refresca sola; la escrita a mano lleva mes y medio.
		expect(desk.staleRates).toBe(1);
	});

	it('devuelve la fecha más antigua de cada grupo', () => {
		const desk = summarizeDesk(input, NOW);

		expect(desk.oldestPriceAt).toBe(ago(30));
		expect(desk.nextExpiryAt).toBe(ago(-1));
	});

	// Un hueco sin fecha gana a cualquier fecha: nunca se tocó.
	it('marca como sin fecha el grupo donde falta alguna', () => {
		const desk = summarizeDesk(
			{ ...input, assets: [{ id: 'a3', priceUpdatedAt: null }] as Asset[] },
			NOW
		);

		expect(desk.stalePrices).toBe(1);
		expect(desk.oldestPriceAt).toBeNull();
	});
});
