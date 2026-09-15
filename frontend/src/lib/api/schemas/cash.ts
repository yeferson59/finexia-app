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
	 * Los intereses abonados al saldo, en `currency`: todos, y los fechados en el
	 * mes en curso (UTC). `interestThisMonthValue` es esa parte en
	 * `displayCurrency`, con el mismo contrato de `fxConverted` que `value`.
	 */
	interestEarned: z.string().default('0'),
	interestThisMonth: z.string().default('0'),
	interestThisMonthValue: z.string().default('0'),
	/** Último día con intereses calculados; `null` si la cuenta nunca rindió. */
	lastAccrualDate: z.string().nullable().default(null)
});

/**
 * Un movimiento sobre un saldo (`GET /portfolios/cash/movements`).
 *
 * `kind` es lo que hizo el dueño; `type`, la transacción con que se guardó.
 * `other` es un movimiento anotado sobre el saldo desde la posición —un
 * interés cobrado fuera, una comisión suelta— que esta pantalla enseña pero no
 * sabe reescribir: por eso llega con `editable: false`.
 */
export const cashMovementSchema = z.object({
	id: z.string(),
	entryId: z.string(),
	type: z.string(),
	kind: z.enum(['deposit', 'withdrawal', 'interest', 'other']),
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
	/** Tasa efectiva anual en porcentaje: `"9.25"` es 9,25 % E.A. */
	annualRatePct: z.string(),
	/** Parte de los intereses que se retiene, en porcentaje. */
	withholdingPct: z.string(),
	posting: z.enum(['daily', 'monthly']),
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
