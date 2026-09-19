/**
 * Schema Zod de la cabecera de un artículo.
 *
 * El blog no tiene formularios; el borde que valida aquí es el otro, el de
 * entrada: un `.md` del repositorio que `posts.ts` convierte en una página
 * pública. Si a un artículo le falta la descripción, lo que se publica es una
 * página sin `meta description`, así que el build tiene que pararse antes.
 */

import { z } from 'zod';

/** Fecha de publicación en ISO corto; el orden del índice depende de ella. */
const ISO_DATE = /^\d{4}-\d{2}-\d{2}$/;

export const postFrontmatterSchema = z.object({
	title: z.string().min(1, 'El artículo necesita un `title`.'),
	description: z
		.string()
		.min(1, 'El artículo necesita una `description`: es su resumen y su meta description.'),
	date: z.string().regex(ISO_DATE, 'La `date` va en formato AAAA-MM-DD.'),
	tags: z.array(z.string().min(1)).min(1, 'El artículo necesita al menos una etiqueta.'),
	// Firma el equipo salvo que el artículo diga otra cosa.
	author: z.string().min(1).default('Equipo Finexia'),
	// Un borrador se ve en `pnpm dev` y no se publica.
	draft: z.boolean().default(false)
});
