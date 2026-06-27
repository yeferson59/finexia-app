import type { LayoutServerLoad } from './$types';
import { isRedirect } from '@sveltejs/kit';
import { authedFetch } from '$lib/server/api';

interface Asset {
	id: string;
	ticker: string;
	name: string;
	assetType: string;
	exchange: string;
	currency: string;
	currentPrice: { value: string; currency: string } | null;
	createdAt: string;
	updatedAt: string;
}

interface Platform {
	id: string;
	name: string;
}

export const load: LayoutServerLoad = async ({ cookies, fetch }) => {
	try {
		const [platformsRes, assetsRes] = await Promise.all([
			authedFetch({ cookies, fetch }, '/portfolios/sources'),
			authedFetch({ cookies, fetch }, '/portfolios/assets')
		]);

		const platformsData = await platformsRes.json();
		const assetsData = await assetsRes.json();

		const platforms: Platform[] = platformsData.success ? platformsData.data : [];
		const assets: Asset[] = assetsData.success ? assetsData.data : [];

		return { platforms, assets };
	} catch (error) {
		// A 401 redirect must not be swallowed by the graceful fallback below.
		if (isRedirect(error)) throw error;
		console.error('Error loading platforms or assets:', error);
		return { platforms: [], assets: [] };
	}
};
