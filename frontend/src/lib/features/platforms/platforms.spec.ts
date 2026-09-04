import { describe, it, expect } from 'vitest';
import { formatSourceType, rankByShare, shareTint } from './platforms';
import type { Platform } from './platforms';

const platform = (over: Partial<Platform> & { id: string; totalValue: string }): Platform => ({
	name: over.id,
	description: '',
	sourceType: 'broker',
	isActive: true,
	createdAt: '2026-01-15T00:00:00Z',
	investments: 1,
	...over
});

describe('formatSourceType', () => {
	it('traduce el `sourceType` del backend', () => {
		expect(formatSourceType('crypto_wallet')).toBe('Billetera Cripto');
	});

	it('devuelve el valor crudo cuando no conoce el tipo', () => {
		expect(formatSourceType('algo_nuevo')).toBe('algo_nuevo');
	});
});

describe('rankByShare', () => {
	it('ordena de mayor a menor y numera los puestos', () => {
		const ranked = rankByShare([
			platform({ id: 'chica', totalValue: '1000' }),
			platform({ id: 'grande', totalValue: '9000' })
		]);

		expect(ranked.map((r) => r.platform.id)).toEqual(['grande', 'chica']);
		expect(ranked.map((r) => r.rank)).toEqual([0, 1]);
	});

	it('reparte el 100% cuando el backend no informa `percent`', () => {
		const ranked = rankByShare([
			platform({ id: 'a', totalValue: '7500' }),
			platform({ id: 'b', totalValue: '2500' })
		]);

		expect(ranked.map((r) => r.share)).toEqual([75, 25]);
	});

	// El backend calcula `percent` contra su propio total, que puede incluir
	// cosas que este listado no trae: si lo informa, manda él.
	it('respeta el `percent` del backend cuando viene', () => {
		const ranked = rankByShare([platform({ id: 'a', totalValue: '5000', percent: 12.5 })]);

		expect(ranked[0].share).toBe(12.5);
	});

	it('no divide por cero con una cuenta vacía', () => {
		const ranked = rankByShare([platform({ id: 'a', totalValue: '0' })]);

		expect(ranked[0].share).toBe(0);
	});

	it('deja intacto el array que recibe', () => {
		const input = [
			platform({ id: 'chica', totalValue: '1000' }),
			platform({ id: 'grande', totalValue: '9000' })
		];
		rankByShare(input);

		expect(input.map((p) => p.id)).toEqual(['chica', 'grande']);
	});
});

describe('shareTint', () => {
	it('pinta entera la mayor y va bajando con el puesto', () => {
		expect(shareTint(0, 4)).toBe(1);
		expect(shareTint(3, 4)).toBeCloseTo(0.38);
		expect(shareTint(1, 4)).toBeGreaterThan(shareTint(2, 4));
	});

	// El tramo más pequeño tiene que seguir viéndose sobre el fondo casi negro.
	it('nunca baja del suelo de visibilidad', () => {
		for (const count of [2, 5, 20]) {
			expect(shareTint(count - 1, count)).toBeGreaterThanOrEqual(0.38);
		}
	});

	it('pinta entera la única plataforma de la cuenta', () => {
		expect(shareTint(0, 1)).toBe(1);
	});
});
