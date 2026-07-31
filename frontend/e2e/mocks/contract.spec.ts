/**
 * El stub de e2e cumple el contrato que tipa la app.
 *
 * `mock-api.mjs` se escribió contra `docs/API.md` y toda la suite E2E confía en
 * él: si sus fixtures se salen del contrato, los flujos pasan en verde sobre
 * formas que el backend real nunca enviaría. Aquí se validan con los mismos
 * schemas de los que salen los tipos de `$lib/api/types`, de modo que una
 * divergencia falla en los unit tests, sin necesidad de levantar nada.
 *
 * Cubre las fixtures que el stub exporta; las respuestas que construye en
 * línea (sesiones, 2FA, listados de admin) las ejercita la suite E2E.
 */
import { describe, it, expect } from 'vitest';
import {
	assetSchema,
	exchangeRateSchema,
	holdingSchema,
	platformSchema,
	portfolioGrowthSchema,
	portfolioSummarySchema,
	userTransactionSchema
} from '$lib/api/schemas';
import {
	assets,
	exchangeRates,
	growth,
	holdings,
	portfolioSummary,
	sources,
	transactions
} from './mock-api.mjs';

describe('las fixtures del stub cumplen el contrato', () => {
	it.each([
		['holdings', holdingSchema.array(), holdings],
		['portfolios/summary', portfolioSummarySchema.array(), portfolioSummary('USD')],
		['portfolios/transactions', userTransactionSchema.array(), transactions],
		['portfolios/growth', portfolioGrowthSchema, growth],
		['portfolios/assets', assetSchema.array(), assets],
		['portfolios/sources', platformSchema.array(), sources],
		['exchange-rates', exchangeRateSchema.array(), exchangeRates]
	])('%s', (_name, schema, fixture) => {
		const parsed = schema.safeParse(fixture);
		expect(parsed.success ? [] : parsed.error.issues).toEqual([]);
	});
});
