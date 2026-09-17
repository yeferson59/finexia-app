import { describe, it, expect } from 'vitest';
import { HOLDINGS_TOTAL, formatUsd, layoutHoldings, layoutScale, type Holding } from './holdings';
import { TOUR_BREAKDOWN, TOUR_PORTFOLIOS } from './product-tour';

const dollars = (text: string) => Number(text.replace(/[$,]/g, ''));

describe('posiciones del mapa del hero', () => {
	it('cierran en el patrimonio del resumen', () => {
		expect(HOLDINGS_TOTAL).toBe(248500);
	});

	it('cada plataforma suma lo que dice el reparto del recorrido', () => {
		const { groups } = layoutHoldings('platform');
		expect(groups.map((g) => g.label)).toEqual(TOUR_BREAKDOWN.map((r) => r.name));
		TOUR_BREAKDOWN.forEach((row, i) => {
			expect(groups[i].value).toBe(dollars(row.value));
			expect(groups[i].count).toBe(Number.parseInt(row.detail, 10));
		});
	});

	it('cada portafolio suma lo que dice su fila del listado', () => {
		const { groups } = layoutHoldings('portfolio');
		expect(groups.map((g) => g.label)).toEqual(TOUR_PORTFOLIOS.map((p) => p.name));
		TOUR_PORTFOLIOS.forEach((portfolio, i) => {
			expect(groups[i].value).toBe(portfolio.value);
			expect(portfolio.detail).toContain(`${groups[i].count} posiciones`);
		});
	});

	it('los tipos reparten el mismo total', () => {
		const { groups } = layoutHoldings('type');
		expect(groups.reduce((sum, g) => sum + g.value, 0)).toBe(HOLDINGS_TOTAL);
	});
});

describe('layoutHoldings', () => {
	const sample: Holding[] = [
		{ short: 'A', name: 'A', platform: 'Degiro', portfolio: 'Reserva', type: 'ETFs', value: 10 },
		{ short: 'B', name: 'B', platform: 'Degiro', portfolio: 'Jubilación', type: 'ETFs', value: 20 },
		{ short: 'C', name: 'C', platform: 'Banco', portfolio: 'Jubilación', type: 'Bonos', value: 40 }
	];

	it('mide cada bloque sobre el grupo más grande de todas las vistas', () => {
		// Jubilación (60) es el mayor grupo de las tres vistas.
		expect(layoutScale(sample)).toBe(60);
		const { blocks } = layoutHoldings('platform', sample);
		expect(blocks.find((b) => b.short === 'C')?.w).toBeCloseTo(40 / 60);
	});

	it('ordena los bloques de una fila por portafolio y los coloca uno tras otro', () => {
		const { blocks } = layoutHoldings('platform', sample);
		const degiro = blocks.filter((b) => b.row === 0);
		expect(degiro.map((b) => b.short)).toEqual(['B', 'A']);
		expect(degiro[0].x).toBe(0);
		expect(degiro[1].x).toBeCloseTo(20 / 60);
	});

	it('omite las filas vacías y conserva la clave de cada bloque', () => {
		const { groups, blocks } = layoutHoldings('platform', sample);
		expect(groups.map((g) => g.label)).toEqual(['Degiro', 'Banco']);
		expect(blocks.find((b) => b.short === 'C')).toMatchObject({ id: 2, row: 1 });
	});
});

describe('formatUsd', () => {
	it('escribe el dólar como lo lee en-US, sin céntimos', () => {
		expect(formatUsd(248500)).toBe('$248,500');
		expect(formatUsd(3800.4)).toBe('$3,800');
	});
});
