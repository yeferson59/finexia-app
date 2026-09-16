import { describe, it, expect } from 'vitest';
import {
	addCalendarDays,
	daysBetween,
	describeCashAccountRate,
	describeFixedDeposit,
	projectInterestOverDays,
	type CashRate
} from './index';
import type { CashPocket } from '$lib/api/types';
import {
	cashDepositCloseSchema,
	cashDepositCreateSchema,
	cashDepositErrorMessage
} from './schemas';

const PORTFOLIO = '1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d';
const SOURCE = '7a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d';
const POCKET = '6f1e2d3c-4b5a-4c7d-8e9f-0a1b2c3d4e5f';

const pocket = (over: Partial<CashPocket> = {}): CashPocket => ({
	id: POCKET,
	sourceId: SOURCE,
	sourceName: 'Nu',
	currency: 'COP',
	name: 'CDT 90 días',
	kind: 'fixed',
	openedOn: '2026-09-01T00:00:00Z',
	maturesOn: '2026-11-30T00:00:00Z',
	closedOn: null,
	balance: '10036624.23',
	balances: 1,
	movements: 15,
	createdAt: '2026-09-15T00:00:00Z',
	updatedAt: '2026-09-15T00:00:00Z',
	...over
});

const rate = (over: Partial<CashRate> = {}): CashRate => ({
	id: 'r1',
	sourceId: SOURCE,
	sourceName: 'Nu',
	currency: 'COP',
	pocketId: POCKET,
	pocketName: 'CDT 90 días',
	annualRatePct: '10',
	withholdingPct: '4',
	posting: 'daily',
	tiers: [],
	effectiveFrom: '2026-09-01T00:00:00Z',
	endedOn: '2026-11-29T00:00:00Z',
	latest: true,
	accruedThrough: null,
	createdAt: '2026-09-15T00:00:00Z',
	updatedAt: '2026-09-15T00:00:00Z',
	...over
});

const deposit = (over: Record<string, unknown> = {}) => ({
	portfolioId: PORTFOLIO,
	sourceId: SOURCE,
	currency: 'COP',
	name: 'CDT 90 días',
	amount: '10000000',
	openedOn: '2026-09-01',
	maturesOn: '2026-11-30',
	annualRatePct: '10',
	withholdingPct: '4',
	posting: 'daily',
	...over
});

describe('lo que rinde un depósito', () => {
	// El ejemplo del plan: diez millones al 10 % E.A., 90 días, 4 % de retención.
	it('los 90 días netos de retención', () => {
		expect(projectInterestOverDays(10_000_000, 10, 90, 4)).toBeCloseTo(228_176, 0);
	});

	// Sin retención, lo mismo en bruto.
	it('los 90 días en bruto', () => {
		expect(projectInterestOverDays(10_000_000, 10, 90)).toBeCloseTo(237_794.68, 1);
	});

	// Los catorce días que ya pasaron cuando se registra tarde.
	it('los días que lleva abierto', () => {
		expect(projectInterestOverDays(10_000_000, 10, 14)).toBeCloseTo(36_624.23, 1);
	});

	it('sin dinero, sin tasa o sin días no rinde nada', () => {
		expect(projectInterestOverDays(0, 10, 90)).toBe(0);
		expect(projectInterestOverDays(10_000, 0, 90)).toBe(0);
		expect(projectInterestOverDays(10_000, 10, 0)).toBe(0);
	});
});

describe('el calendario del plazo', () => {
	it('cuenta los días entre dos fechas, en UTC', () => {
		expect(daysBetween('2026-09-01', '2026-11-30')).toBe(90);
		expect(daysBetween('2026-09-01T00:00:00Z', '2026-09-01')).toBe(0);
	});

	it('suma días a una fecha, en UTC', () => {
		expect(addCalendarDays('2026-09-01', 90)).toBe('2026-11-30');
		// Y cruza el cambio de año sin tropezar.
		expect(addCalendarDays('2026-12-31', 1)).toBe('2027-01-01');
	});
});

