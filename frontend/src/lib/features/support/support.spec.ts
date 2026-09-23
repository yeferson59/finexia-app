import { describe, expect, it } from 'vitest';
import { parsePesos, supportResult } from './support';
import { supportAmountSchema } from './schemas';

describe('parsePesos', () => {
	it('lee el monto como lo escribe la gente', () => {
		expect(parsePesos('20.000')).toBe(20000);
		expect(parsePesos('$ 15 000')).toBe(15000);
		expect(parsePesos('abc')).toBeNaN();
	});
});

describe('supportAmountSchema', () => {
	it('toma el monto sugerido o el que se escribe en «Otro»', () => {
		expect(supportAmountSchema.parse({ preset: '50000' })).toBe(50000);
		expect(supportAmountSchema.parse({ preset: 'otro', custom: '7.500' })).toBe(7500);
	});

	it('rechaza montos fuera de rango o vacíos con un mensaje claro', () => {
		const low = supportAmountSchema.safeParse({ preset: 'otro', custom: '500' });
		const high = supportAmountSchema.safeParse({ preset: 'otro', custom: '9000000' });
		const empty = supportAmountSchema.safeParse({ preset: 'otro', custom: '' });
		// es-CO separa el símbolo con un espacio duro: «$ 1.000».
		expect(low.error?.issues[0].message).toMatch(/^El aporte mínimo es de \$\s1\.000\.$/);
		expect(high.error?.issues[0].message).toMatch(/^El aporte máximo es de \$\s2\.000\.000\.$/);
		expect(empty.error?.issues[0].message).toBe('Escribe cuánto quieres aportar.');
	});
});

describe('supportResult', () => {
	it('usa el estado que guarda el backend', () => {
		expect(supportResult('approved', 20000, 'rejected')).toEqual({
			outcome: 'approved',
			total: 20000
		});
		expect(supportResult('pending', null, null)?.outcome).toBe('pending');
		expect(supportResult('rejected', null, 'approved')?.outcome).toBe('failed');
		expect(supportResult('voided', null, null)?.outcome).toBe('failed');
	});

	it('una orden que sigue en «created» se lee de la URL, sin aprobarla', () => {
		expect(supportResult('created', null, 'approved')).toEqual({ outcome: 'pending', total: null });
		expect(supportResult('created', null, null)).toBeNull();
	});

	it('sin respuesta del backend, la URL nunca da un pago por aprobado', () => {
		expect(supportResult(null, null, 'approved')).toEqual({ outcome: 'pending', total: null });
		expect(supportResult(null, null, 'rejected')?.outcome).toBe('failed');
		expect(supportResult(null, null, null)).toBeNull();
	});
});
