// Backend stub for the e2e suite. Serves the subset of the HTTP contract
// (docs/API.md) that the SvelteKit server consumes, with fixed fixtures, so
// the smoke tests exercise the frontend's loaders/actions/session handling
// without a real Go backend or database. Playwright starts it via the
// `webServer` array in playwright.config.ts and points BASE_API at it.
import { createServer } from 'node:http';

const PORT = Number(process.env.MOCK_API_PORT ?? 4174);
const API_PREFIX = '/api/v1';

export const PASSWORD = 'Password123!';

const IDS = {
	portfolio: '11111111-1111-4111-8111-111111111111',
	assetAapl: '22222222-2222-4222-8222-222222222222',
	assetBtc: '22222222-2222-4222-8222-222222222223',
	platform: '33333333-3333-4333-8333-333333333333',
	risk: '44444444-4444-4444-8444-444444444444',
	entry: '55555555-5555-4555-8555-555555555555'
};

const NOW = '2026-07-01T00:00:00Z';
const FUTURE = '2027-01-01T00:00:00Z';

const ACCOUNTS = {
	'user@finexia.test': {
		accessToken: 'access-user',
		refreshToken: 'refresh-user',
		user: {
			name: 'Usuaria Prueba',
			email: 'user@finexia.test',
			emailVerified: true,
			image: '',
			role: 'customer',
			preferredCurrency: 'USD',
			createdAt: NOW,
			updatedAt: NOW
		},
		session: {
			id: 'session-user',
			userId: 'user-1',
			expiresAt: FUTURE,
			ipAddress: null,
			userAgent: null,
			createdAt: NOW
		}
	},
	'admin@finexia.test': {
		accessToken: 'access-admin',
		refreshToken: 'refresh-admin',
		user: {
			name: 'Admin Prueba',
			email: 'admin@finexia.test',
			emailVerified: true,
			image: '',
			role: 'admin',
			preferredCurrency: 'USD',
			createdAt: NOW,
			updatedAt: NOW
		},
		session: {
			id: 'session-admin',
			userId: 'admin-1',
			expiresAt: FUTURE,
			ipAddress: null,
			userAgent: null,
			createdAt: NOW
		}
	}
};

const holdings = [
	{
		id: IDS.entry,
		assetId: IDS.assetAapl,
		ticker: 'AAPL',
		name: 'Apple Inc.',
		assetType: 'stock',
		exchange: 'NASDAQ',
		currency: 'USD',
		quantity: '10',
		price: '150.00',
		marketPrice: '190.00',
		costCurrency: 'USD',
		category: 'stock',
		entryDate: '2026-05-01',
		notes: ''
	}
];

const portfolioSummary = (displayCurrency = 'USD') => [
	{
		id: IDS.portfolio,
		name: 'Cartera Principal',
		description: 'Portafolio de prueba e2e',
		type: 'personal',
		baseCurrency: 'USD',
		displayCurrency,
		isDefault: true,
		riskId: IDS.risk,
		riskName: 'Moderado',
		totalPositions: 1,
		totalCostBase: '1500.00',
		totalMarketValue: '1900.00',
		totalGainLoss: '400.00',
		totalGainLossPct: '26.67',
		createdAt: NOW
	}
];

const transactions = [
	{
		id: 'txn-1',
		entryId: IDS.entry,
		type: 'buy',
		quantity: '10',
		price: '150.00',
		currency: 'USD',
		fees: '1.00',
		transactionDate: '2026-05-01T00:00:00Z',
		notes: '',
		createdAt: '2026-05-01T00:00:00Z',
		assetTicker: 'AAPL',
		assetName: 'Apple Inc.'
	},
	{
		id: 'txn-2',
		entryId: IDS.entry,
		type: 'sell',
		quantity: '0.01',
		price: '65000.00',
		currency: 'USD',
		fees: '2.00',
		transactionDate: '2026-06-01T00:00:00Z',
		notes: '',
		createdAt: '2026-06-01T00:00:00Z',
		assetTicker: 'BTC',
		assetName: 'Bitcoin'
	}
];

const growth = {
	points: [
		{
			date: '2026-05-01',
			totalValue: '1500.00',
			totalCostBase: '1500.00',
			gainLoss: '0',
			gainLossPct: '0'
		},
		{
			date: '2026-06-01',
			totalValue: '1900.00',
			totalCostBase: '1500.00',
			gainLoss: '400.00',
			gainLossPct: '26.67'
		}
	],
	summary: {
		firstDate: '2026-05-01',
		initialValue: '1500.00',
		currentValue: '1900.00',
		totalGrowthPct: '26.67'
	}
};

