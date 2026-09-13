import { describe, it, expect } from 'vitest';
import {
	describeSectorBreakdown,
	formatSector,
	formatSectorBreakdown,
	SECTOR_OPTIONS,
	sectorWeightsTotal
} from './sector';

describe('formatSector', () => {
	it('traduce el vocabulario del backend', () => {
		expect(formatSector('consumer_discretionary')).toBe('Consumo discrecional');
	});

	// La parte de un fondo que está en bonos o en caja no es ninguna industria
	// GICS, y sin estas dos no había dónde escribirla.
	it('traduce renta fija y efectivo', () => {
		expect(formatSector('fixed_income')).toBe('Renta fija');
		expect(formatSector('cash')).toBe('Efectivo');
	});

	// Una industria que este lado no conoce sigue siendo dinero del usuario: se
	// enseña en crudo en vez de desaparecer de la fila.
	it('conserva en crudo una industria que no conoce', () => {
		expect(formatSector('quantum_widgets')).toBe('quantum_widgets');
	});
});

describe('SECTOR_OPTIONS', () => {
	// Son las casillas del desglose: si faltan, un fondo mixto solo se puede
	// transcribir por su mitad en acciones.
	it('ofrece renta fija y efectivo', () => {
		const values = SECTOR_OPTIONS.map((o) => o.value);
		expect(values).toContain('fixed_income');
		expect(values).toContain('cash');
	});

	it('no ofrece los cubos derivados', () => {
		const values = SECTOR_OPTIONS.map((o) => o.value);
		expect(values).not.toContain('unclassified');
		expect(values).not.toContain('not_applicable');
	});
});

describe('formatSectorBreakdown', () => {
	// Es la etiqueta que distingue un fondo bien clasificado de un activo sin
	// clasificar, que traen los dos el `sector` vacío.
	it('cuenta las industrias de un fondo', () => {
		expect(
			formatSectorBreakdown([
				{ sector: 'technology', weight: 33.1 },
				{ sector: 'financials', weight: 13.8 }
			])
		).toBe('2 industrias');
	});

	// Un desglose de una sola fila dice lo mismo que el campo `sector`, así que
	// se lee como él: «1 industria» sería contar por contar.
	it('escribe el nombre cuando solo hay una', () => {
		expect(formatSectorBreakdown([{ sector: 'technology', weight: 100 }])).toBe('Tecnología');
	});

	it('no dice nada cuando no hay desglose', () => {
		expect(formatSectorBreakdown([])).toBe('');
	});
});

describe('describeSectorBreakdown', () => {
	it('escribe el desglose entero, en el orden en que llega', () => {
		expect(
			describeSectorBreakdown([
				{ sector: 'technology', weight: 33.1 },
				{ sector: 'financials', weight: 13.75 }
			])
		).toBe('Tecnología 33,1% · Finanzas 13,8%');
	});
});

describe('sectorWeightsTotal', () => {
	// No tiene por qué dar 100: la ficha de un fondo real deja unas décimas en
	// caja y el reparto normaliza sobre lo que haya. El total está para que
	// quien lo escribe vea cuánto lleva.
	it('suma los pesos tal como están', () => {
		expect(
			sectorWeightsTotal([
				{ sector: 'technology', weight: 33.1 },
				{ sector: 'financials', weight: 13.8 }
			])
		).toBeCloseTo(46.9);
	});

	it('ignora un peso ilegible en vez de contaminar el total con NaN', () => {
		expect(sectorWeightsTotal([{ sector: 'technology', weight: Number.NaN }])).toBe(0);
	});
});
