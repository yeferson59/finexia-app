/**
 * El escritorio de administración: la edad de los datos y lo que está abierto.
 *
 * Separado de `admin.ts` —que son las constantes y los formateadores del
 * dominio— porque responde a otra pregunta. Ahí está cómo se escribe un precio;
 * aquí, cuánto lleva ese precio sin tocarse y si eso es un problema.
 */

import type { Asset, ExchangeRate, InvitationItem, UserItem, WaitlistItem } from '$lib/api/types';

// --- Antigüedad -------------------------------------------------------------

/**
 * «1 activo» / «4 activos», que es lo que evita el «1 activos».
 *
 * Es una copia de la de `features/dashboard`: una feature no importa de otra
 * (docs/FRONTEND_ARCHITECTURE.md), y bajar tres líneas a `lib/shared` para
 * compartirlas cuesta más de lo que ahorra.
 */
export function plural(count: number, one: string, many: string): string {
	return `${count} ${count === 1 ? one : many}`;
}

/**
 * A partir de cuándo un dato del catálogo deja de servir.
 *
 * Una semana no es un número redondo cualquiera: el feed de la TRM corre cada
 * hora y los precios de mercado los sincroniza cada usuario con su clave, así
 * que lo único que llega aquí con siete días encima es lo que nadie mantiene.
 */
export const STALE_AFTER_DAYS = 7;

const DAY_MS = 86_400_000;

/** Días completos desde `iso`. Negativo si la fecha está por venir. */
export function daysSince(iso: string | null | undefined, now: Date = new Date()): number | null {
	if (!iso) return null;
	const then = new Date(iso).getTime();
	if (isNaN(then)) return null;
	return Math.floor((now.getTime() - then) / DAY_MS);
}

/**
 * La edad de un dato en palabras: «hoy», «hace 5 días», «hace 3 semanas».
 *
 * Es lo que sustituye al sello de fecha en las tablas del catálogo. Un
 * «14/08/26 09:12» obliga a restar mentalmente contra el día de hoy en cada
 * fila; la pregunta que se hace un administrador no es qué día se tocó un
 * precio, sino cuánto lleva sin tocarse. La fecha exacta sigue estando, en el
 * `title` de la celda.
 */
export function formatAge(iso: string | null | undefined, now: Date = new Date()): string {
	const days = daysSince(iso, now);
	if (days === null) return 'nunca';
	if (days <= 0) return 'hoy';
	if (days === 1) return 'ayer';
	if (days < 7) return `hace ${days} días`;
	if (days < 30) {
		const weeks = Math.floor(days / 7);
		return weeks === 1 ? 'hace una semana' : `hace ${weeks} semanas`;
	}
	const months = Math.floor(days / 30);
	if (months >= 12) return 'hace más de un año';
	return months === 1 ? 'hace un mes' : `hace ${months} meses`;
}

/** Si un dato lleva sin actualizarse más de la cuenta. Sin fecha cuenta como sí. */
export function isStale(
	iso: string | null | undefined,
	after: number = STALE_AFTER_DAYS,
	now: Date = new Date()
): boolean {
	const days = daysSince(iso, now);
	return days === null || days >= after;
}

/**
 * Cuánto le queda a una invitación: «caduca en 3 días», «caducó ayer».
 *
 * Igual que la edad de un dato, cambia de unidad en cuanto la cuenta de días
 * deja de decir algo: nadie lee «caduca en 116 días», y menos para decidir si
 * hay que reenviarla hoy.
 */
export function formatDeadline(iso: string, now: Date = new Date()): string {
	const days = daysSince(iso, now);
	if (days === null) return 'sin fecha de caducidad';
	if (days <= -30) {
		const months = Math.floor(-days / 30);
		return months === 1 ? 'caduca dentro de un mes' : `caduca dentro de ${months} meses`;
	}
	if (days <= -7) {
		const weeks = Math.floor(-days / 7);
		return weeks === 1 ? 'caduca dentro de una semana' : `caduca dentro de ${weeks} semanas`;
	}
	if (days < -1) return `caduca en ${-days} días`;
	if (days === -1) return 'caduca mañana';
	if (days === 0) return 'caduca hoy';
	if (days === 1) return 'caducó ayer';
	return `caducó ${formatAge(iso, now)}`;
}

