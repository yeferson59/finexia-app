import type { Actions, PageServerLoad } from './$types';
import { z } from 'zod';
import { fail } from '@sveltejs/kit';
import { authedFetch, authedFetchSafe } from '$lib/server/api';

const ALLOWED_IMAGE_TYPES = ['image/jpeg', 'image/png', 'image/webp'];

export interface ActiveSession {
	id: string;
	ipAddress: string | null;
	userAgent: string | null;
	location: string | null;
	createdAt: string;
	lastActiveAt: string;
	expiresAt: string;
	current: boolean;
}

export interface TwoFactorStatus {
	enabled: boolean;
	pendingSetup: boolean;
	recoveryCodesLeft: number;
}

export const load: PageServerLoad = async ({ locals, fetch, cookies }) => {
	let sessions: ActiveSession[] = [];
	// 2FA is off by default; the null fallback just hides the section's state
	// details if the backend can't be reached.
	let twoFactor: TwoFactorStatus = { enabled: false, pendingSetup: false, recoveryCodesLeft: 0 };

	const [sessionsRes, twoFactorRes] = await Promise.all([
		authedFetchSafe({ cookies, fetch }, '/auth/sessions'),
		authedFetchSafe({ cookies, fetch }, '/auth/2fa')
	]);

	if (sessionsRes?.ok) {
		const body = await sessionsRes.json().catch(() => null);
		sessions = body?.data ?? [];
	}
	if (twoFactorRes?.ok) {
		const body = await twoFactorRes.json().catch(() => null);
		if (body?.data) twoFactor = body.data;
	}

	return { user: locals.user, sessions, twoFactor };
};

