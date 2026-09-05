/*
 * Las tres lecturas del patrimonio: dónde está custodiado (plataforma), cómo lo
 * agrupó el usuario (portafolio) y en qué está invertido (tipo de activo).
 *
 * Son el mismo total leído de tres formas, y por eso viven juntas: comparten el
 * reparto porcentual, el trato de las filas que no se pudieron convertir y la
 * regla de cuándo una cifra de rendimiento se puede enseñar. Antes el panel
 * repetía el mismo total en tres tarjetas distintas sin descomponerlo ni una
 * vez.
 *
 * Lógica pura y sin Svelte: es aritmética de dinero, que es justo lo que hay
 * que poder probar sin montar un componente.
 */
import type { AllocationItem, Platform, PortfolioSummary } from '$lib/api/types';
import { partitionByCurrency } from '$lib/shared/currency';
import { formatAssetType } from '$lib/shared/format/asset-type';
import { formatPortfolioType } from '$lib/shared/format/portfolio-type';

/** Cuál de las tres lecturas se está mirando. */
export type CutId = 'platform' | 'portfolio' | 'type';

export const CUTS: { id: CutId; label: string }[] = [
	{ id: 'platform', label: 'Plataforma' },
	{ id: 'portfolio', label: 'Portafolio' },
	{ id: 'type', label: 'Tipo de activo' }
];

/** Una fila del reparto: una plataforma, un portafolio o una clase de activo. */
export interface BreakdownRow {
	/** Id de la entidad: con él, y con el corte, se arma el enlace de la fila. */
	key: string;
	/** Nombre propio; es el texto del enlace cuando la fila lleva a algún sitio. */
	label: string;
	/** Segunda línea, en prosa. Vacía cuando no hay nada que añadir. */
	detail: string;
	/** Importe en la moneda de visualización. */
	value: number;
	/** Parte del total enseñado, de 0 a 1. Es lo que mide la barra. */
	share: number;
	/**
	 * Rendimiento de la fila, o `null` cuando no hay ninguno que enseñar.
	 *
	 * `null` no es cero. Una plataforma cuyas posiciones se valoran todas a
	 * coste informa una ganancia de cero porque no tiene precio de mercado, no
	 * porque no se haya movido; enseñar «0,00 %» ahí sería presentar la falta de
	 * dato como un resultado.
	 */
	gainPct: number | null;
}

/** El reparto completo, con lo que se quedó fuera y por qué. */
export interface Breakdown {
	rows: BreakdownRow[];
	/** Suma de las filas enseñadas. */
	total: number;
	/** Filas que no están en la moneda pedida y por eso no se suman. */
	excluded: number;
	/** Posiciones que sí se suman, pero sin convertir: el total mezcla monedas. */
	unconverted: number;
	/**
	 * Qué mide la última columna. El reparto por tipo de activo sale de un
	 * endpoint que no devuelve ganancias, así que allí la columna enseña la
	 * participación en vez de inventar un rendimiento.
	 */
	trailing: 'gain' | 'share';
}

const num = (value: string | undefined): number => parseFloat(value || '0') || 0;

/** «1 posición» / «4 posiciones», que es lo que evita el «1 posiciones». */
export function plural(count: number, one: string, many: string): string {
	return `${count} ${count === 1 ? one : many}`;
}

/** Reparte el peso de cada fila y ordena de mayor a menor. */
function finish(
	rows: Omit<BreakdownRow, 'share'>[],
	excluded: number,
	unconverted: number,
	trailing: 'gain' | 'share'
): Breakdown {
	const total = rows.reduce((acc, row) => acc + row.value, 0);
	return {
		// Sin total no hay reparto que medir: la barra se queda a cero en vez de
		// dividir por cero y pintar todas las filas llenas.
		rows: rows
			.map((row) => ({ ...row, share: total > 0 ? row.value / total : 0 }))
			.sort((a, b) => b.value - a.value),
		total,
		excluded,
		unconverted,
		trailing
	};
}

/**
 * Dónde está custodiado el dinero.
 *
 * `marketValue` es opcional en backends anteriores; sin él queda lo invertido,
 * que es el dato que esa versión sabe dar.
 */
export function platformBreakdown(platforms: Platform[], currency: string): Breakdown {
	const split = partitionByCurrency(platforms, currency);
	const unconverted = split.converted.reduce((acc, p) => acc + (p.positionsUnconverted ?? 0), 0);

	const rows = split.converted.map((platform) => {
		// Todas las posiciones valoradas a su propio coste: la ganancia que
		// informa el backend es cero por falta de precio, no por falta de
		// movimiento, y las dos cosas no se pueden enseñar igual.
		const allAtCost =
			platform.investments > 0 && (platform.positionsAtCost ?? 0) === platform.investments;

		return {
			key: platform.id,
			label: platform.name,
			detail: plural(platform.investments, 'posición abierta', 'posiciones abiertas'),
			value: num(platform.marketValue ?? platform.totalValue),
			gainPct: platform.gainLossPct === undefined || allAtCost ? null : platform.gainLossPct
		};
	});

	return finish(rows, split.unconverted.length, unconverted, 'gain');
}

/** Cómo agrupó el usuario ese mismo dinero. */
export function portfolioBreakdown(summaries: PortfolioSummary[], currency: string): Breakdown {
	const split = partitionByCurrency(summaries, currency);
	const unconverted = split.converted.reduce((acc, s) => acc + (s.positionsUnconverted ?? 0), 0);

	const rows = split.converted.map((summary) => ({
		key: summary.id,
		label: summary.name,
		detail: formatPortfolioType(summary.type),
		value: num(summary.totalMarketValue),
		gainPct: num(summary.totalGainLossPct)
	}));

	return finish(rows, split.unconverted.length, unconverted, 'gain');
}

/**
 * En qué está invertido.
 *
 * Estas filas no llevan enlace: no hay una página por clase de activo, y quien
 * pinta el reparto lo sabe por el corte que le pidieron.
 */
export function typeBreakdown(allocation: AllocationItem[], currency: string): Breakdown {
	// Todas las categorías vienen en la misma moneda; si no es la pedida, el
	// reparto entero se queda fuera en vez de enseñarse bajo el símbolo ajeno.
	const usable = allocation.filter((item) => (item.currency ?? currency) === currency);
	const unconverted = usable.reduce((acc, item) => acc + (item.positionsUnconverted ?? 0), 0);

	const rows = usable.map((item) => ({
		key: item.category,
		label: formatAssetType(item.category),
		detail: '',
		value: num(item.marketValue),
		gainPct: null
	}));

	return finish(rows, allocation.length - usable.length, unconverted, 'share');
}

/** El reparto pedido, ya listo para pintar. */
export function breakdownFor(
	cut: CutId,
	data: { platforms: Platform[]; summaries: PortfolioSummary[]; allocation: AllocationItem[] },
	currency: string
): Breakdown {
	if (cut === 'platform') return platformBreakdown(data.platforms, currency);
	if (cut === 'portfolio') return portfolioBreakdown(data.summaries, currency);
	return typeBreakdown(data.allocation, currency);
}
