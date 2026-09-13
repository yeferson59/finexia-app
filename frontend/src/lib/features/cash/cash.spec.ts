import { describe, it, expect } from 'vitest';
import {
	cashAccountLabel,
	cashErrorMessage,
	cashKindSign,
	formatCashKind,
	summarizeCash,
	type CashBalance
} from './cash';

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
	...over
});

describe('summarizeCash', () => {
	it('suma los saldos convertidos y los reparte por moneda', () => {
		const summary = summarizeCash(
			[
				balance({ balance: '4000000', currency: 'COP', value: '1000' }),
				balance({ balance: '250', currency: 'USD' }),
				balance({ balance: '1000000', currency: 'COP', value: '250', sourceName: 'Bancolombia' })
			],
			'USD'
		);

		expect(summary.total).toBe(1500);
		expect(summary.currency).toBe('USD');
		expect(summary.funded).toBe(3);
		expect(summary.byCurrency).toEqual([
			{ currency: 'COP', balance: 5000000, value: 1250, accounts: 2 },
			{ currency: 'USD', balance: 250, value: 250, accounts: 1 }
		]);
	});

	it('deja fuera del total lo que no tenía tasa, pero no del reparto', () => {
		const summary = summarizeCash(
			[
				balance({ balance: '100', currency: 'USD' }),
				balance({ balance: '900', currency: 'CHF', value: '900', fxConverted: false })
			],
			'USD'
		);

		expect(summary.total).toBe(100);
		expect(summary.unconverted).toBe(1);
		expect(summary.byCurrency.find((g) => g.currency === 'CHF')).toMatchObject({
			balance: 900,
			value: 0
		});
	});

	it('no avisa de conversión por una cuenta vacía', () => {
		const summary = summarizeCash(
			[balance({ balance: '0', currency: 'CHF', value: '0', fxConverted: false })],
			'USD'
		);

		expect(summary.unconverted).toBe(0);
		expect(summary.funded).toBe(0);
	});

	it('sin saldos usa la moneda de reserva', () => {
		expect(summarizeCash([], 'COP')).toMatchObject({ total: 0, currency: 'COP', byCurrency: [] });
	});
});

describe('formatCashKind y cashKindSign', () => {
	it('nombra y firma los tres movimientos', () => {
		expect(formatCashKind('deposit')).toBe('Depósito');
		expect(formatCashKind('interest')).toBe('Intereses');
		expect(cashKindSign('deposit')).toBe(1);
		expect(cashKindSign('interest')).toBe(1);
		expect(cashKindSign('withdrawal')).toBe(-1);
	});

	it('un movimiento ajeno no suma ni resta', () => {
		expect(formatCashKind('other')).toBe('Otro movimiento');
		expect(cashKindSign('other')).toBe(0);
	});
});

describe('cashAccountLabel', () => {
	it('pone la plataforma delante del portafolio', () => {
		expect(cashAccountLabel({ sourceName: 'Nu', portfolioName: 'Ahorro' })).toBe('Nu · Ahorro');
	});
});

describe('cashErrorMessage', () => {
	it('explica un sobregiro', () => {
		expect(cashErrorMessage(409, 'insufficient cash balance: the balance holds 10 USD')).toContain(
			'no alcanza'
		);
	});

	it('distingue un portafolio perdido de un movimiento perdido', () => {
		expect(cashErrorMessage(404, 'portfolio or source not found')).toContain('portafolio');
		expect(cashErrorMessage(404, 'cash movement not found')).toContain('ya no existe');
	});

	it('traduce las reglas de la comisión', () => {
		expect(cashErrorMessage(400, 'invalid cash movement: interest is recorded net of fees')).toBe(
			'Los intereses se anotan netos, sin comisión.'
		);
	});

	it('no inventa un motivo que no conoce', () => {
		expect(cashErrorMessage(500)).toContain('Vuelve a intentarlo');
	});
});
