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
	version: '1.6',
	date: 'Agosto 2026',
	bytes: 2827259,
	generatedAt: '2026-08-23T15:18:31.804Z',
	sourceHash: 'd1779c9113d0168159bc24330fdeaf13acc1c7d5d3d5364ba927066a1b114f00',
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
			title: 'Plataformas'
		},
		{
			number: 9,
			title: 'Transacciones'
		},
		{
			number: 10,
			title: 'Importación masiva de transacciones (Excel/CSV)'
		},
		{
			number: 11,
			title: 'Reportes y exportaciones'
		},
		{
			number: 12,
			title: 'Notificaciones'
		},
		{
			number: 13,
			title: 'Configuración de la cuenta'
		},
		{
			number: 14,
			title: 'Seguridad: 2FA y sesiones'
		},
		{
			number: 15,
			title: 'Preguntas frecuentes (FAQ)'
		},
		{
			number: 16,
			title: 'Solución de problemas'
		},
		{
			number: 17,
			title: 'Glosario'
		}
	]
};
