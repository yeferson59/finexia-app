import { env } from '$env/dynamic/private';
import { error } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

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

export const load: PageServerLoad = async ({ cookies, fetch, params }) => {
	const accessToken = cookies.get('access_token_finexia');

	const response = await fetch(`${env.BASE_API}/portfolios/${params.id}`, {
		headers: {
			Authorization: `Bearer ${accessToken}`
		}
	});

	if (!response.ok) {
		if (response.status === 404) {
			error(404, 'Portafolio no encontrado');
		}
		return { portfolio: null };
	}

	const { data, success } = await response.json();

	if (!success || !data) {
		return { portfolio: null };
	}

	return { portfolio: data as PortfolioDetail };
};