describe('cómo se lee un depósito', () => {
	it('la tasa, que es fija, y cuándo vence', () => {
		expect(describeFixedDeposit(pocket(), rate())).toBe(
			'10% E.A. · tasa fija · vence el 30 de nov'
		);
	});

	it('dice que abona al vencer cuando lo hace', () => {
		expect(describeFixedDeposit(pocket(), rate({ posting: 'at_maturity' }))).toContain(
			'abono al vencer'
		);
	});

	it('un depósito sin plazo va hasta que lo canceles', () => {
		expect(describeFixedDeposit(pocket({ maturesOn: null }), rate())).toContain('sin plazo');
	});

	it('uno cerrado dice cuándo se cerró, no cuándo vencía', () => {
		const closed = pocket({ closedOn: '2026-10-15T00:00:00Z' });
		expect(describeFixedDeposit(closed, rate())).toContain('cerrado el 15 de oct');
		expect(describeFixedDeposit(closed, rate())).not.toContain('vence');
	});

	// La cuenta corriente también sabe decir que abona al vencer, aunque solo un
	// depósito pueda tener esa tasa.
	it('la línea de una cuenta nombra el abono al vencer', () => {
		const account = { current: rate({ posting: 'at_maturity' }), upcoming: null, latest: null };
		expect(describeCashAccountRate({ ...account, accruedThrough: null })).toContain(
			'abono al vencer'
		);
	});
});

describe('abrir un depósito', () => {
	it('acepta el ejemplo del plan', () => {
		const parsed = cashDepositCreateSchema.safeParse(deposit());
		expect(parsed.success).toBe(true);
		if (parsed.success) {
			expect(parsed.data.amount).toBe(10_000_000);
			expect(parsed.data.maturesOn).toBe('2026-11-30');
			expect(parsed.data.posting).toBe('daily');
		}
	});

	it('un depósito sin plazo deja el vencimiento vacío', () => {
		const parsed = cashDepositCreateSchema.safeParse(deposit({ maturesOn: '' }));
		expect(parsed.success).toBe(true);
		if (parsed.success) expect(parsed.data.maturesOn).toBeUndefined();
	});

	it.each([
		['sin nombre', { name: '   ' }],
		['sin dinero', { amount: '0' }],
		['sin día de apertura', { openedOn: '' }],
		['que vence el día que se abrió', { maturesOn: '2026-09-01' }],
		['que vence antes de abrirse', { maturesOn: '2026-08-01' }],
		['sin tasa', { annualRatePct: '0' }],
		['con una tasa de más decimales', { annualRatePct: '10.123456' }],
		['que abona al vencer sin plazo', { maturesOn: '', posting: 'at_maturity' }]
	])('rechaza uno %s', (_, over) => {
		expect(cashDepositCreateSchema.safeParse(deposit(over)).success).toBe(false);
	});
});

describe('cancelar un depósito', () => {
	it('el día y lo que cobra la entidad', () => {
		const parsed = cashDepositCloseSchema.safeParse({
			id: POCKET,
			closesOn: '2026-10-15',
			penalty: '50000'
		});
		expect(parsed.success).toBe(true);
		if (parsed.success) expect(parsed.data.penalty).toBe(50_000);
	});

	it('sin penalidad es cero', () => {
		const parsed = cashDepositCloseSchema.safeParse({ id: POCKET, closesOn: '2026-10-15' });
		expect(parsed.success).toBe(true);
		if (parsed.success) expect(parsed.data.penalty).toBe(0);
	});

	it.each([
		['sin depósito', { closesOn: '2026-10-15' }],
		['sin día', { id: POCKET }],
		['con una penalidad negativa', { id: POCKET, closesOn: '2026-10-15', penalty: '-1' }]
	])('rechaza una cancelación %s', (_, body) => {
		expect(cashDepositCloseSchema.safeParse(body).success).toBe(false);
	});
});

describe('lo que se dice cuando un depósito no se puede guardar', () => {
	it.each([
		['takes no movements or rate versions of its own', 'no admite movimientos'],
		['the deposit is already closed', 'ya está cerrado'],
		['openedOn cannot be in the future', 'no puede ser futura'],
		['openedOn cannot be more than 5 years ago', 'más de cinco años'],
		['maturesOn must be after openedOn', 'posterior al día en que lo abriste'],
		['maturesOn cannot be before 2026-09-14', 'ya venció'],
		['posting at_maturity needs a maturesOn to credit on', 'ponerle un plazo'],
		['the platform is inactive; activate it', 'está inactiva'],
		['closesOn cannot be in the future', 'aún no ha pasado'],
		['closesOn cannot be before 2026-09-14', 'Elige hoy o ayer']
	])('traduce %s', (details, expected) => {
		expect(cashDepositErrorMessage(409, details)).toContain(expected);
	});

	it('dice desde cuándo se puede cancelar cuando el backend lo sabe', () => {
		const message = cashDepositErrorMessage(
			409,
			'the cash rate already earned interest: interest through 2026-09-14 was computed at it; the deposit can be cancelled from 2026-09-16'
		);
		expect(message).toContain('2026-09-16');
	});

	it('una cuenta que no existe se recarga', () => {
		expect(cashDepositErrorMessage(404, '')).toContain('Recarga');
	});
});
