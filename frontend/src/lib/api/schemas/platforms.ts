/**
 * Plataformas — contratos HTTP como schemas Zod.
 *
 * Parte de `lib/api/schemas`: el porqué de que sean schemas y no interfaces
 * está en el `index.ts` de la carpeta.
 */

import { z } from 'zod';

// ---------------------------------------------------------------------------
// Plataformas / fuentes
// ---------------------------------------------------------------------------

/** Plataforma / fuente (`GET /portfolios/sources`). */
export const platformSchema = z.object({
	id: z.string(),
	name: z.string(),
	description: z.string(),
	sourceType: z.string(),
	/** Alias histórico de `sourceType` en algunas vistas/formularios. */
	type: z.string().optional(),
	isActive: z.boolean(),
	/** Posiciones abiertas: las vendidas del todo ya no son algo que se tenga. */
	investments: z.number(),
	/**
	 * Sobre cuántos activos y portafolios se reparten esas posiciones. Diez
	 * posiciones son una cuenta distinta si son diez empresas que si son una
	 * empresa repetida en diez portafolios.
	 */
	assets: z.number().optional(),
	portfolios: z.number().optional(),
	/**
	 * Lo invertido: cantidad × coste medio ponderado, en `displayCurrency`.
	 *
	 * Es el coste de lo que **sigue en cartera** — `quantity` es lo que queda
	 * tras las ventas — y no incluye comisiones ni lo que una venta realizó.
	 */
	totalValue: z.string(),
	/**
	 * Métricas de la plataforma, todas en `displayCurrency` y sobre las mismas
	 * posiciones que `totalValue`, así que la resta entre coste y mercado es una
	 * ganancia y no el choque de dos alcances distintos.
	 *
	 * `percent` es la parte del total invertido de la cuenta que vive en esta
	 * plataforma — lo que hace legible el orden — y viene en 0 cuando la
	 * plataforma se lee sola, porque una participación necesita el conjunto.
	 *
	 * Opcionales mientras convivan backends viejos: sin ellas la vista enseña lo
	 * invertido y calla el resto, en vez de inventar una ganancia de cero.
	 */
	marketValue: z.string().optional(),
	gainLoss: z.string().optional(),
	gainLossPct: z.number().optional(),
	percent: z.number().optional(),
	/**
	 * Moneda en la que vienen los importes. El total suma posiciones compradas
	 * en monedas distintas, así que sin esto la cifra no tiene unidad — y la
	 * vista le ponía un "$" fijo. Opcional mientras convivan backends viejos.
	 */
	displayCurrency: z.string().optional(),
	/**
	 * Posiciones que siguen contadas en `totalValue` a valor nominal porque no
	 * había tasa para su moneda: si es > 0 el total mezcla monedas.
	 */
	positionsUnconverted: z.number().optional(),
	/**
	 * De dónde salió `marketValue`, y suman `investments`.
	 *
	 * Una posición sin precio de mercado se valora a su propio coste, que es
	 * justo contra lo que se la compara: aporta cero a la ganancia. Así que una
	 * `gainLoss` de cero es lo que informa una plataforma plana y también una
	 * cuyas posiciones no tienen precio, y solo `positionsAtCost` distingue «no
	 * se movió» de «no está valorada».
	 */
	positionsPricedOwn: z.number().optional(),
	positionsPricedManual: z.number().optional(),
	positionsAtCost: z.number().optional(),
	createdAt: z.string()
});
