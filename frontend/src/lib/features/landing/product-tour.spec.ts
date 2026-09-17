import { describe, it, expect } from 'vitest';
import { TOUR_ICONS, TOUR_NAV, TOUR_VIEWS } from './product-tour';

describe('menú de la maqueta del panel', () => {
	it('tiene un icono para cada sección', () => {
		for (const item of TOUR_NAV) {
			expect(TOUR_ICONS[item.icon], item.label).toBeDefined();
		}
	});

	it('marca en cada vista una sección que existe en el menú', () => {
		const labels = TOUR_NAV.map((item) => item.label);
		for (const view of TOUR_VIEWS) {
			expect(labels).toContain(view.nav);
		}
	});
});
