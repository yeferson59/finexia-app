import type { PageServerLoad } from './$types';
import * as portfolio from '$lib/api/portfolio';
import {
	buildGrowthProjection,
	buildKeyStatistics,
	buildPerformanceCalendars,
	buildRecordSummary,
	historySpanDays,
	type GrowthProjectionSeries,
	type KeyStat,
	type PerformanceCalendar,
	type RecordSummary
} from '$lib/features/reports';
import type { GrowthDataPoint, GrowthSummary } from '$lib/api/types';

export const load: PageServerLoad = async ({ cookies, fetch }) => {
	const empty = {
		/**
		 * La página en blanco decía lo mismo cuando el historial estaba vacío que
		 * cuando la petición se cayó, y no son lo mismo: uno invita a registrar
		 * una posición y el otro tiene que decir que el fallo no es del usuario.
		 */
		failed: false,
		record: null as RecordSummary | null,
		performanceCalendars: [] as PerformanceCalendar[],
		keyStatistics: [] as KeyStat[],
		growthProjection: null as GrowthProjectionSeries | null,
		historyDays: 0
	};

	const growthRes = await portfolio.getAggregateGrowth({ cookies, fetch }, { period: 'ALL' });

	if (!growthRes.ok || !growthRes.success || !growthRes.data) return { ...empty, failed: true };

	const data = growthRes.data;
	const points: GrowthDataPoint[] = Array.isArray(data.points) ? data.points : [];
	const summary: GrowthSummary = data.summary ?? {
		firstDate: '',
		initialValue: '0',
		currentValue: '0',
		totalGrowthPct: '0'
	};

	return {
		failed: false,
		record: buildRecordSummary(points, summary),
		performanceCalendars: buildPerformanceCalendars(points),
		keyStatistics: buildKeyStatistics(points),
		growthProjection: buildGrowthProjection(points, summary),
		// La proyección lo usa para decir cuánto historial falta.
		historyDays: historySpanDays(points)
	};
};
