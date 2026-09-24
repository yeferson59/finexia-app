import type { Actions, PageServerLoad } from './$types';
import { fail } from '@sveltejs/kit';
import * as funds from '$lib/api/funds';
import * as portfolio from '$lib/api/portfolio';
import * as platforms from '$lib/api/platforms';
import { resolveDisplayCurrency } from '$lib/shared/currency';
import {
	fundCreateSchema,
	fundDeleteSchema,
	fundErrorMessage,
	fundMarkDeleteSchema,
	fundMarkErrorMessage,
	fundMarkSchema,
	toFundDateTime,
	type FundMark
} from '$lib/features/funds';

export const load: PageServerLoad = async ({ cookies, fetch, locals }) => {
	const event = { cookies, fetch };

	const [fundsRes, portfoliosRes, platformsRes] = await Promise.all([
		funds.getFunds(event),
		portfolio.getSummaries(event),
		platforms.getSources(event)
	]);

	const list = fundsRes.success ? (fundsRes.data ?? []) : [];

	// Las marcas de cada fondo, para el historial del diálogo. Un usuario sigue
	// pocos fondos, así que se piden todas a la vez en vez de al abrirlo.
	const marksRes = await Promise.all(list.map((f) => funds.getMarks(event, f.assetId)));
	const marks: Record<string, FundMark[]> = {};
	list.forEach((f, i) => {
		marks[f.assetId] = marksRes[i].success ? (marksRes[i].data ?? []) : [];
	});

	return {
		currency: resolveDisplayCurrency(null, locals.user?.preferredCurrency),
		loadFailed: !fundsRes.success,
		funds: list,
		marks,
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

/*
 * Cada acción devuelve `fail` con el motivo ya escrito para el usuario, como en
 * el efectivo: el diálogo lo lee del resultado de su envío.
 */
export const actions = {
	createFund: async ({ request, cookies, fetch }) => {
		const formData = await request.formData();

		const parsed = fundCreateSchema.safeParse({
			portfolioId: formData.get('portfolioId'),
			sourceId: formData.get('sourceId'),
			currency: formData.get('currency'),
			name: formData.get('name'),
			date: formData.get('date'),
			units: formData.get('units'),
			unitValue: formData.get('unitValue'),
			currentUnitValue: formData.get('currentUnitValue'),
			currentDate: formData.get('currentDate')
		});

		if (!parsed.success) {
			return fail(400, { error: parsed.error.issues[0].message });
		}

		const { date, currentUnitValue, currentDate, ...fund } = parsed.data;
		const res = await funds.createFund(
			{ cookies, fetch },
			{
				...fund,
				tracking: 'units',
				date: toFundDateTime(date),
				// Sin valor de hoy no viaja fecha: el backend fecharía un valor que no existe.
				...(currentUnitValue
					? {
							currentUnitValue,
							...(currentDate ? { currentDate: toFundDateTime(currentDate) } : {})
						}
					: {})
			}
		);

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
