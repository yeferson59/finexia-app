/**
 * Shared, rune-based store that holds the investment products shown across the
 * dashboard (list, detail and the "add product" form). It replaces the
 * duplicated mock arrays that previously lived inside the route components and
 * makes the "Agregar Producto" flow actually persist for the session.
 */

export interface Investment {
	id: string;
	name: string;
	description: string;
	type: string;
	category: string;
	riskLevel: string;
	/** Expected return on investment, as a percentage (e.g. 15.2). */
	expectedROI: number;
	/** Investment horizon, in months. */
	horizon: number;
	/** Minimum investment in USD; 0 when not specified. */
	minimumInvestment: number;
	status: string;
}

/** Fields collected by the add-product form. The id is assigned by the store. */
export type NewInvestment = Omit<Investment, 'id'>;

const seed: Investment[] = [
	{
		id: '1',
		name: 'Fondo Crecimiento Tecnológico',
		description: 'Fondo diversificado enfocado en empresas tecnológicas de alto crecimiento.',
		type: 'Fondos',
		category: 'Tecnología',
		riskLevel: 'Medio',
		expectedROI: 15.2,
		horizon: 24,
		minimumInvestment: 5000,
		status: 'Activo'
	},
	{
		id: '2',
		name: 'ETF Mercados Emergentes',
		description: 'Exposición amplia a mercados emergentes de rápido crecimiento.',
		type: 'ETF',
		category: 'Mercados Emergentes',
		riskLevel: 'Alto',
		expectedROI: 18.5,
		horizon: 36,
		minimumInvestment: 1000,
		status: 'Activo'
	},
	{
		id: '3',
		name: 'Energía Renovable',
		description: 'Cartera centrada en proyectos de energía limpia y sostenible.',
		type: 'Fondos',
		category: 'Energía Renovable',
		riskLevel: 'Bajo',
		expectedROI: 8.1,
		horizon: 24,
		minimumInvestment: 2500,
		status: 'Activo'
	}
];

class InvestmentStore {
	items = $state<Investment[]>(seed);

	/** Adds a new investment product and returns its generated id. */
	addInvestment(data: NewInvestment): string {
		const id =
			typeof crypto !== 'undefined' && 'randomUUID' in crypto
				? crypto.randomUUID()
				: String(Date.now());
		this.items.push({ ...data, id });
		return id;
	}

	/** Looks up a single investment by id. */
	getById(id: string): Investment | undefined {
		return this.items.find((item) => item.id === id);
	}
}

export const investmentStore = new InvestmentStore();
