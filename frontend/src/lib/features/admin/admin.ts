/**
 * Constantes, formateadores y tipos del área de administración.
 *
 * Los contratos del backend no se redeclaran aquí: se reexportan de
 * `$lib/api/types` para que las tres pantallas de admin (usuarios, activos y
 * tasas de cambio) hablen del mismo shape que la capa de API.
 */

import type { ImportResult } from '$lib/api/types';

export type {
	Asset,
	AssetPrice,
	ExchangeRate,
	ImportResult,
	InvitationItem,
	PageMeta,
	UserItem,
	WaitlistItem
} from '$lib/api/types';

/** Tono de `Badge` admitido por los estados que pinta el admin. */
export type BadgeTone = 'neutral' | 'amber' | 'success' | 'warning' | 'danger' | 'info';

/** Tipos de activo que acepta el catálogo compartido. */
export const ASSET_TYPES = [
	{ value: 'stock', label: 'Acción (Stock)' },
	{ value: 'etf', label: 'ETF' },
	{ value: 'crypto', label: 'Cripto' },
	{ value: 'bond', label: 'Bono' },
	{ value: 'real_estate', label: 'Bienes raíces' },
	{ value: 'commodity', label: 'Commodities' },
	{ value: 'cash', label: 'Efectivo' },
	{ value: 'other', label: 'Otro' }
];

/** Roles que se pueden asignar al invitar a alguien. */
export const INVITE_ROLES = [
	{ value: 'customer', label: 'Usuario' },
	{ value: 'admin', label: 'Administrador' }
];

export const INVITATION_STATUS_LABELS: Record<string, string> = {
	pending: 'Pendiente',
	expired: 'Expirada',
	accepted: 'Aceptada',
	revoked: 'Revocada'
};

export function invitationStatusLabel(status: string): string {
	return INVITATION_STATUS_LABELS[status] ?? status;
}

export function invitationStatusTone(status: string): BadgeTone {
	if (status === 'accepted') return 'success';
	if (status === 'revoked') return 'danger';
	if (status === 'expired') return 'neutral';
	return 'amber';
}

const dayFormatter = new Intl.DateTimeFormat('es', { dateStyle: 'medium' });
const dateTimeFormatter = new Intl.DateTimeFormat('es', {
	dateStyle: 'short',
	timeStyle: 'short'
});

/** Fecha sin hora (altas de usuario, caducidad de invitaciones). */
export function formatDay(iso: string): string {
	return dayFormatter.format(new Date(iso));
}

/** Fecha con hora; `—` cuando el backend no manda valor. */
export function formatDateTime(iso: string | null): string {
	if (!iso) return '—';
	return dateTimeFormatter.format(new Date(iso));
}

/**
 * Precio de un activo con su moneda.
 *
 * El backend usa `XXX` para «sin moneda»; en ese caso vale la del activo. Si el
 * código no es válido, `Intl` lanza y se cae a un formato plano en vez de
 * romper la tabla entera.
 */
export function formatPrice(
	price: { value: string; currency: string } | null,
	assetCurrency: string | undefined = undefined
): string {
	if (!price) return '—';
	const num = parseFloat(price.value);
	if (isNaN(num)) return price.value;
	const currency =
		price.currency && price.currency !== 'XXX' ? price.currency : assetCurrency || 'USD';
	try {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency,
			currencyDisplay: 'narrowSymbol',
			minimumFractionDigits: 2,
			maximumFractionDigits: 4
		}).format(num);
	} catch {
		return `${currency} ${num.toFixed(2)}`;
	}
}

/** Tasa de cambio con separadores de millar. */
export function formatRate(rate: string): string {
	const num = parseFloat(rate);
	if (isNaN(num)) return rate;
	return num.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 6 });
}

/** Resumen en una línea de un import masivo, con los plurales resueltos. */
export function summarizeImport(result: ImportResult): string {
	const rows = `fila${result.totalRows === 1 ? '' : 's'}`;
	const imported = `importada${result.imported === 1 ? '' : 's'}`;
	const skipped =
		result.skipped > 0 ? `, ${result.skipped} omitida${result.skipped === 1 ? '' : 's'}` : '';
	return `${result.imported} de ${result.totalRows} ${rows} ${imported}${skipped}.`;
}
