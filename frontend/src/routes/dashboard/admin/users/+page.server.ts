import type { Actions, PageServerLoad } from './$types';
import { authedFetch } from '$lib/server/api';
import { fail } from '@sveltejs/kit';

interface UserItem {
	id: string;
	name: string;
	email: string;
	emailVerified: boolean;
	createdAt: string;
	role: { name: string };
}

export const load: PageServerLoad = async ({ cookies, fetch, url }) => {
	const event = { cookies, fetch };
	const page = Number(url.searchParams.get('page') ?? '1');

	const res = await authedFetch(event, `/users?page=${page}&limit=20`);
	const { data, success } = await res.json();

	return {
		users: success ? ((data?.items ?? []) as UserItem[]) : [],
		meta: success
			? (data?.metaData ?? { currentPage: 1, totalPages: 1, previous: false, next: false })
			: { currentPage: 1, totalPages: 1, previous: false, next: false }
	};
};

export const actions = {
	deleteUser: async ({ request, cookies, fetch }) => {
		const fd = await request.formData();
		const id = fd.get('id') as string;
		if (!id) return fail(400, { error: 'ID de usuario requerido' });

		const res = await authedFetch({ cookies, fetch }, `/users/${id}`, { method: 'DELETE' });
		if (!res.ok) return fail(res.status, { error: 'No se pudo eliminar el usuario' });
		return { success: true };
	},

	createUser: async ({ request, cookies, fetch }) => {
		const fd = await request.formData();
		const name = (fd.get('name') as string | null)?.trim();
		const email = (fd.get('email') as string | null)?.trim();

		if (!name || !email) return fail(400, { error: 'Nombre y correo son requeridos' });

		const res = await authedFetch(
			{ cookies, fetch },
			'/users',
			{
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ name, email })
			}
		);

		if (!res.ok) {
			const body = await res.json().catch(() => ({}));
			return fail(res.status, { error: body.details ?? 'Error al crear el usuario' });
		}

		return { success: true };
	}
} satisfies Actions;
