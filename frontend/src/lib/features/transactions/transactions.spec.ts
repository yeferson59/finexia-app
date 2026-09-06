import { describe, it, expect } from 'vitest';
import {
	CATEGORY_OPTIONS,
	TXN_TYPE_LABELS,
	TXN_TYPE_OPTIONS,
	sortByDateDesc,
	transactionTotal
} from './transactions';
import type { UserTransaction } from '$lib/api/types';

describe('catálogos del import de transacciones', () => {
	it('el índice de etiquetas cubre todos los tipos ofrecidos', () => {
		// El wizard ofrece los tipos por `TXN_TYPE_OPTIONS` y pinta el preview con
		// `TXN_TYPE_LABELS`: si se editara solo uno, las filas saldrían en crudo.
		for (const { value, label } of TXN_TYPE_OPTIONS) {
			expect(TXN_TYPE_LABELS[value]).toBe(label);
		}
		expect(Object.keys(TXN_TYPE_LABELS)).toHaveLength(TXN_TYPE_OPTIONS.length);
	});

	it('incluye los tipos que el backend acepta', () => {
		expect(TXN_TYPE_OPTIONS.map((t) => t.value)).toEqual([
			'buy',
			'sell',
			'dividend',
			'interest',
			'transfer_in',
			'transfer_out',
			'fee',
			'split'
		]);
	});

	it('no ofrece categorías ni tipos duplicados', () => {
		const types = TXN_TYPE_OPTIONS.map((t) => t.value);
		const categories = CATEGORY_OPTIONS.map((c) => c.value);
		expect(new Set(types).size).toBe(types.length);
		expect(new Set(categories).size).toBe(categories.length);
	});
});

/** Transacción mínima; cada prueba pisa lo que le importa. */
function txn(over: Partial<UserTransaction> = {}): UserTransaction {
	return {
		id: 'a',
		entryId: 'e',
		type: 'buy',
		quantity: '10',
		price: '100',
		currency: 'USD',
		fees: '0',
		transactionDate: '2026-01-01',
		notes: '',
		createdAt: '2026-01-01T00:00:00Z',
		assetTicker: 'AAPL',
		assetName: 'Apple Inc.',
		...over
	};
}

describe('transactionTotal', () => {
	it('multiplica cantidad por precio cuando no hay conversión', () => {
		const total = transactionTotal(txn({ quantity: '10', price: '182.4' }));

		expect(total.amount).toBeCloseTo(1824);
		expect(total.currency).toBe('USD');
		expect(total.converted).toBe(false);
	});

	/* Lo que el listado no hacía: una compra cotizada en euros y liquidada en
	   dólares se leía como si el bróker hubiera cobrado la cifra en euros. */
	it('lleva el importe a la moneda en que se liquidó', () => {
		const total = transactionTotal(
			txn({ quantity: '5', price: '100', currency: 'EUR', costCurrency: 'USD', fxRate: '1.1' })
		);

		expect(total.amount).toBeCloseTo(550);
		expect(total.currency).toBe('USD');
		expect(total.quoteCurrency).toBe('EUR');
		expect(total.converted).toBe(true);
		expect(total.rate).toBe(1.1);
	});

	it('asume tasa 1 y la moneda de la operación si el backend no las manda', () => {
		const total = transactionTotal(txn({ fxRate: undefined, costCurrency: undefined }));

		expect(total.rate).toBe(1);
		expect(total.currency).toBe('USD');
		expect(total.converted).toBe(false);
	});

	it('no revienta con cifras vacías o ilegibles', () => {
		const total = transactionTotal(txn({ quantity: '', price: 'x', fxRate: '0' }));

		expect(total.amount).toBe(0);
		expect(total.rate).toBe(1);
	});
});

describe('sortByDateDesc', () => {
	it('deja la más reciente primero', () => {
		const ordered = sortByDateDesc([
			txn({ id: 'vieja', transactionDate: '2026-02-20' }),
			txn({ id: 'nueva', transactionDate: '2026-06-24' }),
			txn({ id: 'media', transactionDate: '2026-05-19' })
		]);

		expect(ordered.map((t) => t.id)).toEqual(['nueva', 'media', 'vieja']);
	});

	it('no toca el array que recibe', () => {
		const original = [
			txn({ id: 'a', transactionDate: '2026-01-01' }),
			txn({ id: 'b', transactionDate: '2026-09-01' })
		];
		sortByDateDesc(original);

		expect(original.map((t) => t.id)).toEqual(['a', 'b']);
	});
});
