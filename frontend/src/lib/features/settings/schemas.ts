/**
 * Schemas Zod de los formularios de ajustes.
 *
 * Estaban declarados dentro de cada action de `settings/+page.server.ts`; aquí
 * viven juntos para que las reglas de un mismo campo (longitud de contraseña,
 * formato del código 2FA) no diverjan entre acciones.
 */

import { z } from 'zod';

/**
 * El login exige `max(20)`: sin ese límite aquí el usuario podría fijar una
 * contraseña con la que luego no puede iniciar sesión.
 */
const password = z.string().min(8, 'La contraseña debe tener al menos 8 caracteres').max(20);

/** Código del autenticador o de recuperación. */
const otpCode = z.string().trim().min(6, 'Ingresa un código válido').max(20);

export const profileSchema = z.object({
	name: z.string().min(2, 'El nombre debe tener al menos 2 caracteres').max(254),
	preferredCurrency: z
		.string()
		.length(3, 'La moneda debe ser un código de 3 caracteres')
		.toUpperCase(),
	image: z.string().optional()
});

export const changePasswordSchema = z
	.object({
		currentPassword: password,
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

export const sessionIdSchema = z.uuid();

/** Contraseña que autoriza a empezar el alta de 2FA. */
export const setupTwoFactorSchema = z.string().min(8, 'Ingresa tu contraseña actual').max(20);

/** Código que confirma el alta de 2FA. */
export const enableTwoFactorSchema = z
	.string()
	.trim()
	.min(6, 'Ingresa el código de 6 dígitos')
	.max(20);

/** Contraseña + código: desactivar 2FA y regenerar los códigos piden ambos. */
export const twoFactorChallengeSchema = z.object({
	password: z.string().min(8, 'Ingresa tu contraseña actual').max(20),
	code: otpCode
});

/** Proveedores para los que el backend acepta una clave. */
export const marketProviderSchema = z.enum(['finnhub', 'alphavantage']);

export const marketKeySchema = z.object({
	provider: marketProviderSchema,
	apiKey: z.string().trim().min(8, 'La clave es demasiado corta').max(256)
});

/** Tipos de imagen que el backend acepta como avatar. */
export const ALLOWED_AVATAR_TYPES = ['image/jpeg', 'image/png', 'image/webp'];

/** Tamaño máximo del avatar antes de comprimir, en bytes. */
export const MAX_AVATAR_BYTES = 5 * 1024 * 1024;
