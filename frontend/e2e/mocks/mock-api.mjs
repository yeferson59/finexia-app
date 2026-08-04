// Backend stub for the e2e suite. Serves the subset of the HTTP contract
// (docs/API.md) that the SvelteKit server consumes, with fixed fixtures, so
// the smoke tests exercise the frontend's loaders/actions/session handling
// without a real Go backend or database. Playwright starts it via the
// `webServer` array in playwright.config.ts and points BASE_API at it.
import { createServer } from 'node:http';
import { pathToFileURL } from 'node:url';

const PORT = Number(process.env.MOCK_API_PORT ?? 4174);
const API_PREFIX = '/api/v1';

export const PASSWORD = 'Password123!';

// Las fixtures viven aparte: describen una cuenta completa y las usan tanto
// esta suite como el generador de capturas del manual (`pnpm manual:shots`).
import {
	FUTURE,
	IDS,
	NOW,
	allocation,
	assets,
	exchangeRates,
	growth,
	growthFor,
	holdings,
	importPreview,
	PORTFOLIOS,
	portfolioSummary,
	risks,
	sources,
	topTransaction,
	transactions
} from './fixtures.mjs';

// Se reexportan para `contract.spec.ts`, que valida las fixtures contra los
// schemas Zod de los que salen los tipos de la aplicación.
export {
	assets,
	exchangeRates,
	growth,
	holdings,
	portfolioSummary,
	sources,
	transactions,
	allocation
};

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
	if (route === 'GET /users/me/preferences') {
		return send(
			res,
			200,
			envelope({ userId: account.session.userId, emailAlerts: true, weeklySummary: false })
		);
	}
	if (route === 'PATCH /users/me/preferences') {
		const body = JSON.parse((await readBody(req)).toString() || '{}');
		return send(res, 200, envelope({ userId: account.session.userId, ...body }));
	}
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
		return send(res, 200, envelope(risks));
	}
	if (route === 'GET /portfolios/summary') {
		return send(res, 200, envelope(portfolioSummary(url.searchParams.get('currency') ?? 'USD')));
	}
	if (route === 'GET /portfolios/transactions') {
		return send(res, 200, envelope(transactions));
	}
	if (route === 'GET /portfolios/allocation') {
		return send(res, 200, envelope(allocation));
	}
	if (route === 'GET /portfolios/growth') {
		return send(res, 200, envelope(growth));
	}
	if (req.method === 'GET' && /^\/portfolios\/[0-9a-f-]{36}\/growth$/.test(path)) {
		return send(res, 200, envelope(growthFor(path.split('/')[2])));
	}
	if (route === 'GET /portfolios/sources') {
		return send(res, 200, envelope(sources));
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
	if (route === 'GET /exchange-rates') {
		return send(res, 200, envelope(exchangeRates));
	}
	if (route === 'POST /portfolios/entries') {
		await readBody(req);
		return send(res, 201, envelope({ id: IDS.entry }, 'entry created'));
	}
	if (route === 'POST /portfolios/transactions/import/preview') {
		await readBody(req);
		return send(res, 200, envelope(importPreview));
	}
	if (
		req.method === 'GET' &&
		/^\/portfolios\/[0-9a-f-]{36}\/assets\/[^/]+\/transactions$/.test(path)
	) {
		const symbol = decodeURIComponent(path.split('/')[4]);
		const rows = transactions.filter((t) => t.assetTicker === symbol);
		return send(
			res,
			200,
			envelope({ data: rows, total: rows.length, page: 1, limit: 20, totalPages: 1 })
		);
	}
	if (req.method === 'GET' && /^\/portfolios\/[0-9a-f-]{36}\/top-transaction$/.test(path)) {
		const top = topTransaction(path.split('/')[2]);
		if (!top) return send(res, 404, errorEnvelope('no transactions'));
		return send(res, 200, envelope(top));
	}
	if (req.method === 'GET' && /^\/portfolios\/[0-9a-f-]{36}$/.test(path)) {
		const portfolio = PORTFOLIOS.find((p) => p.id === path.split('/')[2]);
		if (!portfolio) return send(res, 404, errorEnvelope('portfolio not found'));
		return send(
			res,
			200,
			envelope({
				id: portfolio.id,
				userId: account.session.userId,
				name: portfolio.name,
				description: portfolio.description,
				type: portfolio.type,
				baseCurrency: 'USD',
				isDefault: portfolio.isDefault,
				riskId: portfolio.riskId,
				riskName: portfolio.riskName,
				createdAt: NOW,
				updatedAt: NOW,
				holdings: portfolio.holdings
			})
		);
	}

	await readBody(req);
	return send(res, 404, errorEnvelope(`no mock for ${route}`));
});

// Solo escucha cuando se ejecuta como programa (así lo arranca Playwright);
// importarlo desde un test trae las fixtures sin abrir un puerto.
if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
	server.listen(PORT, () => {
		console.log(`mock backend listening on http://127.0.0.1:${PORT}${API_PREFIX}`);
	});
}
