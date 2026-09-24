/**
 * Helpers puros de la rentabilidad de un fondo y de la tabla que se pega desde
 * un extracto.
 *
 * La rentabilidad la calcula el backend (`GET /portfolios/funds/:id/performance`);
 * aquí solo se ordena y se nombra. La tabla pegada sí se lee aquí: cada entidad
 * exporta fechas y cifras a su manera, y el backend recibe ya días y números.
 */
import type { FundPeriod } from '$lib/api/types';

export type { FundPerformance, FundPeriod } from '$lib/api/types';

/** Cómo se llama cada periodo en pantalla, en el orden en que se enseñan. */
export const FUND_PERIOD_LABELS: Record<FundPeriod['key'], string> = {
	'30d': '30 días',
	'90d': '90 días',
	'180d': '180 días',
	'365d': '1 año',
	ytd: 'Año corrido',
	inception: 'Desde el inicio'
};

/** Un periodo listo para la tabla: cifras como números, o `null` sin historial. */
export interface FundPeriodRow {
	key: FundPeriod['key'];
	label: string;
	pct: number | null;
	ea: number | null;
	days: number | null;
	from: string | null;
}

const toNumber = (v: string | null) => (v === null ? null : Number(v));

/** Los periodos en el orden de la tabla, con sus cifras como números. */
export function fundPeriodRows(periods: FundPeriod[]): FundPeriodRow[] {
	const byKey = new Map(periods.map((p) => [p.key, p]));

	return (Object.keys(FUND_PERIOD_LABELS) as FundPeriod['key'][]).flatMap((key) => {
		const p = byKey.get(key);
		if (!p) return [];

		return [
			{
				key,
				label: FUND_PERIOD_LABELS[key],
				pct: toNumber(p.pct),
				ea: toNumber(p.eaPct),
				days: p.days,
				from: p.from
			}
		];
	});
}

/**
 * La cifra que resume un fondo en su tarjeta: la de 30 días, como la publica la
 * entidad, o la de desde el inicio cuando aún no tiene un mes de historia.
 * `null` si no hay nada que medir.
 */
export function headlinePeriod(periods: FundPeriod[]): FundPeriodRow | null {
	const rows = fundPeriodRows(periods);

	return (
		rows.find((r) => r.key === '30d' && r.pct !== null) ??
		rows.find((r) => r.key === 'inception' && r.pct !== null) ??
		null
	);
}

// ---------------------------------------------------------------------------
// La tabla pegada
// ---------------------------------------------------------------------------

/** Una fila leída: el día (`YYYY-MM-DD`) y la cifra. */
export interface PastedMark {
	date: string;
	value: number;
}

export interface PastedTable {
	rows: PastedMark[];
	/** Las líneas que no se pudieron leer, numeradas desde 1. */
	errors: { line: number; text: string }[];
}

/**
 * Una cifra como la escribe un extracto: «12.431,22», «12,431.22», «12431.22»,
 * «$ 12.431,22». Con los dos separadores, el último es el decimal. Con uno solo,
 * repetido es de miles («1.234.567») y suelto es el decimal («1,008», «100.8»):
 * un valor de unidad lleva decimales mucho más a menudo que miles sin ellos.
 */
export function parseStatementNumber(raw: string): number | null {
	const clean = raw.replace(/[^\d.,-]/g, '');
	if (!/\d/.test(clean)) return null;

	const dot = clean.lastIndexOf('.');
	const comma = clean.lastIndexOf(',');
	let normalized: string;

	if (dot >= 0 && comma >= 0) {
		const decimal = dot > comma ? '.' : ',';
		const thousands = decimal === '.' ? ',' : '.';
		normalized = clean.split(thousands).join('').replace(decimal, '.');
	} else {
		const sep = dot >= 0 ? '.' : comma >= 0 ? ',' : '';
		const parts = sep ? clean.split(sep) : [clean];
		normalized = parts.length > 2 ? parts.join('') : parts.join('.');
	}

	const value = Number(normalized);

	return Number.isFinite(value) ? value : null;
}

/**
 * Un día como lo escribe un extracto: «2026-09-30», «30/09/2026» o «30-09-2026».
 * Día antes que mes, como se escribe en Colombia. Un día que no existe —el 31 de
 * septiembre— no se lee.
 */
export function parseStatementDate(raw: string): string | null {
	const s = raw.trim();
	let y: number, m: number, d: number;

	let match = /^(\d{4})-(\d{1,2})-(\d{1,2})$/.exec(s);
	if (match) {
		[y, m, d] = [Number(match[1]), Number(match[2]), Number(match[3])];
	} else {
		match = /^(\d{1,2})[/-](\d{1,2})[/-](\d{4})$/.exec(s);
		if (!match) return null;
		[d, m, y] = [Number(match[1]), Number(match[2]), Number(match[3])];
	}

	const date = new Date(Date.UTC(y, m - 1, d));
	if (date.getUTCFullYear() !== y || date.getUTCMonth() !== m - 1 || date.getUTCDate() !== d) {
		return null;
	}

	return date.toISOString().slice(0, 10);
}

/**
 * Lee la tabla que se pega desde un extracto o una hoja de cálculo: una línea
 * por día, la fecha en la primera columna y la cifra en la última. Las columnas
 * se separan con tabulador, punto y coma o espacios.
 *
 * Una primera línea que no se lee se toma por el encabezado y se salta en
 * silencio; cualquier otra se reporta. Un día repetido se queda con la última
 * cifra, como pasaría al guardarlas una por una.
 */
export function parseMarksTable(text: string): PastedTable {
	const byDay = new Map<string, number>();
	const errors: PastedTable['errors'] = [];

	const lines = text.split(/\r?\n/);
	let first = true;

	lines.forEach((raw, i) => {
		const line = raw.trim();
		if (!line) return;

		const cells = line.split(/\t|;|\s{2,}|\s(?=[$\d-])/).filter((c) => c.trim() !== '');
		const date = cells.length >= 2 ? parseStatementDate(cells[0]) : null;
		const value = date ? parseStatementNumber(cells[cells.length - 1]) : null;

		if (date && value !== null && value > 0) {
			byDay.set(date, value);
		} else if (!first) {
			errors.push({ line: i + 1, text: line });
		}

		first = false;
	});

	const rows = [...byDay.entries()]
		.map(([date, value]) => ({ date, value }))
		.sort((a, b) => a.date.localeCompare(b.date));

	return { rows, errors };
}
