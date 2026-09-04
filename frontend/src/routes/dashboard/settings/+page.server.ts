import type { Actions, PageServerLoad } from './$types';
import { fail } from '@sveltejs/kit';
import * as user from '$lib/api/user';
import * as market from '$lib/api/market';
import { env } from '$env/dynamic/private';
import type { ActiveSession, MarketCredential, MCPToken, TwoFactorStatus } from '$lib/api/types';
import {
	ALLOWED_AVATAR_TYPES,
	MAX_AVATAR_BYTES,
	changePasswordSchema,
	enableTwoFactorSchema,
	marketKeySchema,
	marketProviderSchema,
	mcpTokenError,
	mcpTokenExpirySchema,
	mcpTokenIdSchema,
	mcpTokenNameSchema,
	profileSchema,
	sessionIdSchema,
	setupTwoFactorSchema,
	twoFactorChallengeSchema
} from '$lib/features/settings';

export const load: PageServerLoad = async ({ locals, fetch, cookies }) => {
	const event = { cookies, fetch };

	let sessions: ActiveSession[] = [];
	// 2FA is off by default; the null fallback just hides the section's state
	// details if the backend can't be reached.
	let twoFactor: TwoFactorStatus = { enabled: false, pendingSetup: false, recoveryCodesLeft: 0 };
	// Nunca contiene la clave: solo proveedor, last4 y estado.
	let marketCredentials: MarketCredential[] = [];
	// Tampoco contiene el secreto: solo nombre, last4 y fechas.
	let mcpTokens: MCPToken[] = [];

	const [sessionsRes, twoFactorRes, credentialsRes, mcpTokensRes] = await Promise.all([
		user.getSessions(event),
		user.getTwoFactorStatus(event),
		market.getMarketCredentials(event),
		user.getMCPTokens(event)
	]);

	if (sessionsRes.ok) sessions = sessionsRes.data ?? [];
	if (twoFactorRes.ok && twoFactorRes.data) twoFactor = twoFactorRes.data;
	if (credentialsRes.ok) marketCredentials = credentialsRes.data ?? [];
	if (mcpTokensRes.ok) mcpTokens = mcpTokensRes.data ?? [];

	return {
		user: locals.user,
		sessions,
		twoFactor,
		marketCredentials,
		mcpTokens,
		// El cliente MCP habla con el backend directamente, no con esta app, así
		// que la URL que hay que pegar en su configuración es la del API. Se
		// resuelve aquí porque BASE_API es privada y no llega al navegador.
		mcpUrl: `${env.BASE_API}/mcp`
	};
};

