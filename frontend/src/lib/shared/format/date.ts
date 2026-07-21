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