// ---------------------------------------------------------------------------
// Cómo está cada cosa, en una frase
// ---------------------------------------------------------------------------

/*
 * Cada bloque de administración abre diciendo qué le pasa a lo que contiene, y
 * no con una etiqueta que repite el nombre de la tabla. Es lo que sustituye a
 * recorrer siete columnas para averiguar si hay alguien esperando o cuántos
 * precios se quedaron viejos. La misma idea que los resúmenes de configuración,
 * con una diferencia: aquí la frase cuenta trabajo pendiente, no estado.
 */

/** Quién pidió acceso y sigue esperando. */
export function describeWaitlist(waitlist: WaitlistItem[], now: Date = new Date()): string {
	if (waitlist.length === 0) return 'Nadie espera invitación ahora mismo.';

	const oldest = waitlist.reduce((a, b) => (a.createdAt <= b.createdAt ? a : b));
	const age = formatAge(oldest.createdAt, now);

	return waitlist.length === 1
		? `Una persona pidió acceso ${age} y sigue esperando.`
		: `${waitlist.length} personas esperan invitación. La primera pidió acceso ${age}.`;
}

/** Invitaciones que siguen en pie y cuándo se cae la primera. */
export function describeInvitations(invitations: InvitationItem[], now: Date = new Date()): string {
	const pending = invitations.filter((i) => i.status === 'pending');
	if (pending.length === 0) return 'Ninguna invitación sigue en pie.';

	const soonest = pending.reduce((a, b) => (a.expiresAt <= b.expiresAt ? a : b));
	const deadline = formatDeadline(soonest.expiresAt, now);

	return pending.length === 1
		? `Una invitación sin aceptar: ${deadline}.`
		: `${pending.length} invitaciones sin aceptar. La primera ${deadline}.`;
}

/**
 * Las cuentas y lo que no está en orden en ellas.
 *
 * `total` viene de la paginación y `users` es solo la página que se está
 * viendo, así que cuando hay más de una la frase dice de dónde salen las
 * cuentas pequeñas en vez de dar a entender que son de todo el sistema.
 */
export function describeUsers(users: UserItem[], total: number, morePages = false): string {
	const head = plural(total, 'cuenta', 'cuentas');
	const banned = users.filter((u) => !!u.bannedAt).length;
	const unverified = users.filter((u) => !u.emailVerified).length;

	const notes: string[] = [];
	if (banned > 0) notes.push(plural(banned, 'baneada', 'baneadas'));
	if (unverified > 0) notes.push(`${unverified} sin verificar el correo`);

	if (notes.length === 0) {
		return morePages
			? `${head}. Ninguna de las que se ven aquí está baneada.`
			: `${head}, todas activas y con el correo verificado.`;
	}

	const list = notes.join(' y ');
	return morePages ? `${head}. En esta página, ${list}.` : `${head}: ${list}.`;
}

/** El catálogo compartido: cuánto hay y cuánto se quedó viejo. */
export function describeAssets(assets: Asset[], now: Date = new Date()): string {
	if (assets.length === 0) return 'El catálogo está vacío.';

	const head = plural(assets.length, 'activo', 'activos');
	const stale = assets.filter((a) => isStale(a.priceUpdatedAt, STALE_AFTER_DAYS, now)).length;
	const contributed = assets.filter((a) => a.isCurated === false).length;

	const notes: string[] = [];
	if (stale > 0) {
		const which =
			stale === 1
				? 'uno lleva'
				: stale === assets.length
					? `los ${stale} llevan`
					: `${stale} llevan`;
		notes.push(`${which} más de una semana con el mismo precio`);
	}
	if (contributed > 0) {
		notes.push(
			contributed === 1
				? 'uno lo aportó un usuario y solo lo ve quien lo aportó'
				: `${contributed} los aportaron usuarios y solo los ven quienes los aportaron`
		);
	}

	if (notes.length === 0) return `${head}, todos con un precio de esta semana.`;
	return `${head}: ${notes.join('; ')}.`;
}

