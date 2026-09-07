import type { PageServerLoad } from './$types';
import * as user from '$lib/api/user';
import * as market from '$lib/api/market';
import { summarizeDesk } from '$lib/features/admin';
import type { Asset, ExchangeRate, InvitationItem, WaitlistItem } from '$lib/api/types';

export const load: PageServerLoad = async ({ cookies, fetch }) => {
	const event = { cookies, fetch };

	/*
	 * La portada abre diciendo qué hay pendiente, así que necesita las cuatro
	 * listas y no solo sus totales: quién espera invitación, qué invitación se
	 * cae, qué precios se quedaron viejos y qué tasas manuales llevan un mes.
	 * Van en paralelo porque ninguna depende de otra.
	 */
	const [usersRes, assetsRes, ratesRes, invitationsRes, waitlistRes] = await Promise.all([
		user.getUsers(event, { page: 1, limit: 1 }),
		market.getAssets(event, { page: 1, limit: 100 }),
		market.getExchangeRates(event, { page: 1, limit: 100 }),
		user.getInvitations(event, { page: 1, limit: 50 }),
		user.getWaitlist(event, { page: 1, limit: 50 })
	]);

	let totalUsers = 0;
	if (usersRes.ok && usersRes.success && usersRes.data?.metaData) {
		totalUsers = Number(usersRes.data.metaData.totalUsers ?? 0);
	}

	let assets: Asset[] = [];
	if (assetsRes.ok && assetsRes.success && Array.isArray(assetsRes.data)) assets = assetsRes.data;

	let rates: ExchangeRate[] = [];
	if (ratesRes.ok && ratesRes.success && Array.isArray(ratesRes.data)) rates = ratesRes.data;

	const invitations: InvitationItem[] = invitationsRes.success
		? (invitationsRes.data?.items ?? [])
		: [];
	const waitlist: WaitlistItem[] = waitlistRes.success ? (waitlistRes.data?.items ?? []) : [];

	return {
		totalUsers,
		totalAssets: assets.length,
		totalRates: rates.length,
		desk: summarizeDesk({ assets, rates, invitations, waitlist })
	};
};
