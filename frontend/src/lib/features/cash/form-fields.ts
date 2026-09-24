/**
 * Campos que comparten varios formularios de efectivo: el bolsillo sobre el que
 * cae algo, y los valores de una tasa, que piden igual la tasa de una cuenta y
 * la de un depósito a plazo.
 */

import { z } from 'zod';

/**
 * Un bolsillo del formulario: vacío es la cuenta principal, que es donde cae
 * todo mientras la cuenta no tenga cajitas.
 */
export const pocketField = z
	.union([z.uuid('Elige el bolsillo.'), z.literal('')])
	.nullish()
	.transform((v) => v || undefined);

/** Si `value` no tiene más decimales de los que caben, sin tropezar con la coma flotante. */
function hasAtMostDecimals(value: number, decimals: number): boolean {
	const scaled = value * 10 ** decimals;
	return Math.abs(scaled - Math.round(scaled)) < 1e-6;
}

/**
 * Los valores de una versión de la tasa, en porcentaje como los publica la
 * entidad: `9.25` es 9,25 % E.A. Los decimales son los que guarda el backend.
 */
export const rateValues = {
	annualRatePct: z.coerce
		.number('Escribe la tasa con números.')
		.positive('La tasa tiene que ser mayor que cero.')
		.max(100, 'La tasa no puede pasar de 100 %.')
		.refine((v) => hasAtMostDecimals(v, 4), 'Escribe la tasa con hasta cuatro decimales.'),
	// Vacía es que no hay retención.
	withholdingPct: z.coerce
		.number('Escribe la retención con números.')
		.min(0, 'La retención no puede ser negativa.')
		.lt(100, 'La retención tiene que ser menor que 100 %.')
		.refine((v) => hasAtMostDecimals(v, 2), 'Escribe la retención con hasta dos decimales.')
		.default(0),
	/** Cada cuánto la entidad abona lo que la cuenta rinde. */
	posting: z
		.enum(['daily', 'monthly'], 'Elige cada cuánto se abonan los intereses.')
		.default('daily'),
	/**
	 * Los tramos por encima de la tasa, del más bajo al más alto: desde qué saldo
	 * de la cuenta rige cada uno y a qué tasa. Un tramo al 0 % es un tope. Sin
	 * tramos, la cuenta rinde la tasa sobre todo el saldo.
	 */
	tiers: z
		.array(
			z.object({
				fromBalance: z.coerce
					.number('Escribe el saldo de cada tramo con números.')
					.positive('El saldo desde el que rige un tramo tiene que ser mayor que cero.')
					.lt(1e12, 'El saldo de un tramo es demasiado grande.')
					.refine(
						(v) => hasAtMostDecimals(v, 8),
						'Escribe el saldo de cada tramo con hasta ocho decimales.'
					),
				annualRatePct: z.coerce
					.number('Escribe la tasa de cada tramo con números.')
					.min(0, 'La tasa de un tramo no puede ser negativa.')
					.max(100, 'La tasa de un tramo no puede pasar de 100 %.')
					.refine(
						(v) => hasAtMostDecimals(v, 4),
						'Escribe la tasa de cada tramo con hasta cuatro decimales.'
					)
			})
		)
		.max(10, 'Una tasa tiene como mucho 10 tramos.')
		.refine(
			(tiers) => tiers.every((t, i) => i === 0 || t.fromBalance > tiers[i - 1].fromBalance),
			'Cada tramo tiene que empezar en un saldo mayor que el anterior.'
		)
		.default([])
};