/** Las tasas compartidas, separadas por quién las mantiene. */
export function describeRates(rates: ExchangeRate[]): string {
	if (rates.length === 0) return 'No hay ninguna tasa guardada.';

	const head = plural(rates.length, 'tasa compartida', 'tasas compartidas');
	const auto = rates.filter((r) => r.source !== 'manual').length;
	const manual = rates.length - auto;

	if (auto === 0) return `${head}, y todas las mantienes tú.`;
	if (manual === 0) return `${head}, todas del feed público: cada refresco las reescribe.`;

	const fromFeed =
		auto === 1 ? 'Una la reescribe el feed cada hora' : `${auto} las reescribe el feed cada hora`;
	const byHand = manual === 1 ? 'la otra la mantienes tú' : `las otras ${manual} las mantienes tú`;

	return `${head}. ${fromFeed}; ${byHand}.`;
}

// ---------------------------------------------------------------------------
// Lo que hay que hacer hoy
// ---------------------------------------------------------------------------

/**
 * Una tasa escrita a mano no la refresca nadie, así que envejece de verdad. Se
 * le da un mes —y no la semana del catálogo— porque una paridad entre divisas
 * se mueve mucho más despacio que el precio de una acción.
 */
export const STALE_RATE_AFTER_DAYS = 30;

/** Cuántos días antes de que caduque una invitación empieza a ser trabajo. */
export const EXPIRY_WARNING_DAYS = 3;

/** A qué pantalla lleva una tarea; la ruta la resuelve quien la pinta. */
export type AdminTaskKey = 'waitlist' | 'invitations' | 'prices' | 'rates';

export interface AdminTask {
	key: AdminTaskKey;
	/** Qué hacer, en imperativo. */
	title: string;
	/** El dato que lo justifica: cuánto lleva esperando o cuándo caduca. */
	detail: string;
	/** En sustantivo, para encadenarlo en la frase de portada. */
	summary: string;
	/** Cuántas cosas esperan. */
	count: number;
}

/** Lo que hace falta saber del sistema para saber qué está pendiente. */
export interface DeskState {
	waiting: number;
	oldestWaitAt: string | null;
	expiringInvites: number;
	nextExpiryAt: string | null;
	stalePrices: number;
	oldestPriceAt: string | null;
	staleRates: number;
	oldestRateAt: string | null;
}

/**
 * Las tareas abiertas, y solo esas.
 *
 * La portada de administración enseñaba cuatro cifras y tres atajos, los mismos
 * estuviera el sistema al día o hubiera doce precios de hace un mes. Aquí una
 * fila solo existe mientras haya algo que hacer, así que una lista vacía es una
 * respuesta —no queda nada— y no una pantalla en blanco.
 */
