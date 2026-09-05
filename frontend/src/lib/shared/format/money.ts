// Locale each supported currency reads most naturally in (grouping/decimal
// separators, symbol placement). Falls back to 'es-CO' for anything else.
const CURRENCY_LOCALES: Record<string, string> = {
	USD: 'en-US',
	COP: 'es-CO'
};

/**
 * Formats a monetary amount with its currency symbol, using the locale that
 * currency is conventionally displayed in (e.g. "$1,234.50" for USD,
 * "$1.234" for COP — COP has no minor unit in everyday use).
 *
 * `maxDigits` sube el techo de decimales para un precio unitario más pequeño
 * que la unidad mínima de su moneda: el interés diario de una cuenta o el
 * dividendo por acción salían como «$0.00» al lado de su propio total. El
 * mínimo no cambia, así que un importe normal sigue escribiéndose igual y no
 * arrastra ceros.
 */
export function formatCurrency(value: number, currencyCode: string, maxDigits?: number): string {
	const locale = CURRENCY_LOCALES[currencyCode] ?? 'es-CO';
	const digits = currencyCode === 'COP' ? 0 : 2;

	return new Intl.NumberFormat(locale, {
		style: 'currency',
		currency: currencyCode,
		minimumFractionDigits: digits,
		maximumFractionDigits: Math.max(digits, maxDigits ?? digits)
	}).format(value);
}

/**
 * Símbolo de la moneda, para quien compone el importe por su cuenta (ejes
 * abreviados, "$1,2k") y no puede usar `formatCurrency`.
 *
 * Sale de la misma tabla de locales, así que un importe formateado a mano lleva
 * el mismo símbolo que uno formateado entero. Cae al dólar ante un código que
 * no sea ISO de tres letras: `Intl.NumberFormat` lanza con cualquier otra cosa
 * y un gráfico no debería caerse por una moneda mal escrita.
 */
export function currencySymbol(currencyCode: string): string {
	const code = currencyCode.trim().toUpperCase();
	if (!/^[A-Z]{3}$/.test(code)) return '$';

	const locale = CURRENCY_LOCALES[code] ?? 'es-CO';
	const parts = new Intl.NumberFormat(locale, { style: 'currency', currency: code }).formatToParts(
		0
	);

	return parts.find((part) => part.type === 'currency')?.value ?? '$';
}

/**
 * Importe abreviado: «$1,2k», «$89k», «$1,4M».
 *
 * Para donde no cabe la cifra entera —las marcas de un eje, el centro de un
 * donut—, no para leerla con precisión: eso es `formatCurrency`, y suele estar
 * al lado.
 *
 * Entre 1.000 y 10.000 hace falta el decimal: redondeando a miles, una serie de
 * 1.500 a 1.900 pintaba «$2k» en las cinco marcas del eje.
 *
 * No enmascara: quien lo muestre lo pasa por `privacy.money`, igual que a
 * cualquier otro importe.
 */
export function formatCompactCurrency(value: number, currencyCode: string): string {
	const symbol = currencySymbol(currencyCode);
	const abs = Math.abs(value);

	if (abs >= 1_000_000) return `${symbol}${(value / 1_000_000).toFixed(1)}M`;
	if (abs >= 10_000) return `${symbol}${(value / 1_000).toFixed(0)}k`;
	if (abs >= 1_000) return `${symbol}${(value / 1_000).toFixed(1)}k`;

	return `${symbol}${value.toFixed(0)}`;
}
