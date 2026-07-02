import type { Actions, PageServerLoad } from './$types';
import { env } from '$env/dynamic/private';
import { redirect } from '@sveltejs/kit';
import { authedFetchSafe } from '$lib/server/api';
import { ACCESS_COOKIE, REFRESH_COOKIE, clearSessionCookies } from '$lib/server/session';

interface PortfolioSummary {
	id: string;
	name: string;
	type: string;
	baseCurrency: string;
	totalPositions: number;
	totalCostBase: string;
	totalMarketValue: string;
	totalGainLoss: string;
	totalGainLossPct: string;
}

interface AllocationItem {
	category: string;
	marketValue: string;
	percent: number;
}

interface Transaction {
	id: string;
	entryId: string;
	type: string;
	quantity: string;
	price: string;
	currency: string;
	fees: string;
	transactionDate: string;
	notes: string;
	createdAt: string;
	assetTicker: string;
	assetName: string;
}

interface GrowthDataPoint {
	date: string;
	totalValue: string;
	totalCostBase: string;
	gainLoss: string;
	gainLossPct: string;
}

interface GrowthSummary {
	firstDate: string;
	initialValue: string;
	currentValue: string;
	totalGrowthPct: string;
}

export const load: PageServerLoad = async ({ cookies, fetch }) => {
	const event = { cookies, fetch };

	const [transactionsRes, summaryRes, allocationRes, growthRes] = await Promise.all([
		authedFetchSafe(event, '/portfolios/transactions'),
		authedFetchSafe(event, '/portfolios/summary'),
		authedFetchSafe(event, '/portfolios/allocation'),
		authedFetchSafe(event, '/portfolios/growth')
	]);

	let recentTransactions: Transaction[] = [];
	if (transactionsRes?.ok) {
		const { data, success } = await transactionsRes.json();
		if (success && Array.isArray(data)) recentTransactions = data.slice(0, 5);
	}

	let portfolioSummaries: PortfolioSummary[] = [];
	if (summaryRes?.ok) {
		const { data, success } = await summaryRes.json();
		if (success && Array.isArray(data)) portfolioSummaries = data;
	}

	let allocation: AllocationItem[] = [];
	if (allocationRes?.ok) {
		const { data, success } = await allocationRes.json();
		if (success && Array.isArray(data)) allocation = data;
	}

	let portfolioGrowth: { points: GrowthDataPoint[]; summary: GrowthSummary } = {
		points: [],
		summary: { firstDate: '', initialValue: '0', currentValue: '0', totalGrowthPct: '0' }
	};
	if (growthRes?.ok) {
		const { data, success } = await growthRes.json();
		if (success && data) portfolioGrowth = data;
	}

	return { recentTransactions, portfolioSummaries, allocation, portfolioGrowth };
};

export const actions = {
	logout: async ({ cookies, fetch }) => {
		const token = cookies.get(ACCESS_COOKIE);

		if (!token) return { success: false };

		const refreshToken = cookies.get(REFRESH_COOKIE);

		await fetch(`${env.BASE_API}/auth/logout`, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json',
				Authorization: `Bearer ${token}`,
				...(refreshToken ? { Cookie: `${REFRESH_COOKIE}=${refreshToken}` } : {})
			}
		});

		clearSessionCookies(cookies);

		return redirect(302, '/auth');
	}
} satisfies Actions;
