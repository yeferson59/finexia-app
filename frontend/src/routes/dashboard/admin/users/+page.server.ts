import type { Actions, PageServerLoad } from './$types';
import { authedFetch, authedFetchSafe } from '$lib/server/api';
import { fail } from '@sveltejs/kit';

interface UserItem {
	id: string;
	name: string;
	email: string;
	emailVerified: boolean;
	createdAt: string;
	bannedAt: string | null;
	role: { name: string };
}

interface InvitationItem {
	id: string;
	email: string;
	name: string;
	role: string;
	status: 'pending' | 'expired' | 'accepted' | 'revoked';
	expiresAt: string;
	createdAt: string;
}

interface WaitlistItem {
	id: string;
	email: string;
	status: 'pending' | 'invited' | 'registered';
	invitedAt: string | null;
	createdAt: string;
}

async function loadList<T>(
	event: { cookies: import('@sveltejs/kit').Cookies; fetch: typeof fetch },
	path: string
): Promise<T[]> {
	const res = await authedFetchSafe(event, path);
	if (!res || !res.ok) return [];
	const { data, success } = await res.json().catch(() => ({ success: false }));
	return success ? ((data?.items ?? []) as T[]) : [];
}

export const load: PageServerLoad = async ({ cookies, fetch, url }) => {
	const event = { cookies, fetch };
	const page = Number(url.searchParams.get('page') ?? '1');

	const res = await authedFetch(event, `/users?page=${page}&limit=20`);
	const { data, success } = await res.json();

	const [invitations, waitlist] = await Promise.all([
		loadList<InvitationItem>(event, '/users/invitations?page=1&limit=50'),
		loadList<WaitlistItem>(event, '/users/waitlist?page=1&limit=50')
	]);

	return {
		users: success ? ((data?.items ?? []) as UserItem[]) : [],
		meta: success
			? (data?.metaData ?? { currentPage: 1, totalPages: 1, previous: false, next: false })
			: { currentPage: 1, totalPages: 1, previous: false, next: false },
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

		const res = await authedFetch({ cookies, fetch }, '/users/invitations', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ email, name, role })
		});

		if (!res.ok) {
			const body = await res.json().catch(() => ({}));
			return fail(res.status, {
				error: body.details ?? body.message ?? 'No se pudo enviar la invitación'
			});
		}

		return { success: true, invited: email };
	},

	resendInvitation: async ({ request, cookies, fetch }) => {
		const fd = await request.formData();
		const id = fd.get('id') as string;
		if (!id) return fail(400, { inviteError: 'ID requerido', inviteId: '' });

		const res = await authedFetch({ cookies, fetch }, `/users/invitations/${id}/resend`, {
			method: 'POST'
		});
		if (!res.ok) {
			const body = await res.json().catch(() => ({}));
			return fail(res.status, {
				inviteError: body.details ?? 'No se pudo reenviar la invitación',
				inviteId: id
			});
		}
		return { inviteSuccess: true, inviteId: id, inviteAction: 'resent' as const };
	},

	revokeInvitation: async ({ request, cookies, fetch }) => {
		const fd = await request.formData();
		const id = fd.get('id') as string;
		if (!id) return fail(400, { inviteError: 'ID requerido', inviteId: '' });

		const res = await authedFetch({ cookies, fetch }, `/users/invitations/${id}`, {
			method: 'DELETE'
		});
		if (!res.ok) {
			const body = await res.json().catch(() => ({}));
			return fail(res.status, {
				inviteError: body.details ?? 'No se pudo revocar la invitación',
				inviteId: id
			});
		}
		return { inviteSuccess: true, inviteId: id, inviteAction: 'revoked' as const };
	},

	deleteUser: async ({ request, cookies, fetch }) => {
		const fd = await request.formData();
		const id = fd.get('id') as string;
		if (!id) return fail(400, { error: 'ID de usuario requerido' });

		const res = await authedFetch({ cookies, fetch }, `/users/${id}`, { method: 'DELETE' });
		if (!res.ok) return fail(res.status, { error: 'No se pudo eliminar el usuario' });
		return { success: true };
	},

	banUser: async ({ request, cookies, fetch }) => {
		const fd = await request.formData();
		const id = fd.get('id') as string;
		const ban = fd.get('ban') === 'true';
		if (!id) return fail(400, { banError: 'ID requerido', banId: '' });

		const res = await authedFetch({ cookies, fetch }, `/users/${id}/ban`, {
			method: 'PATCH',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ ban })
		});

		if (!res.ok) {
			const body = await res.json().catch(() => ({}));
			return fail(res.status, {
				banError: body.details ?? 'No se pudo actualizar el estado',
				banId: id
			});
		}

		return { banSuccess: true, banId: id, banned: ban };
	}
} satisfies Actions;
