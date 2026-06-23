import { env } from '$env/dynamic/private';
import type { LayoutServerLoad } from './$types';

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
	const accessToken = cookies.get('access_token_finexia');

	const response = await fetch(`${env.BASE_API}/portfolios/summary`, {
		headers: { Authorization: `Bearer ${accessToken}` }
	});

	const { data, success } = await response.json();

	if (!success) {
		return { portfolios: [] as PortfolioSummary[], success: false };
	}

	return { portfolios: data as PortfolioSummary[], success: true };
};
