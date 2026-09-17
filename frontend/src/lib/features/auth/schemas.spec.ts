import { describe, it, expect } from 'vitest';
import { loginSchema, registerSchema, twoFactorSchema } from './schemas';

function messages(result: { success: boolean; error?: { issues: { message: string }[] } }) {
	return result.error?.issues.map((issue) => issue.message) ?? [];
}

describe('mensajes de validación de auth', () => {
	it('explican en español qué falta, no el texto de Zod', () => {
		const result = loginSchema.safeParse({ email: 'no-es-correo', password: 'corta' });
		expect(messages(result)).toEqual([
			'Escribe un correo válido, como tu@correo.com.',
			'La contraseña tiene entre 8 y 20 caracteres.'
		]);
	});

	it('cubren un campo que no llegó en el formulario', () => {
		const result = loginSchema.safeParse({ email: null, password: null });
		expect(messages(result)).toEqual([
			'Escribe un correo válido, como tu@correo.com.',
			'Escribe la contraseña.'
		]);
	});

	it('usan el mismo límite de contraseña por arriba', () => {
		const result = registerSchema.safeParse({
			name: 'Laura',
			email: 'laura@correo.com',
			password: 'x'.repeat(21),
			confirmPassword: 'x'.repeat(21),
			terms: 'on'
		});
		expect(messages(result)).toContain('La contraseña tiene entre 8 y 20 caracteres.');
	});

	it('siguen aceptando lo que acepta el backend', () => {
		expect(loginSchema.safeParse({ email: 'laura@correo.com', password: '12345678' }).success).toBe(
			true
		);
		expect(twoFactorSchema.safeParse({ token: 't', code: 'ABCDE-12345' }).success).toBe(true);
	});
});
