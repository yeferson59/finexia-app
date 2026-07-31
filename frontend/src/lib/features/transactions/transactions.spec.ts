import { describe, it, expect } from 'vitest';
import { CATEGORY_OPTIONS, TXN_TYPE_LABELS, TXN_TYPE_OPTIONS } from './transactions';

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