export const actions = {
	updateProfile: async ({ request, fetch, cookies }) => {
		const formData = await request.formData();

		const parsed = profileSchema.safeParse({
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

		if (!ALLOWED_AVATAR_TYPES.includes(file.type)) {
			return fail(400, {
				action: 'uploadAvatar',
				error: 'Solo se permiten imágenes JPEG, PNG o WebP'
			});
		}

		if (file.size > MAX_AVATAR_BYTES) {
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

		const parsed = changePasswordSchema.safeParse({
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

		const parsed = sessionIdSchema.safeParse(sessionId);
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
		const parsed = setupTwoFactorSchema.safeParse(formData.get('password'));

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
		const parsed = enableTwoFactorSchema.safeParse(formData.get('code'));

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
		const parsed = twoFactorChallengeSchema.safeParse({
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
		const parsed = twoFactorChallengeSchema.safeParse({
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

	// --- Datos de mercado (BYO-key) ----------------------------------------
	//
	// La clave viaja del formulario al backend y ahí se sella; no se guarda en
	// ninguna cookie ni se devuelve nunca. Ninguna de estas acciones incluye la
	// clave en su valor de retorno, que es lo que acaba en `form` y por tanto en
	// el HTML de la página.

	saveMarketKey: async ({ request, fetch, cookies }) => {
		const formData = await request.formData();

		const parsed = marketKeySchema.safeParse({
			provider: formData.get('provider'),
			apiKey: formData.get('apiKey')
		});

		if (!parsed.success) {
			return fail(400, {
				action: 'saveMarketKey',
				marketProvider: formData.get('provider'),
				marketError: parsed.error.issues[0].message
			});
		}

		const res = await market.saveMarketCredential(
			{ cookies, fetch },
			parsed.data.provider,
			parsed.data.apiKey
		);

		if (!res.ok) {
			// El backend verifica la clave contra el proveedor antes de guardarla,
			// así que un 400 aquí significa que el proveedor la rechazó.
			const error =
				res.status === 400
					? 'El proveedor rechazó esta clave. Compruébala y vuelve a intentarlo.'
					: 'No se pudo guardar la clave. Inténtalo de nuevo.';

			return fail(res.status, {
				action: 'saveMarketKey',
				marketProvider: parsed.data.provider,
				marketError: error
			});
		}

		return {
			action: 'saveMarketKey',
			marketProvider: parsed.data.provider,
			marketSuccess: true,
			marketMessage: 'Clave verificada y guardada cifrada.'
		};
	},

	verifyMarketKey: async ({ request, fetch, cookies }) => {
		const formData = await request.formData();
		const parsed = marketProviderSchema.safeParse(formData.get('provider'));

		if (!parsed.success) {
			return fail(400, { action: 'verifyMarketKey', marketError: 'Proveedor no válido' });
		}

		const res = await market.verifyMarketCredential({ cookies, fetch }, parsed.data);

		if (!res.ok) {
			return fail(res.status, {
				action: 'verifyMarketKey',
				marketProvider: parsed.data,
				marketError: 'No se pudo verificar la clave.'
			});
		}

		const status = res.data?.status;
		const message =
			status === 'active'
				? 'La clave funciona.'
				: status === 'rate_limited'
					? 'La clave es válida, pero su cuota está agotada.'
					: 'El proveedor rechazó esta clave.';

		return {
			action: 'verifyMarketKey',
			marketProvider: parsed.data,
			marketSuccess: true,
			marketMessage: message
		};
	},

	deleteMarketKey: async ({ request, fetch, cookies }) => {
		const formData = await request.formData();
		const parsed = marketProviderSchema.safeParse(formData.get('provider'));

		if (!parsed.success) {
			return fail(400, { action: 'deleteMarketKey', marketError: 'Proveedor no válido' });
		}

		const res = await market.deleteMarketCredential({ cookies, fetch }, parsed.data);

		if (!res.ok) {
			return fail(res.status, {
				action: 'deleteMarketKey',
				marketProvider: parsed.data,
				marketError: 'No se pudo eliminar la clave.'
			});
		}

		return {
			action: 'deleteMarketKey',
			marketProvider: parsed.data,
			marketSuccess: true,
			marketMessage: 'Clave eliminada.'
		};
	},

	syncMarketData: async ({ fetch, cookies }) => {
		const res = await market.syncMarketData({ cookies, fetch });

		if (!res.ok) {
			const error =
				res.status === 400
					? 'Configura una clave antes de sincronizar.'
					: 'La sincronización falló. Inténtalo de nuevo en unos minutos.';

			return fail(res.status, { action: 'syncMarketData', marketSyncError: error });
		}

		return {
			action: 'syncMarketData',
			marketSyncSuccess: true,
			marketSyncCount: res.data?.prices?.length ?? 0,
			marketSyncRateCount: res.data?.rates?.length ?? 0
		};
	},

	// --- Tokens MCP --------------------------------------------------------
	//
	// El secreto solo existe en la respuesta de crear y de rotar: acaba en el
	// `form` de la página, que lo muestra una vez, y en ningún otro sitio.

	createMcpToken: async ({ request, fetch, cookies }) => {
		const action = 'createMcpToken';
		const formData = await request.formData();

		const name = mcpTokenNameSchema.safeParse(formData.get('name'));
		if (!name.success) return fail(400, { action, error: name.error.issues[0].message });

		const days = mcpTokenExpirySchema.safeParse(formData.get('expiresInDays'));
		if (!days.success) return fail(400, { action, error: days.error.issues[0].message });

		const res = await user.createMCPToken(
			{ cookies, fetch },
			{ name: name.data, expiresInDays: days.data }
		);

		if (!res.ok) {
			return fail(res.status, { action, error: mcpTokenError(action, res.status, res.details) });
		}

		return { action, success: true, mcpToken: res.data };
	},

	rotateMcpToken: async ({ request, fetch, cookies }) => {
		const action = 'rotateMcpToken';
		const formData = await request.formData();

		const tokenId = mcpTokenIdSchema.safeParse(formData.get('tokenId'));
		if (!tokenId.success) return fail(400, { action, error: 'Token no válido' });

		const days = mcpTokenExpirySchema.safeParse(formData.get('expiresInDays'));
		if (!days.success) return fail(400, { action, error: days.error.issues[0].message });

		const res = await user.rotateMCPToken({ cookies, fetch }, tokenId.data, {
			expiresInDays: days.data
		});

		if (!res.ok) {
			return fail(res.status, { action, error: mcpTokenError(action, res.status, res.details) });
		}

		return { action, success: true, mcpToken: res.data };
	},

	deleteMcpToken: async ({ request, fetch, cookies }) => {
		const action = 'deleteMcpToken';
		const formData = await request.formData();

		const tokenId = mcpTokenIdSchema.safeParse(formData.get('tokenId'));
		if (!tokenId.success) return fail(400, { action, error: 'Token no válido' });

		const res = await user.deleteMCPToken({ cookies, fetch }, tokenId.data);

		if (!res.ok) {
			return fail(res.status, { action, error: mcpTokenError(action, res.status, res.details) });
		}

		return { action, success: true };
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
