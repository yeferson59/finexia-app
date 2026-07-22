import type { PageServerLoad } from './$types';
import * as user from '$lib/api/user';
import * as market from '$lib/api/market';
import type { Asset, ExchangeRate } from '$lib/api/types';

export const load: PageServerLoad = async ({ cookies, fetch }) => {
	const event = { cookies, fetch };

	const [usersRes, assetsRes, ratesRes] = await Promise.all([
		user.getUsers(event, { page: 1, limit: 1 }),
		market.getAssets(event, { page: 1, limit: 100 }),
		market.getExchangeRates(event, { page: 1, limit: 100 })
	]);

	let totalUsers = 0;
	if (usersRes.ok && usersRes.success && usersRes.data?.metaData) {
		totalUsers = Number(usersRes.data.metaData.totalUsers ?? 0);
	}

	let assets: Asset[] = [];
	if (assetsRes.ok && assetsRes.success && Array.isArray(assetsRes.data)) assets = assetsRes.data;

	let rates: ExchangeRate[] = [];
	if (ratesRes.ok && ratesRes.success && Array.isArray(ratesRes.data)) rates = ratesRes.data;

	const lastSync = assets.reduce<string | null>((latest, a) => {
		if (!a.priceUpdatedAt) return latest;
		if (!latest || a.priceUpdatedAt > latest) return a.priceUpdatedAt;
		return latest;
	}, null);

	return { totalUsers, totalAssets: assets.length, totalRates: rates.length, lastSync };
};
