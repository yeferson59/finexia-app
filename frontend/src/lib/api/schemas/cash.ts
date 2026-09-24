/**
 * Efectivo — contratos HTTP como schemas Zod.
 *
 * Parte de `lib/api/schemas`: el porqué de que sean schemas y no interfaces
 * está en el `index.ts` de la carpeta.
 */

import { z } from 'zod';

/**
 * Un saldo de efectivo (`GET /portfolios/cash`): lo que guarda una plataforma
 * en una moneda para un portafolio.
 *
 * Es una posición como cualquier otra —un activo de tipo efectivo que vale uno
 * de su moneda por unidad—, así que el mismo dinero aparece también entre las
 * posiciones de su portafolio y suma en sus totales.
 */
export const cashBalanceSchema = z.object({
	/** La posición que guarda el saldo. */
	entryId: z.string(),
	portfolioId: z.string(),
	portfolioName: z.string(),
	sourceId: z.string(),
	sourceName: z.string(),
	assetId: z.string(),
	/** `CASH-USD` para los saldos que abre la aplicación. */
	ticker: z.string(),
	name: z.string(),
	/** Lo que hay, en `currency`, la moneda del propio saldo. */
	balance: z.string(),
	currency: z.string(),
	/**
	 * `balance` en `displayCurrency`, para poder sumar saldos en monedas
	 * distintas. Con `fxConverted: false` no había tasa: el importe es el nominal
	 * y no se puede sumar con los demás.
	 */
	value: z.string(),
	displayCurrency: z.string(),
	fxConverted: z.boolean(),
	/** Distinguen una cuenta vaciada (tiene movimientos) de una sin estrenar. */
	movements: z.number(),
	lastMovementDate: z.string().nullable(),
	/**
	 * El bolsillo de la cuenta en que está el saldo; `null` es la cuenta
	 * principal, que es donde estaba todo antes de que hubiera bolsillos. Con
	 * `null`, `pocketName` y `pocketKind` van vacíos.
	 */
	pocketId: z.string().nullable().default(null),
	pocketName: z.string().default(''),
	pocketKind: z.string().default(''),
	/**
	 * Los intereses abonados al saldo, en `currency`: todos, y los fechados en el
	 * mes en curso (UTC). `interestThisMonthValue` es esa parte en
	 * `displayCurrency`, con el mismo contrato de `fxConverted` que `value`.
	 */
	interestEarned: z.string().default('0'),
	interestThisMonth: z.string().default('0'),
	interestThisMonthValue: z.string().default('0'),
	/**
	 * Lo calculado y todavía sin abonar, en `currency`: los días que una tasa de
	 * abono mensual guarda hasta que cierra el mes. No está dentro de `balance`.
	 */
	pendingInterest: z.string().default('0'),
	/** Último día con intereses calculados; `null` si la cuenta nunca rindió. */
	lastAccrualDate: z.string().nullable().default(null)
});

/**
 * Un movimiento sobre un saldo (`GET /portfolios/cash/movements`).
 *
 * `kind` es lo que hizo el dueño; `type`, la transacción con que se guardó.
 * `other` es un movimiento anotado sobre el saldo desde la posición —un
 * interés cobrado fuera, una comisión suelta— que esta pantalla enseña pero no
 * sabe reescribir: por eso llega con `editable: false`. `dividend` y `sale` son
 * el dinero que una acción pagó a esta cuenta —un dividendo, lo recibido por
 * una venta— y `purchase` el que salió de ella para comprar una: se registran,
 * cambian y borran desde esa transacción, nunca desde aquí.
 */
export const cashMovementSchema = z.object({
	id: z.string(),
	entryId: z.string(),
	type: z.string(),
	kind: z.enum(['deposit', 'withdrawal', 'interest', 'dividend', 'sale', 'purchase', 'other']),
	/** Lo que se movió, en `currency`, la moneda del saldo. */
	amount: z.string(),
	currency: z.string(),
	fees: z.string(),
	feesCurrency: z.string(),
	date: z.string(),
	notes: z.string(),
	editable: z.boolean(),
	/** Lo abonó sola la tasa de la cuenta, no el dueño. */
	automatic: z.boolean().default(false),
	/** El activo cuyo dividendo o venta es este abono, en el mismo portafolio; vacío en los demás. */
	originTicker: z.string().default(''),
	portfolioId: z.string(),
	portfolioName: z.string(),
	sourceId: z.string(),
	sourceName: z.string(),
	ticker: z.string(),
	createdAt: z.string()
});

/** `data` de los movimientos paginados. */
export const pagedCashMovementsSchema = z.object({
	data: z.array(cashMovementSchema),
	total: z.number(),
	page: z.number(),
	limit: z.number(),
	totalPages: z.number()
});

/**
 * Una versión de la tasa que rinde una cuenta (`GET /portfolios/cash/rates`).
 *
 * La tasa es de la cuenta —plataforma y moneda—, no del portafolio. Cambiarla
 * es anotar una versión nueva desde un día: la anterior termina la víspera y
 * los días pasados conservan la suya.
 */
