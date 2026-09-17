import { z } from 'zod';

/**
 * Schemas Zod de los formularios de auth, centralizados aquí para que las form
 * actions de `routes/auth/**` validen todas contra la misma fuente de verdad.
 * Los límites reflejan los DTO del backend (`docs/API.md`): las contraseñas van
 * de 8 a 20 caracteres para que un valor aceptado aquí nunca lo rechace el login.
 *
 * Cada regla lleva su mensaje. Sin él, el formulario pintaba el de Zod tal cual
 * —«Too small: expected string to have >=8 characters»— debajo del campo.
 */

const EMAIL_ERROR = 'Escribe un correo válido, como tu@correo.com.';
const PASSWORD_LENGTH_ERROR = 'La contraseña tiene entre 8 y 20 caracteres.';
const LINK_ERROR = 'El enlace no es válido. Ábrelo de nuevo desde el correo.';

const email = z.email({ error: EMAIL_ERROR }).min(2, { error: EMAIL_ERROR });

// El backend (LoginRequestDTO, RegisterRequestDTO) valida min=8,max=20.
const password = z
	.string({ error: 'Escribe la contraseña.' })
	.min(8, { error: PASSWORD_LENGTH_ERROR })
	.max(20, { error: PASSWORD_LENGTH_ERROR });

const token = z.string({ error: LINK_ERROR }).min(1, { error: LINK_ERROR });

const name = z
	.string({ error: 'Escribe tu nombre.' })
	.min(2, { error: 'Escribe tu nombre, de al menos 2 letras.' });

export const loginSchema = z.object({ email, password });

export const twoFactorSchema = z.object({
	token,
	// 6 dígitos TOTP o un código de recuperación XXXXX-XXXXX.
	code: z
		.string({ error: 'Escribe el código.' })
		.trim()
		.min(6, { error: 'Escribe los 6 dígitos o un código de recuperación.' })
		.max(20, { error: 'Escribe los 6 dígitos o un código de recuperación.' })
});

export const registerSchema = z.object({
	name,
	email,
	password,
	confirmPassword: password,
	terms: z.coerce.boolean()
});

export const forgotPasswordSchema = z.object({ email });

export const resetPasswordSchema = z.object({
	token,
	// Mirror the backend bounds (min=8,max=20) so login never rejects it.
	password,
	confirmPassword: password
});

export const acceptInviteSchema = z.object({
	token,
	name: name.max(254, { error: 'El nombre es demasiado largo.' }),
	// Mirror the backend bounds (min=8,max=20) so login never rejects it.
	password,
	confirmPassword: password
});

export const verifyEmailConfirmSchema = z.object({ token });

export const resendVerificationSchema = z.object({ email });
