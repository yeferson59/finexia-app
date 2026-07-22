import { z } from 'zod';

/**
 * Schemas Zod de los formularios de auth, centralizados aquí para que las form
 * actions de `routes/auth/**` validen todas contra la misma fuente de verdad.
 * Los límites reflejan los DTO del backend (`docs/API.md`): las contraseñas van
 * de 8 a 20 caracteres para que un valor aceptado aquí nunca lo rechace el login.
 */

export const loginSchema = z.object({
	email: z.email().min(2),
	// El backend (LoginRequestDTO) valida min=8,max=20.
	password: z.string().min(8).max(20)
});

export const twoFactorSchema = z.object({
	token: z.string().min(1),
	// 6 dígitos TOTP o un código de recuperación XXXXX-XXXXX.
	code: z.string().trim().min(6).max(20)
});

export const registerSchema = z.object({
	name: z.string().min(2),
	email: z.email().min(2),
	// El backend (RegisterRequestDTO) valida min=8,max=20.
	password: z.string().min(8).max(20),
	confirmPassword: z.string().min(8).max(20),
	terms: z.coerce.boolean()
});

export const forgotPasswordSchema = z.object({ email: z.email().min(2) });

export const resetPasswordSchema = z.object({
	token: z.string().min(1),
	// Mirror the backend bounds (min=8,max=20) so login never rejects it.
	password: z.string().min(8).max(20),
	confirmPassword: z.string().min(8).max(20)
});

export const acceptInviteSchema = z.object({
	token: z.string().min(1),
	name: z.string().min(2).max(254),
	// Mirror the backend bounds (min=8,max=20) so login never rejects it.
	password: z.string().min(8).max(20),
	confirmPassword: z.string().min(8).max(20)
});

export const verifyEmailConfirmSchema = z.object({ token: z.string().min(1) });

export const resendVerificationSchema = z.object({ email: z.email().min(2) });
