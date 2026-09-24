import type { Actions, PageServerLoad } from './$types';
import { fail } from '@sveltejs/kit';
import * as funds from '$lib/api/funds';
import * as portfolio from '$lib/api/portfolio';
import * as platforms from '$lib/api/platforms';
import { resolveDisplayCurrency } from '$lib/shared/currency';
import {
	fundBalanceCreateSchema,
	fundContributionSchema,
	fundCreateSchema,
	fundDeleteSchema,
	fundErrorMessage,
	fundMarkDeleteSchema,
	fundMarkErrorMessage,
	fundMarkSchema,
	fundMarksBulkSchema,
	fundMovementDeleteSchema,
	fundMovementErrorMessage,
	fundWithdrawalSchema,
	toFundDateTime,
	type FundMark,
	type FundMovement,
	type FundPerformance
} from '$lib/features/funds';

export const load: PageServerLoad = async ({ cookies, fetch, locals }) => {
	const event = { cookies, fetch };

	const [fundsRes, portfoliosRes, platformsRes] = await Promise.all([
		funds.getFunds(event),
		portfolio.getSummaries(event),
		platforms.getSources(event)
	]);

	const list = fundsRes.success ? (fundsRes.data ?? []) : [];

	// Las marcas de cada fondo, y los movimientos de los que se siguen por saldo,
	// para los diálogos. Un usuario sigue pocos fondos, así que se piden todos a
	// la vez en vez de al abrir cada uno.
	const byBalance = list.filter((f) => f.tracking === 'balance');
	const [marksRes, movementsRes, performanceRes] = await Promise.all([
		Promise.all(list.map((f) => funds.getMarks(event, f.assetId))),
		Promise.all(byBalance.map((f) => funds.getMovements(event, f.assetId))),
		Promise.all(list.map((f) => funds.getPerformance(event, f.assetId)))
	]);

	const marks: Record<string, FundMark[]> = {};
	list.forEach((f, i) => {
		marks[f.assetId] = marksRes[i].success ? (marksRes[i].data ?? []) : [];
	});

	// Sin la rentabilidad de un fondo su tarjeta se ve igual, solo que sin la
	// cifra: no es un fallo de la página.
	const performance: Record<string, FundPerformance> = {};
	list.forEach((f, i) => {
		const res = performanceRes[i];
		if (res.success && res.data) performance[f.assetId] = res.data;
	});

	const movements: Record<string, FundMovement[]> = {};
	byBalance.forEach((f, i) => {
		movements[f.assetId] = movementsRes[i].success ? (movementsRes[i].data ?? []) : [];
	});

	return {
		currency: resolveDisplayCurrency(null, locals.user?.preferredCurrency),
		loadFailed: !fundsRes.success,
		funds: list,
		marks,
		movements,
		performance,
		portfolios: (portfoliosRes.data ?? []).map((p) => ({
			id: p.id,
			name: p.name,
			isDefault: p.isDefault === true
		})),
		// Una plataforma inactiva no recibe dinero nuevo.
		platforms: (platformsRes.data ?? [])
			.filter((p) => p.isActive)
			.map((p) => ({ id: p.id, name: p.name }))
	};
};

function failed(res: { status: number }, error: string) {
	return fail(res.status >= 400 ? res.status : 500, { error });
}

/** Lo que va al backend al crear un fondo, según cómo se sigue. */
function createBody(formData: FormData) {
	const common = {
		portfolioId: formData.get('portfolioId'),
		sourceId: formData.get('sourceId'),
		currency: formData.get('currency'),
		name: formData.get('name'),
		date: formData.get('date'),
		currentDate: formData.get('currentDate')
	};

	if (formData.get('tracking') === 'balance') {
		const parsed = fundBalanceCreateSchema.safeParse({
			...common,
			amount: formData.get('amount'),
			currentBalance: formData.get('currentBalance')
		});
		if (!parsed.success) return { error: parsed.error.issues[0].message };

		const { date, currentBalance, currentDate, ...fund } = parsed.data;
		return {
			body: {
				...fund,
				tracking: 'balance',
				date: toFundDateTime(date),
				// Sin saldo de hoy no viaja fecha: el backend fecharía un saldo que no existe.
				...(currentBalance
					? {
							currentBalance,
							...(currentDate ? { currentDate: toFundDateTime(currentDate) } : {})
						}
					: {})
			}
		};
	}

	const parsed = fundCreateSchema.safeParse({
		...common,
		units: formData.get('units'),
		unitValue: formData.get('unitValue'),
		currentUnitValue: formData.get('currentUnitValue')
	});
	if (!parsed.success) return { error: parsed.error.issues[0].message };

	const { date, currentUnitValue, currentDate, ...fund } = parsed.data;
	return {
		body: {
			...fund,
			tracking: 'units',
			date: toFundDateTime(date),
			...(currentUnitValue
				? {
						currentUnitValue,
						...(currentDate ? { currentDate: toFundDateTime(currentDate) } : {})
					}
				: {})
		}
	};
}

/*
 * Cada acción devuelve `fail` con el motivo ya escrito para el usuario, como en
 * el efectivo: el diálogo lo lee del resultado de su envío.
 */
