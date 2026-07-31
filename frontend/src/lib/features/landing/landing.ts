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
