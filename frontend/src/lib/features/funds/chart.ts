/**
 * Geometría de la gráfica de un fondo: el valor de unidad contra el tiempo.
 *
 * El eje horizontal es de tiempo, no de índices: las marcas de un fondo llegan
 * cuando llega el extracto —una semana seguida de días, luego un mes sin nada—
 * y repartirlas a partes iguales dibujaría un mes callado igual que un día.
 */

export interface FundChartPoint {
	date: string;
	value: number;
}

export const FUND_CHART = {
	width: 640,
	height: 220,
	padL: 64,
	padR: 16,
	padT: 14,
	padB: 30
} as const;

const DAY_MS = 86_400_000;

const time = (iso: string) => Date.parse(`${iso.slice(0, 10)}T00:00:00Z`);

/**
 * Marcas redondas para el eje vertical: pasos de 1, 2, 2,5 o 5 por una potencia
 * de diez, de modo que cubran el rango con unas `count` marcas. Una serie plana
 * se abre un 1 % a cada lado para que la línea no quede pegada a un borde.
 */
export function niceTicks(min: number, max: number, count = 4): number[] {
	let lo = min;
	let hi = max;

	if (hi - lo <= Math.abs(hi) * 1e-9) {
		const pad = Math.abs(hi) * 0.01 || 1;
		lo -= pad;
		hi += pad;
	}

	const raw = (hi - lo) / count;
	const power = 10 ** Math.floor(Math.log10(raw));
	const step = ([1, 2, 2.5, 5, 10].find((m) => m * power >= raw) ?? 10) * power;

	const first = Math.floor(lo / step) * step;
	const last = Math.ceil(hi / step) * step;
	const ticks: number[] = [];

	for (let v = first; v <= last + step / 2; v += step) {
		ticks.push(Number(v.toPrecision(12)));
	}

	return ticks;
}

export interface FundChartGeometry {
	x: (iso: string) => number;
	y: (value: number) => number;
	yTicks: number[];
	/** Hasta cuatro fechas repartidas por igual en el tiempo, para el eje. */
	xTicks: string[];
}

/** Escalas y marcas de la gráfica para unos puntos ya ordenados por fecha. */
export function fundChartGeometry(points: FundChartPoint[]): FundChartGeometry {
	const { width, height, padL, padR, padT, padB } = FUND_CHART;
	const plotW = width - padL - padR;
	const plotH = height - padT - padB;

	const values = points.map((p) => p.value);
	const yTicks = niceTicks(Math.min(...values), Math.max(...values));
	const lo = yTicks[0];
	const hi = yTicks[yTicks.length - 1];

	const t0 = time(points[0].date);
	const t1 = time(points[points.length - 1].date);
	const span = t1 - t0;

	const x = (iso: string) =>
		span === 0 ? padL + plotW / 2 : padL + ((time(iso) - t0) / span) * plotW;
	const y = (value: number) => padT + (1 - (value - lo) / (hi - lo)) * plotH;

	const days = Math.round(span / DAY_MS);
	const steps = span === 0 ? 0 : Math.min(3, days);
	const xTicks = Array.from({ length: steps + 1 }, (_, k) =>
		new Date(t0 + (steps === 0 ? 0 : Math.round((k * days) / steps) * DAY_MS))
			.toISOString()
			.slice(0, 10)
	);

	return { x, y, yTicks, xTicks };
}

/** El punto más cercano a una x del lienzo: el que fija el cursor. */
export function nearestPoint(
	points: FundChartPoint[],
	geometry: FundChartGeometry,
	x: number
): number {
	let best = 0;

	points.forEach((p, i) => {
		if (Math.abs(geometry.x(p.date) - x) < Math.abs(geometry.x(points[best].date) - x)) best = i;
	});

	return best;
}