const assets = [
	{
		id: IDS.assetAapl,
		ticker: 'AAPL',
		name: 'Apple Inc.',
		assetType: 'stock',
		exchange: 'NASDAQ',
		currency: 'USD',
		currentPrice: { value: '190.00', currency: 'USD' }
	},
	{
		id: IDS.assetBtc,
		ticker: 'BTC',
		name: 'Bitcoin',
		assetType: 'crypto',
		exchange: '',
		currency: 'USD',
		currentPrice: { value: '65000.00', currency: 'USD' }
	}
];

const importPreview = {
	sheets: ['Hoja1'],
	sheet: 'Hoja1',
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
	totalRows: 2,
	validRows: 2,
	invalidRows: 0,
	rows: [
		{
			rowNumber: 2,
			raw: ['2026-05-01', 'buy', 'AAPL', '10', '150'],
			date: '2026-05-01',
			type: 'buy',
			ticker: 'AAPL',
			assetName: 'Apple Inc.',
			quantity: '10',
			price: '150',
			fees: '',
			currency: 'USD',
			category: 'stock',
			notes: '',
			valid: true,
			errors: []
		},
		{
			rowNumber: 3,
			raw: ['2026-06-01', 'sell', 'BTC', '0.01', '65000'],
			date: '2026-06-01',
			type: 'sell',
			ticker: 'BTC',
			assetName: 'Bitcoin',
			quantity: '0.01',
			price: '65000',
			fees: '',
			currency: 'USD',
			category: 'crypto',
			notes: '',
			valid: true,
			errors: []
		}
	]
};

function envelope(data, message = 'ok') {
	return { success: true, message, details: '', data, timestamp: NOW };
}

function errorEnvelope(message) {
	return { success: false, message, details: '', timestamp: NOW };
}

function send(res, status, body, headers = {}) {
	res.writeHead(status, { 'content-type': 'application/json', ...headers });
	res.end(JSON.stringify(body));
}

function readBody(req) {
	return new Promise((resolve) => {
		const chunks = [];
		req.on('data', (c) => chunks.push(c));
		req.on('end', () => resolve(Buffer.concat(chunks)));
	});
}

function accountByToken(req) {
	const auth = req.headers.authorization ?? '';
	const token = auth.replace(/^Bearer\s+/i, '');
	return Object.values(ACCOUNTS).find((a) => a.accessToken === token) ?? null;
}

function accountByRefreshCookie(req) {
	const cookie = req.headers.cookie ?? '';
	const match = cookie.match(/refresh_token=([^;\s]+)/);
	if (!match) return null;
	return Object.values(ACCOUNTS).find((a) => a.refreshToken === match[1]) ?? null;
}

function refreshSetCookie(account) {
	return `refresh_token=${account.refreshToken}; Path=/; HttpOnly; SameSite=Strict; Max-Age=2592000`;
}

