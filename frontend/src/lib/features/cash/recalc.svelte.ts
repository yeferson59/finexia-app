/**
 * El último recálculo de intereses de la página del efectivo: el aviso que se
 * queda arriba hasta que se cierra, y la fila de su cuenta, que destella un
 * momento para que se vea dónde cambió.
 *
 * Es un solo estado de módulo, como `syncing` en `shared/optimistic`: lo
 * escribe el formulario de la tasa y lo leen las cuentas, sin pasar por la
 * página. Solo cambia en el navegador, al confirmar un recálculo, así que en el
 * servidor siempre está vacío.
 */
import { cashRecalcChange, type CashRecalcChange, type CashRecalcDone } from './interest';

/** Lo que dura el destello de la fila, algo más que su animación. */
const HIGHLIGHT_MS = 2500;

export class CashRecalcFeedback {
	done = $state<CashRecalcDone | null>(null);
	/** La clave de la cuenta cuya fila destella, mientras lo hace. */
	highlight = $state<string | null>(null);
	#timer: ReturnType<typeof setTimeout> | undefined;

	get change(): CashRecalcChange | null {
		return this.done ? cashRecalcChange(this.done) : null;
	}

	show(done: CashRecalcDone): void {
		this.done = done;
		clearTimeout(this.#timer);
		this.highlight = done.key;
		this.#timer = setTimeout(() => (this.highlight = null), HIGHLIGHT_MS);
	}

	dismiss(): void {
		this.done = null;
	}
}

export const cashRecalc = new CashRecalcFeedback();
