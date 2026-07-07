import z from 'zod';
import { error, fail } from '@sveltejs/kit';
import { authedFetch } from '$lib/server/api';
import type { Actions, PageServerLoad } from './$types';

interface Holding {
	id: string;
	assetId: string;
	ticker: string;
	name: string;
	assetType: string;
	exchange: string;
	currency: string;
	quantity: string;
	price: string;
	marketPrice: string;
	costCurrency: string;
	category: string;
	entryDate: string;
	notes: string;
}

interface PortfolioDetail {
	id: string;
	userId: string;
	name: string;
	description: string;
	type: string;
	baseCurrency: string;
	isDefault: boolean;
	riskId: string;
	riskName: string;
	createdAt: string;
	updatedAt: string;
	holdings: Holding[];
}

interface Risk {
	id: string;
	name: string;
	description: string;
}

interface TopTransaction {
	value: string;
	type: string;
	currency: string;
	assetTicker: string;
	assetName: string;
	transactionDate: string;
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

export const load: PageServerLoad = async ({ cookies, fetch, params }) => {
	const [portfolioRes, risksRes, topTxRes, growthRes] = await Promise.all([
		authedFetch({ cookies, fetch }, `/portfolios/${params.id}`),
		authedFetch({ cookies, fetch }, '/portfolios/risks'),
		authedFetch({ cookies, fetch }, `/portfolios/${params.id}/top-transaction`),
		authedFetch({ cookies, fetch }, `/portfolios/${params.id}/growth`)
	]);

	if (!portfolioRes.ok) {
		if (portfolioRes.status === 404) {
			error(404, 'Portafolio no encontrado');
		}
		return { portfolio: null, risks: [], topTransaction: null, growth: null };
	}

	const { data, success } = await portfolioRes.json();

	if (!success || !data) {
		return { portfolio: null, risks: [], topTransaction: null, growth: null };
	}

	let risks: Risk[] = [];
	if (risksRes.ok) {
		const risksJson = await risksRes.json();
		risks = risksJson.data ?? [];
	}

	let topTransaction: TopTransaction | null = null;
	if (topTxRes.ok) {
		const topTxJson = await topTxRes.json();
		const tx = topTxJson.data as TopTransaction;
		topTransaction = tx?.assetTicker ? tx : null;
	}

	let growth: { points: GrowthDataPoint[]; summary: GrowthSummary } | null = null;
	if (growthRes.ok) {
		const growthJson = await growthRes.json();
		if (growthJson.success && growthJson.data) growth = growthJson.data;
	}

	return { portfolio: data as PortfolioDetail, risks, topTransaction, growth };
};

export const actions: Actions = {
	updatePortfolio: async ({ request, cookies, fetch, params }) => {
		const formData = await request.formData();

		const {
			success,
			error: zodError,
			data
		} = await z
			.object({
				name: z.string().min(2, 'El nombre debe tener al menos 2 caracteres'),
				description: z.string().optional().default(''),
				type: z.string().min(1),
				riskId: z.string().uuid(),
				isDefault: z.coerce.boolean()
			})
			.safeParseAsync({
				name: formData.get('name'),
				description: formData.get('description'),
				type: formData.get('type'),
				riskId: formData.get('riskId'),
				isDefault: formData.get('isDefault')
			});

		if (!success) {
			return fail(400, { action: 'updatePortfolio', success: false, error: zodError.message });
		}

		const res = await authedFetch({ cookies, fetch }, `/portfolios/${params.id}`, {
			method: 'PATCH',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(data)
		});

		if (!res.ok) {
			return fail(400, {
				action: 'updatePortfolio',
				success: false,
				error: 'Error al actualizar el portafolio'
			});
		}

		const json = await res.json();
		return { action: 'updatePortfolio', success: json.success ?? false };
	}
};
