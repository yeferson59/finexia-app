// GENERADO por `pnpm manual:build` — no editar a mano.
//
// Describe el PDF que hay en `static/manual-usuario.pdf`. `sourceHash` es la
// huella del manual y sus capturas: `pnpm check:manual` la recalcula y falla
// si el PDF publicado ya no corresponde al manual del repositorio.

export interface ManualMeta {
	/** Versión declarada en la portada del manual. */
	version: string;
	/** Fecha declarada en la portada. */
	date: string;
	/** Tamaño del PDF en bytes. */
	bytes: number;
	/** Fecha de generación del PDF, en ISO. */
	generatedAt: string;
	/** Huella de `docs/MANUAL_DE_USUARIO.md` y sus capturas. */
	sourceHash: string;
	/** Secciones de primer nivel, para el índice de la página de la guía. */
	sections: { number: number; title: string }[];
}

export const manual: ManualMeta = {
	version: '1.9',
	date: 'Septiembre 2026',
	bytes: 2700662,
	generatedAt: '2026-09-06T03:03:09.229Z',
	sourceHash: '6b92a5240a5cb4556fc1041db1df1c1a553104688eec68b0a84579067a8a6694',
	sections: [
		{
			number: 1,
			title: 'Introducción'
		},
		{
			number: 2,
			title: 'Requisitos y acceso'
		},
		{
			number: 3,
			title: 'Primeros pasos: registro e inicio de sesión'
		},
		{
			number: 4,
			title: 'Interfaz general de la aplicación'
		},
		{
			number: 5,
			title: 'Dashboard (panel principal)'
		},
		{
			number: 6,
			title: 'Portafolios'
		},
		{
			number: 7,
			title: 'Posiciones y activos'
		},
		{
			number: 8,
			title: 'Mis Activos (vista consolidada)'
		},
		{
			number: 9,
			title: 'Plataformas'
		},
		{
			number: 10,
			title: 'Transacciones'
		},
		{
			number: 11,
			title: 'Importación masiva de transacciones (Excel/CSV)'
		},
		{
			number: 12,
			title: 'Reportes y exportaciones'
		},
		{
			number: 13,
			title: 'Notificaciones'
		},
		{
			number: 14,
			title: 'Configuración de la cuenta'
		},
		{
			number: 15,
			title: 'Seguridad: 2FA y sesiones'
		},
		{
			number: 16,
			title: 'Conectar un asistente de IA (MCP)'
		},
		{
			number: 17,
			title: 'Preguntas frecuentes (FAQ)'
		},
		{
			number: 18,
			title: 'Solución de problemas'
		},
		{
			number: 19,
			title: 'Glosario'
		}
	]
};
