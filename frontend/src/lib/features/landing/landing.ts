/**
 * Helpers puros de la landing.
 *
 * La cuenta atrás del lanzamiento vivía dentro del `onMount` de su componente,
 * donde no había forma de probarla: aritmética de fechas que se rompe con un
 * milisegundo mal puesto y que nadie nota hasta que el contador va torcido.
 */

/** Fecha de lanzamiento que anuncia la landing (hora local). */
export const LAUNCH_DATE = '2026-10-01T09:00:00';

/** Cuenta atrás ya formateada a dos dígitos, lista para pintar. */
export interface Countdown {
	days: string;
	hours: string;
	mins: string;
	secs: string;
}

function pad(n: number): string {
	return String(n).padStart(2, '0');
}

/**
 * Tiempo que falta entre `now` y `target`. Nunca cuenta hacia atrás: pasada la
 * fecha se queda en ceros en vez de mostrar un negativo.
 */
export function countdownBetween(target: number, now: number): Countdown {
	const diff = Math.max(0, target - now);
	return {
		days: pad(Math.floor(diff / 86400000)),
		hours: pad(Math.floor((diff % 86400000) / 3600000)),
		mins: pad(Math.floor((diff % 3600000) / 60000)),
		secs: pad(Math.floor((diff % 60000) / 1000))
	};
}

/**
 * Días enteros que faltan para `target`, para la línea «Faltan N días» que
 * acompaña al formulario. La cuenta atrás al segundo sobraba: a dos semanas del
 * lanzamiento, lo que se lee es el día.
 */
export function daysUntil(target: number, now: number): number {
	return Number(countdownBetween(target, now).days);
}

/** La frase de la cuenta atrás, con el singular y el día del lanzamiento. */
export function launchCountdownText(days: number): string {
	if (days <= 0) return 'Abrimos el 1 de octubre.';
	return days === 1
		? 'Abrimos el 1 de octubre. Falta 1 día.'
		: `Abrimos el 1 de octubre. Faltan ${days} días.`;
}
