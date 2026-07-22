/**
 * Resultado de las form actions de login/registro (`routes/auth/+page.server.ts`),
 * consumido por el contenedor `login-register` y sus sub-formularios para pintar
 * errores y estados especiales (2FA, correo sin verificar, email duplicado…).
 */
export type AuthActionResult = {
	type: 'login' | 'register';
	errors: Record<string, string> | Array<{ path: PropertyKey[]; message: string }>;
	unverified?: boolean;
	duplicateEmail?: boolean;
	disabled?: boolean;
	twoFactorRequired?: boolean;
	twoFactorToken?: string;
} | null;
