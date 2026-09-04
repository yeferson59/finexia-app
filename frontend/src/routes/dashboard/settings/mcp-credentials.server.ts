/**
 * Acciones de ajustes para las credenciales que usan los clientes MCP: los
 * tokens personales, y las aplicaciones que se conectaron por OAuth.
 *
 * Están fuera de `+page.server.ts` porque son un dominio propio —la única parte
 * de los ajustes que emite y revoca credenciales para terceros— y porque esa
 * página ya rozaba el presupuesto de 500 líneas que comprueba `check:arch`.
 * Se recomponen allí con un spread, así que para SvelteKit no hay diferencia.
 */
import type { Actions } from './$types';
import { fail } from '@sveltejs/kit';
import * as user from '$lib/api/user';
import {
	mcpTokenError,
	mcpTokenExpirySchema,
	mcpTokenIdSchema,
	mcpTokenNameSchema,
	oauthGrantIdSchema
} from '$lib/features/settings';

export const mcpCredentialActions = {
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

	// --- Aplicaciones conectadas (OAuth) -----------------------------------

	revokeOAuthGrant: async ({ request, fetch, cookies }) => {
		const action = 'revokeOAuthGrant';
		const formData = await request.formData();

		const grantId = oauthGrantIdSchema.safeParse(formData.get('grantId'));
		if (!grantId.success) return fail(400, { action, error: 'Aplicación no válida' });

		const res = await user.revokeOAuthGrant({ cookies, fetch }, grantId.data);

		if (!res.ok) {
			return fail(res.status, {
				action,
				error:
					res.status === 404
						? 'Esa aplicación ya estaba desconectada.'
						: 'No se pudo desconectar la aplicación. Inténtalo de nuevo.'
			});
		}

		return { action, success: true };
	}
} satisfies Actions;
