import type { Actions, PageServerLoad } from './$types';
import { env } from '$env/dynamic/private';
import { redirect } from '@sveltejs/kit';

export const load: PageServerLoad = async ({ cookies, fetch }) => {
	const accessToken = cookies.get('access_token_finexia');
	const headers = { Authorization: `Bearer ${accessToken}` };

	const [transactionsRes, summaryRes, allocationRes] = await Promise.all([
		fetch(`${env.BASE_API}/portfolios/transactions`, { headers }).catch(() => null),
		fetch(`${env.BASE_API}/portfolios/summary`, { headers }).catch(() => null),
		fetch(`${env.BASE_API}/portfolios/allocation`, { headers }).catch(() => null)
	]);

	let recentTransactions: unknown[] = [];
	if (transactionsRes?.ok) {
		const { data, success } = await transactionsRes.json();
		if (success && Array.isArray(data)) recentTransactions = data.slice(0, 5);
	}

	let portfolioSummaries: unknown[] = [];
	if (summaryRes?.ok) {
		const { data, success } = await summaryRes.json();
		if (success && Array.isArray(data)) portfolioSummaries = data;
	}

	let allocation: unknown[] = [];
	if (allocationRes?.ok) {
		const { data, success } = await allocationRes.json();
		if (success && Array.isArray(data)) allocation = data;
	}

	return { recentTransactions, portfolioSummaries, allocation };
};

export const actions = {
	logout: async ({ cookies, fetch }) => {
		const token = cookies.get('access_token_finexia');

		if (!token) return { success: false };

		const response = await fetch(`${env.BASE_API}/auth/logout`, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json',
				Authorization: `Bearer ${token}`
			}
		});

		if (!response.ok) {
			return { success: false };
		}

		const { success } = await response.json();

		if (!success) {
			return { success: false };
		}

		cookies.delete('access_token_finexia', { path: '/' });

		return redirect(302, '/auth');
	}
} satisfies Actions;
