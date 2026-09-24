/**
 * Envíos optimistas: la pantalla cambia al pulsar, no cuando el servidor acaba.
 *
 * Las pantallas pesadas —la ficha de un activo, el efectivo— guardaban en tres
 * tiempos que el usuario esperaba enteros con el diálogo abierto: la action, el
 * `update()` de `use:enhance`, que vuelve a correr todos los `load` de la
 * página y del layout (cotizaciones incluidas), y a veces un `invalidateAll()`
 * o un `goto()` detrás que los corría otra vez. Guardar una nota tardaba lo
 * mismo que abrir la página dos veces.
 *
 * Ahora el diálogo se cierra y la fila aparece, cambia o se va en el momento;
 * la página se refresca una sola vez, de fondo, y si el servidor rechaza el
 * envío se deshace lo pintado y el formulario vuelve con lo escrito y el motivo.
 */
import { applyAction } from '$app/forms';
import { invalidateAll } from '$app/navigation';
import { untrack } from 'svelte';
import type { ActionResult, SubmitFunction } from '@sveltejs/kit';

let inFlight = $state(0);
let orphanError = $state<string | null>(null);
let orphanTimer: ReturnType<typeof setTimeout> | undefined;

/**
 * El estado de los envíos optimistas, para la cabecera del panel: si alguno
 * sigue sin confirmar —lo que se ve puede no ser aún lo guardado— y el rechazo
 * que no tiene dónde enseñarse.
 */
export const syncing = {
	get active(): boolean {
		return inFlight > 0;
	},

	get error(): string | null {
		return orphanError;
	},

	/**
	 * Un rechazo que llega cuando su diálogo ya se usa para otra cosa: se abrió
	 * otro movimiento mientras el primero se guardaba. Enseñarlo en el diálogo
	 * lo pondría junto a datos que no son los suyos; perderlo sería peor.
	 */
	report(message: string): void {
		clearTimeout(orphanTimer);
		orphanError = message;
		orphanTimer = setTimeout(() => (orphanError = null), 8000);
	},

	dismiss(): void {
		clearTimeout(orphanTimer);
		orphanError = null;
	}
};

/**
 * El motivo por el que un envío no se hizo, o `null` si salió bien.
 *
 * Las actions de este panel rechazan de dos maneras —`fail()` y un `200` con
 * `{ success: false, error }`—, y las dos tienen que deshacer lo pintado.
 */
export function submitError(result: ActionResult, fallback: string): string | null {
	if (result.type === 'error') return fallback;
	if (result.type === 'redirect') return null;

	const data = result.data as { success?: unknown; error?: unknown } | undefined;
	const message = typeof data?.error === 'string' && data.error ? data.error : null;

	if (result.type === 'failure') return message ?? fallback;
	if (data?.success === false || message) return message ?? fallback;

	return null;
}

interface OptimisticSubmit {
	/**
	 * Lo que se pinta al enviar, antes de la respuesta. Devuelve cómo quitarlo:
	 * se llama al fallar el envío y también cuando la página ya trae lo real.
	 */
	apply: (formData: FormData) => (() => void) | void;
	/** El servidor lo rechazó: lo pintado ya se deshizo. */
	onError?: (message: string) => void;
	/** El servidor lo aceptó, antes del refresco de fondo. */
	onSuccess?: (data: Record<string, unknown> | undefined) => void;
	/**
	 * El servidor lo aceptó y la página ya trae lo real: para lo que compara
	 * cómo estaba antes con cómo quedó.
	 */
	onRefreshed?: (data: Record<string, unknown> | undefined) => void;
	/** Lo que se dice si el rechazo no trae motivo. */
	fallbackError: string;
	/**
	 * Pasar el resultado al `form` de la página, como haría `update()` pero sin
	 * su recarga. Para las pantallas cuyos avisos ya se leen de `form` —ajustes,
	 * notificaciones—: siguen funcionando igual y solo cambia cuándo se pinta.
	 */
	syncForm?: boolean;
}

/**
 * `SubmitFunction` para `use:enhance` que no espera al servidor para cambiar
 * la pantalla. No hace el `update()` de SvelteKit: el formulario no se vacía
 * —si falla, vuelve con lo escrito— y el `form` de la página no se toca, así
 * que quien lo usa lee el resultado de sus callbacks, no de `form`.
 */
export function optimisticSubmit(options: OptimisticSubmit): SubmitFunction {
	return ({ formData }) => {
		const undo = options.apply(formData);
		inFlight++;

		return async ({ result }) => {
			try {
				if (result.type === 'redirect') {
					undo?.();
					await applyAction(result);
					return;
				}

				// Un `error` no se aplica: pintaría la página de error por un envío.
				if (options.syncForm && result.type !== 'error') await applyAction(result);

				const error = submitError(result, options.fallbackError);
				if (error !== null) {
					undo?.();
					// Sin quien lo enseñe —un fallo que no llegó a `form`—, lo dice la
					// cabecera: la fila volvió a su sitio y hay que saber por qué.
					if (options.onError) options.onError(error);
					else if (result.type === 'error') syncing.report(error);
					return;
				}

				const data = result.type === 'success' ? result.data : undefined;
				options.onSuccess?.(data);
				// Lo pintado se queda hasta que llega lo real: quitarlo antes haría
				// parpadear la fila de vuelta a como estaba.
				await invalidateAll();
				undo?.();
				options.onRefreshed?.(data);
			} finally {
				inFlight--;
			}
		};
	};
}

