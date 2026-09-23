import type { PageServerLoad } from './$types';
import * as support from '$lib/api/support';
import type { PageMeta } from '$lib/api/types';
import { parseContributionStatus } from '$lib/features/admin';

const DEFAULT_META: PageMeta = { currentPage: 1, totalPages: 1, previous: false, next: false };

export const load: PageServerLoad = async ({ cookies, fetch, url }) => {
	const event = { cookies, fetch };
	const page = Math.max(1, Number(url.searchParams.get('page')) || 1);
	const status = parseContributionStatus(url.searchParams.get('status'));

	const [listRes, summaryRes] = await Promise.all([
		support.getAdminContributions(event, { page, limit: 25, status }),
		support.getSupportSummary(event)
	]);

	return {
		contributions: listRes.success ? (listRes.data?.items ?? []) : [],
		meta: listRes.success ? (listRes.data?.metaData ?? DEFAULT_META) : DEFAULT_META,
		summary: summaryRes.success ? summaryRes.data : null,
		status,
		unavailable: !listRes.ok
	};
};
