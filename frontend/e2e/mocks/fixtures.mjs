/**
 * Fixtures del stub de la API.
 *
 * Vivían dentro de `mock-api.mjs`, mezcladas con el enrutado. Se separan porque
 * han dejado de ser cuatro constantes: describen una cuenta completa —tres
 * portafolios, veinte posiciones, catorce meses de historial— y de ellas salen
 * tanto los e2e como las capturas del manual de usuario
 * (`pnpm manual:shots`). Con un solo portafolio y dos movimientos, el manual
 * enseñaba gráficas vacías y una proyección que la aplicación se negaba a
 * calcular por falta de historial.
 *
 * Todo lo que se puede derivar, se deriva: los totales de cada portafolio salen
 * de sus posiciones y la serie de crecimiento termina exactamente en el valor
 * que suman. Así las cifras cuadran entre pantallas y nadie tiene que
 * recalcular a mano una fixture cuando toca otra.
 */

export const NOW = '2026-07-01T00:00:00Z';
export const FUTURE = '2027-01-01T00:00:00Z';

export const IDS = {
	portfolio: '11111111-1111-4111-8111-111111111111',
	portfolioCrypto: '11111111-1111-4111-8111-111111111112',
	portfolioReserve: '11111111-1111-4111-8111-111111111113',
	platform: '33333333-3333-4333-8333-333333333333',
	platformExchange: '33333333-3333-4333-8333-333333333334',
	platformBank: '33333333-3333-4333-8333-333333333335',
	riskModerate: '44444444-4444-4444-8444-444444444444',
	riskAggressive: '44444444-4444-4444-8444-444444444445',
	riskConservative: '44444444-4444-4444-8444-444444444446',
	entry: '55555555-5555-4555-8555-555555555555'
};

const money = (value) => value.toFixed(2);

// --- Catálogo de activos ----------------------------------------------------

/** `[ticker, nombre, tipo, mercado, categoría, precio de mercado]`. */
const CATALOG = [
	['AAPL', 'Apple Inc.', 'stock', 'NASDAQ', 'stocks', 214.35],
	['MSFT', 'Microsoft Corp.', 'stock', 'NASDAQ', 'stocks', 438.9],
	['NVDA', 'NVIDIA Corp.', 'stock', 'NASDAQ', 'stocks', 126.4],
	['VWCE', 'Vanguard FTSE All-World UCITS ETF', 'etf', 'XETRA', 'etfs', 128.62],
	['CSPX', 'iShares Core S&P 500 UCITS ETF', 'etf', 'XETRA', 'etfs', 568.2],
	['BTC', 'Bitcoin', 'crypto', '', 'cryptos', 67240.0],
	['ETH', 'Ethereum', 'crypto', '', 'cryptos', 3482.5],
	['SOL', 'Solana', 'crypto', '', 'cryptos', 168.9],
	['TLT', 'iShares 20+ Year Treasury Bond ETF', 'bond', 'NASDAQ', 'bonds', 92.15],
	['USD', 'Efectivo en dólares', 'cash', '', 'cash', 1.0]
];

const asset = (ticker) => CATALOG.find((row) => row[0] === ticker);

/** `GET /portfolios/assets` — el catálogo que ve el buscador de activos. */
export const assets = CATALOG.map(([ticker, name, assetType, exchange, , price], index) => ({
	id: `22222222-2222-4222-8222-2222222222${String(index).padStart(2, '0')}`,
	ticker,
	name,
	assetType,
	exchange,
	currency: 'USD',
	currentPrice: { value: money(price), currency: 'USD' },
	// BTC lo aportó la usuaria y no tiene precio manual fechado del catálogo.
	priceUpdatedAt: ticker === 'BTC' ? null : NOW,
	isCurated: ticker !== 'BTC'
}));

const assetId = (ticker) => assets.find((a) => a.ticker === ticker).id;

// --- Posiciones -------------------------------------------------------------

let entrySeq = 0;

/** Posición dentro de un portafolio: cantidad y precio medio de compra. */
function holding(ticker, quantity, price, entryDate, notes = '') {
	const [, name, assetType, exchange, category, marketPrice] = asset(ticker);
	entrySeq += 1;
	return {
		// La primera posición conserva el id histórico: los e2e lo usan.
		id:
			entrySeq === 1
				? IDS.entry
				: `55555555-5555-4555-8555-5555555555${String(entrySeq).padStart(2, '0')}`,
		assetId: assetId(ticker),
		ticker,
		name,
		assetType,
		exchange,
		currency: 'USD',
		quantity: String(quantity),
		price: money(price),
		marketPrice: money(marketPrice),
		costCurrency: 'USD',
		category,
		entryDate,
		notes
	};
}