/**
 * Un diálogo que es dueño de su formulario —los del efectivo— y lo envía de
 * forma optimista: se oculta al pulsar sin desmontarse, se cierra cuando el
 * servidor confirma y reaparece con lo escrito y el motivo si lo rechaza.
 *
 * `current` es lo que tiene abierto el diálogo (su `target`). Cambiarlo lo
 * devuelve visible y en blanco, y un envío anterior que responda después ya no
 * cierra ni pinta errores sobre el nuevo: su rechazo va a `syncing.report`.
 * Se crea durante la inicialización del componente, porque vigila `current`.
 */
export class OptimisticDialog {
	hidden = $state(false);
	submitting = $state(false);
	error = $state('');
	#current: () => unknown;

	constructor(current: () => unknown) {
		this.#current = current;
		$effect.pre(() => {
			current();
			untrack(() => this.reset());
		});
	}

	/** Visible, sin error y sin envío en curso. */
	reset(): void {
		this.hidden = false;
		this.submitting = false;
		this.error = '';
	}

	submit(options: {
		/** Una función si depende de lo que se envía (el modo del formulario). */
		fallbackError: string | (() => string);
		/** Lo que se pinta fuera del diálogo al enviar; devuelve cómo quitarlo. */
		apply?: (formData: FormData) => (() => void) | void;
		/** El servidor confirmó: cerrar el diálogo. */
		onDone: () => void;
		/**
		 * Lo guardado ya está en la página. Llega aunque el diálogo ya tenga
		 * otra cosa abierta: lo que se hace aquí es de la página, no de él.
		 */
		onSaved?: (data: Record<string, unknown> | undefined) => void;
	}): SubmitFunction {
		return (input) => {
			const sent = this.#current();
			const isCurrent = () => this.#current() === sent;

			const { fallbackError } = options;

			return optimisticSubmit({
				fallbackError: typeof fallbackError === 'function' ? fallbackError() : fallbackError,
				apply: (formData) => {
					this.submitting = true;
					this.error = '';
					this.hidden = true;
					return options.apply?.(formData);
				},
				onError: (message) => {
					if (!isCurrent()) {
						syncing.report(message);
						return;
					}
					this.submitting = false;
					this.error = message;
					this.hidden = false;
				},
				onSuccess: () => {
					if (!isCurrent()) return;
					this.submitting = false;
					options.onDone();
				},
				onRefreshed: options.onSaved
			})(input);
		};
	}
}

/**
 * Lo pintado sobre una lista del servidor mientras sus envíos no se confirman:
 * filas nuevas arriba, filas cambiadas y filas que ya no están. Cada operación
 * devuelve cómo retirarse, que es lo que espera `optimisticSubmit`.
 */
export class OptimisticList<T extends { id: string }> {
	// Se comparan por id y por versión, no por identidad: `$state` guarda un
	// proxy del objeto, y el que se pasó ya no es `===` al guardado.
	#added = $state<T[]>([]);
	#patched = $state<Record<string, { changes: Partial<T>; version: number }>>({});
	#removed = $state<string[]>([]);
	#version = 0;

	/** La lista del servidor con lo pendiente encima. */
	view(items: T[]): T[] {
		const current = items
			.filter((item) => !this.#removed.includes(item.id))
			.map((item) => {
				const patch = this.#patched[item.id];
				return patch ? { ...item, ...patch.changes } : item;
			});

		return [...this.#added, ...current];
	}

	/** La fila todavía no está confirmada: no se puede editar ni borrar. */
	isPending(id: string): boolean {
		return this.#added.some((item) => item.id === id) || id in this.#patched;
	}

	/** Cuánto cambia el total de la lista por lo pendiente. */
	get delta(): number {
		return this.#added.length - this.#removed.length;
	}

	add(item: T): () => void {
		this.#added = [item, ...this.#added];
		return () => {
			this.#added = this.#added.filter((i) => i.id !== item.id);
		};
	}

	patch(id: string, changes: Partial<T>): () => void {
		const version = ++this.#version;
		this.#patched = { ...this.#patched, [id]: { changes, version } };
		return () => {
			// Una edición posterior de la misma fila ya la sustituyó.
			if (this.#patched[id]?.version !== version) return;
			const rest = { ...this.#patched };
			delete rest[id];
			this.#patched = rest;
		};
	}

	remove(id: string): () => void {
		this.#removed = [...this.#removed, id];
		return () => {
			this.#removed = this.#removed.filter((r) => r !== id);
		};
	}
}

/** Id de una fila que aún no existe en el servidor. */
export function pendingId(): string {
	return `pending-${crypto.randomUUID()}`;
}
