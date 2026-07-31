import type { PageServerLoad } from './$types';
import * as portfolio from '$lib/api/portfolio';
import {
	buildGrowthProjection,
	buildKeyStatistics,
	buildPerformanceCalendars,
	type GrowthProjectionEntry,
	type KeyStat,
	type PerformanceCalendar
} from '$lib/features/reports';
import type { GrowthDataPoint, GrowthSummary } from '$lib/api/types';

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