export function buildWorklist(state: DeskState, now: Date = new Date()): AdminTask[] {
	const tasks: AdminTask[] = [];

	if (state.waiting > 0) {
		tasks.push({
			key: 'waitlist',
			count: state.waiting,
			title: state.waiting === 1 ? 'Invitar a quien espera' : 'Invitar a quienes esperan',
			detail: `La solicitud más antigua entró ${formatAge(state.oldestWaitAt, now)}.`,
			summary:
				state.waiting === 1
					? 'una persona esperando invitación'
					: `${state.waiting} personas esperando invitación`
		});
	}

	if (state.expiringInvites > 0) {
		tasks.push({
			key: 'invitations',
			count: state.expiringInvites,
			title: 'Reenviar las invitaciones que se caen',
			detail: state.nextExpiryAt
				? `La primera ${formatDeadline(state.nextExpiryAt, now)}.`
				: 'Nadie las ha aceptado todavía.',
			summary:
				state.expiringInvites === 1
					? 'una invitación a punto de caducar'
					: `${state.expiringInvites} invitaciones a punto de caducar`
		});
	}

	if (state.stalePrices > 0) {
		tasks.push({
			key: 'prices',
			count: state.stalePrices,
			title: 'Actualizar los precios del catálogo',
			detail: state.oldestPriceAt
				? `El más viejo no cambia desde ${formatAge(state.oldestPriceAt, now)}.`
				: 'Al menos uno no tiene precio todavía.',
			summary:
				state.stalePrices === 1
					? 'un precio sin actualizar'
					: `${state.stalePrices} precios sin actualizar`
		});
	}

	if (state.staleRates > 0) {
		tasks.push({
			key: 'rates',
			count: state.staleRates,
			title: 'Revisar las tasas escritas a mano',
			detail: `La más vieja se escribió ${formatAge(state.oldestRateAt, now)}.`,
			summary:
				state.staleRates === 1
					? 'una tasa manual vieja'
					: `${state.staleRates} tasas manuales viejas`
		});
	}

	return tasks;
}

/** El estado del escritorio en una frase, encadenando lo que hay abierto. */
export function describeDesk(tasks: AdminTask[]): string {
	if (tasks.length === 0) return 'Nada pendiente: el acceso y el catálogo están al día.';

	const parts = tasks.map((task) => task.summary);
	const last = parts.pop() as string;

	return parts.length === 0 ? `Hay ${last}.` : `Hay ${parts.join(', ')} y ${last}.`;
}

/**
 * El estado del escritorio a partir de lo que devuelve el backend.
 *
 * Vive aquí y no en el `load` de la ruta porque es criterio de dominio —cuándo
 * un precio está viejo, cuándo una invitación corre peligro—, no orquestación.
 */
export function summarizeDesk(
	input: {
		assets: Asset[];
		rates: ExchangeRate[];
		invitations: InvitationItem[];
		waitlist: WaitlistItem[];
	},
	now: Date = new Date()
): DeskState {
	/**
	 * La fecha más antigua de una lista. Un hueco sin fecha gana a cualquiera
	 * —nunca se tocó—, y se devuelve como `null` para que quien lo pinte lo diga
	 * con palabras en vez de con una fecha que no existe.
	 */
	const earliest = (dates: (string | null)[]): string | null => {
		if (dates.length === 0 || dates.some((date) => !date)) return null;
		return (dates as string[]).reduce((oldest, date) => (date < oldest ? date : oldest));
	};

	const waiting = input.waitlist.filter((w) => w.status === 'pending');

	// Una invitación no es trabajo pendiente hasta que está a punto de caerse:
	// hasta entonces lo normal es que la persona todavía no la haya abierto.
	const pending = input.invitations.filter((i) => i.status === 'pending');
	const expiring = pending.filter(
		(i) => (daysSince(i.expiresAt, now) ?? 0) >= -EXPIRY_WARNING_DAYS
	);

	const stalePrices = input.assets.filter((a) => isStale(a.priceUpdatedAt, STALE_AFTER_DAYS, now));
	const staleRates = input.rates.filter(
		(r) => r.source === 'manual' && isStale(r.rateDate, STALE_RATE_AFTER_DAYS, now)
	);

	return {
		waiting: waiting.length,
		oldestWaitAt: earliest(waiting.map((w) => w.createdAt)),
		expiringInvites: expiring.length,
		nextExpiryAt: earliest(pending.map((i) => i.expiresAt)),
		stalePrices: stalePrices.length,
		oldestPriceAt: earliest(stalePrices.map((a) => a.priceUpdatedAt)),
		staleRates: staleRates.length,
		oldestRateAt: earliest(staleRates.map((r) => r.rateDate))
	};
}
