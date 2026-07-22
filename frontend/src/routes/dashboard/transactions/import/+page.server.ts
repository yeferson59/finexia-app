import type { PageServerLoad } from './$types';
import * as portfolio from '$lib/api/portfolio';
import * as platforms from '$lib/api/platforms';

export const load: PageServerLoad = async ({ cookies, fetch }) => {
	const event = { cookies, fetch };

	const [portfoliosRes, platformsRes] = await Promise.all([
		portfolio.getSummaries(event),
		platforms.getSources(event)
	]);

	const portfolios =
		portfoliosRes.ok && portfoliosRes.success && Array.isArray(portfoliosRes.data)
			? portfoliosRes.data
			: [];

	const platformList =
		platformsRes.ok && platformsRes.success && Array.isArray(platformsRes.data)
			? platformsRes.data.filter((p) => p.isActive !== false)
			: [];

	return { portfolios, platforms: platformList };
};
