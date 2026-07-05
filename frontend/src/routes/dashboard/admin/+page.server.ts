import type { PageServerLoad } from './$types';
import { authedFetchSafe } from '$lib/server/api';

interface Asset {
	id: string;
	ticker: string;
	name: string;
	currentPrice: { value: string; currency: string } | null;
	priceUpdatedAt: string | null;
}

interface ExchangeRate {
	id: string;
	fromCurrency: string;
	toCurrency: string;
	rate: string;
	rateDate: string;
	createdAt: string;
}

export const load: PageServerLoad = async ({ cookies, fetch }) => {
	const event = { cookies, fetch };

	const [usersRes, assetsRes, ratesRes] = await Promise.all([
		authedFetchSafe(event, '/users?page=1&limit=1'),
		authedFetchSafe(event, '/portfolios/assets?page=1&limit=100'),
		authedFetchSafe(event, '/exchange-rates?page=1&limit=100')
	]);

	let totalUsers = 0;
	if (usersRes?.ok) {
		const { data, success } = await usersRes.json();
		if (success && data?.metaData) totalUsers = data.metaData.totalUsers ?? 0;
	}

	let assets: Asset[] = [];
	if (assetsRes?.ok) {
		const { data, success } = await assetsRes.json();
		if (success && Array.isArray(data)) assets = data;
	}

	let rates: ExchangeRate[] = [];
	if (ratesRes?.ok) {
		const { data, success } = await ratesRes.json();
		if (success && Array.isArray(data)) rates = data;
	}

	const lastSync = assets.reduce<string | null>((latest, a) => {
		if (!a.priceUpdatedAt) return latest;
		if (!latest || a.priceUpdatedAt > latest) return a.priceUpdatedAt;
		return latest;
	}, null);

	return { totalUsers, totalAssets: assets.length, totalRates: rates.length, lastSync };
};
