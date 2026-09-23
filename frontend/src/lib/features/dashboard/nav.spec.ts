import { describe, expect, it } from 'vitest';
import { activeItem, sectionTitle } from './nav';

describe('activeItem', () => {
	it('marca solo la entrada más concreta', () => {
		expect(activeItem('/dashboard/admin/users')?.label).toBe('Usuarios');
		expect(activeItem('/dashboard/admin/contributions')?.label).toBe('Aportes');
		expect(activeItem('/dashboard/admin')?.label).toBe('Panel de administración');
	});

	it('el resumen solo en su propia página, no en las hijas', () => {
		expect(activeItem('/dashboard')?.label).toBe('Resumen');
		expect(activeItem('/dashboard/portfolios/p1')?.label).toBe('Portafolios');
	});

	it('nada fuera del menú', () => {
		expect(activeItem('/apoyar')).toBeUndefined();
		expect(sectionTitle('/apoyar')).toBe('Panel');
	});

	it('la cabecera dice lo mismo que el menú', () => {
		expect(sectionTitle('/dashboard/admin/exchange-rates')).toBe('Tasas de cambio');
	});
});