export const PORTFOLIOS = [
	{
		id: IDS.portfolio,
		name: 'Cartera Principal',
		description: 'Acciones y ETFs a largo plazo',
		type: 'stocks_etfs',
		riskId: IDS.riskModerate,
		riskName: 'Moderado',
		isDefault: true,
		holdings: [
			holding('AAPL', 42, 168.4, '2025-06-12', 'Núcleo de la cartera'),
			holding('MSFT', 18, 372.5, '2025-07-03'),
			holding('VWCE', 120, 108.75, '2025-06-12', 'Aporte mensual'),
			holding('CSPX', 9, 492.3, '2025-09-18'),
			holding('NVDA', 60, 98.2, '2026-01-15')
		]
	},
	{
		id: IDS.portfolioCrypto,
		name: 'Cripto',
		description: 'Posición especulativa, revisada cada trimestre',
		type: 'cryptos',
		riskId: IDS.riskAggressive,
		riskName: 'Agresivo',
		isDefault: false,
		holdings: [
			holding('BTC', 0.15, 54300.0, '2025-08-04'),
			holding('ETH', 2.2, 2980.0, '2025-08-04'),
			holding('SOL', 25, 142.6, '2026-02-20')
		]
	},
	{
		id: IDS.portfolioReserve,
		name: 'Reserva',
		description: 'Colchón de liquidez y renta fija',
		type: 'bonds',
		riskId: IDS.riskConservative,
		riskName: 'Conservador',
		isDefault: false,
		holdings: [holding('TLT', 140, 96.4, '2025-10-09'), holding('USD', 9500, 1.0, '2025-06-01')]
	}
];

/** Todas las posiciones, sin importar el portafolio. */
export const holdings = PORTFOLIOS.flatMap((p) => p.holdings);

const costOf = (h) => Number(h.quantity) * Number(h.price);
const valueOf = (h) => Number(h.quantity) * Number(h.marketPrice);

const sum = (values) => values.reduce((acc, v) => acc + v, 0);

export const TOTAL_COST = sum(holdings.map(costOf));
export const TOTAL_VALUE = sum(holdings.map(valueOf));

/** `GET /portfolios/summary` — totales derivados de las posiciones. */
export const portfolioSummary = (displayCurrency = 'USD') =>
	PORTFOLIOS.map((p) => {
		const cost = sum(p.holdings.map(costOf));
		const value = sum(p.holdings.map(valueOf));
		const gain = value - cost;
		return {
			id: p.id,
			name: p.name,
			description: p.description,
			type: p.type,
			baseCurrency: 'USD',
			displayCurrency,
			isDefault: p.isDefault,
			riskId: p.riskId,
			riskName: p.riskName,
			totalPositions: p.holdings.length,
			totalCostBase: money(cost),
			totalMarketValue: money(value),
			totalGainLoss: money(gain),
			totalGainLossPct: (cost > 0 ? (gain / cost) * 100 : 0).toFixed(2),
			createdAt: NOW
		};
	});

/** `GET /portfolios/allocation` — reparto por categoría, en el orden del donut. */
export const allocation = (() => {
	const byCategory = new Map();
	for (const h of holdings) {
		byCategory.set(h.category, (byCategory.get(h.category) ?? 0) + valueOf(h));
	}
	return [...byCategory.entries()]
		.sort(([, a], [, b]) => b - a)
		.map(([category, value]) => ({
			category,
			marketValue: money(value),
			percent: Number(((value / TOTAL_VALUE) * 100).toFixed(2))
		}));
})();

// --- Serie de crecimiento ---------------------------------------------------

/*
 * Catorce meses, cinco puntos por mes. Las curvas son fracciones del valor y
 * del coste finales, así que la serie termina exactamente en lo que suman las
 * posiciones. El último punto de cada mes cae justo en su ancla: el calendario
 * de rentabilidad de reportes lee ese punto, y así sus porcentajes son los de
 * las curvas y no un artefacto de la interpolación.
 */
const VALUE_CURVE = [
	0.78, 0.8, 0.83, 0.81, 0.85, 0.88, 0.86, 0.9, 0.93, 0.95, 0.93, 0.96, 0.98, 1.0
];
const COST_CURVE = [
	0.74, 0.76, 0.79, 0.79, 0.82, 0.85, 0.85, 0.88, 0.91, 0.93, 0.93, 0.96, 0.98, 1.0
];

/** Ondulación fija dentro del mes: la serie tiene que ser reproducible. */
const WOBBLE = [0.004, -0.003, 0.006, -0.005, 0.002, -0.004, 0.005, -0.002];

const FIRST_MONTH = { year: 2025, month: 5 }; // junio de 2025 (0-based)
const DAYS = [1, 8, 15, 22, 28];

function isoDate(year, month, day) {
	const date = new Date(Date.UTC(year, month, day));
	return date.toISOString().slice(0, 10);
}

