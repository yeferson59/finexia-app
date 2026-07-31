import { describe, it, expect } from 'vitest';
import { countdownBetween } from './landing';

const at = (iso: string) => new Date(iso).getTime();

describe('countdownBetween', () => {
	it('descompone el tiempo restante en días, horas, minutos y segundos', () => {
		expect(countdownBetween(at('2026-10-01T09:00:00Z'), at('2026-09-29T07:58:57Z'))).toEqual({
			days: '02',
			hours: '01',
			mins: '01',
			secs: '03'
		});
	});

	it('rellena a dos dígitos', () => {
		expect(countdownBetween(at('2026-10-01T09:00:09Z'), at('2026-10-01T09:00:00Z'))).toEqual({
			days: '00',
			hours: '00',
			mins: '00',
			secs: '09'
		});
	});

	it('se queda en ceros pasada la fecha, sin contar hacia atrás', () => {
		expect(countdownBetween(at('2026-10-01T09:00:00Z'), at('2026-12-01T00:00:00Z'))).toEqual({
			days: '00',
			hours: '00',
			mins: '00',
			secs: '00'
		});
	});
});
