/**
 * Helpers puros de los intereses que abona el efectivo: los abonos automáticos
 * agrupados en el extracto. Sin dependencias de Svelte ni de red; el contrato
 * `CashMovement` viene de `$lib/api/types`.
 */

import type { CashMovement } from '$lib/api/types';
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