export const growth = (() => {
	const points = [];
	let wobbleIndex = 0;

	for (let m = 0; m < VALUE_CURVE.length; m++) {
		const prevValue = m === 0 ? VALUE_CURVE[0] * 0.98 : VALUE_CURVE[m - 1];
		const prevCost = m === 0 ? COST_CURVE[0] * 0.99 : COST_CURVE[m - 1];

		for (const day of DAYS) {
			const t = day / 28;
			const last = day === 28;
			const noise = last ? 0 : WOBBLE[wobbleIndex++ % WOBBLE.length];

			const value = TOTAL_VALUE * (prevValue + (VALUE_CURVE[m] - prevValue) * t + noise);
			const cost = TOTAL_COST * (prevCost + (COST_CURVE[m] - prevCost) * t);
			const gain = value - cost;

			points.push({
				date: isoDate(FIRST_MONTH.year, FIRST_MONTH.month + m, day),
				totalValue: money(value),
				totalCostBase: money(cost),
				gainLoss: money(gain),
				gainLossPct: (cost > 0 ? (gain / cost) * 100 : 0).toFixed(2)
			});
		}
	}

	const first = points[0];
	const last = points[points.length - 1];
	const initial = Number(first.totalValue);
	const current = Number(last.totalValue);

	return {
		points,
		summary: {
			firstDate: first.date,
			initialValue: first.totalValue,
			currentValue: last.totalValue,
			totalGrowthPct: (((current - initial) / initial) * 100).toFixed(2)
		}
	};
})();

/** Serie de un portafolio: la agregada, a escala de lo que pesa dentro. */
export function growthFor(portfolioId) {
	const summary = portfolioSummary().find((p) => p.id === portfolioId);
	if (!summary) return growth;

	const share = Number(summary.totalMarketValue) / TOTAL_VALUE;
	const points = growth.points.map((point) => {
		const value = Number(point.totalValue) * share;
		const cost = Number(point.totalCostBase) * share;
		return {
			date: point.date,
			totalValue: money(value),
			totalCostBase: money(cost),
			gainLoss: money(value - cost),
			gainLossPct: (cost > 0 ? ((value - cost) / cost) * 100 : 0).toFixed(2)
		};
	});

	return {
		points,
		summary: {
			firstDate: points[0].date,
			initialValue: points[0].totalValue,
			currentValue: points[points.length - 1].totalValue,
			totalGrowthPct: growth.summary.totalGrowthPct
		}
	};
}

// --- Movimientos ------------------------------------------------------------

const entryIdOf = (ticker) => holdings.find((h) => h.ticker === ticker).id;

/** `[fecha, tipo, ticker, cantidad, precio, comisión, nota]`. */
const MOVEMENTS = [
	['2026-06-24', 'buy', 'NVDA', '20', '121.80', '1.20', 'Aporte mensual'],
	['2026-06-18', 'dividend', 'AAPL', '42', '0.26', '0.00', 'Dividendo trimestral'],
	['2026-06-11', 'buy', 'VWCE', '15', '126.40', '1.00', ''],
	['2026-06-02', 'interest', 'USD', '9500', '0.0021', '0.00', 'Intereses de la cuenta'],
	['2026-05-27', 'sell', 'ETH', '0.6', '3410.00', '4.10', 'Toma de beneficios'],
	['2026-05-19', 'buy', 'BTC', '0.03', '61200.00', '9.80', ''],
	['2026-05-08', 'dividend', 'TLT', '140', '0.31', '0.00', 'Cupón mensual'],
	['2026-04-30', 'fee', 'USD', '1', '12.50', '0.00', 'Custodia trimestral'],
	['2026-04-21', 'buy', 'CSPX', '2', '541.60', '1.50', ''],
	['2026-04-09', 'transfer_in', 'USD', '2500', '1.00', '0.00', 'Traspaso desde el banco'],
	['2026-03-17', 'buy', 'SOL', '10', '151.20', '2.40', ''],
	['2026-02-20', 'buy', 'AAPL', '12', '186.90', '1.20', '']
];

export const transactions = MOVEMENTS.map(
	([date, type, ticker, quantity, price, fees, notes], index) => ({
		id: `txn-${index + 1}`,
		entryId: entryIdOf(ticker),
		type,
		quantity,
		price,
		currency: 'USD',
		fees,
		transactionDate: `${date}T00:00:00Z`,
		notes,
		createdAt: `${date}T00:00:00Z`,
		assetTicker: ticker,
		assetName: asset(ticker)[1]
	})
);