export const actions = {
	updateProfile: async ({ request, fetch, cookies }) => {
		const formData = await request.formData();

		const schema = z.object({
			name: z.string().min(2, 'El nombre debe tener al menos 2 caracteres').max(254),
			preferredCurrency: z
				.string()
				.length(3, 'La moneda debe ser un código de 3 caracteres')
				.toUpperCase(),
			image: z.string().optional()
		});

		const parsed = schema.safeParse({
			name: formData.get('name'),
			preferredCurrency: formData.get('preferredCurrency'),
			image: formData.get('image') || undefined
		});

		if (!parsed.success) {
			return fail(400, { action: 'updateProfile', error: parsed.error.issues[0].message });
		}

		const res = await authedFetch({ cookies, fetch }, '/users/me', {
			method: 'PATCH',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(parsed.data)
		});

		if (!res.ok) {
			const body = await res.json().catch(() => ({}));
			return fail(res.status, {
				action: 'updateProfile',
				error: body.details ?? 'Error al actualizar el perfil'
			});
		}

		return { action: 'updateProfile', success: true };
	},

	uploadAvatar: async ({ request, fetch, cookies }) => {
		const formData = await request.formData();
		const file = formData.get('avatar');

		if (!file || !(file instanceof File) || file.size === 0) {
			return fail(400, { action: 'uploadAvatar', error: 'Selecciona una imagen para subir' });
		}

		if (!ALLOWED_IMAGE_TYPES.includes(file.type)) {
			return fail(400, {
				action: 'uploadAvatar',
				error: 'Solo se permiten imágenes JPEG, PNG o WebP'
			});
		}

		if (file.size > 5 * 1024 * 1024) {
			return fail(400, { action: 'uploadAvatar', error: 'La imagen no puede superar 5 MB' });
		}

		const body = new FormData();
		body.append('avatar', file);

		const res = await authedFetch({ cookies, fetch }, '/users/me/avatar', {
			method: 'POST',
			body
		});

		if (!res.ok) {
			const err = await res.json().catch(() => ({}));
			return fail(res.status, {
				action: 'uploadAvatar',
				error: err.details ?? 'Error al subir la imagen'
			});
		}

		const { data } = await res.json();
		return { action: 'uploadAvatar', success: true, imageUrl: data?.image ?? '' };
	},

	changePassword: async ({ request, fetch, cookies }) => {
		const formData = await request.formData();

		const schema = z
			.object({
				currentPassword: z.string().min(8, 'La contraseña debe tener al menos 8 caracteres'),
				// El login exige max 20; sin este límite el usuario podría fijar una
				// contraseña con la que luego no puede iniciar sesión.
				newPassword: z
					.string()
					.min(8, 'La nueva contraseña debe tener al menos 8 caracteres')
					.max(20, 'La nueva contraseña no puede superar 20 caracteres'),
				confirmPassword: z.string()
			})
			.refine((d) => d.newPassword === d.confirmPassword, {
				message: 'Las contraseñas no coinciden',
				path: ['confirmPassword']
			});

		const parsed = schema.safeParse({
			currentPassword: formData.get('currentPassword'),
			newPassword: formData.get('newPassword'),
			confirmPassword: formData.get('confirmPassword')
		});

		if (!parsed.success) {
			return fail(400, { action: 'changePassword', error: parsed.error.issues[0].message });
		}

		const res = await authedFetch({ cookies, fetch }, '/users/me/password', {
			method: 'PATCH',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({
				currentPassword: parsed.data.currentPassword,
				newPassword: parsed.data.newPassword
			})
		});

		if (!res.ok) {
			const errorMsg =
				res.status === 400 ? 'Contraseña actual incorrecta' : 'Error al cambiar la contraseña';
			return fail(res.status, { action: 'changePassword', error: errorMsg });
		}

		return { action: 'changePassword', success: true };
	},

	revokeSession: async ({ request, fetch, cookies }) => {
		const formData = await request.formData();
		const sessionId = formData.get('sessionId');

		const parsed = z.uuid().safeParse(sessionId);
		if (!parsed.success) {
			return fail(400, { action: 'revokeSession', error: 'Sesión inválida' });
		}

		const res = await authedFetch({ cookies, fetch }, `/auth/sessions/${parsed.data}`, {
			method: 'DELETE'
		});

		if (!res.ok) {
			return fail(res.status, {
				action: 'revokeSession',
				error: 'No se pudo cerrar la sesión. Inténtalo de nuevo.'
			});
		}

		return { action: 'revokeSession', success: true };
	},

	setup2fa: async ({ request, fetch, cookies }) => {
		const formData = await request.formData();
		const parsed = z
			.string()
			.min(8, 'Ingresa tu contraseña actual')
			.max(20)
			.safeParse(formData.get('password'));

		if (!parsed.success) {
			return fail(400, { action: 'setup2fa', error: parsed.error.issues[0].message });
		}

		const res = await authedFetch({ cookies, fetch }, '/auth/2fa/setup', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ password: parsed.data })
		});

		if (!res.ok) {
			const body = await res.json().catch(() => ({}));
			const error =
				body.action === 'auth:2fa:already-enabled'
					? 'La verificación en dos pasos ya está activada.'
					: 'Contraseña incorrecta';
			return fail(res.status, { action: 'setup2fa', error });
		}

		const body = await res.json().catch(() => null);
		return {
			action: 'setup2fa',
			success: true,
			secret: (body?.data?.secret as string) ?? '',
			otpauthUrl: (body?.data?.otpauthUrl as string) ?? ''
		};
	},

	enable2fa: async ({ request, fetch, cookies }) => {
		const formData = await request.formData();
		const parsed = z
			.string()
			.trim()
			.min(6, 'Ingresa el código de 6 dígitos')
			.max(20)
			.safeParse(formData.get('code'));

		if (!parsed.success) {
			return fail(400, { action: 'enable2fa', error: parsed.error.issues[0].message });
		}

		const res = await authedFetch({ cookies, fetch }, '/auth/2fa/enable', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ code: parsed.data })
		});

		if (!res.ok) {
			return fail(res.status, {
				action: 'enable2fa',
				error: 'Código incorrecto. Comprueba tu aplicación de autenticación e inténtalo de nuevo.'
			});
		}

		const body = await res.json().catch(() => null);
		return {
			action: 'enable2fa',
			success: true,
			recoveryCodes: (body?.data?.recoveryCodes as string[]) ?? []
		};
	},

	disable2fa: async ({ request, fetch, cookies }) => {
		const formData = await request.formData();
		const parsed = z
			.object({
				password: z.string().min(8, 'Ingresa tu contraseña actual').max(20),
				code: z.string().trim().min(6, 'Ingresa un código válido').max(20)
			})
			.safeParse({
				password: formData.get('password'),
				code: formData.get('code')
			});

		if (!parsed.success) {
			return fail(400, { action: 'disable2fa', error: parsed.error.issues[0].message });
		}

		const res = await authedFetch({ cookies, fetch }, '/auth/2fa/disable', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(parsed.data)
		});

		if (!res.ok) {
			return fail(res.status, {
				action: 'disable2fa',
				error: 'Contraseña o código incorrecto.'
			});
		}

		return { action: 'disable2fa', success: true };
	},

	regenerate2faCodes: async ({ request, fetch, cookies }) => {
		const formData = await request.formData();
		const parsed = z
			.object({
				password: z.string().min(8, 'Ingresa tu contraseña actual').max(20),
				code: z.string().trim().min(6, 'Ingresa un código válido').max(20)
			})
			.safeParse({
				password: formData.get('password'),
				code: formData.get('code')
			});

		if (!parsed.success) {
			return fail(400, { action: 'regenerate2faCodes', error: parsed.error.issues[0].message });
		}

		const res = await authedFetch({ cookies, fetch }, '/auth/2fa/recovery-codes', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(parsed.data)
		});

		if (!res.ok) {
			return fail(res.status, {
				action: 'regenerate2faCodes',
				error: 'Contraseña o código incorrecto.'
			});
		}

		const body = await res.json().catch(() => null);
		return {
			action: 'regenerate2faCodes',
			success: true,
			recoveryCodes: (body?.data?.recoveryCodes as string[]) ?? []
		};
	},

	revokeOtherSessions: async ({ fetch, cookies }) => {
		const res = await authedFetch({ cookies, fetch }, '/auth/sessions/revoke-others', {
			method: 'POST'
		});

		if (!res.ok) {
			return fail(res.status, {
				action: 'revokeOtherSessions',
				error: 'No se pudieron cerrar las demás sesiones. Inténtalo de nuevo.'
			});
		}

		const body = await res.json().catch(() => null);
		return {
			action: 'revokeOtherSessions',
			success: true,
			revoked: body?.data?.revoked ?? 0
		};
	}
} satisfies Actions;
