import { error } from '@sveltejs/kit';
import { features } from '$/config/features';
import type { LayoutLoad } from './$types';

/**
 * Guards every `/dashboard/investments` route behind the `investments` feature
 * flag. While the flag is off the section is treated as non-existent (404) so
 * the unreleased feature stays unreachable even via direct URL.
 */
export const load: LayoutLoad = () => {
	if (!features.investments) {
		error(404, 'Página no encontrada');
	}
};
