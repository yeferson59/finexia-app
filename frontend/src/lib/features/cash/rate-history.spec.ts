import { describe, it, expect } from 'vitest';
import { cashRateHistory, type CashRateHistoryInput } from './rate-history';
import type { CashRate } from '$lib/api/types';

const latest = {
	annualRatePct: '9',
	effectiveFrom: '2026-09-16T00:00:00Z'
} as CashRate;

const input = (over: Partial<CashRateHistoryInput> = {}): CashRateHistoryInput => ({
	mode: 'new',
	latest,
	used: true,
	stopped: false,
	pending: false,
	computed: '2026-09-22T00:00:00Z',
	startDay: '2026-09-23',
	today: '2026-09-23',
	...over
});

describe('cashRateHistory', () => {
	it('una tasa nueva desde un día ya calculado rehace esos días', () => {
		const history = cashRateHistory(input({ startDay: '2026-09-20' }));
		expect(history.recomputes).toBe(true);
		expect(history.hint).toMatch(/desde el 20 de septiembre de 2026 se borran/);
	});

	it('una tasa nueva desde hoy no toca nada', () => {
		const history = cashRateHistory(input());
		expect(history.recomputes).toBe(false);
		expect(history.hint).not.toMatch(/se borran|se calculan al guardar/);
	});

	it('mover hacia atrás una tasa que ya rindió rehace desde el día nuevo', () => {
		const history = cashRateHistory(input({ mode: 'move', startDay: '2026-09-11' }));
		expect(history.recomputes).toBe(true);
		expect(history.hint).toMatch(/desde el 11 de septiembre de 2026 se borran/);
	});

	it('mover hacia adelante una tasa que ya rindió rehace desde su inicio de antes', () => {
		const history = cashRateHistory(
			input({ mode: 'move', startDay: '2026-09-30', computed: '2026-09-22T00:00:00Z' })
		);
		expect(history.recomputes).toBe(true);
		expect(history.hint).toMatch(/desde el 16 de septiembre de 2026 se borran/);
	});

	it('mover una tasa que no ha rendido a un día pasado calcula esos días', () => {
		const history = cashRateHistory(
			input({ mode: 'move', used: false, computed: null, startDay: '2026-09-11' })
		);
		expect(history.recomputes).toBe(false);
		expect(history.hint).toMatch(/hasta el último día terminado se calculan al guardar/);
	});

	it('recalcular dice hasta dónde están calculados los intereses', () => {
		const history = cashRateHistory(input({ mode: 'recalc' }));
		expect(history.recomputes).toBe(false);
		expect(history.hint).toMatch(/calculados hasta el 22 de septiembre de 2026/);
	});

	it('sin tasa solo dice desde cuándo rinde', () => {
		expect(cashRateHistory(input({ latest: null, used: false, computed: null })).hint).toBe(
			'Desde ese día, la cuenta rinde esta tasa.'
		);
	});
});
