import type { Actions, PageServerLoad } from './$types';
import * as user from '$lib/api/user';
import { fail } from '@sveltejs/kit';
import type { PageMeta } from '$lib/api/types';
import { inviteUserSchema, rowIdSchema } from '$lib/features/admin';

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

export const actions = {
	inviteUser: async ({ request, cookies, fetch }) => {
		const fd = await request.formData();

		const parsed = inviteUserSchema.safeParse({
			email: fd.get('email') ?? '',
			name: fd.get('name') ?? '',
			role: fd.get('role') ?? ''
		});

		if (!parsed.success) return fail(400, { error: parsed.error.issues[0].message });

		const res = await user.inviteUser({ cookies, fetch }, parsed.data);

		if (!res.ok) {
			return fail(res.status, {
				error: res.details ?? res.message ?? 'No se pudo enviar la invitación'
			});
		}

		return { success: true, invited: parsed.data.email };
	},

	resendInvitation: async ({ request, cookies, fetch }) => {
		const fd = await request.formData();
		const parsed = rowIdSchema.safeParse(fd.get('id') ?? '');
		if (!parsed.success) return fail(400, { inviteError: 'ID requerido', inviteId: '' });

		const res = await user.resendInvitation({ cookies, fetch }, parsed.data);
		if (!res.ok) {
			return fail(res.status, {
				inviteError: res.details ?? 'No se pudo reenviar la invitación',
				inviteId: parsed.data
			});
		}
		return { inviteSuccess: true, inviteId: parsed.data, inviteAction: 'resent' as const };
	},

	revokeInvitation: async ({ request, cookies, fetch }) => {
		const fd = await request.formData();
		const parsed = rowIdSchema.safeParse(fd.get('id') ?? '');
		if (!parsed.success) return fail(400, { inviteError: 'ID requerido', inviteId: '' });

		const res = await user.revokeInvitation({ cookies, fetch }, parsed.data);
		if (!res.ok) {
			return fail(res.status, {
				inviteError: res.details ?? 'No se pudo revocar la invitación',
				inviteId: parsed.data
			});
		}
		return { inviteSuccess: true, inviteId: parsed.data, inviteAction: 'revoked' as const };
	},

	deleteWaitlist: async ({ request, cookies, fetch }) => {
		const fd = await request.formData();
		const parsed = rowIdSchema.safeParse(fd.get('id') ?? '');
		if (!parsed.success) return fail(400, { waitlistError: 'ID requerido', waitlistId: '' });

		const res = await user.deleteWaitlistEntry({ cookies, fetch }, parsed.data);
		if (!res.ok) {
			return fail(res.status, {
				waitlistError: res.details ?? 'No se pudo eliminar de la lista de espera',
				waitlistId: parsed.data
			});
		}
		return { waitlistSuccess: true, waitlistId: parsed.data };
	},

	deleteUser: async ({ request, cookies, fetch }) => {
		const fd = await request.formData();
		const parsed = rowIdSchema.safeParse(fd.get('id') ?? '');
		if (!parsed.success) return fail(400, { error: 'ID de usuario requerido' });

		const res = await user.deleteUser({ cookies, fetch }, parsed.data);
		if (!res.ok) return fail(res.status, { error: 'No se pudo eliminar el usuario' });
		return { success: true };
	},

	banUser: async ({ request, cookies, fetch }) => {
		const fd = await request.formData();
		const parsed = rowIdSchema.safeParse(fd.get('id') ?? '');
		const ban = fd.get('ban') === 'true';

		if (!parsed.success) return fail(400, { banError: 'ID requerido', banId: '' });

		const res = await user.banUser({ cookies, fetch }, parsed.data, { ban });

		if (!res.ok) {
			return fail(res.status, {
				banError: res.details ?? 'No se pudo actualizar el estado',
				banId: parsed.data
			});
		}

		return { banSuccess: true, banId: parsed.data, banned: ban };
	}
} satisfies Actions;
