/**
 * Porcentajes con la coma decimal de es-CO.
 *
 * El resto de cifras de la aplicación pasan por `Intl` con esa configuración;
 * los porcentajes se escapaban con `toFixed`, que escribe un punto, y en una
 * misma tarjeta convivían «+12.35%» y «$1.234,50». Aquí se centraliza el
 * formato para que todos digan la coma.
 *
 * La entrada va en puntos porcentuales (`1.2` es 1,2 %), no en fracción: es la
 * unidad en la que llegan las cifras del backend (`gainLossPct`) y la que
 * usan las tarjetas.
 */

/** Un `Intl.NumberFormat` por número de decimales; construirlos no es gratis. */
const formatters = new Map<number, Intl.NumberFormat>();

function formatterFor(digits: number): Intl.NumberFormat {
	let formatter = formatters.get(digits);
	if (!formatter) {
		formatter = new Intl.NumberFormat('es-CO', {
			minimumFractionDigits: digits,
			maximumFractionDigits: digits
		});
		formatters.set(digits, formatter);
	}
	return formatter;
}

/** Si la cifra redondea a cero con esos decimales. */
function roundsToZero(value: number, digits: number): boolean {
	return Math.abs(value) < 0.5 * 10 ** -digits;
}

/**
 * Porcentaje sin signo explícito: `12.345` → «12,3 %».
 *
 * Una cifra que redondea a cero se normaliza a cero para que no salga
 * «-0,0%», que se lee como una pérdida que no existe.
 */
export function formatPercent(value: number, digits = 1): string {
	if (!Number.isFinite(value)) return '—';
	return `${formatterFor(digits).format(roundsToZero(value, digits) ? 0 : value)}%`;
}

/**
 * Como `formatPercent`, con el `+` que pide una cifra de rendimiento.
 *
 * El signo no aparece en la banda que redondea a cero: un «+0,0%» promete una
 * ganancia que la propia cifra desmiente.
 */
export function formatSignedPercent(value: number, digits = 1): string {
	if (!Number.isFinite(value)) return '—';
	const sign = !roundsToZero(value, digits) && value > 0 ? '+' : '';
	return `${sign}${formatPercent(value, digits)}`;
}
