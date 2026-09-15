import type { Actions, PageServerLoad } from './$types';
import { fail } from '@sveltejs/kit';
import * as cash from '$lib/api/cash';
import * as portfolio from '$lib/api/portfolio';
import * as platforms from '$lib/api/platforms';
import { resolveDisplayCurrency } from '$lib/shared/currency';
import {
	cashErrorMessage,
	cashMovementCreateSchema,
	cashMovementDeleteSchema,
	cashMovementUpdateSchema,
	cashRateCreateSchema,
	cashRateDeleteSchema,
	cashRateEndSchema,
	cashRateErrorMessage,
	cashRateUpdateSchema,
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

	const [balancesRes, movementsRes, ratesRes, portfoliosRes, platformsRes] = await Promise.all([
		cash.getBalances(event, currency),
		cash.getMovements(event, 1, MOVEMENTS_LIMIT),
		cash.getRates(event),
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

function movementFields(formData: FormData) {
	return {
		kind: formData.get('kind'),
		amount: formData.get('amount'),
		fees: formData.get('fees'),
		date: formData.get('date'),
		notes: formData.get('notes')
	};
}

function rateFields(formData: FormData) {
	return {
		annualRatePct: formData.get('annualRatePct'),
		withholdingPct: formData.get('withholdingPct'),
		posting: formData.get('posting') ?? 'daily',
		maxBalance: formData.get('maxBalance')
	};
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
			...movementFields(formData),
			portfolioId: formData.get('portfolioId'),
			sourceId: formData.get('sourceId'),
			currency: formData.get('currency')
		});

		if (!parsed.success) {
			return fail(400, { error: parsed.error.issues[0].message });
		}

		const { portfolioId, sourceId, currency, ...movement } = parsed.data;
		const res = await cash.createMovement(
			{ cookies, fetch },
			{ portfolioId, sourceId, currency, ...toCashMovementBody(movement) }
		);

		if (!res.ok || !res.success) {
			return fail(res.status >= 400 ? res.status : 500, {
				error: cashErrorMessage(res.status, res.details)
			});
		}

		return { success: true };
	},

	update: async ({ request, cookies, fetch }) => {
		const formData = await request.formData();

		const parsed = cashMovementUpdateSchema.safeParse({
			...movementFields(formData),
			id: formData.get('id')
		});

		if (!parsed.success) {
			return fail(400, { error: parsed.error.issues[0].message });
		}

		const { id, ...movement } = parsed.data;
		const res = await cash.updateMovement({ cookies, fetch }, id, toCashMovementBody(movement));

		if (!res.ok || !res.success) {
			return fail(res.status >= 400 ? res.status : 500, {
				error: cashErrorMessage(res.status, res.details)
			});
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
			return fail(res.status >= 400 ? res.status : 500, {
				error:
					res.status === 409
						? 'No se puede borrar: ese dinero ya salió en un retiro y el saldo quedaría en negativo. Borra o reduce antes el retiro.'
						: cashErrorMessage(res.status, res.details)
			});
		}

		return { success: true };
	},

	createRate: async ({ request, cookies, fetch }) => {
		const formData = await request.formData();

		const parsed = cashRateCreateSchema.safeParse({
			...rateFields(formData),
			sourceId: formData.get('sourceId'),
			currency: formData.get('currency'),
			effectiveFrom: formData.get('effectiveFrom')
		});

		if (!parsed.success) {
			return fail(400, { error: parsed.error.issues[0].message });
		}

		const { sourceId, currency, effectiveFrom, ...values } = parsed.data;
		const res = await cash.createRate(
			{ cookies, fetch },
			{
				sourceId,
				currency,
				effectiveFrom: toCalendarDateTime(effectiveFrom),
				...toCashRateBody(values)
			}
		);

		if (!res.ok || !res.success) {
			return fail(res.status >= 400 ? res.status : 500, {
				error: cashRateErrorMessage(res.status, res.details)
			});
		}

		return { success: true };
	},

	updateRate: async ({ request, cookies, fetch }) => {
		const formData = await request.formData();

		const parsed = cashRateUpdateSchema.safeParse({
			...rateFields(formData),
			id: formData.get('id')
		});

		if (!parsed.success) {
			return fail(400, { error: parsed.error.issues[0].message });
		}

		const { id, ...values } = parsed.data;
		const res = await cash.updateRate({ cookies, fetch }, id, toCashRateBody(values));

		if (!res.ok || !res.success) {
			return fail(res.status >= 400 ? res.status : 500, {
				error: cashRateErrorMessage(res.status, res.details)
			});
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
			return fail(res.status >= 400 ? res.status : 500, {
				error: cashRateErrorMessage(res.status, res.details)
			});
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
				from: toCalendarDateTime(parsed.data.from)
			}
		);

		if (!res.ok || !res.success) {
			return fail(res.status >= 400 ? res.status : 500, {
				error: cashRateErrorMessage(res.status, res.details, CASH_RECALCULATE_FALLBACK)
			});
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
			return fail(res.status >= 400 ? res.status : 500, {
				error: cashRateErrorMessage(res.status, res.details)
			});
		}

		return { success: true };
	}
} satisfies Actions;