export const actions = {
	createFund: async ({ request, cookies, fetch }) => {
		const built = createBody(await request.formData());

		if (!built.body) {
			return fail(400, { error: built.error });
		}

		const res = await funds.createFund({ cookies, fetch }, built.body);

		if (!res.ok || !res.success) {
			return failed(res, fundErrorMessage(res.status, res.details));
		}

		return { success: true };
	},

	saveMark: async ({ request, cookies, fetch }) => {
		const formData = await request.formData();

		const parsed = fundMarkSchema.safeParse({
			id: formData.get('id'),
			date: formData.get('date'),
			unitValue: formData.get('unitValue'),
			balance: formData.get('balance'),
			notes: formData.get('notes')
		});

		if (!parsed.success) {
			return fail(400, { error: parsed.error.issues[0].message });
		}

		const { id, date, ...mark } = parsed.data;
		const res = await funds.saveMark({ cookies, fetch }, id, {
			...mark,
			date: toFundDateTime(date)
		});

		if (!res.ok || !res.success) {
			return failed(res, fundMarkErrorMessage(res.status, res.details));
		}

		return { success: true };
	},

	saveMarks: async ({ request, cookies, fetch }) => {
		const formData = await request.formData();

		const parsed = fundMarksBulkSchema.safeParse({
			id: formData.get('id'),
			tracking: formData.get('tracking'),
			marks: formData.get('marks')
		});

		if (!parsed.success) {
			return fail(400, { error: parsed.error.issues[0].message });
		}

		const { id, tracking, marks } = parsed.data;
		const res = await funds.saveMarks({ cookies, fetch }, id, {
			marks: marks.map((m) => ({
				date: toFundDateTime(m.date),
				...(tracking === 'balance' ? { balance: m.value } : { unitValue: m.value })
			}))
		});

		if (!res.ok || !res.success) {
			return failed(res, fundMarkErrorMessage(res.status, res.details));
		}

		return { success: true, saved: res.data?.saved ?? marks.length };
	},

	deleteMark: async ({ request, cookies, fetch }) => {
		const formData = await request.formData();

		const parsed = fundMarkDeleteSchema.safeParse({
			id: formData.get('id'),
			date: formData.get('date')
		});

		if (!parsed.success) {
			return fail(400, { error: parsed.error.issues[0].message });
		}

		const res = await funds.deleteMark({ cookies, fetch }, parsed.data.id, parsed.data.date);

		if (!res.ok || !res.success) {
			return failed(res, fundMarkErrorMessage(res.status, res.details));
		}

		return { success: true };
	},

	contribute: async ({ request, cookies, fetch }) => {
		const formData = await request.formData();

		const parsed = fundContributionSchema.safeParse({
			id: formData.get('id'),
			portfolioId: formData.get('portfolioId'),
			sourceId: formData.get('sourceId'),
			date: formData.get('date'),
			amount: formData.get('amount'),
			balanceBefore: formData.get('balanceBefore'),
			notes: formData.get('notes')
		});

		if (!parsed.success) {
			return fail(400, { error: parsed.error.issues[0].message });
		}

		const { id, date, ...contribution } = parsed.data;
		const res = await funds.contribute({ cookies, fetch }, id, {
			...contribution,
			date: toFundDateTime(date)
		});

		if (!res.ok || !res.success) {
			return failed(res, fundMovementErrorMessage(res.status, res.details));
		}

		return { success: true };
	},

	withdraw: async ({ request, cookies, fetch }) => {
		const formData = await request.formData();

		const parsed = fundWithdrawalSchema.safeParse({
			id: formData.get('id'),
			entryId: formData.get('entryId'),
			date: formData.get('date'),
			amount: formData.get('amount'),
			fees: formData.get('fees') || 0,
			all: formData.get('all'),
			notes: formData.get('notes')
		});

		if (!parsed.success) {
			return fail(400, { error: parsed.error.issues[0].message });
		}

		const { id, date, ...withdrawal } = parsed.data;
		const res = await funds.withdraw({ cookies, fetch }, id, {
			...withdrawal,
			date: toFundDateTime(date)
		});

		if (!res.ok || !res.success) {
			return failed(res, fundMovementErrorMessage(res.status, res.details));
		}

		return { success: true };
	},

	deleteMovement: async ({ request, cookies, fetch }) => {
		const formData = await request.formData();

		const parsed = fundMovementDeleteSchema.safeParse({ txnId: formData.get('txnId') });

		if (!parsed.success) {
			return fail(400, { error: parsed.error.issues[0].message });
		}

		const res = await funds.deleteMovement({ cookies, fetch }, parsed.data.txnId);

		if (!res.ok || !res.success) {
			return failed(res, fundMovementErrorMessage(res.status, res.details));
		}

		return { success: true };
	},

	deleteFund: async ({ request, cookies, fetch }) => {
		const formData = await request.formData();

		const parsed = fundDeleteSchema.safeParse({ id: formData.get('id') });

		if (!parsed.success) {
			return fail(400, { error: parsed.error.issues[0].message });
		}

		const res = await funds.deleteFund({ cookies, fetch }, parsed.data.id);

		if (!res.ok || !res.success) {
			return failed(res, fundErrorMessage(res.status, res.details));
		}

		return { success: true };
	}
} satisfies Actions;
