import type { PageServerLoad } from './$types';
import { authedFetchSafe } from '$lib/server/api';

export interface PortfolioOption {
	id: string;
	name: string;
	baseCurrency: string;
	isDefault: boolean;
}

export interface PlatformOption {
	id: string;
	name: string;
	isActive: boolean;
}

export const load: PageServerLoad = async ({ cookies, fetch }) => {
	const [portfoliosRes, platformsRes] = await Promise.all([
		authedFetchSafe({ cookies, fetch }, '/portfolios/summary'),
		authedFetchSafe({ cookies, fetch }, '/portfolios/sources')
	]);

	let portfolios: PortfolioOption[] = [];
	if (portfoliosRes?.ok) {
		const { data, success } = await portfoliosRes.json();
		if (success && Array.isArray(data)) portfolios = data;
	}

	let platforms: PlatformOption[] = [];
	if (platformsRes?.ok) {
		const { data, success } = await platformsRes.json();
		if (success && Array.isArray(data)) {
			platforms = data.filter((p: PlatformOption) => p.isActive !== false);
		}
	}

	return { portfolios, platforms };
};
