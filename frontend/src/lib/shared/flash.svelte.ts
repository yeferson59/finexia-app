/**
 * Acuse temporal: un mensaje que se retira solo pasado un rato.
 *
 * Nació repetido tres veces en `routes/dashboard` —el alta de activos, la de
 * tasas y el guardado de un portafolio— y las tres copias tenían el mismo par
 * de fallos: el temporizador no se cancelaba al desmontar, así que disparaba
 * contra un componente que ya no existía, y dos acuses seguidos compartían
 * reloj, con lo que el primero borraba el segundo antes de tiempo.
 *
 * `show()` se llama desde el callback de `use:enhance`, que es donde de verdad
 * ocurre el envío: el acuse no se deduce del `form` de la página, que es común
 * a todas sus actions y no distingue de quién es el resultado.
 */
export function flash(ms = 4000) {
	let text = $state<string | null>(null);
	let timer: ReturnType<typeof setTimeout> | undefined;

	// El temporizador es un recurso de fuera de Svelte: hay que soltarlo al
	// desmontar. El efecto no tiene dependencias, solo su limpieza.
	$effect(() => () => clearTimeout(timer));

	return {
		get text(): string | null {
			return text;
		},

		/** Muestra `message` y reinicia la cuenta atrás si ya había otro acuse. */
		show(message: string): void {
			clearTimeout(timer);
			text = message;
			timer = setTimeout(() => (text = null), ms);
		},

		/** Retira el acuse ahora, sin esperar. */
		clear(): void {
			clearTimeout(timer);
			text = null;
		}
	};
}
