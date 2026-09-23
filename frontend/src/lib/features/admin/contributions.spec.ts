import { describe, expect, it } from 'vitest';
import {
	contributionStatusLabel,
	contributionStatusTone,
	describeContributions,
	parseContributionStatus,
	paymentMethodLabel,
	type SupportSummary
} from './contributions';

function summary(overrides: Partial<SupportSummary> = {}): SupportSummary {
	return {
		counts: { created: 0, pending: 0, rejected: 0, approved: 0, voided: 0 },
		approvedTotal: 0,
		approvedRecent: 0,
		lastApprovedAt: null,
		...overrides
	};
}

// es-CO separa el símbolo con un espacio duro: «$ 20.000».
const sp = ' ';

describe('describeContributions', () => {
	it('lo dice cuando todavía no hay nada aprobado', () => {
		expect(describeContributions(summary())).toBe('Todavía no hay aportes aprobados.');
	});

	it('suma lo aprobado y separa lo reciente', () => {
		const text = describeContributions(
			summary({
				counts: { created: 9, pending: 0, rejected: 1, approved: 2, voided: 0 },
				approvedTotal: 70_000,
				approvedRecent: 20_000
			})
		);
		expect(text).toBe(
			`$${sp}70.000 en 2 aportes aprobados, $${sp}20.000 de ellos en los últimos 30 días.`
		);
	});

	it('avisa de lo que sigue en proceso, y calla lo sin pagar', () => {
		const text = describeContributions(
			summary({
				counts: { created: 40, pending: 1, rejected: 0, approved: 1, voided: 0 },
				approvedTotal: 10_000
			})
		);
		expect(text).toContain('ninguno en los últimos 30 días');
		expect(text).toContain('1 pago sigue en proceso en Bold.');
		expect(text).not.toContain('40');
	});
});

describe('estados y medios', () => {
	it('nombra cada estado y solo lo aprobado va en verde', () => {
		expect(contributionStatusLabel('created')).toBe('Sin pagar');
		expect(contributionStatusTone('approved')).toBe('success');
		expect(contributionStatusTone('rejected')).not.toBe('danger');
	});

	it('lee el filtro de la URL y descarta lo desconocido', () => {
		expect(parseContributionStatus('approved')).toBe('approved');
		expect(parseContributionStatus('paid')).toBeUndefined();
		expect(parseContributionStatus(null)).toBeUndefined();
	});

	it('nombra los medios de Bold con los dos vocabularios', () => {
		expect(paymentMethodLabel('CARD')).toBe('Tarjeta');
		expect(paymentMethodLabel('CREDIT_CARD')).toBe('Tarjeta');
		expect(paymentMethodLabel('pse')).toBe('PSE');
		expect(paymentMethodLabel('')).toBe('—');
		expect(paymentMethodLabel('CRYPTO')).toBe('CRYPTO');
	});
});
