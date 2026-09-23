import type { Actions, PageServerLoad } from './$types';
import { fail } from '@sveltejs/kit';
import * as cash from '$lib/api/cash';
import * as portfolio from '$lib/api/portfolio';
import * as platforms from '$lib/api/platforms';
import { resolveDisplayCurrency } from '$lib/shared/currency';
import {
	cashErrorMessage,
	cashMovementFields,
	cashMoveSchema,
	cashMovementCreateSchema,
	cashMovementDeleteSchema,
	cashMovementUpdateSchema,
	cashRateCreateSchema,
	cashRateDeleteSchema,
	cashRateEndSchema,
	cashRateErrorMessage,
	cashRateRescheduleSchema,
	cashRateFields,
	cashRateUpdateSchema,
	cashPocketCreateSchema,
	cashPocketDeleteSchema,
	cashPocketErrorMessage,
	cashPocketRenameSchema,
	cashDepositCloseSchema,
	cashDepositCreateSchema,
	cashDepositErrorMessage,
	cashRecalculateSchema,
	CASH_RECALCULATE_FALLBACK,
	toCalendarDateTime,
	toCashMovementBody,
	toCashRateBody
} from '$lib/features/cash';

/**
 * Los movimientos que trae la página, de los más recientes. Es el máximo que el
 * backend sirve por petición, y se paginan en la pantalla.
 */
const MOVEMENTS_LIMIT = 100;

export const load: PageServerLoad = async ({ cookies, fetch, url, locals }) => {
	const event = { cookies, fetch };

	// Los saldos están en monedas distintas, así que se piden convertidos a la
	// de la cuenta —o a la de `?currency=`— para que el total se pueda sumar.
	const currency = resolveDisplayCurrency(
		url.searchParams.get('currency'),
		locals.user?.preferredCurrency
	);

	const [balancesRes, movementsRes, ratesRes, pocketsRes, portfoliosRes, platformsRes] =
		await Promise.all([
			cash.getBalances(event, currency),
			cash.getMovements(event, 1, MOVEMENTS_LIMIT),
			cash.getRates(event),
			cash.getPockets(event),
			portfolio.getSummaries(event),
			platforms.getSources(event)
		]);

	return {
		currency,
		loadFailed: !balancesRes.success,
		balances: balancesRes.success ? (balancesRes.data ?? []) : [],
		movements: movementsRes.success ? (movementsRes.data?.data ?? []) : [],
		movementsTotal: movementsRes.success ? (movementsRes.data?.total ?? 0) : 0,
		// Sin tasas las cuentas se ven igual, solo que sin rentabilidad: no es un
		// fallo de la página.
		rates: ratesRes.success ? (ratesRes.data ?? []) : [],
		// Los bolsillos vacíos no tienen saldo, así que no llegan con los saldos:
		// esta es la lista que permite enseñarlos y darles tasa antes de meterles
		// dinero.
		pockets: pocketsRes.success ? (pocketsRes.data ?? []) : [],
		portfolios: (portfoliosRes.data ?? []).map((p) => ({
			id: p.id,
			name: p.name,
			isDefault: p.isDefault === true
		})),
		// Una plataforma inactiva no recibe dinero nuevo; sus saldos siguen en la lista.
		platforms: (platformsRes.data ?? [])
			.filter((p) => p.isActive)
			.map((p) => ({ id: p.id, name: p.name }))
	};
};

/**
 * El `fail` de una respuesta que no salió: el estado que dio el backend, o un
 * 500 cuando lo que falló fue el viaje y no hay ninguno.
 */
function failed(res: { status: number }, error: string) {
	return fail(res.status >= 400 ? res.status : 500, { error });
}

/*
 * Cada acción devuelve `fail` con el motivo ya escrito para el usuario. El
 * formulario lo lee del resultado, no del `form` de la página, así que una
 * ventana que se cierra y se vuelve a abrir no arrastra el error anterior.
 */
