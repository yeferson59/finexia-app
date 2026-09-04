/**
 * Paginación — contratos HTTP como schemas Zod.
 *
 * Parte de `lib/api/schemas`: el porqué de que sean schemas y no interfaces
 * está en el `index.ts` de la carpeta.
 */

import { z } from 'zod';

// ---------------------------------------------------------------------------
// Paginación (§1.5 de docs/API.md)
// ---------------------------------------------------------------------------

/**
 * Bloque de metadatos de las rutas paginadas. Conserva nombres históricos por
 * área (`usersForPage`/`totalUsers`, …), que llegan como claves extra.
 */
export const pageMetaSchema = z
	.object({
		currentPage: z.number(),
		totalPages: z.number(),
		previous: z.boolean(),
		next: z.boolean(),
		offset: z.number().optional()
	})
	.catchall(z.union([z.number(), z.boolean()]));

/** `data` de una ruta paginada: lista de items + metadatos. */
export function paginatedSchema<T extends z.ZodType>(item: T) {
	return z.object({ items: z.array(item), metaData: pageMetaSchema });
}
