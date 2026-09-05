/*
 * El mapa de navegación del panel: una sola lista que alimenta el menú lateral
 * y el nombre de sección que enseña la cabecera.
 *
 * Estaba solo dentro del menú, así que la cabecera no tenía forma de saber en
 * qué página estaba el usuario — y no decía nada. Con la lista aquí las dos
 * partes del chrome hablan del mismo sitio y no pueden discrepar.
 */
import { resolve } from '$app/paths';
import { features } from '$lib/shared/config/features';

export interface NavItem {
	label: string;
	icon: string;
	href: string;
}

/** Menú principal. «Inversiones» solo aparece con su feature flag encendida. */
export const MAIN_NAV: NavItem[] = [
	{ label: 'Resumen', icon: 'grid', href: resolve('/dashboard') },
	{ label: 'Portafolios', icon: 'briefcase', href: resolve('/dashboard/portfolios') },
	{ label: 'Mis activos', icon: 'pie', href: resolve('/dashboard/assets') },
	...(features.investments
		? [{ label: 'Inversiones', icon: 'trending', href: resolve('/dashboard/investments') }]
		: []),
	{ label: 'Plataformas', icon: 'layers', href: resolve('/dashboard/platforms') },
	{ label: 'Transacciones', icon: 'exchange', href: resolve('/dashboard/transactions') },
	{ label: 'Reportes', icon: 'bars', href: resolve('/dashboard/reports') },
	{ label: 'Notificaciones', icon: 'bell', href: resolve('/dashboard/notifications') },
	{ label: 'Guía de usuario', icon: 'book', href: resolve('/dashboard/guia') },
	{ label: 'Configuración', icon: 'gear', href: resolve('/dashboard/settings') }
];

/** Solo para administradores; va en su propio grupo porque es otro permiso. */
export const ADMIN_NAV: NavItem[] = [
	{ label: 'Panel de administración', icon: 'shield', href: resolve('/dashboard/admin') },
	{ label: 'Usuarios', icon: 'users', href: resolve('/dashboard/admin/users') },
	{ label: 'Activos', icon: 'database', href: resolve('/dashboard/admin/assets') },
	{ label: 'Tasas de cambio', icon: 'rates', href: resolve('/dashboard/admin/exchange-rates') }
];

const HOME = resolve('/dashboard');

/** Si `href` es la sección que se está mirando. */
export function isActive(href: string, pathname: string): boolean {
	// El resumen es la única ruta que no manda sobre sus hijas: cualquier página
	// del panel empieza por `/dashboard` y si no, saldría siempre marcada.
	return href === HOME ? pathname === href : pathname.startsWith(href);
}

/**
 * Nombre de la sección abierta, para la cabecera.
 *
 * Gana la coincidencia más larga: `/dashboard/admin/users` es «Usuarios» y no
 * «Panel de administración», aunque las dos rutas encajen.
 */
export function sectionTitle(pathname: string): string {
	const match = [...MAIN_NAV, ...ADMIN_NAV]
		.filter((item) => isActive(item.href, pathname))
		.sort((a, b) => b.href.length - a.href.length)[0];

	return match?.label ?? 'Panel';
}
