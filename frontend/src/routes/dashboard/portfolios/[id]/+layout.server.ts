import type { LayoutServerLoad } from './$types';
import { isRedirect } from '@sveltejs/kit';
import { authedFetch } from '$lib/server/api';

interface Platform {
	id: string;
	name: string;
}

export const load: LayoutServerLoad = async ({ cookies, fetch }) => {
	try {
		const platformsRes = await authedFetch({ cookies, fetch }, '/portfolios/sources');
		const platformsData = await platformsRes.json();
		const platforms: Platform[] = platformsData.success ? platformsData.data : [];
		return { platforms };
	} catch (error) {
		if (isRedirect(error)) throw error;
		console.error('Error loading platforms:', error);
		return { platforms: [] };
	}
};
