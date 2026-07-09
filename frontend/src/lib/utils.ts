/**
 * Utility function to conditionally combine classNames
 */
export function cn(...classes: (string | Record<string, boolean> | undefined | null)[]): string {
	return classes
		.flatMap((cls) => {
			if (!cls) return [];
			if (typeof cls === 'string') return cls;
			return Object.entries(cls)
				.filter(([, value]) => value)
				.map(([key]) => key);
		})
		.filter(Boolean)
		.join(' ');
}

/**
 * Formats a calendar-only date value ("2026-07-07" or a UTC-midnight ISO
 * timestamp like "2026-07-07T00:00:00Z") without letting the browser's
 * local timezone shift it to the previous or next day.
 */
export function formatCalendarDate(
	value: string,
	options: Intl.DateTimeFormatOptions,
	locale = 'es-CO'
): string {
	const [year, month, day] = value.split('T')[0].split('-').map(Number);
	return new Date(Date.UTC(year, month - 1, day)).toLocaleDateString(locale, {
		...options,
		timeZone: 'UTC'
	});
}

/**
 * Returns "today" as "YYYY-MM-DD" in the browser's local timezone, for use
 * as a default value in date-only form fields. Avoids `toISOString()`,
 * which reflects the UTC date and can be off by a day for the user's
 * local calendar date.
 */
export function todayLocalDateString(): string {
	const now = new Date();
	const month = String(now.getMonth() + 1).padStart(2, '0');
	const day = String(now.getDate()).padStart(2, '0');
	return `${now.getFullYear()}-${month}-${day}`;
}

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