export const cashRateSchema = z.object({
	id: z.string(),
	sourceId: z.string(),
	sourceName: z.string(),
	currency: z.string(),
	/**
	 * El bolsillo que rinde esta tasa; `null` es la cuenta principal. Una
	 * plataforma puede pagar una tasa en la cuenta y otra en su bolsillo, y cada
	 * una rinde sobre sus propios saldos.
	 */
	pocketId: z.string().nullable().default(null),
	pocketName: z.string().default(''),
	/** Tasa efectiva anual en porcentaje: `"9.25"` es 9,25 % E.A. */
	annualRatePct: z.string(),
	/** Parte de los intereses que se retiene, en porcentaje. */
	withholdingPct: z.string(),
	/**
	 * `daily` abona cada día; `monthly`, el último día de cada mes;
	 * `at_maturity` calcula cada día y lo abona todo el último, y solo lo tiene
	 * un depósito a tasa fija, que es el único con un último día.
	 */
	posting: z.enum(['daily', 'monthly', 'at_maturity']),
	/**
	 * Los tramos por encima de `annualRatePct`, del más bajo al más alto: la
	 * parte de la cuenta por encima de `fromBalance` rinde su tasa, hasta el
	 * siguiente. Un tramo al 0 % es un tope. Son de la cuenta, así que sus saldos
	 * se los reparten en proporción a lo que guarda cada uno. Vacío si la cuenta
	 * rinde `annualRatePct` sobre todo el saldo.
	 */
	tiers: z.array(z.object({ fromBalance: z.string(), annualRatePct: z.string() })).default([]),
	/** Primer día de la versión, a medianoche UTC. */
	effectiveFrom: z.string(),
	/** Último día que rinde; `null` mientras no tiene fin. */
	endedOn: z.string().nullable(),
	/** La versión más reciente de su cuenta: la única que se pausa, corrige o borra. */
	latest: z.boolean(),
	/**
	 * Último día con intereses calculados a esta versión. Con él la versión ya no
	 * se corrige ni se borra, y no puede terminar antes.
	 */
	accruedThrough: z.string().nullable().default(null),
	createdAt: z.string(),
	updatedAt: z.string()
});

/**
 * Un bolsillo de una cuenta (`GET /portfolios/cash/pockets`): la «cajita»
 * dentro de la cuenta de ahorros, el subsaldo dentro del bróker.
 *
 * Pertenece a una plataforma y una moneda, y su dinero cuenta dentro de esa
 * plataforma: ninguna cifra por plataforma cambia porque el dinero esté en un
 * bolsillo. Lo propio del bolsillo es su tasa.
 */
export const cashPocketSchema = z.object({
	id: z.string(),
	sourceId: z.string(),
	sourceName: z.string(),
	currency: z.string(),
	name: z.string(),
	/** `flexible` admite depósitos y retiros; `fixed` es un depósito a plazo. */
	kind: z.enum(['flexible', 'fixed']),
	/** El día en que empezó a guardar dinero. */
	openedOn: z.string(),
	/** Cuándo vence un depósito a tasa fija; `null` en uno flexible. */
	maturesOn: z.string().nullable().default(null),
	/** El día en que se cerró; `null` mientras sigue abierto. */
	closedOn: z.string().nullable().default(null),
	/** Lo que guardan sus saldos juntos, en `currency`, y en cuántos portafolios. */
	balance: z.string().default('0'),
	balances: z.number().default(0),
	/** Cuántos movimientos se anotaron en él: con alguno ya no se puede borrar. */
	movements: z.number().default(0),
	createdAt: z.string(),
	updatedAt: z.string()
});

/** Las dos patas de un movimiento entre bolsillos (`POST /movements/move`). */
export const cashMoveSchema = z.object({
	from: cashMovementSchema,
	to: cashMovementSchema
});

/**
 * Una racha de días de intereses: cuántos, de cuándo a cuándo y lo que
 * rindieron netos de retención, como decimal en texto.
 */
const cashInterestDaysSchema = z.object({
	days: z.number(),
	from: z.string().nullable(),
	through: z.string().nullable(),
	net: z.string()
});

/**
 * Lo que hizo un recálculo de intereses (`POST /cash/interest/recalculate`).
 *
 * Llega hasta el último día que la cuenta ya tenía calculado, no más: el día
 * siguiente lo calcula el proceso nocturno. `before` es lo que rendían los días
 * que se tiraron y `after` lo que rinden ahora esos mismos días: la diferencia
 * es lo que cambió el recálculo. `new` son días que ningún saldo tenía
 * calculados todavía, aparte porque suman sin que nada haya cambiado.
 */
export const cashRecalculationSchema = z.object({
	cleared: z.object({
		from: z.string(),
		balances: z.number(),
		days: z.number()
	}),
	through: z.string(),
	credited: z.number(),
	recomputed: z.number(),
	before: cashInterestDaysSchema,
	after: cashInterestDaysSchema,
	new: cashInterestDaysSchema
});
