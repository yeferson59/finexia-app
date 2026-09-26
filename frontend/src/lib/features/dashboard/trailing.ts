/**
 * La rentabilidad por periodo del resumen: cuánto se ganó el último día, la
 * última semana, el último mes…
 *
 * El backend ya manda cada ventana calculada (`returns` de
 * `GET /portfolios/growth`): la ganancia en dinero descontando aportes y
 * retiros, y el porcentaje ponderado por tiempo, el mismo que dibuja la vista
 * `%` de la gráfica. Aquí solo se les pone nombre, tono y la nota que explica
 * desde cuándo se mide o por qué todavía no hay cifra.
 */

import type { TrailingPeriod, TrailingReturn } from '$lib/api/types';
import { formatCalendarDate } from '$lib/shared/format/date';

/** Cómo se llama cada ventana de cara al usuario. */
export const TRAILING_LABELS: Record<TrailingPeriod, string> = {
	'1D': '1 día',
	'1W': '7 días',
	'1M': '1 mes',
	'3M': '3 meses',
	YTD: 'En el año',
	'1Y': '1 año',
	ALL: 'Desde el inicio'
};

export type TrailingTone = 'up' | 'down' | 'neutral';

/** Una ventana lista para pintar. Las cifras son `null` cuando no hay dato. */
export interface TrailingCell {
	period: TrailingPeriod;
	label: string;
	available: boolean;
	/** En puntos porcentuales (`0.41` es +0,41 %), como el resto de porcentajes. */
	pct: number | null;
	gain: number | null;
	pctTone: TrailingTone;
	gainTone: TrailingTone;
	/** Desde cuándo se mide, o por qué todavía no se puede. */
	note: string;
}

function toFinite(raw: string | undefined): number | null {
	if (raw === undefined || raw === '') return null;
	const value = Number.parseFloat(raw);
	return Number.isFinite(value) ? value : null;
}

/**
 * Tono de una cifra con signo. Lo que redondea a cero en pantalla va neutro:
 * un +0,00 % en verde promete una ganancia que la propia cifra desmiente.
 */
export function trailingTone(value: number | null): TrailingTone {
	if (value === null || Math.abs(value) < 0.005) return 'neutral';
	return value > 0 ? 'up' : 'down';
}

function shortDate(iso: string): string {
	return formatCalendarDate(iso, { day: 'numeric', month: 'short', year: 'numeric' });
}

function noteFor(entry: TrailingReturn): string {
	if (entry.available) return entry.from ? `Desde el ${shortDate(entry.from)}` : '';
	if (entry.historyStart) return `El historial empieza el ${shortDate(entry.historyStart)}`;
	return 'Todavía no hay historial';
}

/**
 * Si hay algo que enseñar: al menos una ventana con cifra.
 *
 * Una cuenta recién abierta, con un solo día de historia, no tiene ninguna, y
 * siete rayas no dicen nada que la cifra de arriba no diga ya. Un backend
 * anterior no manda `returns` y el resultado es el mismo. Quien envuelve la
 * franja lo pregunta antes, para no dejar un filete sobre un hueco.
 */
export function hasTrailingReturns(returns: TrailingReturn[] | undefined): boolean {
	return returns?.some((entry) => entry.available) ?? false;
}

/**
 * La ventana que se lee al abrir: el último mes, que es la que menos baila de
 * un día para otro sin quedarse tan lejos que ya no diga nada del presente. Si
 * el historial no llega a un mes, la primera que tenga cifra.
 */
export function defaultTrailingPeriod(cells: TrailingCell[]): TrailingPeriod | null {
	const month = cells.find((cell) => cell.period === '1M' && cell.available);
	return (month ?? cells.find((cell) => cell.available) ?? cells[0])?.period ?? null;
}

/**
 * La frase que lee la ventana elegida, partida alrededor del importe para que
 * quien la pinta lo formatee (y lo tape en modo privado) a su manera.
 */
export interface TrailingReading {
	before: string;
	/** Sin signo: el verbo ya dice hacia dónde fue. `null` si no hay importe. */
	amount: number | null;
	tone: TrailingTone;
	after: string;
}

const FLOWS_CAVEAT = ', sin contar lo que metiste o sacaste.';

export function trailingReading(cell: TrailingCell): TrailingReading {
	const none = { amount: null, tone: 'neutral' as const, after: '' };

	if (!cell.available) {
		return { ...none, before: `${cell.note}, así que todavía no hay cifra para este periodo.` };
	}

	const lead = cell.note || 'En este periodo';

	if (cell.gain === null) {
		return { ...none, before: `${lead} no hay cifra en dinero para este periodo.` };
	}
	if (cell.gainTone === 'neutral') {
		return { ...none, before: `${lead} no ganaste ni perdiste${FLOWS_CAVEAT}` };
	}

	return {
		before: `${lead} ${cell.gainTone === 'up' ? 'ganaste' : 'perdiste'} `,
		amount: Math.abs(cell.gain),
		tone: cell.gainTone,
		after: FLOWS_CAVEAT
	};
}

/**
 * Las ventanas en el orden en que llegan, que es de la más corta a la más
 * larga; ninguna mientras `hasTrailingReturns` diga que no hay qué enseñar.
 */
export function toTrailingCells(returns: TrailingReturn[] | undefined): TrailingCell[] {
	if (!returns || !hasTrailingReturns(returns)) return [];

	return returns.map((entry) => {
		const pct = entry.available ? toFinite(entry.returnPct) : null;
		const gain = entry.available ? toFinite(entry.gain) : null;

		return {
			period: entry.period,
			label: TRAILING_LABELS[entry.period],
			available: entry.available,
			pct,
			gain,
			pctTone: trailingTone(pct),
			gainTone: trailingTone(gain),
			note: noteFor(entry)
		};
	});
}
