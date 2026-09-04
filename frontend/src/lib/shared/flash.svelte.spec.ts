import { describe, it, expect, vi, afterEach } from 'vitest';
import { flushSync } from 'svelte';
import { flash } from './flash.svelte';

describe('flash — acuse temporal', () => {
	afterEach(() => vi.useRealTimers());

	it('muestra el mensaje y lo retira pasado el plazo', () => {
		vi.useFakeTimers();

		const destroy = $effect.root(() => {
			const acuse = flash(1000);
			expect(acuse.text).toBeNull();

			acuse.show('Listo.');
			expect(acuse.text).toBe('Listo.');

			vi.advanceTimersByTime(1000);
			expect(acuse.text).toBeNull();
		});

		destroy();
	});

	it('reinicia la cuenta atrás en vez de compartir reloj entre dos acuses', () => {
		vi.useFakeTimers();

		const destroy = $effect.root(() => {
			const acuse = flash(1000);
			acuse.show('Primero.');
			vi.advanceTimersByTime(900);

			acuse.show('Segundo.');
			// Con un solo temporizador compartido, el del primer acuse vencía
			// aquí y borraba el segundo a los 100 ms de aparecer.
			vi.advanceTimersByTime(200);
			expect(acuse.text).toBe('Segundo.');

			vi.advanceTimersByTime(800);
			expect(acuse.text).toBeNull();
		});

		destroy();
	});

	it('cancela el temporizador pendiente al desmontar', () => {
		vi.useFakeTimers();

		const destroy = $effect.root(() => {
			flash(1000).show('Listo.');
		});
		flushSync();

		expect(vi.getTimerCount()).toBe(1);

		destroy();

		expect(vi.getTimerCount()).toBe(0);
	});
});
