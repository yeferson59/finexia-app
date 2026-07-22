import type { Actions, PageServerLoad } from './$types';
import * as user from '$lib/api/user';
import { fail } from '@sveltejs/kit';
import type { PageMeta } from '$lib/api/types';

const DEFAULT_META: PageMeta = { currentPage: 1, totalPages: 1, previous: false, next: false };

export const load: PageServerLoad = async ({ cookies, fetch, url }) => {
	const event = { cookies, fetch };
	const page = Number(url.searchParams.get('page') ?? '1');

	const [usersRes, invitationsRes, waitlistRes] = await Promise.all([
		user.getUsers(event, { page, limit: 20 }),
		user.getInvitations(event, { page: 1, limit: 50 }),
		user.getWaitlist(event, { page: 1, limit: 50 })
	]);

	const invitations = invitationsRes.success ? (invitationsRes.data?.items ?? []) : [];
	const waitlist = waitlistRes.success ? (waitlistRes.data?.items ?? []) : [];

	return {
		users: usersRes.success ? (usersRes.data?.items ?? []) : [],
		meta: usersRes.success ? (usersRes.data?.metaData ?? DEFAULT_META) : DEFAULT_META,
		invitations,
		// Only pending entries are actionable from here; invited/registered ones
		// already moved down the funnel.
		waitlist: waitlist.filter((w) => w.status === 'pending')
	};
};

const ROLES = new Set(['customer', 'admin']);

export const actions = {
	inviteUser: async ({ request, cookies, fetch }) => {
		const fd = await request.formData();
		const email = (fd.get('email') as string | null)?.trim();
		const name = (fd.get('name') as string | null)?.trim() ?? '';
		const role = ((fd.get('role') as string | null)?.trim() || 'customer').toLowerCase();

		if (!email) return fail(400, { error: 'El correo es requerido' });
		if (!ROLES.has(role)) return fail(400, { error: 'Rol inválido' });

		const res = await user.inviteUser({ cookies, fetch }, { email, name, role });

		if (!res.ok) {
			return fail(res.status, {
				error: res.details ?? res.message ?? 'No se pudo enviar la invitación'
			});
		}

		return { success: true, invited: email };
	},

	resendInvitation: async ({ request, cookies, fetch }) => {
		const fd = await request.formData();
		const id = fd.get('id') as string;
		if (!id) return fail(400, { inviteError: 'ID requerido', inviteId: '' });

		const res = await user.resendInvitation({ cookies, fetch }, id);
		if (!res.ok) {
			return fail(res.status, {
				inviteError: res.details ?? 'No se pudo reenviar la invitación',
				inviteId: id
			});
		}
		return { inviteSuccess: true, inviteId: id, inviteAction: 'resent' as const };
	},

	revokeInvitation: async ({ request, cookies, fetch }) => {
		const fd = await request.formData();
		const id = fd.get('id') as string;
		if (!id) return fail(400, { inviteError: 'ID requerido', inviteId: '' });

		const res = await user.revokeInvitation({ cookies, fetch }, id);
		if (!res.ok) {
			return fail(res.status, {
				inviteError: res.details ?? 'No se pudo revocar la invitación',
				inviteId: id
			});
		}
		return { inviteSuccess: true, inviteId: id, inviteAction: 'revoked' as const };
	},

	deleteUser: async ({ request, cookies, fetch }) => {
		const fd = await request.formData();
		const id = fd.get('id') as string;
		if (!id) return fail(400, { error: 'ID de usuario requerido' });

		const res = await user.deleteUser({ cookies, fetch }, id);
		if (!res.ok) return fail(res.status, { error: 'No se pudo eliminar el usuario' });
		return { success: true };
	},

	banUser: async ({ request, cookies, fetch }) => {
		const fd = await request.formData();
		const id = fd.get('id') as string;
		const ban = fd.get('ban') === 'true';
		if (!id) return fail(400, { banError: 'ID requerido', banId: '' });

		const res = await user.banUser({ cookies, fetch }, id, { ban });

		if (!res.ok) {
			return fail(res.status, {
				banError: res.details ?? 'No se pudo actualizar el estado',
				banId: id
			});
		}

		return { banSuccess: true, banId: id, banned: ban };
	}
} satisfies Actions;
