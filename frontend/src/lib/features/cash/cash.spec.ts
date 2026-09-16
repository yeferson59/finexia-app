import { describe, it, expect } from 'vitest';
import {
	cashAccountLabel,
	cashErrorMessage,
	cashKindSign,
	formatCashKind,
	groupCashAccounts,
	groupCashMovementsByMonth,
	suggestCashPortfolio,
	summarizeCash,
	type CashBalance,
	type CashMovement
} from './cash';
import { groupCashPlatforms } from './pockets';

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

describe('summarizeCash', () => {
	it('suma los saldos convertidos y los reparte por moneda', () => {
		const summary = summarizeCash(
			[
				balance({ balance: '4000000', currency: 'COP', value: '1000' }),
				balance({ balance: '250', currency: 'USD' }),
				balance({
					balance: '1000000',
					currency: 'COP',
					value: '250',
					sourceId: 's2',
					sourceName: 'Bancolombia'
				})
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

	it('cuenta cuentas, no saldos: una repartida entre portafolios es una', () => {
		const summary = summarizeCash(
			[
				balance({ balance: '700', currency: 'USD' }),
				balance({ balance: '300', currency: 'USD', portfolioId: 'p2', portfolioName: 'Retiro' })
			],
			'USD'
		);

		expect(summary.total).toBe(1000);
		expect(summary.funded).toBe(1);
		expect(summary.byCurrency).toEqual([
			{ currency: 'USD', balance: 1000, value: 1000, accounts: 1 }
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

describe('groupCashAccounts', () => {
	it('junta en una cuenta los saldos de una plataforma y una moneda', () => {
		const accounts = groupCashAccounts([
			balance({
				balance: '300',
				currency: 'USD',
				portfolioId: 'p2',
				portfolioName: 'Retiro',
				lastMovementDate: '2026-08-15T00:00:00Z'
			}),
			balance({ balance: '700', currency: 'USD' }),
			balance({ balance: '50', currency: 'USD', sourceId: 's2', sourceName: 'IBKR' })
		]);

		expect(accounts).toHaveLength(2);
		expect(accounts[0]).toMatchObject({
			key: 's1:USD',
			sourceName: 'Nu',
			balance: 1000,
			value: 1000,
			fxConverted: true,
			lastMovementDate: '2026-09-01T00:00:00Z'
		});
		expect(accounts[0].balances.map((b) => b.portfolioName)).toEqual(['Ahorro', 'Retiro']);
		expect(accounts[1]).toMatchObject({ key: 's2:USD', balance: 50 });
	});

	it('separa las monedas de una misma plataforma', () => {
		const accounts = groupCashAccounts([
			balance({ balance: '10', currency: 'USD' }),
			balance({ balance: '40000', currency: 'COP', value: '10' })
		]);

		expect(accounts.map((a) => a.key).sort()).toEqual(['s1:COP', 's1:USD']);
	});

	it('no suma al valor lo que no se pudo convertir', () => {
		const [account] = groupCashAccounts([
			balance({ balance: '900', currency: 'CHF', value: '900', fxConverted: false })
		]);

		expect(account).toMatchObject({ balance: 900, value: 0, fxConverted: false });
	});
});

describe('groupCashPlatforms', () => {
	it('junta las monedas de una plataforma y ordena por lo que guarda', () => {
		const platforms = groupCashPlatforms([
			balance({ balance: '100', currency: 'USD', sourceId: 's2', sourceName: 'Bróker' }),
			balance({ balance: '4000000', currency: 'COP', value: '1000' }),
			balance({ balance: '50', currency: 'USD' })
		]);

		expect(platforms.map((p) => p.sourceName)).toEqual(['Nu', 'Bróker']);
		expect(platforms[0]).toMatchObject({ value: 1050, partial: false });
		expect(platforms[0].accounts.map((a) => a.currency)).toEqual(['COP', 'USD']);
	});

	it('marca la plataforma a la que le falta una cuenta sin tasa', () => {
		const [platform] = groupCashPlatforms([
			balance({ balance: '50', currency: 'USD' }),
			balance({ balance: '900', currency: 'CHF', value: '900', fxConverted: false })
		]);

		expect(platform).toMatchObject({ value: 50, partial: true });
	});

	it('una cuenta vacía sin tasa no deja el total a medias', () => {
		const [platform] = groupCashPlatforms([
			balance({ balance: '50', currency: 'USD' }),
			balance({ balance: '0', currency: 'CHF', value: '0', fxConverted: false })
		]);

		expect(platform.partial).toBe(false);
	});
});

describe('groupCashMovementsByMonth', () => {
	const movement = (id: string, date: string) => ({ id, date }) as CashMovement;

	it('agrupa por mes, con el mes escrito', () => {
		const months = groupCashMovementsByMonth([
			movement('a', '2026-09-11T00:00:00Z'),
			movement('b', '2026-09-01T00:00:00Z'),
			movement('c', '2026-08-28T00:00:00Z')
		]);

		expect(months.map((m) => [m.key, m.label, m.movements.map((x) => x.id)])).toEqual([
			['2026-09', 'Septiembre de 2026', ['a', 'b']],
			['2026-08', 'Agosto de 2026', ['c']]
		]);
	});

	it('no repite un mes aunque sus movimientos no lleguen seguidos', () => {
		const months = groupCashMovementsByMonth([
			movement('a', '2026-09-11T00:00:00Z'),
			movement('b', '2026-08-28T00:00:00Z'),
			movement('c', '2026-09-02T00:00:00Z')
		]);

		expect(months.map((m) => m.key)).toEqual(['2026-09', '2026-08']);
		expect(months[0].movements.map((x) => x.id)).toEqual(['a', 'c']);
	});
});

describe('suggestCashPortfolio', () => {
	const balances = [
		balance({ balance: '300', currency: 'USD', portfolioId: 'p2' }),
		balance({ balance: '700', currency: 'USD', portfolioId: 'p1' }),
		balance({ balance: '5', currency: 'COP', portfolioId: 'p3' })
	];

	it('propone el portafolio del saldo mayor de la cuenta', () => {
		expect(suggestCashPortfolio(balances, 's1', 'USD', 'p9')).toBe('p1');
		expect(suggestCashPortfolio(balances, 's1', 'COP', 'p9')).toBe('p3');
	});

	it('si la cuenta no existe, deja lo que estaba elegido', () => {
		expect(suggestCashPortfolio(balances, 's2', 'USD', 'p9')).toBe('p9');
		expect(suggestCashPortfolio(balances, 's1', 'EUR', 'p9')).toBe('p9');
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

	it('sin portafolio nombra solo la plataforma', () => {
		expect(cashAccountLabel({ sourceName: 'Nu', portfolioName: 'Ahorro' }, false)).toBe('Nu');
		expect(cashAccountLabel({ sourceName: '', portfolioName: 'Ahorro' }, false)).toBe(
			'Sin plataforma'
		);
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
