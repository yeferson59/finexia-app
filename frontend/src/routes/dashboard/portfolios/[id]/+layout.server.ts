import { env } from '$env/dynamic/private';
import type { LayoutServerLoad } from './$types';

interface Asset {
	id: string;
	ticker: string;
	name: string;
	assetType: string;
	exchange: string;
	currency: string;
	currentPrice: { value: string; currency: string } | null;
	createdAt: string;
	updatedAt: string;
}

interface Platform {
	id: string;
	name: string;
}

export const load: LayoutServerLoad = async ({ cookies, fetch }) => {
	const accessToken = cookies.get('access_token_finexia');

	try {
		const [platformsRes, assetsRes] = await Promise.all([
			fetch(`${env.BASE_API}/portfolios/sources`, {
				headers: {
					Authorization: `Bearer ${accessToken}`
				}
			}),
			fetch(`${env.BASE_API}/portfolios/assets`, {
				headers: {
					Authorization: `Bearer ${accessToken}`
				}
			})
		]);

		const platformsData = await platformsRes.json();
		const assetsData = await assetsRes.json();

		const platforms: Platform[] = platformsData.success ? platformsData.data : [];
		const assets: Asset[] = assetsData.success ? assetsData.data : [];

		return { platforms, assets };
	} catch (error) {
		console.error('Error loading platforms or assets:', error);
		return { platforms: [], assets: [] };
	}
};
