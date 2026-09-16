import { describe, it, expect } from 'vitest';
import { groupCashAccounts, type CashBalance } from './cash';
import type { CashRate } from './rates';
import {
	cashCurrencyName,
	cashDepositTerm,
	cashRateLines,
	cashRateScale,
	cashYieldMap
} from './yield';

const balance = (
	over: Partial<CashBalance> & { balance: string; currency: string }
): CashBalance => ({
	entryId: crypto.randomUUID(),
	portfolioId: 'p1',
	portfolioName: 'Ahorro',
	sourceId: 's1',
	sourceName: 'Nu',
	assetId: 'a1',
	ticker: `CASH-${over.currency}`,
	name: 'Efectivo',
	value: over.balance,
	displayCurrency: 'USD',
	fxConverted: true,
	movements: 1,
	lastMovementDate: '2026-09-01T00:00:00Z',
	interestEarned: '0',
	interestThisMonth: '0',
	interestThisMonthValue: '0',
	pendingInterest: '0',
	lastAccrualDate: null,
	pocketId: null,
	pocketName: '',
	pocketKind: '',
	...over
});

const rate = (
	over: Partial<CashRate> & { annualRatePct: string; effectiveFrom: string }
): CashRate => ({
	id: crypto.randomUUID(),
	sourceId: 's1',
	sourceName: 'Nu',
	currency: 'COP',
	pocketId: null,
	pocketName: '',
	withholdingPct: '0',
	posting: 'daily',
	tiers: [],
	endedOn: null,
	latest: true,
	createdAt: '2026-09-01T00:00:00Z',
	updatedAt: '2026-09-01T00:00:00Z',
	accruedThrough: null,
	...over
});

const TODAY = '2026-09-14';
const money = (amount: number) => `$ ${amount.toLocaleString('es-CO')}`;

describe('cashYieldMap', () => {
	it('ordena de la tasa más alta a la más baja y deja lo parado al final', () => {
		const accounts = groupCashAccounts([
			balance({ balance: '300', currency: 'USD', sourceId: 'broker' }),
			balance({ balance: '4000000', currency: 'COP', value: '1000' }),
			balance({
				balance: '2000000',
				currency: 'COP',
				value: '500',
				pocketId: 'k1',
				pocketName: 'Viajes',
				pocketKind: 'flexible'
			})
		]);
		const map = cashYieldMap(
			accounts,
			[
				rate({ annualRatePct: '9', effectiveFrom: '2026-01-01T00:00:00Z' }),
				rate({ annualRatePct: '11', effectiveFrom: '2026-01-01T00:00:00Z', pocketId: 'k1' })
			],
			TODAY
		);

		expect(map.blocks.map((b) => [b.pocketName || b.currency, b.pct])).toEqual([
			['Viajes', 11],
			['COP', 9],
			['USD', 0]
		]);
		expect(map.earningShare).toBeCloseTo(1500 / 1800);
		expect(map.idleShare).toBeCloseTo(300 / 1800);
		expect(map.top).toBe(11);
	});

	it('deja fuera las cuentas vacías y las que no tienen tasa de cambio', () => {
		const map = cashYieldMap(
			groupCashAccounts([
				balance({ balance: '0', currency: 'USD' }),
				balance({ balance: '900', currency: 'EUR', fxConverted: false, sourceId: 's2' }),
				balance({ balance: '100', currency: 'USD', sourceId: 's3' })
			]),
			[],
			TODAY
		);

		expect(map.blocks).toHaveLength(1);
		expect(map.blocks[0].share).toBe(1);
		expect(map.idleShare).toBe(1);
	});

	it('mide una cuenta con tramos por la tasa a la que rinde todo su saldo', () => {
		const [block] = cashYieldMap(
			groupCashAccounts([balance({ balance: '8000000', currency: 'COP', value: '2000' })]),
			[
				rate({
					annualRatePct: '12',
					effectiveFrom: '2026-01-01T00:00:00Z',
					tiers: [{ fromBalance: '5000000', annualRatePct: '8' }]
				})
			],
			TODAY
		).blocks;

		expect(block.pct).toBeGreaterThan(10);
		expect(block.pct).toBeLessThan(11);
		expect(block.yearly).toBeGreaterThan(0);
	});

	it('marca los depósitos a tasa fija', () => {
		const [block] = cashYieldMap(
			groupCashAccounts([
				balance({
					balance: '1000',
					currency: 'USD',
					pocketId: 'd1',
					pocketName: 'CDT',
					pocketKind: 'fixed'
				})
			]),
			[],
			TODAY
		).blocks;

		expect(block.fixed).toBe(true);
	});

	it('sin cuentas con dinero no hay nada parado ni rindiendo', () => {
		expect(cashYieldMap([], [], TODAY)).toEqual({
			blocks: [],
			earningShare: 0,
			idleShare: 0,
			top: 0
		});
	});
});

