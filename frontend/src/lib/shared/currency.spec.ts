import { describe, it, expect } from 'vitest';
import {
	FALLBACK_CURRENCY,
	SUPPORTED_CURRENCIES,
	isSupportedCurrency,
	partitionByCurrency,
	resolveDisplayCurrency
} from './currency';

describe('isSupportedCurrency', () => {
	it('accepts every listed code', () => {
		for (const code of SUPPORTED_CURRENCIES) {
			expect(isSupportedCurrency(code)).toBe(true);
		}
	});

	// ARS existe en ISO 4217 pero ninguna fuente publica su par contra el dólar:
	// la lista es de monedas convertibles, no de monedas válidas.
	it('rejects a real currency with no rate source, and anything unnormalized', () => {
		for (const code of ['ARS', 'usd', ' USD', 'XXX', '', null, undefined]) {
			expect(isSupportedCurrency(code)).toBe(false);
		}
	});
});

describe('resolveDisplayCurrency', () => {
	it('prefers the requested currency over the account one', () => {
		expect(resolveDisplayCurrency('cop', 'EUR')).toBe('COP');
	});

	it('falls back to the account currency when nothing is requested', () => {
		expect(resolveDisplayCurrency(null, 'eur')).toBe('EUR');
		expect(resolveDisplayCurrency('', ' eur ')).toBe('EUR');
	});

	// Una preferencia guardada antes de que el backend validara el campo no
	// puede dejar el panel sin moneda: cae al dólar en vez de romper la página.
	it('ignores unsupported values on either side', () => {
		expect(resolveDisplayCurrency('ARS', 'EUR')).toBe('EUR');
		expect(resolveDisplayCurrency(null, 'ARS')).toBe(FALLBACK_CURRENCY);
		expect(resolveDisplayCurrency(undefined, undefined)).toBe(FALLBACK_CURRENCY);
	});
});

describe('partitionByCurrency', () => {
	const usd = { id: 'a', displayCurrency: 'USD', baseCurrency: 'USD' };
	const convertedToUsd = { id: 'b', displayCurrency: 'USD', baseCurrency: 'EUR' };
	const leftInEur = { id: 'c', displayCurrency: 'EUR', baseCurrency: 'EUR' };

	it('separates rows the backend could not convert', () => {
		const { converted, unconverted } = partitionByCurrency([usd, convertedToUsd, leftInEur], 'USD');

		expect(converted.map((r) => r.id)).toEqual(['a', 'b']);
		expect(unconverted.map((r) => r.id)).toEqual(['c']);
	});

	it('reads the base currency when no display currency came back', () => {
		const { converted, unconverted } = partitionByCurrency(
			[
				{ id: 'a', baseCurrency: 'USD' },
				{ id: 'b', baseCurrency: 'COP' }
			],
			'USD'
		);

		expect(converted.map((r) => r.id)).toEqual(['a']);
		expect(unconverted.map((r) => r.id)).toEqual(['b']);
	});

	// Un backend que no informe la moneda deja los importes tal cual venían, que
	// es lo que la página ya asumía: contarlos fuera vaciaría los totales.
	it('assumes a row with no currency at all is already in the target', () => {
		const { converted, unconverted } = partitionByCurrency([{ id: 'a' }], 'USD');

		expect(converted).toHaveLength(1);
		expect(unconverted).toHaveLength(0);
	});
});
