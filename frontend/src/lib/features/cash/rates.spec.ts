import { describe, it, expect } from 'vitest';
import {
	annualFromNominal,
	CASH_RATE_FALLBACK,
	CASH_RECALCULATE_FALLBACK,
	cashAccountRate,
	cashRateErrorMessage,
	cashYield,
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
	maxBalance: null,
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

	// Con tope solo rinde la parte del saldo que cabe en él.
	it('solo proyecta hasta el tope', () => {
		expect(projectInterest(10_000_000, 9, 0, 5_000_000).day).toBeCloseTo(1180.66, 2);
		// Por debajo del tope, rinde todo.
		expect(projectInterest(4_000_000, 9, 0, 5_000_000).day).toBeCloseTo(
			projectInterest(4_000_000, 9).day,
			10
		);
		// Un tope de cero es «sin tope», que es como llega un campo vacío.
		expect(projectInterest(1_000_000, 9, 0, 0).day).toBeCloseTo(
			projectInterest(1_000_000, 9).day,
			10
		);
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

	// Lo que dice si hay días que recalcular: el más lejano de cualquier versión.
	it('lleva el último día calculado de la cuenta', () => {
		const first = rate({
			annualRatePct: '9',
			effectiveFrom: '2026-08-01T00:00:00Z',
			endedOn: '2026-08-31T00:00:00Z',
			accruedThrough: '2026-08-31T00:00:00Z'
		});
		const now = rate({
			annualRatePct: '10',
			effectiveFrom: '2026-09-01T00:00:00Z',
			accruedThrough: '2026-09-13T00:00:00Z',
			latest: true
		});

		expect(cashAccountRate([now, first], 's1', 'COP', TODAY).accruedThrough).toBe(
			'2026-09-13T00:00:00Z'
		);
		expect(cashAccountRate([first], 's1', 'COP', TODAY).accruedThrough).toBe(
			'2026-08-31T00:00:00Z'
		);
	});

	it('una cuenta sin tasas no tiene ninguna', () => {
		expect(cashAccountRate([], 's1', 'COP', TODAY)).toEqual({
			current: null,
			upcoming: null,
			accruedThrough: null,
			latest: null
		});
	});
});

describe('annualFromNominal', () => {
	// Lo que dice el folleto: 12 % nominal mes vencido no son 12 % al año.
	it('pasa una nominal a efectiva anual', () => {
		expect(annualFromNominal(12, 12)).toBeCloseTo(12.6825, 4);
		expect(annualFromNominal(12, 4)).toBeCloseTo(12.5509, 4);
		expect(annualFromNominal(12, 2)).toBeCloseTo(12.36, 4);
		expect(annualFromNominal(12, 365)).toBeCloseTo(12.7475, 4);
	});

	it('sin tasa o sin periodos no convierte nada', () => {
		expect(annualFromNominal(0, 12)).toBe(0);
		expect(annualFromNominal(-1, 12)).toBe(0);
		expect(annualFromNominal(12, 0)).toBe(0);
	});

	// La vuelta: capitalizar la efectiva que sale da la misma plata.
	it('la efectiva que devuelve rinde lo mismo que la nominal', () => {
		const monthly = 1 + 12 / 100 / 12;
		expect(1 + annualFromNominal(12, 12) / 100).toBeCloseTo(monthly ** 12, 10);
	});
});

describe('describeCashAccountRate', () => {
	const current = rate({ annualRatePct: '9.25', effectiveFrom: '2026-09-01T00:00:00Z' });
	const upcoming = rate({ annualRatePct: '8.75', effectiveFrom: '2026-09-20T00:00:00Z' });

	it('dice la tasa de hoy', () => {
		expect(
			describeCashAccountRate({ current, upcoming: null, latest: current, accruedThrough: null })
		).toBe('9,25% E.A.');
	});

	it('avisa del cambio que viene', () => {
		expect(
			describeCashAccountRate({ current, upcoming, latest: upcoming, accruedThrough: null })
		).toMatch(/^9,25% E\.A\. · 8,75% E\.A\. desde el 20 /);
		expect(
			describeCashAccountRate({ current: null, upcoming, latest: upcoming, accruedThrough: null })
		).toMatch(/^8,75% E\.A\. desde el 20 /);
	});

	it('dice hasta cuándo rinde una tasa pausada a futuro', () => {
		const ending = { ...current, endedOn: '2026-09-30T00:00:00Z' };
		expect(
			describeCashAccountRate({
				current: ending,
				upcoming: null,
				latest: ending,
				accruedThrough: null
			})
		).toMatch(/^9,25% E\.A\. hasta el 30 /);
	});

	// Terminó el 13: el primer día sin rentabilidad es el 14.
	it('dice desde cuándo no rinde una tasa que ya terminó', () => {
		const ended = { ...current, endedOn: '2026-09-13T00:00:00Z' };
		expect(
			describeCashAccountRate({
				current: null,
				upcoming: null,
				latest: ended,
				accruedThrough: null
			})
		).toMatch(/^Sin rentabilidad desde el 14 /);
	});

	it('sin tasa no dice nada', () => {
		expect(
			describeCashAccountRate({ current: null, upcoming: null, latest: null, accruedThrough: null })
		).toBeNull();
	});

	// El abono mensual cambia cuándo llega el dinero, así que se dice en la línea.
	it('avisa del abono mensual', () => {
		const each = { ...current, posting: 'monthly' as const };
		expect(
			describeCashAccountRate({ current: each, upcoming: null, latest: each, accruedThrough: null })
		).toBe('9,25% E.A. · abono mensual');
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
		expect(cashRateErrorMessage(500)).toBe(CASH_RATE_FALLBACK);
	});

	// Los motivos son los mismos —la cuenta es la misma— pero lo que se
	// intentaba hacer no, así que la última frase la pone quien llama.
	it('el recálculo comparte los motivos y pone su propia frase final', () => {
		expect(cashRateErrorMessage(500, '', CASH_RECALCULATE_FALLBACK)).toBe(
			CASH_RECALCULATE_FALLBACK
		);
		expect(cashRateErrorMessage(404, '', CASH_RECALCULATE_FALLBACK)).toMatch(/esa cuenta/);
		expect(cashRateErrorMessage(404, 'platform not found', CASH_RECALCULATE_FALLBACK)).toMatch(
			/esa plataforma/
		);
		expect(
			cashRateErrorMessage(400, 'from cannot be in the future', CASH_RECALCULATE_FALLBACK)
		).toMatch(/no hay intereses que recalcular/);
	});
});

describe('cashYield', () => {
	const account = (over: { sourceId?: string; currency?: string; balance: number }) => ({
		sourceId: 's1',
		currency: 'COP',
		value: over.balance,
		...over
	});

	const nine = rate({ annualRatePct: '9', effectiveFrom: '2026-09-01T00:00:00Z', latest: true });
	const four = rate({
		annualRatePct: '4',
		effectiveFrom: '2026-09-01T00:00:00Z',
		sourceId: 's2',
		latest: true
	});

	it('pondera por lo que guarda cada cuenta que rinde', () => {
		// (30 × 9 + 10 × 4) / 40 = 7,75.
		expect(
			cashYield(
				[account({ balance: 30_000_000 }), account({ sourceId: 's2', balance: 10_000_000 })],
				[nine, four],
				TODAY
			)
		).toEqual({ pct: 7.75, idle: 0 });
	});

	// Una cuenta parada se cuenta aparte, no se promedia con un cero.
	it('deja fuera de la media las cuentas sin tasa', () => {
		expect(
			cashYield(
				[account({ balance: 10_000_000 }), account({ sourceId: 's3', balance: 90_000_000 })],
				[nine],
				TODAY
			)
		).toEqual({ pct: 9, idle: 1 });
	});

	it('sin ninguna cuenta con tasa la media es cero', () => {
		expect(cashYield([account({ sourceId: 's3', balance: 1000 })], [nine], TODAY)).toEqual({
			pct: 0,
			idle: 1
		});
	});

	it('una cuenta vacía no cuenta, y sin ninguna no hay cifra', () => {
		expect(cashYield([account({ balance: 0 })], [nine], TODAY)).toBeNull();
		expect(cashYield([], [], TODAY)).toBeNull();
	});
});
