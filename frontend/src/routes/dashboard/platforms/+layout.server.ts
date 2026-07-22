import type { LayoutServerLoad } from './$types';
import * as platforms from '$lib/api/platforms';

export const load: LayoutServerLoad = async ({ cookies, fetch }) => {
	const res = await platforms.getSources({ cookies, fetch });

	if (!res.success) {
		return { platforms: [] };
	}

	return { platforms: res.data ?? [] };
};
