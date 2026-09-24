import { describe, expect, it } from 'vitest';
import { FUND_CHART, fundChartGeometry, nearestPoint, niceTicks } from './chart';

describe('niceTicks', () => {
	it('covers the range with round steps', () => {
		expect(niceTicks(100, 101.0005)).toEqual([100, 100.5, 101, 101.5]);
		expect(niceTicks(12345.68, 12431.22)).toEqual([12325, 12350, 12375, 12400, 12425, 12450]);
	});

	it('opens a flat series around its value', () => {
		const ticks = niceTicks(100, 100);

		expect(ticks[0]).toBeLessThan(100);
		expect(ticks[ticks.length - 1]).toBeGreaterThan(100);
	});
});

describe('fundChartGeometry', () => {
	const points = [
		{ date: '2026-07-01', value: 100 },
		{ date: '2026-07-31', value: 100.8 },
		{ date: '2026-08-31', value: 101.0005305 },
		{ date: '2026-09-30', value: 100.53828551 }
	];

	it('places points by time, not by position', () => {
		const g = fundChartGeometry(points);
		const { padL, padR, width } = FUND_CHART;

		expect(g.x('2026-07-01')).toBe(padL);
		expect(g.x('2026-09-30')).toBe(width - padR);
		// 30 of 91 days in.
		expect(g.x('2026-07-31')).toBeCloseTo(padL + (30 / 91) * (width - padL - padR), 6);
	});

	it('spreads the date labels evenly in time', () => {
		expect(fundChartGeometry(points).xTicks).toEqual([
			'2026-07-01',
			'2026-07-31',
			'2026-08-31',
			'2026-09-30'
		]);
	});

	it('centers a lone point', () => {
		const g = fundChartGeometry([{ date: '2026-09-30', value: 100 }]);

		expect(g.x('2026-09-30')).toBe(
			FUND_CHART.padL + (FUND_CHART.width - FUND_CHART.padL - FUND_CHART.padR) / 2
		);
		expect(g.xTicks).toEqual(['2026-09-30']);
	});

	it('snaps the cursor to the nearest point', () => {
		const g = fundChartGeometry(points);

		expect(nearestPoint(points, g, g.x('2026-08-01'))).toBe(1);
		expect(nearestPoint(points, g, 9999)).toBe(3);
	});
});
