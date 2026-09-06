import { describe, it, expect } from 'vitest';
import { formatBytes, formatGeneratedAt, groupSections } from './guide';
import { manual } from './manual-meta';

describe('formatBytes', () => {
	it('escala a la unidad que toca', () => {
		expect(formatBytes(512)).toBe('512 B');
		expect(formatBytes(2048)).toBe('2 KB');
		expect(formatBytes(7_340_032)).toBe('7,0 MB');
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

describe('groupSections', () => {
	const section = (number: number) => ({ number, title: `Capítulo ${number}` });

	it('reparte los capítulos en los bloques del índice', () => {
		const groups = groupSections([1, 5, 13, 17].map(section));

		expect(groups.map((g) => g.label)).toEqual([
			'Para empezar',
			'El día a día',
			'Tu cuenta',
			'Si te atascas'
		]);
		expect(groups.map((g) => g.sections.map((s) => s.number))).toEqual([[1], [5], [13], [17]]);
	});

	it('no pierde ningún capítulo del manual', () => {
		const groups = groupSections(manual.sections);
		const numbers = groups.flatMap((g) => g.sections.map((s) => s.number));

		expect(numbers).toEqual(manual.sections.map((s) => s.number));
	});

	/* Un capítulo nuevo se añade al final del manual, y tiene que caer con los
	   de su alrededor sin que haya que tocar la lista de bloques. */
	it('mete un capítulo nuevo en el último bloque', () => {
		const groups = groupSections([...manual.sections, section(20)]);

		expect(groups.at(-1)?.sections.at(-1)?.number).toBe(20);
	});

	it('no pinta bloques vacíos', () => {
		const groups = groupSections([section(1), section(2)]);

		expect(groups).toHaveLength(1);
		expect(groups[0].label).toBe('Para empezar');
	});
});