export const actions = {
	create: async ({ request, cookies, fetch }) => {
		const formData = await request.formData();

		const parsed = cashMovementCreateSchema.safeParse({
			...cashMovementFields(formData),
			portfolioId: formData.get('portfolioId'),
			sourceId: formData.get('sourceId'),
			currency: formData.get('currency'),
			pocketId: formData.get('pocketId')
		});

		if (!parsed.success) {
			return fail(400, { error: parsed.error.issues[0].message });
		}

		const { portfolioId, sourceId, currency, pocketId, ...movement } = parsed.data;
		const res = await cash.createMovement(
			{ cookies, fetch },
			{ portfolioId, sourceId, currency, pocketId, ...toCashMovementBody(movement) }
		);

		if (!res.ok || !res.success) {
			return failed(res, cashErrorMessage(res.status, res.details));
		}

		return { success: true };
	},

	update: async ({ request, cookies, fetch }) => {
		const formData = await request.formData();

		const parsed = cashMovementUpdateSchema.safeParse({
			...cashMovementFields(formData),
			id: formData.get('id')
		});

		if (!parsed.success) {
			return fail(400, { error: parsed.error.issues[0].message });
		}

		const { id, ...movement } = parsed.data;
		const res = await cash.updateMovement({ cookies, fetch }, id, toCashMovementBody(movement));

		if (!res.ok || !res.success) {
			return failed(res, cashErrorMessage(res.status, res.details));
		}

		return { success: true };
	},

	delete: async ({ request, cookies, fetch }) => {
		const formData = await request.formData();

		const parsed = cashMovementDeleteSchema.safeParse({ id: formData.get('id') });

		if (!parsed.success) {
			return fail(400, { error: parsed.error.issues[0].message });
		}

		const res = await cash.deleteMovement({ cookies, fetch }, parsed.data.id);

		if (!res.ok) {
			return failed(
				res,
				res.status === 409
					? 'No se puede borrar: ese dinero ya salió en un retiro y el saldo quedaría en negativo. Borra o reduce antes el retiro.'
					: cashErrorMessage(res.status, res.details)
			);
		}

		return { success: true };
	},

	createRate: async ({ request, cookies, fetch }) => {
		const formData = await request.formData();

		const parsed = cashRateCreateSchema.safeParse({
			...cashRateFields(formData),
			sourceId: formData.get('sourceId'),
			currency: formData.get('currency'),
			pocketId: formData.get('pocketId'),
			effectiveFrom: formData.get('effectiveFrom'),
			recompute: formData.get('recompute')
		});

		if (!parsed.success) {
			return fail(400, { error: parsed.error.issues[0].message });
		}

		const { sourceId, currency, pocketId, effectiveFrom, recompute, ...values } = parsed.data;
		const res = await cash.createRate(
			{ cookies, fetch },
			{
				sourceId,
				currency,
				pocketId,
				effectiveFrom: toCalendarDateTime(effectiveFrom),
				recompute,
				...toCashRateBody(values)
			}
		);

		if (!res.ok || !res.success) {
			return failed(res, cashRateErrorMessage(res.status, res.details));
		}

		return { success: true };
	},

	updateRate: async ({ request, cookies, fetch }) => {
		const formData = await request.formData();

		const parsed = cashRateUpdateSchema.safeParse({
			...cashRateFields(formData),
			id: formData.get('id')
		});

		if (!parsed.success) {
			return fail(400, { error: parsed.error.issues[0].message });
		}

		const { id, ...values } = parsed.data;
		const res = await cash.updateRate({ cookies, fetch }, id, toCashRateBody(values));

		if (!res.ok || !res.success) {
			return failed(res, cashRateErrorMessage(res.status, res.details));
		}

		return { success: true };
	},

	endRate: async ({ request, cookies, fetch }) => {
		const formData = await request.formData();

		const parsed = cashRateEndSchema.safeParse({
			id: formData.get('id'),
			endsOn: formData.get('endsOn')
		});

		if (!parsed.success) {
			return fail(400, { error: parsed.error.issues[0].message });
		}

		const res = await cash.endRate({ cookies, fetch }, parsed.data.id, {
			endsOn: toCalendarDateTime(parsed.data.endsOn)
		});

		if (!res.ok || !res.success) {
			return failed(res, cashRateErrorMessage(res.status, res.details));
		}

		return { success: true };
	},

	/* Mover el día en que empieza la versión más reciente. */
	rescheduleRate: async ({ request, cookies, fetch }) => {
		const formData = await request.formData();

		const parsed = cashRateRescheduleSchema.safeParse({
			id: formData.get('id'),
			effectiveFrom: formData.get('effectiveFrom'),
			recompute: formData.get('recompute')
		});

		if (!parsed.success) {
			return fail(400, { error: parsed.error.issues[0].message });
		}

		const res = await cash.rescheduleRate({ cookies, fetch }, parsed.data.id, {
			effectiveFrom: toCalendarDateTime(parsed.data.effectiveFrom),
			recompute: parsed.data.recompute
		});

		if (!res.ok || !res.success) {
			return failed(res, cashRateErrorMessage(res.status, res.details));
		}

		return { success: true };
	},

	/*
	 * Recalcular no es un cambio de tasa: tira los días que la cuenta ya tenía
	 * calculados desde una fecha y los vuelve a calcular sobre lo que guarda hoy.
	 * Es la salida cuando se anota un movimiento con fecha pasada.
	 */
	recalculateInterest: async ({ request, cookies, fetch }) => {
		const formData = await request.formData();

		const parsed = cashRecalculateSchema.safeParse({
			sourceId: formData.get('sourceId'),
			currency: formData.get('currency'),
			pocketId: formData.get('pocketId'),
			from: formData.get('from')
		});

		if (!parsed.success) {
			return fail(400, { error: parsed.error.issues[0].message });
		}

		const res = await cash.recalculateInterest(
			{ cookies, fetch },
			{
				sourceId: parsed.data.sourceId,
				currency: parsed.data.currency,
				pocketId: parsed.data.pocketId,
				from: toCalendarDateTime(parsed.data.from)
			}
		);

		if (!res.ok || !res.success) {
			return failed(res, cashRateErrorMessage(res.status, res.details, CASH_RECALCULATE_FALLBACK));
		}

		return { success: true };
	},

	/*
	 * Los bolsillos: abrir, renombrar y borrar. Un bolsillo es una subcuenta de
	 * la cuenta —su dinero sigue contando en la plataforma— con su propia tasa.
	 */
	createPocket: async ({ request, cookies, fetch }) => {
		const formData = await request.formData();

		const parsed = cashPocketCreateSchema.safeParse({
			sourceId: formData.get('sourceId'),
			currency: formData.get('currency'),
			name: formData.get('name')
		});

		if (!parsed.success) {
			return fail(400, { error: parsed.error.issues[0].message });
		}

		const res = await cash.createPocket({ cookies, fetch }, parsed.data);

		if (!res.ok || !res.success) {
			return failed(res, cashPocketErrorMessage(res.status, res.details));
		}

		return { success: true };
	},

	renamePocket: async ({ request, cookies, fetch }) => {
		const formData = await request.formData();

		const parsed = cashPocketRenameSchema.safeParse({
			id: formData.get('id'),
			name: formData.get('name')
		});

		if (!parsed.success) {
			return fail(400, { error: parsed.error.issues[0].message });
		}

		const res = await cash.renamePocket({ cookies, fetch }, parsed.data.id, {
			name: parsed.data.name
		});

		if (!res.ok || !res.success) {
			return failed(res, cashPocketErrorMessage(res.status, res.details));
		}

		return { success: true };
	},

	deletePocket: async ({ request, cookies, fetch }) => {
		const formData = await request.formData();

		const parsed = cashPocketDeleteSchema.safeParse({ id: formData.get('id') });

		if (!parsed.success) {
			return fail(400, { error: parsed.error.issues[0].message });
		}

		const res = await cash.deletePocket({ cookies, fetch }, parsed.data.id);

		if (!res.ok) {
			return failed(res, cashPocketErrorMessage(res.status, res.details));
		}

		return { success: true };
	},

	/*
	 * Los depósitos a tasa fija: abrirlos y cancelarlos. Borrar uno entero es la
	 * misma acción que borra un bolsillo, porque un depósito es uno.
	 *
	 * Abrirlo dice el dinero, el plazo y la tasa a la vez: son una misma cosa, y
	 * la tasa no se le da después. La fecha de apertura puede estar en el pasado,
	 * y el backend calcula de una vez los días que ya ganó.
	 */
	openDeposit: async ({ request, cookies, fetch }) => {
		const formData = await request.formData();

		const parsed = cashDepositCreateSchema.safeParse({
			portfolioId: formData.get('portfolioId'),
			sourceId: formData.get('sourceId'),
			currency: formData.get('currency'),
			name: formData.get('name'),
			amount: formData.get('amount'),
			openedOn: formData.get('openedOn'),
			maturesOn: formData.get('maturesOn'),
			annualRatePct: formData.get('annualRatePct'),
			withholdingPct: formData.get('withholdingPct'),
			posting: formData.get('posting') ?? 'daily'
		});

		if (!parsed.success) {
			return fail(400, { error: parsed.error.issues[0].message });
		}

		const { openedOn, maturesOn, ...deposit } = parsed.data;
		const res = await cash.openDeposit(
			{ cookies, fetch },
			{
				...deposit,
				openedOn: toCalendarDateTime(openedOn),
				// Sin plazo viaja como null, que es lo que el backend lee como
				// «hasta que lo canceles».
				maturesOn: maturesOn ? toCalendarDateTime(maturesOn) : null
			}
		);

		if (!res.ok || !res.success) {
			return failed(res, cashDepositErrorMessage(res.status, res.details));
		}

		return { success: true };
	},

	closeDeposit: async ({ request, cookies, fetch }) => {
		const formData = await request.formData();

		const parsed = cashDepositCloseSchema.safeParse({
			id: formData.get('id'),
			closesOn: formData.get('closesOn'),
			penalty: formData.get('penalty')
		});

		if (!parsed.success) {
			return fail(400, { error: parsed.error.issues[0].message });
		}

		const res = await cash.closeDeposit({ cookies, fetch }, parsed.data.id, {
			closesOn: toCalendarDateTime(parsed.data.closesOn),
			penalty: parsed.data.penalty
		});

		if (!res.ok || !res.success) {
			return failed(res, cashDepositErrorMessage(res.status, res.details));
		}

		return { success: true };
	},

	/*
	 * Mover no es un retiro y un depósito anotados a mano: las dos patas van en
	 * una sola transacción y se compensan, así que la rentabilidad no se mueve.
	 * Vale entre cajones de una cuenta y entre dos plataformas, que es el
	 * traslado al bróker antes de comprar; cruzando monedas el formulario dice
	 * lo que sale y lo que llega, que es lo que trae el extracto.
	 */
	move: async ({ request, cookies, fetch }) => {
		const formData = await request.formData();

		const parsed = cashMoveSchema.safeParse({
			portfolioId: formData.get('portfolioId'),
			sourceId: formData.get('sourceId'),
			currency: formData.get('currency'),
			fromPocketId: formData.get('fromPocketId'),
			toSourceId: formData.get('toSourceId'),
			toCurrency: formData.get('toCurrency'),
			toPocketId: formData.get('toPocketId'),
			amount: formData.get('amount'),
			toAmount: formData.get('toAmount'),
			date: formData.get('date'),
			notes: formData.get('notes')
		});

		if (!parsed.success) {
			return fail(400, { error: parsed.error.issues[0].message });
		}

		const { date, ...move } = parsed.data;
		const res = await cash.moveCash(
			{ cookies, fetch },
			{ ...move, date: toCalendarDateTime(date) }
		);

		if (!res.ok || !res.success) {
			return failed(
				res,
				res.status === 409
					? 'No hay tanto dinero en el saldo del que sale. Mira lo que guarda y prueba con menos.'
					: cashErrorMessage(res.status, res.details)
			);
		}

		return { success: true };
	},

	deleteRate: async ({ request, cookies, fetch }) => {
		const formData = await request.formData();

		const parsed = cashRateDeleteSchema.safeParse({ id: formData.get('id') });

		if (!parsed.success) {
			return fail(400, { error: parsed.error.issues[0].message });
		}

		const res = await cash.deleteRate({ cookies, fetch }, parsed.data.id);

		if (!res.ok) {
			return failed(res, cashRateErrorMessage(res.status, res.details));
		}

		return { success: true };
	}
} satisfies Actions;
