import type { LayoutServerLoad } from './$types';
import { authedFetch } from '$lib/server/api';

export interface PortfolioSummary {
	id: string;
	name: string;
	description: string;
	type: string;
	baseCurrency: string;
	isDefault: boolean;
	riskId: string;
	riskName: string;
	totalPositions: number;
	totalCostBase: string;
	totalMarketValue: string;
	totalGainLoss: string;
	totalGainLossPct: string;
	createdAt: string;
}

export const load: LayoutServerLoad = async ({ cookies, fetch }) => {
	const response = await authedFetch({ cookies, fetch }, '/portfolios/summary');

	const { data, success } = await response.json();

	if (!success) {
		return { portfolios: [] as PortfolioSummary[], success: false };
	}

	return { portfolios: data as PortfolioSummary[], success: true };
};
