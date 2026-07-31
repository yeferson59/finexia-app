/**
 * Schemas Zod de los formularios de plataformas.
 *
 * Estaban dentro de las actions de `routes/dashboard/platforms/**`.
 */

import { z } from 'zod';

/** Alta de una plataforma (`routes/dashboard/platforms/add`). */
export const platformCreateSchema = z.object({
	name: z.coerce.string().min(2),
	description: z.coerce.string().optional(),
	type: z.coerce.string().min(2)
});

/** Edición de una plataforma (`routes/dashboard/platforms/[id]`). */
export const platformUpdateSchema = z.object({
	name: z.string().min(2),
	description: z.string().optional().default(''),
	type: z.string().min(2),
	// El <select> envía "true"/"false" como string; z.coerce.boolean()
	// convertiría "false" en true, así que se compara explícitamente.
	isActive: z.enum(['true', 'false']).transform((v) => v === 'true')
});
