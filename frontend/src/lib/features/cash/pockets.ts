/**
 * Helpers puros de los bolsillos (000047): cómo cuelgan de la cuenta a la que
 * pertenecen, cómo suman en su plataforma, y lo que la lista de cuentas
 * necesita para enseñarlos.
 *
 * Un bolsillo es una subcuenta de una cuenta —la «cajita» del banco, el
 * subsaldo del bróker—, no otra plataforma: su dinero sigue contando en ella y
 * lo propio del bolsillo es su tasa. Por eso se anida en lugar de listarse
 * aparte, y por eso todo esto vive junto en lugar de en `cash.ts`.
 */

import type { CashPocket } from '$lib/api/types';
import { formatCalendarDate } from '$lib/shared/format/date';
import { groupCashAccounts, type CashAccount, type CashBalance } from './cash';

/** Una plataforma y lo que guarda en cada moneda. */
export interface CashPlatform {
	sourceId: string;
	sourceName: string;
	displayCurrency: string;
	/** La suma en `displayCurrency` de las cuentas que se pudieron convertir. */
	value: number;
	/** Si a `value` le falta una cuenta con dinero que no tenía tasa. */
	partial: boolean;
	/** De la que más vale a la que menos. */
	accounts: CashAccount[];
}

/**
 * Las cuentas principales con sus bolsillos debajo, como lo enseña el banco: la
 * cuenta y, debajo, sus cajitas.
 *
 * Un bolsillo cuenta dentro de su plataforma —su dinero suma en ella— y lo que
 * es suyo es la tasa. Por eso se anida en lugar de listarse aparte.
 *
 * Una cuenta principal vacía se inventa cuando solo hay bolsillos: el dinero
 * entró directo en la cajita, y sin ella la cajita no tendría dónde colgarse.
 */
export function nestCashPockets(accounts: CashAccount[]): CashAccount[] {
	const main = new Map<string, CashAccount>();

	for (const account of accounts) {
		if (account.pocketId === null) main.set(`${account.sourceId}:${account.currency}`, account);
	}

	for (const account of accounts) {
		if (account.pocketId === null) continue;

		const key = `${account.sourceId}:${account.currency}`;
		let parent = main.get(key);

		if (!parent) {
			parent = {
				...account,
				key,
				pocketId: null,
				pocketName: '',
				pocketKind: '',
				balance: 0,
				value: 0,
				fxConverted: true,
				lastMovementDate: null,
				balances: [],
				pockets: []
			};
			main.set(key, parent);
		}

		parent.pockets.push(account);
	}

	return accounts
		.filter((a) => a.pocketId === null)
		.concat([...main.values()].filter((a) => !accounts.includes(a)));
}

/**
 * Las cuentas agrupadas por plataforma, de la que más guarda a la que menos.
 *
 * Es como las enseña un banco: la entidad y, debajo, cada moneda. Dentro de
 * cada plataforma las cuentas conservan el orden de `groupCashAccounts`.
 */
export function groupCashPlatforms(balances: CashBalance[]): CashPlatform[] {
	const platforms = new Map<string, CashPlatform>();

	for (const account of nestCashPockets(groupCashAccounts(balances))) {
		const platform = platforms.get(account.sourceId) ?? {
			sourceId: account.sourceId,
			sourceName: account.sourceName,
			displayCurrency: account.displayCurrency,
			value: 0,
			partial: false,
			accounts: []
		};

		// Un bolsillo es dinero de la plataforma, así que suma en ella aunque se
		// enseñe anidado bajo su cuenta.
		platform.value += account.value + account.pockets.reduce((sum, p) => sum + p.value, 0);
		for (const drawer of [account, ...account.pockets]) {
			if (!drawer.fxConverted && drawer.balance !== 0) platform.partial = true;
		}
		platform.accounts.push(account);
		platforms.set(account.sourceId, platform);
	}

	return [...platforms.values()].sort(
		(a, b) => b.value - a.value || a.sourceName.localeCompare(b.sourceName)
	);
}

/**
 * Los bolsillos de una cuenta que todavía no guardan nada.
 *
 * No llegan con los saldos —no hay posición que listar—, y aun así tienen que
 * verse: es donde se les da una tasa y desde donde sale el primer traslado. Los
 * cerrados no están: un depósito vencido ya devolvió su dinero.
 */
export function emptyCashPockets(account: CashAccount, pockets: CashPocket[]): CashPocket[] {
	const held = new Set(account.pockets.map((p) => p.pocketId));

	return pockets.filter(
		(p) =>
			p.sourceId === account.sourceId &&
			p.currency === account.currency &&
			p.closedOn === null &&
			!held.has(p.id)
	);
}

/**
 * Los cajones de una cuenta que siguen abiertos.
 *
 * Un depósito cerrado —vencido o cancelado— ya devolvió su dinero a la cuenta
 * principal, así que deja de ser un cajón donde haya nada: sale de la lista y se
 * queda en los movimientos, que es donde vive su historia.
 */
export function openCashPockets(account: CashAccount, pockets: CashPocket[]): CashAccount[] {
	return account.pockets.filter(
		(held) => pockets.find((p) => p.id === held.pocketId)?.closedOn == null
	);
}

/** Un bolsillo sin saldo como una cuenta más, para que su fila sea la misma. */
export function pocketAsCashAccount(account: CashAccount, pocket: CashPocket): CashAccount {
	return {
		...account,
		key: `${account.sourceId}:${account.currency}:${pocket.id}`,
		pocketId: pocket.id,
		pocketName: pocket.name,
		pocketKind: pocket.kind,
		balance: 0,
		value: 0,
		fxConverted: true,
		lastMovementDate: null,
		balances: [],
		pockets: []
	};
}

/**
 * Desde cuándo una cuenta está como está. Una vaciada sigue en la lista —es
 * donde cae el próximo depósito— y decir desde cuándo lo está es lo que la
 * distingue de una sin estrenar.
 */
export function cashAccountHistory(account: CashAccount): string {
	if (!account.lastMovementDate) return 'Sin movimientos todavía';

	const when = formatCalendarDate(account.lastMovementDate, {
		day: 'numeric',
		month: 'short',
		year: 'numeric'
	});

	return account.balance === 0 ? `Vacía desde el ${when}` : `Último movimiento el ${when}`;
}
