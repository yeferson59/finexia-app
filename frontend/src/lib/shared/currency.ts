/**
 * Las monedas en las que la aplicación puede expresar dinero.
 *
 * Es mucho más corta que ISO 4217 y el corte lo decide el dato, no el gusto:
 * cada cifra que se muestra en una moneda distinta a la del portafolio es el
 * producto de una tasa guardada, así que ofrecer una moneda sin tasa detrás no
 * falla de forma visible — enseña importes sin convertir bajo el símbolo
 * equivocado. La lista es lo que publican contra el USD las dos fuentes
 * públicas sin clave: el feed de referencia del BCE y la TRM de dolarapi.
 *
 * Es copia de `internal/platform/currency` en el backend, que es quien manda:
 * valida la moneda del perfil y el `?currency=` de los endpoints. Aquí vive
 * solo para poder pintar un selector sin pedirla por red. Si allá se agrega
 * una moneda, se agrega aquí.
 */
export const SUPPORTED_CURRENCIES = [
	'USD',
	'COP',
	'EUR',
	'GBP',
	'CHF',
	'JPY',
	'CAD',
	'AUD',
	'CNY',
	'MXN',
	'BRL'
] as const;

export type SupportedCurrency = (typeof SUPPORTED_CURRENCIES)[number];

/** La que se usa cuando no hay preferencia utilizable: el eje de las tasas. */
export const FALLBACK_CURRENCY: SupportedCurrency = 'USD';

export function isSupportedCurrency(code: string | null | undefined): code is SupportedCurrency {
	return !!code && (SUPPORTED_CURRENCIES as readonly string[]).includes(code);
}

/**
 * Resuelve la moneda en la que mostrar los totales.
 *
 * El orden es: lo que pide la URL, la preferencia de la cuenta y, si ninguna
 * sirve, el dólar. Una preferencia guardada antes de que el backend validara
 * el campo —o de una lista que después se recortó— cae al fallback en vez de
 * romper la página.
 */
export function resolveDisplayCurrency(
	requested?: string | null,
	preferred?: string | null
): SupportedCurrency {
	const asked = requested?.trim().toUpperCase();
	if (isSupportedCurrency(asked)) return asked;

	const account = preferred?.trim().toUpperCase();
	if (isSupportedCurrency(account)) return account;

	return FALLBACK_CURRENCY;
}

/** Fila con importes cuya moneda el backend informa por separado. */
export interface RowInCurrency {
	/** Moneda real de los importes. `baseCurrency` es el respaldo histórico. */
	displayCurrency?: string;
	baseCurrency?: string;
}

/**
 * Separa las filas que sí están en `currency` de las que no.
 *
 * Cuando se pide una moneda de visualización el backend convierte lo que puede
 * y devuelve el resto en la suya, marcado (`fxConverted: false`). Sumar los dos
 * grupos da un número que no está en ninguna moneda, así que quien agregue
 * totales suma solo `converted` y cuenta los otros aparte para poder decirlo.
 */
export function partitionByCurrency<T extends RowInCurrency>(
	rows: T[],
	currency: string
): { converted: T[]; unconverted: T[] } {
	const converted: T[] = [];
	const unconverted: T[] = [];

	for (const row of rows) {
		const rowCurrency = row.displayCurrency || row.baseCurrency || currency;
		(rowCurrency === currency ? converted : unconverted).push(row);
	}

	return { converted, unconverted };
}
