/**
 * Datos y helpers de la feature `investments` (catálogo de productos de
 * inversión). Hoy el detalle enriquecido es un mock local: el backend todavía
 * no expone este dominio, así que vive junto a la feature en vez de en
 * `lib/api`. Cuando exista el endpoint, esto se sustituye por un módulo de
 * `lib/api` sin tocar los componentes.
 */
import { investmentStore, type Investment } from './state/investments.svelte';

/** Detalle enriquecido de un producto de inversión. */
export interface InvestmentProduct {
	id: string;
	name: string;
	type: string;
	category: string;
	description: string;
	riskLevel: string;
	expectedROI: number;
	currentROI: number;
	horizon: number;
	minimumInvestment: number;
	totalInvested: number;
	currentValue: number;
	investors: number;
	status: string;
	startDate: string;
	maturityDate: string;
	highlights: string[];
}

// Mock data - would come from backend
export const investmentDetails: Record<string, InvestmentProduct> = {
	'1': {
		id: '1',
		name: 'Fondo Crecimiento Tecnológico',
		type: 'Fondo',
		category: 'Tecnología',
		description:
			'Fondo diversificado enfocado en empresas tecnológicas de alto crecimiento. Nuestro equipo de gestores expertos selecciona las mejores oportunidades en el sector tech global.',
		riskLevel: 'Medio',
		expectedROI: 15.2,
		currentROI: 12.8,
		horizon: 24,
		minimumInvestment: 5000,
		totalInvested: 2500000,
		currentValue: 2820000,
		investors: 342,
		status: 'Activo',
		startDate: '2023-01-15',
		maturityDate: '2025-12-31',
		highlights: [
			'Cartera diversificada en 15+ empresas tech líderes',
			'Gestor con 10+ años de experiencia',
			'Comisión de gestión competitiva (1.5% anual)',
			'Rebalanceo trimestral automático'
		]
	},
	'2': {
		id: '2',
		name: 'ETF Mercados Emergentes',
		type: 'ETF',
		category: 'Mercados Emergentes',
		description:
			'Exposición amplia a mercados emergentes de rápido crecimiento. Este ETF rastrea el desempeño de índices de economías emergentes con mayor potencial de apreciación.',
		riskLevel: 'Alto',
		expectedROI: 18.5,
		currentROI: 14.2,
		horizon: 36,
		minimumInvestment: 1000,
		totalInvested: 8750000,
		currentValue: 9980000,
		investors: 1204,
		status: 'Activo',
		startDate: '2022-06-10',
		maturityDate: '2026-06-10',
		highlights: [
			'Rastreo de índices de mercados emergentes',
			'Comisión ultra baja (0.35% anual)',
			'Liquidez diaria',
			'Diversificación en 25+ países'
		]
	}
};

/**
 * Construye el detalle de un producto que solo existe en el store compartido
 * (p. ej. recién creado con el formulario de alta) y por tanto no tiene ficha
 * mock. Los campos que el formulario no recoge caen a valores por defecto.
 */
export function fromStoredInvestment(stored: Investment): InvestmentProduct {
	const start = new Date();
	const maturity = new Date(
		start.getFullYear(),
		start.getMonth() + stored.horizon,
		start.getDate()
	);
	const highlights = [`Tipo de instrumento: ${stored.type}`, `Categoría: ${stored.category}`];
	if (stored.minimumInvestment > 0) {
		highlights.push(`Inversión mínima: $${stored.minimumInvestment.toLocaleString('es-CO')}`);
	}
	return {
		id: stored.id,
		name: stored.name,
		type: stored.type,
		category: stored.category,
		description: stored.description,
		riskLevel: stored.riskLevel,
		expectedROI: stored.expectedROI,
		currentROI: 0,
		horizon: stored.horizon,
		minimumInvestment: stored.minimumInvestment,
		totalInvested: 0,
		currentValue: 0,
		investors: 0,
		status: stored.status,
		startDate: start.toISOString().slice(0, 10),
		maturityDate: maturity.toISOString().slice(0, 10),
		highlights
	};
}

/** Color asociado a cada nivel de riesgo (variable CSS del tema). */
export function getRiskColor(risk: string): string {
	switch (risk) {
		case 'Bajo':
			return 'var(--green)';
		case 'Medio':
			return 'var(--amber-light)';
		case 'Alto':
			return 'var(--amber)';
		case 'Muy Alto':
			return 'var(--red)';
		default:
			return 'var(--text)';
	}
}

/**
 * Busca un producto por id: primero en el catálogo mock, y si no, en el store
 * compartido (productos creados en la sesión). Única fuente de verdad de la
 * resolución, usada por la página (para el `<title>`) y por el detalle.
 */
export function findInvestmentProduct(id: string): InvestmentProduct | null {
	if (!id) return null;
	const stored = investmentStore.getById(id);
	return investmentDetails[id] ?? (stored ? fromStoredInvestment(stored) : null);
}
