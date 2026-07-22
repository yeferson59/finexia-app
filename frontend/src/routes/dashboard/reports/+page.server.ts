import type { PageServerLoad } from './$types';
import * as portfolio from '$lib/api/portfolio';
import type { GrowthDataPoint, GrowthSummary } from '$lib/api/types';

export interface PerformanceCalendar {
	year: string;
	values: (number | null)[];
}

export interface KeyStat {
	label: string;
	value: string;
}

export interface GrowthProjectionEntry {
	period: string;
	value: number;
}

function buildPerformanceCalendars(points: GrowthDataPoint[]): PerformanceCalendar[] {
	if (points.length === 0) return [];

	const byYearMonth = new Map<string, GrowthDataPoint>();
	for (const point of points) {
		byYearMonth.set(point.date.substring(0, 7), point);
	}

	const monthEntries = [...byYearMonth.entries()].sort(([a], [b]) => a.localeCompare(b));
	const byYear = new Map<string, (number | null)[]>();

	for (let i = 0; i < monthEntries.length; i++) {
		const [key] = monthEntries[i];
		const [year, monthStr] = key.split('-');
		const monthIndex = parseInt(monthStr, 10) - 1;

		if (!byYear.has(year)) byYear.set(year, Array(12).fill(null));

		if (i === 0) {
			byYear.get(year)![monthIndex] = null;
		} else {
			const prevVal = parseFloat(monthEntries[i - 1][1].totalValue);
			const currVal = parseFloat(monthEntries[i][1].totalValue);
			byYear.get(year)![monthIndex] =
				prevVal > 0 ? parseFloat(((currVal / prevVal - 1) * 100).toFixed(2)) : null;
		}
	}

	return [...byYear.entries()]
		.sort(([a], [b]) => b.localeCompare(a))
		.map(([year, values]) => ({ year, values }));
}

function buildKeyStatistics(points: GrowthDataPoint[]): KeyStat[] {
	if (points.length === 0) return [];

	let peak = -Infinity;
	let maxDrawdown = 0;
	for (const p of points) {
		const v = parseFloat(p.totalValue);
		if (v > peak) peak = v;
		if (peak > 0) {
			const dd = ((v - peak) / peak) * 100;
			if (dd < maxDrawdown) maxDrawdown = dd;
		}
	}

	const byYearMonth = new Map<string, GrowthDataPoint>();
	for (const p of points) byYearMonth.set(p.date.substring(0, 7), p);

	const monthEntries = [...byYearMonth.entries()].sort(([a], [b]) => a.localeCompare(b));
	const returns: number[] = [];
	for (let i = 1; i < monthEntries.length; i++) {
		const prev = parseFloat(monthEntries[i - 1][1].totalValue);
		const curr = parseFloat(monthEntries[i][1].totalValue);
		if (prev > 0) returns.push((curr / prev - 1) * 100);
	}

	let volatilityStr = 'N/A';
	if (returns.length >= 3) {
		const mean = returns.reduce((s, r) => s + r, 0) / returns.length;
		const variance = returns.reduce((s, r) => s + (r - mean) ** 2, 0) / returns.length;
		volatilityStr = `${(Math.sqrt(variance) * Math.sqrt(12)).toFixed(1)}%`;
	}

	return [
		{ label: 'Max Drawdown', value: `${maxDrawdown.toFixed(1)}%` },
		{ label: 'Volatilidad', value: volatilityStr }
	];
}

function buildGrowthProjection(
	points: GrowthDataPoint[],
	summary: GrowthSummary
): GrowthProjectionEntry[] {
	const initialValue = parseFloat(summary.initialValue);
	const currentValue = parseFloat(summary.currentValue);

	if (!points.length || initialValue <= 0 || currentValue <= 0) return [];

	const firstDate = new Date(summary.firstDate + 'T00:00:00');
	const lastDate = new Date(points[points.length - 1].date + 'T00:00:00');
	const years = (lastDate.getTime() - firstDate.getTime()) / (365.25 * 86_400_000);

	if (years < 0.5) return [];

	const cagr = Math.pow(currentValue / initialValue, 1 / years) - 1;
	if (!isFinite(cagr) || cagr < -0.5 || cagr > 2.0) return [];

	const startYear = lastDate.getFullYear();
	return Array.from({ length: 5 }, (_, i) => ({
		period: String(startYear + i),
		value: Math.round(currentValue * Math.pow(1 + cagr, i))
	}));
}

export const load: PageServerLoad = async ({ cookies, fetch }) => {
	const empty = {
		performanceCalendars: [] as PerformanceCalendar[],
		keyStatistics: [] as KeyStat[],
		growthProjection: [] as GrowthProjectionEntry[]
	};

	const growthRes = await portfolio.getAggregateGrowth({ cookies, fetch }, { period: 'ALL' });

	if (!growthRes.ok || !growthRes.success || !growthRes.data) return empty;

	const data = growthRes.data;
	const points: GrowthDataPoint[] = Array.isArray(data.points) ? data.points : [];
	const summary: GrowthSummary = data.summary ?? {
		firstDate: '',
		initialValue: '0',
		currentValue: '0',
		totalGrowthPct: '0'
	};

	return {
		performanceCalendars: buildPerformanceCalendars(points),
		keyStatistics: buildKeyStatistics(points),
		growthProjection: buildGrowthProjection(points, summary)
	};
};
