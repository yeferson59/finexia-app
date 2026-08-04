import { describe, it, expect } from 'vitest';
import { formatBytes, formatGeneratedAt } from './guide';
import { manual } from './manual-meta';

describe('formatBytes', () => {
	it('escala a la unidad que toca', () => {
		expect(formatBytes(512)).toBe('512 B');
		expect(formatBytes(2048)).toBe('2 KB');
		expect(formatBytes(7_340_032)).toBe('7.0 MB');
	});

	it('no inventa un tamaño cuando no lo hay', () => {
		expect(formatBytes(0)).toBe('—');
		expect(formatBytes(Number.NaN)).toBe('—');
	});
});

describe('formatGeneratedAt', () => {
	it('formatea la fecha en largo', () => {
		expect(formatGeneratedAt('2026-08-04T10:00:00.000Z')).toMatch(/2026/);
	});

	it('no revienta con una fecha inválida', () => {
		expect(formatGeneratedAt('no es una fecha')).toBe('—');
	});
});

describe('manual-meta', () => {
	it('describe el PDF que sirve la aplicación', () => {
		expect(manual.bytes).toBeGreaterThan(0);
		expect(manual.version).toMatch(/^\d+\.\d+$/);
		expect(manual.sourceHash).toMatch(/^[a-f0-9]{64}$/);
	});

	it('trae el índice de secciones numerado y sin huecos', () => {
		expect(manual.sections.length).toBeGreaterThan(10);
		manual.sections.forEach((section, i) => {
			expect(section.number).toBe(i + 1);
			expect(section.title).not.toBe('');
		});
	});
});
