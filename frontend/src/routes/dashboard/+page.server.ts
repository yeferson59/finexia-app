import type { Actions, PageServerLoad } from './$types';
import { redirect } from '@sveltejs/kit';
import * as auth from '$lib/api/auth';
import * as portfolio from '$lib/api/portfolio';
import * as transactions from '$lib/api/transactions';
import * as platforms from '$lib/api/platforms';
import * as market from '$lib/api/market';
import { ACCESS_COOKIE, REFRESH_COOKIE, clearSessionCookies } from '$lib/server/session';
import { resolveDisplayCurrency } from '$lib/shared/currency';
import type {
	AllocationItem,
	Platform,
	PortfolioGrowth,
	PortfolioSummary,
	UserTransaction
} from '$lib/api/types';

// La moneda desde la que se convierte: las tasas compartidas se guardan en una
// sola dirección (USD → X) y es la que publica el feed público.
const BASE_CURRENCY = 'USD';

export const load: PageServerLoad = async ({ cookies, fetch, url, locals }) => {
	const event = { cookies, fetch };

	// Sin `?currency=` manda la moneda de la cuenta: es lo que el usuario eligió
	// en ajustes y el único sitio donde esa preferencia se aplica sola. El
	// parámetro sigue existiendo para mirar el panel en otra moneda sin cambiar
	// la preferencia.
	const currency = resolveDisplayCurrency(
		url.searchParams.get('currency'),
		locals.user?.preferredCurrency
	);

	const [
		transactionsRes,
		summaryRes,
		allocationRes,
		growthRes,
		platformsRes,
		credentialsRes,
		ratesRes
	] = await Promise.all([
		transactions.getRecent(event),
		portfolio.getSummaries(event, currency),
		portfolio.getAllocation(event, currency),
		portfolio.getAggregateGrowth(event, { currency }),
		// El panel enseña dónde está custodiado el dinero, no solo cómo lo agrupó
		// el usuario: son dos lecturas del mismo total y la de plataformas es la
		// que da sentido a que la aplicación exista.
		platforms.getSources(event, currency),
		market.getMarketCredentials(event),
		market.getLatestExchangeRates(event)
	]);

	const recentTransactions: UserTransaction[] =
		transactionsRes.ok && transactionsRes.success && Array.isArray(transactionsRes.data)
			? transactionsRes.data.slice(0, 5)
			: [];

	const portfolioSummaries: PortfolioSummary[] =
		summaryRes.ok && summaryRes.success && Array.isArray(summaryRes.data) ? summaryRes.data : [];

	const allocation: AllocationItem[] =
		allocationRes.ok && allocationRes.success && Array.isArray(allocationRes.data)
			? allocationRes.data
			: [];

	const userPlatforms: Platform[] =
		platformsRes.ok && platformsRes.success && Array.isArray(platformsRes.data)
			? platformsRes.data
			: [];

	let portfolioGrowth: PortfolioGrowth = {
		points: [],
		summary: { firstDate: '', initialValue: '0', currentValue: '0', totalGrowthPct: '0' }
	};
	if (growthRes.ok && growthRes.success && growthRes.data) portfolioGrowth = growthRes.data;

	// Las cifras de mercado dependen de la clave del propio usuario, así que el
	// dashboard necesita saber si hay una usable para no presentar un valor a
	// coste como si fuera de mercado. Un fallo al leerlas se trata como "sí hay
	// clave": callar es mejor que avisar en falso.
	const credentials = credentialsRes.ok ? (credentialsRes.data ?? []) : null;
	const hasUsableKey = credentials === null || credentials.some((c) => c.status !== 'invalid');
	const hasBrokenKey = credentials !== null && credentials.length > 0 && !hasUsableKey;

	// La tasa con la que se convierten las cifras de arriba, para poder
	// enseñarla junto a ellas. Solo interesa el par entre la moneda base de la
	// aplicación y la que se está mostrando: viéndola en USD no hay conversión
	// que enseñar. Si no hay tasa —o el backend no contesta— se omite la línea
	// en vez de inventar un valor.
	const rates = ratesRes.ok && ratesRes.success ? (ratesRes.data ?? []) : [];
	const displayRate =
		currency === BASE_CURRENCY
			? null
			: (rates.find((r) => r.fromCurrency === BASE_CURRENCY && r.toCurrency === currency) ?? null);

	return {
		recentTransactions,
		portfolioSummaries,
		platforms: userPlatforms,
		allocation,
		portfolioGrowth,
		currency,
		displayRate,
		hasUsableKey,
		hasBrokenKey
	};
};

export const actions = {
	logout: async ({ cookies, fetch }) => {
		const token = cookies.get(ACCESS_COOKIE);

		if (!token) return { success: false };

		const refreshToken = cookies.get(REFRESH_COOKIE);

		await auth.logout(fetch, {
			accessToken: token,
			refreshToken,
			refreshCookieName: REFRESH_COOKIE
		});

		clearSessionCookies(cookies);

		return redirect(302, '/auth');
	}
} satisfies Actions;
