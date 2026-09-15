import { describe, it, expect } from 'vitest';
import {
	cashRateCreateSchema,
	cashRateDeleteSchema,
	cashRateEndSchema,
	cashRateUpdateSchema,
	toCalendarDateTime,
	toCashRateBody
} from './schemas';

const SOURCE = '7a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d';
const RATE = '6f1e2d3c-4b5a-4c7d-8e9f-0a1b2c3d4e5f';

/** Lo que llega de un `FormData`: todo cadenas, y `null` lo que no se envió. */
const form = (over: Record<string, string | null> = {}) => ({
	sourceId: SOURCE,
	currency: 'COP',
	effectiveFrom: '2026-09-15',
	annualRatePct: '9.25',
	withholdingPct: '',
	...over
});

const firstMessage = (result: { success: boolean; error?: { issues: { message: string }[] } }) =>
	result.error?.issues[0]?.message;

describe('cashRateCreateSchema', () => {
	it('acepta una tasa con la retención vacía', () => {
		const result = cashRateCreateSchema.safeParse(form());

		expect(result.success).toBe(true);
		expect(result.data).toMatchObject({
			annualRatePct: 9.25,
			withholdingPct: 0,
			effectiveFrom: '2026-09-15'
		});
	});

	it('pide una tasa mayor que cero y de hasta 100 %', () => {
		expect(firstMessage(cashRateCreateSchema.safeParse(form({ annualRatePct: '' })))).toBe(
			'La tasa tiene que ser mayor que cero.'
		);
		expect(firstMessage(cashRateCreateSchema.safeParse(form({ annualRatePct: '100.5' })))).toBe(
			'La tasa no puede pasar de 100 %.'
		);
	});

	// Los que caben en la base: cuatro para la tasa y dos para la retención.
	it('acepta los decimales que guarda el backend y ninguno más', () => {
		expect(cashRateCreateSchema.safeParse(form({ annualRatePct: '9.2525' })).success).toBe(true);
		expect(firstMessage(cashRateCreateSchema.safeParse(form({ annualRatePct: '9.25251' })))).toBe(
			'Escribe la tasa con hasta cuatro decimales.'
		);
		expect(firstMessage(cashRateCreateSchema.safeParse(form({ withholdingPct: '7.125' })))).toBe(
			'Escribe la retención con hasta dos decimales.'
		);
	});

	it('no deja retener todos los intereses', () => {
		expect(firstMessage(cashRateCreateSchema.safeParse(form({ withholdingPct: '100' })))).toBe(
			'La retención tiene que ser menor que 100 %.'
		);
	});

	it('pide el día desde el que rige', () => {
		expect(firstMessage(cashRateCreateSchema.safeParse(form({ effectiveFrom: null })))).toBe(
			'Elige desde qué día rige la tasa.'
		);
	});
});

describe('cashRateUpdateSchema', () => {
	it('corrige los valores de una versión, sin fechas', () => {
		const result = cashRateUpdateSchema.safeParse({
			id: RATE,
			annualRatePct: '8.5',
			withholdingPct: '7'
		});

		// Omitir el abono y el tope es abono diario y sin tope, que es lo que
		// dice una versión que no los menciona.
		expect(result.data).toEqual({
			id: RATE,
			annualRatePct: 8.5,
			withholdingPct: 7,
			posting: 'daily',
			maxBalance: null
		});
	});

	it('acepta el abono mensual y un tope', () => {
		const result = cashRateUpdateSchema.safeParse({
			id: RATE,
			annualRatePct: '9',
			withholdingPct: '',
			posting: 'monthly',
			maxBalance: '25000000'
		});

		expect(result.data?.posting).toBe('monthly');
		expect(result.data?.maxBalance).toBe(25000000);
	});

	it('rechaza un tope que no es un importe', () => {
		const cap = (maxBalance: string) =>
			cashRateUpdateSchema.safeParse({
				id: RATE,
				annualRatePct: '9',
				withholdingPct: '',
				maxBalance
			}).success;

		expect(cap('')).toBe(true);
		expect(cap('0')).toBe(false);
		expect(cap('-1')).toBe(false);
		expect(cap('hola')).toBe(false);
		expect(cap('1.000000001')).toBe(false);
	});

	it('rechaza un abono que no conoce', () => {
		expect(
			cashRateUpdateSchema.safeParse({
				id: RATE,
				annualRatePct: '9',
				withholdingPct: '',
				posting: 'weekly'
			}).success
		).toBe(false);
	});

	it('pide saber qué versión corregir', () => {
		expect(
			cashRateUpdateSchema.safeParse({ id: 'x', annualRatePct: '8', withholdingPct: '' }).success
		).toBe(false);
	});
});

describe('cashRateEndSchema', () => {
	it('pide el primer día sin rentabilidad', () => {
		expect(cashRateEndSchema.safeParse({ id: RATE, endsOn: '2026-09-30' }).success).toBe(true);
		expect(firstMessage(cashRateEndSchema.safeParse({ id: RATE, endsOn: '' }))).toBe(
			'Elige desde qué día deja de rendir.'
		);
	});
});

describe('cashRateDeleteSchema', () => {
	it('pide un id válido', () => {
		expect(cashRateDeleteSchema.safeParse({ id: 'x' }).success).toBe(false);
	});
});

describe('toCashRateBody', () => {
	it('manda los porcentajes con el abono elegido, y el día a medianoche UTC', () => {
		expect(
			toCashRateBody({ annualRatePct: 9.25, withholdingPct: 0, posting: 'daily', maxBalance: null })
		).toEqual({
			annualRatePct: 9.25,
			withholdingPct: 0,
			posting: 'daily',
			maxBalance: null
		});
		expect(toCalendarDateTime('2026-09-15')).toBe('2026-09-15T00:00:00Z');
	});

	// Una corrección dice la versión entera: un tope que se deja vacío lo quita.
	it('manda el tope y el abono mensual tal cual', () => {
		expect(
			toCashRateBody({
				annualRatePct: 12,
				withholdingPct: 7,
				posting: 'monthly',
				maxBalance: 25000000
			})
		).toEqual({
			annualRatePct: 12,
			withholdingPct: 7,
			posting: 'monthly',
			maxBalance: 25000000
		});
	});
});
