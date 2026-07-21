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
 */
export function formatCurrency(value: number, currencyCode: string): string {
	const locale = CURRENCY_LOCALES[currencyCode] ?? 'es-CO';
	return new Intl.NumberFormat(locale, {
		style: 'currency',
		currency: currencyCode,
		minimumFractionDigits: currencyCode === 'COP' ? 0 : 2,
		maximumFractionDigits: currencyCode === 'COP' ? 0 : 2
	}).format(value);
}
