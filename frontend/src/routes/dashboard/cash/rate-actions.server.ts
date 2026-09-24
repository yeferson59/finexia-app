/**
 * Acciones de la tasa de una cuenta de efectivo: anotarla, corregirla,
 * pausarla, moverla, borrarla y recalcular sus intereses.
 *
 * Están fuera de `+page.server.ts` porque la tasa es un dominio propio dentro
 * del efectivo y porque esa página pasaba el presupuesto de 500 líneas que
 * comprueba `check:arch`. Se recomponen allí con un spread, como las
 * credenciales MCP de los ajustes, así que para SvelteKit no hay diferencia.
 */
import type { Actions } from './$types';
import { fail } from '@sveltejs/kit';
import * as cash from '$lib/api/cash';
import {
	cashRateCreateSchema,
	cashRateDeleteSchema,
	cashRateEndSchema,
	cashRateErrorMessage,
	cashRateRescheduleSchema,
	cashRateFields,
	cashRateUpdateSchema,
	cashRecalculateSchema,
	CASH_RECALCULATE_FALLBACK,
	toCalendarDateTime,
	toCashRateBody
} from '$lib/features/cash';

/**
 * El `fail` de una respuesta que no salió: el estado que dio el backend, o un
 * 500 cuando lo que falló fue el viaje y no hay ninguno.
 */
export function failed(res: { status: number }, error: string) {
	return fail(res.status >= 400 ? res.status : 500, { error });
}

export const cashRateActions = {
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

		// Lo que hizo, para que la página lo diga: sin esto un recálculo que da lo
		// mismo no se distingue de uno que no se hizo.
		return { success: true, recalculation: res.data };
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
