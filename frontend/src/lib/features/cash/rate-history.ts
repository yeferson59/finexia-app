/**
 * Lo que le pasa a la historia de una cuenta con cada opción del formulario de
 * su tasa: si guardar rehace días ya calculados, y la frase que lo dice antes
 * de guardar. Sin dependencias de Svelte; el formulario le pasa lo que tiene
 * abierto.
 */

import type { CashRate } from '$lib/api/types';
import { formatCalendarDate } from '$lib/shared/format/date';
import { formatAnnualRate } from './rates';

/** Las opciones del formulario de la tasa. */
export type CashRateMode = 'new' | 'move' | 'edit' | 'end' | 'delete' | 'recalc';

export interface CashRateHistoryInput {
	mode: CashRateMode;
	/** La versión más reciente de la tasa, `null` si la cuenta no tiene. */
	latest: CashRate | null;
	/** La versión más reciente ya generó intereses. */
	used: boolean;
	/** La versión más reciente terminó antes de hoy. */
	stopped: boolean;
	/** La versión más reciente todavía no empieza. */
	pending: boolean;
	/** El último día que la cuenta tiene calculado, `null` si ninguno. */
	computed: string | null;
	/** El día desde el que rige la tasa al anotarla o moverla. */
	startDay: string;
	today: string;
}

export interface CashRateHistory {
	/**
	 * Guardar rehace días ya calculados: el formulario lo dice y manda el
	 * permiso (`recompute`), sin el que el backend lo rechaza.
	 */
	recomputes: boolean;
	/** Qué le pasa a la historia de la cuenta, para decirlo antes de guardar. */
	hint: string;
}

const longDate = (iso: string) =>
	formatCalendarDate(iso.slice(0, 10), { day: 'numeric', month: 'long', year: 'numeric' });

export function cashRateHistory(input: CashRateHistoryInput): CashRateHistory {
	const { mode, latest, used, computed, startDay, today } = input;
	const dated = mode === 'new' || mode === 'move';

	/*
	 * Si el día cae en días ya calculados, guardar los rehace: se borran los
	 * abonos automáticos desde entonces y se calculan con esta tasa.
	 *
	 * Mover una tasa que ya generó intereses los rehace siempre: hacia atrás gana
	 * días que nunca se calcularon con ella, y hacia adelante los que deja ya no
	 * son suyos.
	 */
	const recomputes =
		(mode === 'move' && used) ||
		(dated && !!computed && !!startDay && startDay <= computed.slice(0, 10));

	/* El primer día que se rehace: al mover hacia adelante, el inicio de antes. */
	const redoFrom =
		mode === 'move' && used && latest && latest.effectiveFrom.slice(0, 10) < startDay
			? latest.effectiveFrom.slice(0, 10)
			: startDay;

	/* Un día pasado sin nada calculado: los intereses se calculan al guardar. */
	const backfills = dated && !recomputes && !!startDay && startDay < today;

	const pastNote = recomputes
		? ` Los intereses ya calculados desde el ${longDate(redoFrom)} se borran y se vuelven a calcular con las fechas nuevas.`
		: backfills
			? ` Los intereses desde el ${longDate(startDay)} hasta el último día terminado se calculan al guardar, sobre lo que la cuenta tenía cada día.`
			: '';

	return { recomputes, hint: hintFor(input, pastNote) };
}

function hintFor(
	{ mode, latest, stopped, pending, computed, today }: CashRateHistoryInput,
	pastNote: string
): string {
	if (!latest) return `Desde ese día, la cuenta rinde esta tasa.${pastNote}`;

	switch (mode) {
		case 'new':
			if (pending) {
				return `Ya hay una tasa anotada desde el ${longDate(latest.effectiveFrom)}: la nueva tiene que empezar después. Para cambiar esa, usa «Corregir» o «Cambiar fecha».`;
			}
			return stopped
				? `Desde ese día la cuenta vuelve a rendir. La pausa se queda como estaba.${pastNote}`
				: `La tasa de ahora, ${formatAnnualRate(latest.annualRatePct)}, termina la víspera. Los días anteriores conservan la suya.${pastNote}`;
		case 'move':
			return `Mueve el inicio de la tasa de ${formatAnnualRate(latest.annualRatePct)}, que hoy empieza el ${longDate(latest.effectiveFrom)}. La tasa anterior rige hasta la víspera del día nuevo.${pastNote}`;
		case 'edit':
			return `Corrige la tasa anotada desde el ${longDate(latest.effectiveFrom)}, sin crear otra. Si la entidad cambió la tasa, usa «Cambiar tasa».`;
		case 'end':
			return 'Desde ese día la cuenta deja de rendir. Los días anteriores conservan su tasa.';
		case 'delete':
			return `Borra la tasa de ${formatAnnualRate(latest.annualRatePct)} anotada desde el ${longDate(latest.effectiveFrom)}. Si al empezar cerró otra, esa vuelve a regir.`;
		case 'recalc':
			return `Los intereses están calculados hasta el ${longDate(computed ?? today)}. Desde el día que elijas hasta ese se borran los abonos automáticos y se vuelven a calcular sobre lo que la cuenta guarda ahora. Úsalo si anotaste un depósito o un retiro con fecha pasada. Si la tasa empezó antes de lo anotado, usa «Cambiar fecha».`;
	}
}