describe('cashRateScale', () => {
	it('pone un techo redondo con como mucho cuatro marcas sobre el cero', () => {
		expect(cashRateScale(10.4)).toEqual({ max: 12, ticks: [0, 3, 6, 9, 12] });
		expect(cashRateScale(4.1)).toEqual({ max: 6, ticks: [0, 2, 4, 6] });
		expect(cashRateScale(12)).toEqual({ max: 12, ticks: [0, 3, 6, 9, 12] });
		expect(cashRateScale(0.3)).toEqual({ max: 0.5, ticks: [0, 0.5] });
	});

	it('sin tasas la escala es solo el suelo', () => {
		expect(cashRateScale(0)).toEqual({ max: 1, ticks: [0] });
	});
});

describe('cashRateLines', () => {
	const status = (over: Partial<Parameters<typeof cashRateLines>[0]>) => ({
		current: null,
		upcoming: null,
		latest: null,
		accruedThrough: null,
		...over
	});

	it('parte la tasa que rige de sus tramos y su forma de abono', () => {
		const current = rate({
			annualRatePct: '9.25',
			effectiveFrom: '2026-01-01T00:00:00Z',
			posting: 'monthly',
			tiers: [{ fromBalance: '20000000', annualRatePct: '6' }]
		});

		expect(cashRateLines(status({ current, latest: current }), money)).toEqual({
			rate: '9,25% E.A.',
			detail: 'hasta $ 20.000.000, 6% después, abono mensual',
			earning: true
		});
	});

	it('no nombra la tasa de un tope', () => {
		const current = rate({
			annualRatePct: '12',
			effectiveFrom: '2026-01-01T00:00:00Z',
			tiers: [
				{ fromBalance: '5000000', annualRatePct: '8' },
				{ fromBalance: '10000000', annualRatePct: '0' }
			]
		});

		expect(cashRateLines(status({ current }), money)?.detail).toBe(
			'hasta $ 5.000.000, 8% hasta $ 10.000.000'
		);
	});

	it('dice lo que viene después', () => {
		const current = rate({ annualRatePct: '9', effectiveFrom: '2026-01-01T00:00:00Z' });
		const upcoming = rate({ annualRatePct: '8', effectiveFrom: '2026-10-01T00:00:00Z' });

		expect(cashRateLines(status({ current, upcoming }))?.detail).toMatch(/^8% E\.A\. desde el 1/);
	});

	it('sin tasa hoy dice cuándo empieza o desde cuándo está en pausa', () => {
		const upcoming = rate({ annualRatePct: '8', effectiveFrom: '2026-10-01T00:00:00Z' });
		expect(cashRateLines(status({ upcoming }))).toMatchObject({
			rate: 'Sin rendir hoy',
			earning: false
		});

		const ended = rate({
			annualRatePct: '8',
			effectiveFrom: '2026-01-01T00:00:00Z',
			endedOn: '2026-08-31T00:00:00Z'
		});
		expect(cashRateLines(status({ latest: ended }))).toMatchObject({
			rate: 'En pausa',
			earning: false
		});
		expect(cashRateLines(status({ latest: ended }))?.detail).toMatch(/desde el 1 .*sept/);
	});

	it('sin ninguna tasa no hay líneas', () => {
		expect(cashRateLines(status({}))).toBeNull();
	});
});

describe('cashDepositTerm', () => {
	it('dice cuánto lleva y cuánto le falta', () => {
		expect(cashDepositTerm('2026-09-01T00:00:00Z', '2026-11-30T00:00:00Z', TODAY)).toEqual({
			total: 90,
			elapsed: 13,
			left: 77,
			progress: 13 / 90
		});
	});

	it('no se sale del plazo antes de abrir ni después de vencer', () => {
		expect(cashDepositTerm('2026-10-01', '2026-10-31', TODAY)?.elapsed).toBe(0);
		expect(cashDepositTerm('2026-06-01', '2026-07-01', TODAY)).toMatchObject({
			left: 0,
			progress: 1
		});
	});

	it('sin vencimiento no hay plazo', () => {
		expect(cashDepositTerm('2026-09-01', null, TODAY)).toBeNull();
	});
});

describe('cashCurrencyName', () => {
	it('nombra la moneda con mayúscula inicial', () => {
		expect(cashCurrencyName('EUR')).toBe('Euro');
		expect(cashCurrencyName('COP')).toBe('Peso colombiano');
	});

	it('deja el código si no es uno que se pueda nombrar', () => {
		expect(cashCurrencyName('no')).toBe('no');
	});
});
