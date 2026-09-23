/**
 * Aportes de `/apoyar` — contratos HTTP como schemas Zod (`docs/API.md` §2.13).
 *
 * Parte de `lib/api/schemas`: el porqué de que sean schemas y no interfaces
 * está en el `index.ts` de la carpeta.
 */

import { z } from 'zod';

/** `GET /support`: si se pueden recibir aportes y entre qué montos. */
export const supportConfigSchema = z.object({
	enabled: z.boolean(),
	currency: z.string(),
	minAmount: z.number(),
	maxAmount: z.number()
});

/**
 * `POST /support/checkouts`: la orden firmada, campo por campo como la recibe
 * `new BoldCheckout(…)` en el navegador. La llave secreta nunca viaja: solo la
 * firma que se calculó con ella.
 */
export const supportCheckoutSchema = z.object({
	orderId: z.string(),
	currency: z.string(),
	amount: z.string(),
	apiKey: z.string(),
	integritySignature: z.string(),
	redirectionUrl: z.string(),
	description: z.string(),
	renderMode: z.literal('embedded')
});

/** Estados de un aporte, tal como los guarda el backend. */
export const supportStatusSchema = z.enum(['created', 'pending', 'rejected', 'approved', 'voided']);

/** `GET /support/checkouts/:orderId`: la orden y lo que Bold dijo de ella. */
export const supportContributionSchema = z.object({
	orderId: z.string(),
	amount: z.number(),
	currency: z.string(),
	status: supportStatusSchema,
	totalCharged: z.number().nullable(),
	approvedAt: z.string().nullable(),
	createdAt: z.string()
});

/**
 * Un aporte en el listado de administración (`GET /support/contributions`):
 * lo público más lo que identifica el pago en el panel de Bold.
 */
export const adminContributionSchema = supportContributionSchema.extend({
	paymentId: z.string(),
	paymentMethod: z.string(),
	updatedAt: z.string()
});

/** `GET /support/contributions/summary`: los aportes en conjunto. */
export const supportSummarySchema = z.object({
	counts: z.record(supportStatusSchema, z.number()),
	approvedTotal: z.number(),
	/** Lo aprobado en los últimos 30 días. */
	approvedRecent: z.number(),
	lastApprovedAt: z.string().nullable()
});
