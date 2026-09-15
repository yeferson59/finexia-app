import { describe, it, expect } from 'vitest';
import {
	cashAccountRate,
	cashRateErrorMessage,
	dailyRateFromAnnual,
	describeCashAccountRate,
	formatAnnualRate,
	projectInterest,
	type CashRate
} from './rates';

const rate = (
	over: Partial<CashRate> & { annualRatePct: string; effectiveFrom: string }
): CashRate => ({
	id: crypto.randomUUID(),
	sourceId: 's1',
	sourceName: 'Nu',
	currency: 'COP',
	withholdingPct: '0',
	posting: 'daily',
	endedOn: null,
	latest: false,
	createdAt: '2026-09-01T00:00:00Z',
	updatedAt: '2026-09-01T00:00:00Z',
	accruedThrough: null,
	...over
});

const TODAY = '2026-09-14';

describe('formatAnnualRate', () => {
	it('escribe la tasa como la publica la entidad', () => {
		expect(formatAnnualRate('9.25')).toBe('9,25% E.A.');
		expect(formatAnnualRate('9')).toBe('9% E.A.');
		expect(formatAnnualRate(13.5)).toBe('13,5% E.A.');
	});
});

describe('dailyRateFromAnnual', () => {
	it('capitalizada 365 días devuelve la tasa efectiva anual', () => {
		const daily = dailyRateFromAnnual(9);

		expect(daily).toBeCloseTo(0.000236131, 9);
		expect(Math.pow(1 + daily, 365) - 1).toBeCloseTo(0.09, 12);
	});
});

describe('projectInterest', () => {
	// El ejemplo del plan: diez millones al 9 % E.A.
	it('proyecta un día, 30 días y un año sobre el saldo de hoy', () => {
		const p = projectInterest(10_000_000, 9);

		expect(p.day).toBeCloseTo(2361.31, 2);
		expect(p.month).toBeCloseTo(71082.43, 1);
		expect(p.year).toBeCloseTo(900000, 2);
	});

	it('descuenta la retención de lo que rinde', () => {
		expect(projectInterest(10_000_000, 9, 7).day).toBeCloseTo(2196.02, 2);
	});

	it('sin saldo o sin tasa no rinde nada', () => {
		expect(projectInterest(0, 9)).toEqual({ day: 0, month: 0, year: 0 });
		expect(projectInterest(1000, 0)).toEqual({ day: 0, month: 0, year: 0 });
	});
});

describe('cashAccountRate', () => {
	it('separa la tasa que rige hoy de la que viene, en la cuenta pedida', () => {
		const old = rate({
			annualRatePct: '9.25',
			effectiveFrom: '2026-09-01T00:00:00Z',
			endedOn: '2026-09-19T00:00:00Z'
		});
		const next = rate({
			annualRatePct: '8.75',
			effectiveFrom: '2026-09-20T00:00:00Z',
			latest: true
		});
		const dollars = rate({
			annualRatePct: '4',
			effectiveFrom: '2026-09-01T00:00:00Z',
			currency: 'USD',
			latest: true
		});

		const status = cashAccountRate([next, old, dollars], 's1', 'COP', TODAY);

		expect(status.current).toBe(old);
		expect(status.upcoming).toBe(next);
		expect(status.latest).toBe(next);
	});

	it('el último día todavía rinde, y el siguiente ya no', () => {
		const lastToday = rate({
			annualRatePct: '9',
			effectiveFrom: '2026-09-01T00:00:00Z',
			endedOn: '2026-09-14T00:00:00Z',
			latest: true
		});
		expect(cashAccountRate([lastToday], 's1', 'COP', TODAY).current).toBe(lastToday);

		const endedYesterday = { ...lastToday, endedOn: '2026-09-13T00:00:00Z' };
		const status = cashAccountRate([endedYesterday], 's1', 'COP', TODAY);
		expect(status.current).toBeNull();
		expect(status.latest).toBe(endedYesterday);
	});

	it('una cuenta sin tasas no tiene ninguna', () => {
		expect(cashAccountRate([], 's1', 'COP', TODAY)).toEqual({
			current: null,
			upcoming: null,
			latest: null
		});
	});
});

describe('describeCashAccountRate', () => {
	const current = rate({ annualRatePct: '9.25', effectiveFrom: '2026-09-01T00:00:00Z' });
	const upcoming = rate({ annualRatePct: '8.75', effectiveFrom: '2026-09-20T00:00:00Z' });

	it('dice la tasa de hoy', () => {
		expect(describeCashAccountRate({ current, upcoming: null, latest: current })).toBe(
			'9,25% E.A.'
		);
	});

	it('avisa del cambio que viene', () => {
		expect(describeCashAccountRate({ current, upcoming, latest: upcoming })).toMatch(
			/^9,25% E\.A\. · 8,75% E\.A\. desde el 20 /
		);
		expect(describeCashAccountRate({ current: null, upcoming, latest: upcoming })).toMatch(
			/^8,75% E\.A\. desde el 20 /
		);
	});

	it('dice hasta cuándo rinde una tasa pausada a futuro', () => {
		const ending = { ...current, endedOn: '2026-09-30T00:00:00Z' };
		expect(describeCashAccountRate({ current: ending, upcoming: null, latest: ending })).toMatch(
			/^9,25% E\.A\. hasta el 30 /
		);
	});

	// Terminó el 13: el primer día sin rentabilidad es el 14.
	it('dice desde cuándo no rinde una tasa que ya terminó', () => {
		const ended = { ...current, endedOn: '2026-09-13T00:00:00Z' };
		expect(describeCashAccountRate({ current: null, upcoming: null, latest: ended })).toMatch(
			/^Sin rentabilidad desde el 14 /
		);
	});

	it('sin tasa no dice nada', () => {
		expect(describeCashAccountRate({ current: null, upcoming: null, latest: null })).toBeNull();
	});
});

describe('cashRateErrorMessage', () => {
	it('traduce los rechazos que puede provocar el formulario', () => {
		expect(
			cashRateErrorMessage(409, 'invalid: only the latest version of a cash rate can change')
		).toMatch(/versión más reciente/);
		expect(
			cashRateErrorMessage(
				409,
				'a version of this cash rate already starts on or after that date: the latest starts on 2026-09-20'
			)
		).toMatch(/empieza ese día o después/);
		expect(
			cashRateErrorMessage(
				400,
				'invalid cash rate: effectiveFrom cannot be before 2026-09-13: past days are not recomputed'
			)
		).toMatch(/hoy o una fecha posterior/);
		expect(cashRateErrorMessage(404, 'cash rate not found')).toMatch(/ya no existe/);
		expect(cashRateErrorMessage(500)).toMatch(/No pudimos guardar la tasa/);
	});
});
