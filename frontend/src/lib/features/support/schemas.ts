/**
 * Schema Zod del formulario de aporte (`/apoyar`, acción `pagar`).
 *
 * El formulario manda `preset` (uno de los montos sugeridos u `otro`) y, si es
 * `otro`, el monto escrito en `custom`. Sale un entero en pesos dentro del
 * rango que acepta Bold.
 */

import { z } from 'zod';
import { formatCurrency } from '$lib/shared/format/money';
import { MAX_AMOUNT, MIN_AMOUNT, parsePesos } from './support';

export const supportAmountSchema = z
	.object({
		preset: z.string(),
		custom: z.string().optional()
	})
	.transform(({ preset, custom }) => parsePesos(preset === 'otro' ? (custom ?? '') : preset))
	.pipe(
		z
			.number({ error: 'Escribe cuánto quieres aportar.' })
			.int()
			.min(MIN_AMOUNT, `El aporte mínimo es de ${formatCurrency(MIN_AMOUNT, 'COP')}.`)
			.max(MAX_AMOUNT, `El aporte máximo es de ${formatCurrency(MAX_AMOUNT, 'COP')}.`)
	);
