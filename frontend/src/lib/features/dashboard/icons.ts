/*
 * Los iconos del panel, como datos.
 *
 * Vivían en dos cadenas de `{#if}` dentro del menú lateral: dieciocho bloques
 * de SVG en línea que ocupaban el 80 % del archivo y escondían la lógica de
 * navegación entre ellos. Aquí son un mapa de trazos, y `icon.svelte` los
 * pinta; añadir una entrada al menú ya no obliga a tocar el marcado.
 *
 * Todos son trazos de 24×24 sin relleno, para que hereden grosor y color del
 * componente en vez de traerlos puestos.
 */
export const ICONS: Record<string, string[]> = {
	grid: ['M3 3h7v7H3zM14 3h7v7h-7zM14 14h7v7h-7zM3 14h7v7H3z'],
	briefcase: [
		'M4 7h16a2 2 0 0 1 2 2v10a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V9a2 2 0 0 1 2-2z',
		'M16 7V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v2'
	],
	pie: ['M21.21 15.89A10 10 0 1 1 8 2.83', 'M22 12A10 10 0 0 0 12 2v10z'],
	trending: ['M23 6l-9.5 9.5-5-5L1 17', 'M17 6h6v6'],
	layers: ['M12 2l8 4v12l-8 4-8-4V6l8-4z', 'M12 22V12', 'M4 6l8 6 8-6', 'M4 18l8-6 8 6'],
	exchange: ['M12 5v14', 'M19 12l-7 7-7-7'],
	bars: ['M12 20V10', 'M18 20V4', 'M6 20v-4'],
	bell: ['M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9', 'M13.73 21a2 2 0 0 1-3.46 0'],
	book: [
		'M4 19.5A2.5 2.5 0 0 1 6.5 17H20',
		'M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z'
	],
	gear: [
		'M15 12a3 3 0 1 1-6 0 3 3 0 0 1 6 0z',
		'M12 1v6m0 6v6M4.22 4.22l4.24 4.24m5.08 5.08l4.24 4.24M1 12h6m6 0h6m-1.78 7.78l-4.24-4.24m-5.08-5.08l-4.24-4.24'
	],
	shield: ['M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z'],
	users: [
		'M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2',
		'M13 7a4 4 0 1 1-8 0 4 4 0 0 1 8 0z',
		'M23 21v-2a4 4 0 0 0-3-3.87',
		'M16 3.13a4 4 0 0 1 0 7.75'
	],
	database: [
		'M21 5a9 3 0 1 1-18 0 9 3 0 0 1 18 0z',
		'M21 12c0 1.66-4 3-9 3s-9-1.34-9-3',
		'M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5'
	],
	rates: [
		'M17 1l4 4-4 4',
		'M3 11V9a4 4 0 0 1 4-4h14',
		'M7 23l-4-4 4-4',
		'M21 13v2a4 4 0 0 1-4 4H3'
	],
	logout: ['M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4', 'M16 17l5-5-5-5', 'M21 12H9'],
	menu: ['M3 12h18', 'M3 6h18', 'M3 18h18'],
	eye: ['M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z', 'M15 12a3 3 0 1 1-6 0 3 3 0 0 1 6 0z'],
	'eye-off': [
		'M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94',
		'M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19',
		'M14.12 14.12a3 3 0 1 1-4.24-4.24',
		'M1 1l22 22'
	],
	alert: ['M12 22a10 10 0 1 1 0-20 10 10 0 0 1 0 20z', 'M12 8v4', 'M12 16h.01']
};
