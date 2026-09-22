import { describe, it, expect } from 'vitest';
import {
	cashAccountRate,
	describeCashAccountRate,
	groupCashAccounts,
	groupCashPlatforms,
	nestCashPockets,
	type CashBalance,
	type CashRate
} from './index';
import {
	cashMoveSchema,
	cashPocketCreateSchema,
	cashPocketErrorMessage,
	cashPocketRenameSchema
} from './schemas';

const PORTFOLIO = '1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d';
const SOURCE = '7a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d';
const POCKET = '6f1e2d3c-4b5a-4c7d-8e9f-0a1b2c3d4e5f';
const BROKER = '9c8b7a6d-5e4f-4a3b-8c2d-1e0f9a8b7c6d';

const balance = (over: Partial<CashBalance> & { balance: string }): CashBalance => ({
	entryId: crypto.randomUUID(),
	portfolioId: 'p1',
	portfolioName: 'Ahorro',
	sourceId: 's1',
	sourceName: 'Nu',
	assetId: 'a1',
	ticker: 'CASH-COP',
	name: 'Efectivo',
	currency: 'COP',
	value: over.balance,
	displayCurrency: 'COP',
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

const rate = (over: Partial<CashRate> & { annualRatePct: string }): CashRate => ({
	id: crypto.randomUUID(),
	sourceId: 's1',
	sourceName: 'Nu',
	currency: 'COP',
	pocketId: null,
	pocketName: '',
	withholdingPct: '0',
	posting: 'daily',
	tiers: [],
	effectiveFrom: '2026-09-01T00:00:00Z',
	endedOn: null,
	latest: true,
	accruedThrough: null,
	createdAt: '2026-09-01T00:00:00Z',
	updatedAt: '2026-09-01T00:00:00Z',
	...over
});

describe('agrupar con bolsillos', () => {
	// Un bolsillo es otro cajón de la misma cuenta, no otra cuenta.
	it('separa el bolsillo de la cuenta principal', () => {
		const accounts = groupCashAccounts([
			balance({ balance: '6000000' }),
			balance({
				balance: '2000000',
				pocketId: POCKET,
				pocketName: 'Viajes',
				pocketKind: 'flexible'
			})
		]);

		expect(accounts.map((a) => a.key)).toEqual([`s1:COP`, `s1:COP:${POCKET}`]);
		expect(accounts[1]).toMatchObject({ pocketName: 'Viajes', balance: 2000000 });
	});

	it('cuelga los bolsillos de su cuenta', () => {
		const nested = nestCashPockets(
			groupCashAccounts([
				balance({ balance: '6000000' }),
				balance({ balance: '2000000', pocketId: POCKET, pocketName: 'Viajes' })
			])
		);

		expect(nested).toHaveLength(1);
		expect(nested[0].pocketId).toBeNull();
		expect(nested[0].pockets.map((p) => p.pocketName)).toEqual(['Viajes']);
	});

	// El dinero entró directo a la cajita: sin una principal inventada, la
	// cajita no tendría de dónde colgar.
	it('inventa la cuenta principal cuando solo hay bolsillo', () => {
		const nested = nestCashPockets(
			groupCashAccounts([balance({ balance: '2000000', pocketId: POCKET, pocketName: 'Viajes' })])
		);

		expect(nested).toHaveLength(1);
		expect(nested[0]).toMatchObject({ key: 's1:COP', pocketId: null, balance: 0 });
		expect(nested[0].pockets.map((p) => p.pocketName)).toEqual(['Viajes']);
	});

	// D1: un bolsillo no se va de su plataforma, así que la plataforma suma los
	// dos aunque la cajita se enseñe aparte.
	it('la plataforma suma la cuenta y sus bolsillos', () => {
		const [platform] = groupCashPlatforms([
			balance({ balance: '6000000' }),
			balance({ balance: '2000000', pocketId: POCKET, pocketName: 'Viajes' })
		]);

		expect(platform.value).toBe(8000000);
		expect(platform.accounts).toHaveLength(1);
	});
});

describe('cashAccountRate con bolsillos', () => {
	const rates = [
		rate({ annualRatePct: '8' }),
		rate({ annualRatePct: '10', pocketId: POCKET, pocketName: 'Viajes' })
	];

	it('cada cajón toma su propia tasa', () => {
		expect(cashAccountRate(rates, 's1', 'COP', '2026-09-15').current?.annualRatePct).toBe('8');
		expect(cashAccountRate(rates, 's1', 'COP', '2026-09-15', POCKET).current?.annualRatePct).toBe(
			'10'
		);
	});

	// Sin bolsillo se pide la principal, que es lo que había antes de que
	// hubiera bolsillos: la tasa de la cajita no se cuela en ella.
	it('la cuenta principal no hereda la tasa de la cajita', () => {
		const onlyPocket = [rate({ annualRatePct: '10', pocketId: POCKET, pocketName: 'Viajes' })];

		expect(cashAccountRate(onlyPocket, 's1', 'COP', '2026-09-15').current).toBeNull();
		expect(
			describeCashAccountRate(cashAccountRate(onlyPocket, 's1', 'COP', '2026-09-15'))
		).toBeNull();
	});
});

describe('formularios de bolsillos', () => {
	it('recorta el nombre y pide uno', () => {
		const ok = cashPocketCreateSchema.safeParse({
			sourceId: SOURCE,
			currency: 'COP',
			name: '  Viajes  '
		});
		expect(ok.success && ok.data.name).toBe('Viajes');

		const blank = cashPocketCreateSchema.safeParse({
			sourceId: SOURCE,
			currency: 'COP',
			name: '   '
		});
		expect(blank.success).toBe(false);

		const long = cashPocketRenameSchema.safeParse({ id: POCKET, name: 'x'.repeat(101) });
		expect(long.success).toBe(false);
	});

	// Vacío es la cuenta principal, y el destino no puede ser el origen: mover
	// el dinero a donde ya está no hace nada.
	it('lee los cajones del traslado', () => {
		const move = {
			portfolioId: PORTFOLIO,
			sourceId: SOURCE,
			currency: 'COP',
			fromPocketId: '',
			toSourceId: SOURCE,
			toCurrency: 'COP',
			toPocketId: POCKET,
			amount: '2000000',
			date: '2026-09-15',
			notes: ''
		};

		const ok = cashMoveSchema.safeParse(move);
		expect(ok.success && ok.data).toMatchObject({
			fromPocketId: undefined,
			toPocketId: POCKET,
			amount: 2000000
		});

		const toItself = cashMoveSchema.safeParse({ ...move, toPocketId: '' });
		expect(toItself.success).toBe(false);

		const samePocket = cashMoveSchema.safeParse({
			...move,
			fromPocketId: POCKET,
			toPocketId: POCKET
		});
		expect(samePocket.success).toBe(false);

		const nothing = cashMoveSchema.safeParse({ ...move, amount: '0' });
		expect(nothing.success).toBe(false);
	});

	// El traslado a otra plataforma: el mismo cajón vacío a los dos lados es otro
	// sitio si la cuenta es otra, y cruzar monedas exige decir a cuánto.
	it('lee el traslado entre plataformas', () => {
		const move = {
			portfolioId: PORTFOLIO,
			sourceId: SOURCE,
			currency: 'COP',
			fromPocketId: '',
			toSourceId: BROKER,
			toCurrency: 'COP',
			toPocketId: '',
			amount: '400000',
			date: '2026-09-21',
			notes: ''
		};

		// La cuenta principal de otra plataforma sí es un destino distinto.
		expect(cashMoveSchema.safeParse(move).success).toBe(true);

		const converted = cashMoveSchema.safeParse({
			...move,
			toCurrency: 'USD',
			toAmount: '98.50'
		});
		expect(converted.success && converted.data).toMatchObject({
			toSourceId: BROKER,
			toCurrency: 'USD',
			toAmount: 98.5
		});

		// Cruzar monedas sin decir lo que llegó trasladaría el importe con otra
		// etiqueta: 400.000 pesos entrando como 400.000 dólares.
		const noArrival = cashMoveSchema.safeParse({ ...move, toCurrency: 'USD' });
		expect(noArrival.success).toBe(false);

		// Y dentro de una misma moneda llega lo mismo que sale.
		const shrunk = cashMoveSchema.safeParse({ ...move, toAmount: '399000' });
		expect(shrunk.success).toBe(false);

		// Sin cruzar monedas queda fuera del cuerpo, que es lo que el backend
		// entiende por «llega lo mismo».
		const plain = cashMoveSchema.safeParse(move);
		expect(plain.success && plain.data.toAmount).toBeUndefined();

		const nothingArrives = cashMoveSchema.safeParse({ ...move, toCurrency: 'USD', toAmount: '0' });
		expect(nothingArrives.success).toBe(false);
	});

	it('traduce los rechazos del backend', () => {
		expect(cashPocketErrorMessage(404)).toMatch(/ya no existe/);
		expect(cashPocketErrorMessage(409, 'this account already has a pocket with that name')).toMatch(
			/ya tiene un bolsillo con ese nombre/
		);
		expect(cashPocketErrorMessage(409, 'cash pocket still holds money')).toMatch(
			/todavía tiene movimientos/
		);
	});
});
