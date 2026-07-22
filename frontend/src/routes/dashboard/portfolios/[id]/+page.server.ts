import z from 'zod';
import { error, fail } from '@sveltejs/kit';
import * as portfolio from '$lib/api/portfolio';
import type { Actions, PageServerLoad } from './$types';
import type { PortfolioGrowth, Risk, TopTransaction } from '$lib/api/types';

export const load: PageServerLoad = async ({ cookies, fetch, params }) => {
	const event = { cookies, fetch };

	const [portfolioRes, risksRes, topTxRes, growthRes] = await Promise.all([
		portfolio.getPortfolio(event, params.id),
		portfolio.getRisks(event),
		portfolio.getTopTransaction(event, params.id),
		portfolio.getPortfolioGrowth(event, params.id)
	]);

	if (!portfolioRes.ok) {
		if (portfolioRes.status === 404) {
			error(404, 'Portafolio no encontrado');
		}
		return { portfolio: null, risks: [], topTransaction: null, growth: null };
	}

	if (!portfolioRes.success || !portfolioRes.data) {
		return { portfolio: null, risks: [], topTransaction: null, growth: null };
	}

	const risks: Risk[] = risksRes.ok ? (risksRes.data ?? []) : [];

	let topTransaction: TopTransaction | null = null;
	if (topTxRes.ok) {
		const tx = topTxRes.data;
		topTransaction = tx?.assetTicker ? tx : null;
	}

	let growth: PortfolioGrowth | null = null;
	if (growthRes.ok && growthRes.success && growthRes.data) growth = growthRes.data;

	return { portfolio: portfolioRes.data, risks, topTransaction, growth };
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

		const res = await portfolio.updatePortfolio({ cookies, fetch }, params.id, data);

		if (!res.ok) {
			return fail(400, {
				action: 'updatePortfolio',
				success: false,
				error: 'Error al actualizar el portafolio'
			});
		}

		return { action: 'updatePortfolio', success: res.success };
	}
};
