import type { Actions, PageServerLoad } from './$types';
import { z } from 'zod';
import { fail } from '@sveltejs/kit';
import * as user from '$lib/api/user';
import type { ActiveSession, TwoFactorStatus } from '$lib/api/types';

const ALLOWED_IMAGE_TYPES = ['image/jpeg', 'image/png', 'image/webp'];

export const load: PageServerLoad = async ({ locals, fetch, cookies }) => {
	const event = { cookies, fetch };

	let sessions: ActiveSession[] = [];
	// 2FA is off by default; the null fallback just hides the section's state
	// details if the backend can't be reached.
	let twoFactor: TwoFactorStatus = { enabled: false, pendingSetup: false, recoveryCodesLeft: 0 };

	const [sessionsRes, twoFactorRes] = await Promise.all([
		user.getSessions(event),
		user.getTwoFactorStatus(event)
	]);

	if (sessionsRes.ok) sessions = sessionsRes.data ?? [];
	if (twoFactorRes.ok && twoFactorRes.data) twoFactor = twoFactorRes.data;

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

		const res = await user.updateProfile({ cookies, fetch }, parsed.data);

		if (!res.ok) {
			return fail(res.status, {
				action: 'updateProfile',
				error: res.details ?? 'Error al actualizar el perfil'
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

		const res = await user.uploadAvatar({ cookies, fetch }, body);

		if (!res.ok) {
			return fail(res.status, {
				action: 'uploadAvatar',
				error: res.details ?? 'Error al subir la imagen'
			});
		}

		return { action: 'uploadAvatar', success: true, imageUrl: res.data?.image ?? '' };
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

		const res = await user.changePassword(
			{ cookies, fetch },
			{ currentPassword: parsed.data.currentPassword, newPassword: parsed.data.newPassword }
		);

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

		const res = await user.revokeSession({ cookies, fetch }, parsed.data);

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

		const res = await user.setupTwoFactor({ cookies, fetch }, { password: parsed.data });

		if (!res.ok) {
			const error =
				res.action === 'auth:2fa:already-enabled'
					? 'La verificación en dos pasos ya está activada.'
					: 'Contraseña incorrecta';
			return fail(res.status, { action: 'setup2fa', error });
		}

		return {
			action: 'setup2fa',
			success: true,
			secret: res.data?.secret ?? '',
			otpauthUrl: res.data?.otpauthUrl ?? ''
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

		const res = await user.enableTwoFactor({ cookies, fetch }, { code: parsed.data });

		if (!res.ok) {
			return fail(res.status, {
				action: 'enable2fa',
				error: 'Código incorrecto. Comprueba tu aplicación de autenticación e inténtalo de nuevo.'
			});
		}

		return {
			action: 'enable2fa',
			success: true,
			recoveryCodes: res.data?.recoveryCodes ?? []
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

		const res = await user.disableTwoFactor({ cookies, fetch }, parsed.data);

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

		const res = await user.regenerateRecoveryCodes({ cookies, fetch }, parsed.data);

		if (!res.ok) {
			return fail(res.status, {
				action: 'regenerate2faCodes',
				error: 'Contraseña o código incorrecto.'
			});
		}

		return {
			action: 'regenerate2faCodes',
			success: true,
			recoveryCodes: res.data?.recoveryCodes ?? []
		};
	},

	revokeOtherSessions: async ({ fetch, cookies }) => {
		const res = await user.revokeOtherSessions({ cookies, fetch });

		if (!res.ok) {
			return fail(res.status, {
				action: 'revokeOtherSessions',
				error: 'No se pudieron cerrar las demás sesiones. Inténtalo de nuevo.'
			});
		}

		return {
			action: 'revokeOtherSessions',
			success: true,
			revoked: res.data?.revoked ?? 0
		};
	}
} satisfies Actions;