const server = createServer(async (req, res) => {
	const url = new URL(req.url, `http://127.0.0.1:${PORT}`);
	if (!url.pathname.startsWith(API_PREFIX)) {
		return send(res, 404, errorEnvelope('not found'));
	}
	const path = url.pathname.slice(API_PREFIX.length) || '/';
	const route = `${req.method} ${path}`;

	// ---- Public auth routes ----
	if (route === 'POST /auth/login') {
		const body = JSON.parse((await readBody(req)).toString() || '{}');
		const account = ACCOUNTS[body.email];
		if (!account || body.password !== PASSWORD) {
			return send(res, 401, errorEnvelope('Credenciales incorrectas'));
		}
		return send(res, 200, envelope({ accessToken: account.accessToken }), {
			'set-cookie': refreshSetCookie(account)
		});
	}

	if (route === 'POST /auth/refresh') {
		const account = accountByRefreshCookie(req);
		if (!account) return send(res, 401, errorEnvelope('invalid refresh token'));
		return send(res, 200, envelope({ accessToken: account.accessToken }), {
			'set-cookie': refreshSetCookie(account)
		});
	}

	// ---- Everything below requires a valid access token ----
	const account = accountByToken(req);
	if (!account) {
		await readBody(req);
		return send(res, 401, errorEnvelope('invalid or missing token'));
	}

	if (route === 'GET /auth/session') {
		return send(res, 200, envelope({ user: account.user, session: account.session }));
	}
	if (route === 'POST /auth/logout') {
		return send(res, 200, envelope(null, 'logged out'));
	}
	if (route === 'GET /auth/sessions') {
		return send(
			res,
			200,
			envelope([
				{
					id: account.session.id,
					ipAddress: '127.0.0.1',
					userAgent: 'Playwright e2e',
					location: null,
					createdAt: NOW,
					lastActiveAt: NOW,
					expiresAt: FUTURE,
					current: true
				}
			])
		);
	}
	if (route === 'GET /auth/2fa') {
		return send(res, 200, envelope({ enabled: false, pendingSetup: false, recoveryCodesLeft: 0 }));
	}

	// ---- Users ----
	if (route === 'PATCH /users/me') {
		const body = JSON.parse((await readBody(req)).toString() || '{}');
		return send(res, 200, envelope({ ...account.user, ...body }));
	}
	if (path === '/users' && req.method === 'GET') {
		if (account.user.role !== 'admin') return send(res, 403, errorEnvelope('forbidden'));
		return send(
			res,
			200,
			envelope({
				items: Object.values(ACCOUNTS).map((a, i) => ({
					id: a.session.userId,
					name: a.user.name,
					email: a.user.email,
					emailVerified: a.user.emailVerified,
					createdAt: NOW,
					bannedAt: null,
					role: { name: a.user.role },
					index: i
				})),
				metaData: {
					currentPage: 1,
					usersForPage: 20,
					offset: 0,
					totalUsers: 2,
					totalPages: 1,
					previous: false,
					next: false
				}
			})
		);
	}
	if (route === 'GET /users/invitations') {
		if (account.user.role !== 'admin') return send(res, 403, errorEnvelope('forbidden'));
		return send(
			res,
			200,
			envelope({
				items: [
					{
						id: 'invite-1',
						email: 'invitada@finexia.test',
						name: 'Invitada',
						role: 'customer',
						status: 'pending',
						expiresAt: FUTURE,
						createdAt: NOW
					}
				]
			})
		);
	}
	if (route === 'GET /users/waitlist') {
		if (account.user.role !== 'admin') return send(res, 403, errorEnvelope('forbidden'));
		return send(
			res,
			200,
			envelope({
				items: [
					{
						id: 'wait-1',
						email: 'espera@finexia.test',
						status: 'pending',
						invitedAt: null,
						createdAt: NOW
					}
				]
			})
		);
	}

	// ---- Portfolios ----
	if (route === 'GET /portfolios/risks') {
		return send(
			res,
			200,
			envelope([{ id: IDS.risk, name: 'Moderado', description: 'Riesgo moderado' }])
		);
	}
	if (route === 'GET /portfolios/summary') {
		return send(res, 200, envelope(portfolioSummary(url.searchParams.get('currency') ?? 'USD')));
	}
	if (route === 'GET /portfolios/transactions') {
		return send(res, 200, envelope(transactions));
	}
	if (route === 'GET /portfolios/allocation') {
		return send(res, 200, envelope([{ category: 'stock', marketValue: '1900.00', percent: 100 }]));
	}
	if (route === 'GET /portfolios/growth' || route === `GET /portfolios/${IDS.portfolio}/growth`) {
		return send(res, 200, envelope(growth));
	}
	if (route === 'GET /portfolios/sources') {
		return send(res, 200, envelope([{ id: IDS.platform, name: 'Broker Demo', isActive: true }]));
	}
	if (route === 'GET /portfolios/assets') {
		const search = (url.searchParams.get('search') ?? '').toLowerCase();
		const filtered = search
			? assets.filter(
					(a) => a.ticker.toLowerCase().includes(search) || a.name.toLowerCase().includes(search)
				)
			: assets;
		return send(res, 200, envelope(filtered));
	}
	if (route === 'POST /portfolios/entries') {
		await readBody(req);
		return send(res, 201, envelope({ id: IDS.entry }, 'entry created'));
	}
	if (route === 'POST /portfolios/transactions/import/preview') {
		await readBody(req);
		return send(res, 200, envelope(importPreview));
	}
	if (route === `GET /portfolios/${IDS.portfolio}/top-transaction`) {
		return send(
			res,
			200,
			envelope({
				value: '1500.00',
				type: 'buy',
				currency: 'USD',
				assetTicker: 'AAPL',
				assetName: 'Apple Inc.',
				transactionDate: '2026-05-01T00:00:00Z'
			})
		);
	}
	if (route === `GET /portfolios/${IDS.portfolio}`) {
		return send(
			res,
			200,
			envelope({
				id: IDS.portfolio,
				userId: account.session.userId,
				name: 'Cartera Principal',
				description: 'Portafolio de prueba e2e',
				type: 'personal',
				baseCurrency: 'USD',
				isDefault: true,
				riskId: IDS.risk,
				riskName: 'Moderado',
				createdAt: NOW,
				updatedAt: NOW,
				holdings
			})
		);
	}
	if (req.method === 'GET' && /^\/portfolios\/[0-9a-f-]{36}$/.test(path)) {
		return send(res, 404, errorEnvelope('portfolio not found'));
	}

	await readBody(req);
	return send(res, 404, errorEnvelope(`no mock for ${route}`));
});

server.listen(PORT, () => {
	console.log(`mock backend listening on http://127.0.0.1:${PORT}${API_PREFIX}`);
});
