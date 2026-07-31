import { describe, it, expect } from 'vitest';
import {
	formatDateTime,
	formatPrice,
	formatRate,
	invitationStatusLabel,
	invitationStatusTone,
	summarizeImport
} from './admin';

describe('estado de una invitación', () => {
	it('traduce los estados conocidos y deja pasar los demás', () => {
		expect(invitationStatusLabel('pending')).toBe('Pendiente');
		expect(invitationStatusLabel('accepted')).toBe('Aceptada');
		expect(invitationStatusLabel('desconocido')).toBe('desconocido');
	});

	it('asigna un tono por estado', () => {
		expect(invitationStatusTone('accepted')).toBe('success');
		expect(invitationStatusTone('revoked')).toBe('danger');
		expect(invitationStatusTone('expired')).toBe('neutral');
		expect(invitationStatusTone('pending')).toBe('amber');
	});
});

describe('formatPrice', () => {
	it('formatea con la moneda del precio', () => {
		expect(formatPrice({ value: '191.5', currency: 'USD' })).toBe('$191.50');
	});

	it('cae a la moneda del activo cuando el precio no la trae', () => {
		expect(formatPrice({ value: '4000', currency: 'XXX' }, 'COP')).toBe('$4,000.00');
	});

	it('no rompe con una moneda inválida ni con un valor no numérico', () => {
		expect(formatPrice({ value: '10', currency: 'NOPE' })).toBe('NOPE 10.00');
		expect(formatPrice({ value: 'n/d', currency: 'USD' })).toBe('n/d');
	});

	it('marca la ausencia de precio', () => {
		expect(formatPrice(null)).toBe('—');
	});
});

describe('formatRate', () => {
	it('agrupa los millares y recorta los decimales sobrantes', () => {
		expect(formatRate('4123.456789123')).toBe('4,123.456789');
	});

	it('devuelve el valor tal cual si no es un número', () => {
		expect(formatRate('sin dato')).toBe('sin dato');
	});
});

describe('formatDateTime', () => {
	it('marca la ausencia de fecha', () => {
		expect(formatDateTime(null)).toBe('—');
	});

	it('formatea una fecha ISO', () => {
		expect(formatDateTime('2026-03-14T09:05:00.000Z')).not.toBe('—');
	});
});

describe('summarizeImport', () => {
	it('resuelve los plurales de un import completo', () => {
		expect(summarizeImport({ totalRows: 5, imported: 5, skipped: 0, errors: [] })).toBe(
			'5 de 5 filas importadas.'
		);
	});

	it('cuenta las filas omitidas cuando las hay', () => {
		expect(summarizeImport({ totalRows: 5, imported: 3, skipped: 2, errors: [] })).toBe(
			'3 de 5 filas importadas, 2 omitidas.'
		);
	});

	it('usa el singular con una sola fila', () => {
		expect(summarizeImport({ totalRows: 1, imported: 1, skipped: 0, errors: [] })).toBe(
			'1 de 1 fila importada.'
		);
		expect(summarizeImport({ totalRows: 2, imported: 1, skipped: 1, errors: [] })).toBe(
			'1 de 2 filas importada, 1 omitida.'
		);
	});
});