/** Mayor movimiento de un portafolio, por importe. */
export function topTransaction(portfolioId) {
	const tickers = new Set(
		(PORTFOLIOS.find((p) => p.id === portfolioId)?.holdings ?? []).map((h) => h.ticker)
	);
	const candidates = transactions.filter((t) => tickers.has(t.assetTicker));
	if (candidates.length === 0) return null;

	const top = candidates.reduce((best, t) =>
		Number(t.quantity) * Number(t.price) > Number(best.quantity) * Number(best.price) ? t : best
	);

	return {
		value: money(Number(top.quantity) * Number(top.price)),
		type: top.type,
		currency: 'USD',
		assetTicker: top.assetTicker,
		assetName: top.assetName,
		transactionDate: top.transactionDate
	};
}

// --- Plataformas y tasas ----------------------------------------------------

/** Valor custodiado en cada plataforma, repartiendo las posiciones por tipo. */
export const sources = [
	{
		id: IDS.platform,
		name: 'Broker Demo',
		description: 'Acciones y ETFs',
		sourceType: 'broker',
		isActive: true,
		investments: 5,
		totalValue: money(sum(PORTFOLIOS[0].holdings.map(valueOf))),
		createdAt: NOW
	},
	{
		id: IDS.platformExchange,
		name: 'Exchange Demo',
		description: 'Criptomonedas',
		sourceType: 'exchange',
		isActive: true,
		investments: 3,
		totalValue: money(sum(PORTFOLIOS[1].holdings.map(valueOf))),
		createdAt: NOW
	},
	{
		id: IDS.platformBank,
		name: 'Banco Demo',
		description: 'Renta fija y efectivo',
		sourceType: 'bank',
		isActive: true,
		investments: 2,
		totalValue: money(sum(PORTFOLIOS[2].holdings.map(valueOf))),
		createdAt: NOW
	}
];

export const exchangeRates = [
	{
		id: '66666666-6666-4666-8666-666666666661',
		fromCurrency: 'USD',
		toCurrency: 'COP',
		rate: '4123.456789',
		rateDate: NOW,
		createdAt: NOW
	},
	{
		id: '66666666-6666-4666-8666-666666666662',
		fromCurrency: 'EUR',
		toCurrency: 'USD',
		rate: '1.085',
		rateDate: NOW,
		createdAt: NOW
	},
	{
		id: '66666666-6666-4666-8666-666666666663',
		fromCurrency: 'GBP',
		toCurrency: 'USD',
		rate: '1.272',
		rateDate: NOW,
		createdAt: NOW
	}
];

export const risks = [
	{ id: IDS.riskConservative, name: 'Conservador', description: 'Prioriza preservar el capital' },
	{
		id: IDS.riskModerate,
		name: 'Moderado',
		description: 'Equilibrio entre crecimiento y estabilidad'
	},
	{ id: IDS.riskAggressive, name: 'Agresivo', description: 'Busca máximo crecimiento' }
];

// --- Importación ------------------------------------------------------------

/*
 * Vista previa con filas descartadas a propósito: el manual explica que la
 * pantalla avisa de las que se omitirán y por qué, y con una previsualización
 * impecable esa parte de la interfaz no se veía nunca.
 */
const IMPORT_ROWS = [
	['2026-06-24', 'buy', 'NVDA', '20', '121.80', true, []],
	['2026-06-18', 'dividend', 'AAPL', '42', '0.26', true, []],
	['2026-06-11', 'buy', 'VWCE', '15', '126.40', true, []],
	['24/06/26', 'buy', 'MSFT', '3', '431.00', false, ['Fecha no reconocida']],
	['2026-05-27', 'sell', 'ETH', '1.5', '3410.00', true, []],
	['2026-05-19', 'buy', 'BTC', '0.06', '61200.00', true, []],
	['2026-05-08', 'buy', 'CSPX', '2', '', false, ['Precio vacío']],
	['2026-04-21', 'buy', 'TLT', '40', '95.10', true, []]
];

export const importPreview = {
	sheets: ['Movimientos', 'Resumen'],
	sheet: 'Movimientos',
	headerRow: 1,
	headers: ['Fecha', 'Tipo', 'Ticker', 'Cantidad', 'Precio'],
	suggestedMapping: {
		date: 0,
		type: 1,
		ticker: 2,
		assetName: null,
		quantity: 3,
		price: 4,
		fees: null,
		currency: null,
		category: null,
		notes: null
	},
	missingFields: [],
	totalRows: IMPORT_ROWS.length,
	validRows: IMPORT_ROWS.filter(([, , , , , valid]) => valid).length,
	invalidRows: IMPORT_ROWS.filter(([, , , , , valid]) => !valid).length,
	rows: IMPORT_ROWS.map(([date, type, ticker, quantity, price, valid, errors], index) => ({
		rowNumber: index + 2,
		raw: [date, type, ticker, quantity, price],
		date: valid ? date : '',
		type,
		ticker,
		assetName: asset(ticker)?.[1] ?? ticker,
		quantity,
		price,
		fees: '',
		currency: 'USD',
		category: asset(ticker)?.[4] ?? 'others',
		notes: '',
		valid,
		errors
	}))
};
