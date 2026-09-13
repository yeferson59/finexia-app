import { describe, it, expect } from 'vitest';
import {
	cashMovementCreateSchema,
	cashMovementDeleteSchema,
	cashMovementUpdateSchema,
	toCashMovementBody
} from './schemas';

const PORTFOLIO = '6f1e2d3c-4b5a-4c7d-8e9f-0a1b2c3d4e5f';
const SOURCE = '7a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d';

/** Lo que llega de un `FormData`: todo cadenas, y `null` lo que no se envió. */
const form = (over: Record<string, string | null> = {}) => ({
	kind: 'deposit',
	portfolioId: PORTFOLIO,
	sourceId: SOURCE,
	currency: 'COP',
	amount: '1500000',
	fees: '',
	date: '2026-09-10',
	notes: null,
	...over
});

const firstMessage = (result: { success: boolean; error?: { issues: { message: string }[] } }) =>
	result.error?.issues[0]?.message;

describe('cashMovementCreateSchema', () => {
	it('acepta un depósito con la comisión vacía y sin nota', () => {
		const result = cashMovementCreateSchema.safeParse(form());

		expect(result.success).toBe(true);
		expect(result.data).toMatchObject({ amount: 1500000, fees: 0, notes: '', currency: 'COP' });
	});

	it('pide un importe mayor que cero', () => {
		expect(firstMessage(cashMovementCreateSchema.safeParse(form({ amount: '0' })))).toBe(
			'El importe tiene que ser mayor que cero.'
		);
	});

	it('rechaza una moneda que no se puede convertir', () => {
		expect(cashMovementCreateSchema.safeParse(form({ currency: 'ARS' })).success).toBe(false);
	});

	it('pide portafolio y plataforma', () => {
		expect(firstMessage(cashMovementCreateSchema.safeParse(form({ sourceId: '' })))).toBe(
			'Elige la plataforma.'
		);
	});

	it('no deja poner comisión a unos intereses', () => {
		expect(
			firstMessage(cashMovementCreateSchema.safeParse(form({ kind: 'interest', fees: '2' })))
		).toBe('Los intereses se anotan netos, sin comisión.');
	});

	it('no deja que la comisión de un retiro supere el retiro', () => {
		const over = form({ kind: 'withdrawal', amount: '100', fees: '101' });

		expect(firstMessage(cashMovementCreateSchema.safeParse(over))).toBe(
			'La comisión no puede ser mayor que el retiro.'
		);
		expect(cashMovementCreateSchema.safeParse({ ...over, fees: '100' }).success).toBe(true);
	});
});

describe('cashMovementUpdateSchema', () => {
	it('no pide saldo, solo el movimiento', () => {
		const result = cashMovementUpdateSchema.safeParse({
			id: PORTFOLIO,
			kind: 'withdrawal',
			amount: '20.5',
			fees: '1',
			date: '2026-09-11',
			notes: '  cajero  '
		});

		expect(result.success).toBe(true);
		expect(result.data?.notes).toBe('cajero');
	});
});

describe('cashMovementDeleteSchema', () => {
	it('pide un id válido', () => {
		expect(cashMovementDeleteSchema.safeParse({ id: 'x' }).success).toBe(false);
	});
});

describe('toCashMovementBody', () => {
	it('manda la fecha como medianoche UTC de ese día', () => {
		const body = toCashMovementBody({
			kind: 'deposit',
			amount: 10,
			fees: 1,
			date: '2026-09-10',
			notes: ''
		});

		expect(body.date).toBe('2026-09-10T00:00:00Z');
		expect(body.fees).toBe(1);
	});

	it('no envía comisión con unos intereses', () => {
		expect(
			toCashMovementBody({ kind: 'interest', amount: 3, fees: 5, date: '2026-09-10', notes: '' })
				.fees
		).toBe(0);
	});
});
