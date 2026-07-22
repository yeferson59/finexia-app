import type { Actions, PageServerLoad } from './$types';
import { redirect } from '@sveltejs/kit';
import * as auth from '$lib/api/auth';
import * as portfolio from '$lib/api/portfolio';
import * as transactions from '$lib/api/transactions';
import { ACCESS_COOKIE, REFRESH_COOKIE, clearSessionCookies } from '$lib/server/session';
import type {
	AllocationItem,
	PortfolioGrowth,
	PortfolioSummary,
	UserTransaction
} from '$lib/api/types';

const SUPPORTED_CURRENCIES = ['USD', 'COP'];

export const load: PageServerLoad = async ({ cookies, fetch, url }) => {
	const event = { cookies, fetch };

	const requestedCurrency = url.searchParams.get('currency')?.toUpperCase() ?? '';
	const currency = SUPPORTED_CURRENCIES.includes(requestedCurrency) ? requestedCurrency : 'USD';

	const [transactionsRes, summaryRes, allocationRes, growthRes] = await Promise.all([
		transactions.getRecent(event),
		portfolio.getSummaries(event, currency),
		portfolio.getAllocation(event),
		portfolio.getAggregateGrowth(event)
	]);

	const recentTransactions: UserTransaction[] =
		transactionsRes.ok && transactionsRes.success && Array.isArray(transactionsRes.data)
			? transactionsRes.data.slice(0, 5)
			: [];

	const portfolioSummaries: PortfolioSummary[] =
		summaryRes.ok && summaryRes.success && Array.isArray(summaryRes.data) ? summaryRes.data : [];

	const allocation: AllocationItem[] =
		allocationRes.ok && allocationRes.success && Array.isArray(allocationRes.data)
			? allocationRes.data
			: [];

	let portfolioGrowth: PortfolioGrowth = {
		points: [],
		summary: { firstDate: '', initialValue: '0', currentValue: '0', totalGrowthPct: '0' }
	};
	if (growthRes.ok && growthRes.success && growthRes.data) portfolioGrowth = growthRes.data;

	return { recentTransactions, portfolioSummaries, allocation, portfolioGrowth, currency };
};

export const actions = {
	logout: async ({ cookies, fetch }) => {
		const token = cookies.get(ACCESS_COOKIE);

		if (!token) return { success: false };

		const refreshToken = cookies.get(REFRESH_COOKIE);

		await auth.logout(fetch, {
			accessToken: token,
			refreshToken,
			refreshCookieName: REFRESH_COOKIE
		});

		clearSessionCookies(cookies);

		return redirect(302, '/auth');
	}
} satisfies Actions;
