/**
 * Helpers puros de los intereses que abona el efectivo: los abonos automáticos
 * agrupados en el extracto, y lo que cambió un recálculo. Sin dependencias de
 * Svelte ni de red; los contratos vienen de `$lib/api/types`.
 */

import type { CashBalance, CashMovement, CashRate, CashRecalculation } from '$lib/api/types';
import { formatCalendarDate } from '$lib/shared/format/date';

/** Los abonos automáticos de un saldo en un mes, como una sola fila. */
export interface CashInterestGroup {
	type: 'interest';
	/** `interest:<saldo>:<mes>`. */
	key: string;
	/** El día del abono más reciente. */
	date: string;
	/** El día del más antiguo. */
	since: string;
	/** La suma, en `currency`. */
	amount: number;
	currency: string;
	sourceName: string;
	portfolioName: string;
	/** Del más reciente al más antiguo. */
	movements: CashMovement[];
}

/** Una fila del extracto: un movimiento, o un grupo de abonos automáticos. */
export type CashLedgerRow =
	{ type: 'movement'; key: string; date: string; movement: CashMovement } | CashInterestGroup;

const movementRow = (movement: CashMovement): CashLedgerRow => ({
	type: 'movement',
	key: movement.id,
	date: movement.date,
	movement
});

/**
 * El extracto con los abonos automáticos de intereses juntos: una fila por
 * saldo y mes, en el sitio del abono más reciente.
 *
 * Una cuenta que rinde abona cada día, y treinta filas iguales al mes taparían
 * los depósitos y retiros. Los intereses anotados a mano no se juntan: son algo
 * que hizo el usuario. Un abono solo en su mes se queda como el movimiento que
 * es.
 *
 * La lista llega de la más reciente a la más antigua y conserva ese orden. El
 * mes es el del día del calendario que guarda el backend.
 */
export function groupAutomaticInterest(movements: CashMovement[]): CashLedgerRow[] {
	const rows: CashLedgerRow[] = [];
	const groups = new Map<string, CashInterestGroup>();

	for (const movement of movements) {
		if (!movement.automatic || movement.kind !== 'interest') {
			rows.push(movementRow(movement));
			continue;
		}

		const key = `interest:${movement.entryId}:${movement.date.slice(0, 7)}`;
		const amount = parseFloat(movement.amount) || 0;
		const group = groups.get(key);

		if (group) {
			group.movements.push(movement);
			group.amount += amount;
			group.since = movement.date;
			continue;
		}

		const created: CashInterestGroup = {
			type: 'interest',
			key,
			date: movement.date,
			since: movement.date,
			amount,
			currency: movement.currency,
			sourceName: movement.sourceName,
			portfolioName: movement.portfolioName,
			movements: [movement]
		};
		groups.set(key, created);
		rows.push(created);
	}

	return rows.map((row) =>
		row.type === 'interest' && row.movements.length === 1 ? movementRow(row.movements[0]) : row
	);
}

/** Las filas de un mes del extracto. */
export interface CashLedgerMonth {
	/** `2026-09`. */
	key: string;
	/** «Septiembre de 2026». */
	label: string;
	rows: CashLedgerRow[];
}

/** Las filas agrupadas por mes, en el orden en que aparece cada mes. */
export function groupCashLedgerByMonth(rows: CashLedgerRow[]): CashLedgerMonth[] {
	const months = new Map<string, CashLedgerMonth>();

	for (const row of rows) {
		const key = row.date.slice(0, 7);
		const month = months.get(key);

		if (month) {
			month.rows.push(row);
			continue;
		}

		const label = formatCalendarDate(`${key}-01`, { month: 'long', year: 'numeric' });
		months.set(key, {
			key,
			label: label.charAt(0).toUpperCase() + label.slice(1),
			rows: [row]
		});
	}

	return [...months.values()];
}

/**
 * Lo que una cuenta lleva generado: lo abonado —solo o a mano— y lo calculado
 * que espera su abono. Es lo que un recálculo puede mover: con abono mensual o
 * al vencer, lo del mes en curso está todo en lo pendiente.
 */
export function cashInterestTotal(
	balances: Pick<CashBalance, 'interestEarned' | 'pendingInterest'>[]
): number {
	return balances.reduce(
		(sum, b) => sum + (parseFloat(b.interestEarned) || 0) + (parseFloat(b.pendingInterest) || 0),
		0
	);
}

/** El primer día en que rinde una cuenta: el de su versión de tasa más antigua. */
export function firstRateDay(
	rates: CashRate[],
	sourceId: string,
	currency: string,
	pocketId: string | null
): string | null {
	return (
		rates
			.filter(
				(r) =>
					r.sourceId === sourceId && r.currency === currency && (r.pocketId ?? null) === pocketId
			)
			.map((r) => r.effectiveFrom.slice(0, 10))
			.sort()[0] ?? null
	);
}

/** Un recálculo hecho, con lo que hace falta para contarlo cuando la página se refresca. */
export interface CashRecalcDone {
	/** La cuenta, con la clave de `groupCashAccounts`. */
	key: string;
	/** «Rappi, education». */
	where: string;
	currency: string;
	/** Lo que llevaba generado antes, según `cashInterestTotal`. */
	before: number;
	/** El día desde el que se pidió. */
	from: string;
	/** El primer día con tasa de la cuenta, si lo tiene. */
	rateFrom: string | null;
	result: CashRecalculation;
}

/** Cómo quedó una cuenta tras recalcular. */
export interface CashRecalcChange {
	after: number;
	/** `after - before`, cero si no se mueve ni medio centavo. */
	delta: number;
	/** Se pidió desde antes de que la cuenta rindiera: esos días no generan nada. */
	beforeRate: boolean;
}

/**
 * Lo que cambió un recálculo: lo generado antes y después, y si se pidió desde
 * un día en que la cuenta todavía no rendía.
 *
 * Una diferencia de menos de medio centavo es ruido de redondeo y se cuenta
 * como ninguna: anunciar «+0,00» diría que cambió algo que no se ve.
 */
export function cashRecalcChange(done: CashRecalcDone, after: number): CashRecalcChange {
	const delta = after - done.before;

	return {
		after,
		delta: Math.abs(delta) < 0.005 ? 0 : delta,
		beforeRate: !!done.rateFrom && done.from < done.rateFrom
	};
}
