import { page } from 'vitest/browser';
import { describe, it, expect } from 'vitest';
import { render } from 'vitest-browser-svelte';
import EntryTotal from './portfolio-entry-total.svelte';

const formatCurrency = (value: number, code: string) => `${code} ${value.toFixed(2)}`;

const base = {
	units: 0.0241,
	unitPrice: 606.6,
	currency: 'EUR',
	costCurrency: 'EUR',
	rate: 1,
	converted: false,
	formatCurrency
};

describe('portfolio-entry-total.svelte', () => {
	it('enseña la cuenta de la que sale el total', async () => {
		render(EntryTotal, { ...base, units: 5, unitPrice: 100, currency: 'USD', costCurrency: 'USD' });

		await expect.element(page.getByText('USD 500.00')).toBeInTheDocument();
		await expect.element(page.getByText('5 × USD 100.00')).toBeInTheDocument();
	});

	/*
	 * 0.0241 × 606.60 × 1.0638 = 15.55, que es el importe que el bróker debitó.
	 * Es el número con el que el usuario contrasta lo que escribió antes de
	 * guardar, así que es el que la prueba fija. Vivía en los campos de compra.
	 */
	it('desglosa la conversión cuando la cuenta liquidó en otra moneda', async () => {
		render(EntryTotal, { ...base, costCurrency: 'USD', rate: 1.0638, converted: true });

		await expect.element(page.getByText('USD 15.55')).toBeInTheDocument();
		await expect
			.element(page.getByText('EUR 14.62 al cambio de 1,0638', { exact: false }))
			.toBeInTheDocument();
	});

	it('no enseña un total mientras falten la cantidad o el precio', async () => {
		render(EntryTotal, { ...base, units: 0, unitPrice: 0 });

		await expect.element(page.getByText('—')).toBeInTheDocument();
		await expect
			.element(page.getByText('Escribe la cantidad y el precio', { exact: false }))
			.toBeInTheDocument();
	});

	/* Con conversión declarada y sin tasa el total daría cero: un número
	   plausible y falso justo donde más se mira. */
	it('no enseña un total de cero cuando falta la tasa', async () => {
		render(EntryTotal, { ...base, costCurrency: 'USD', rate: 0, converted: true });

		await expect.element(page.getByText('—')).toBeInTheDocument();
		await expect.element(page.getByText('Falta la tasa de la operación.')).toBeInTheDocument();
	});
});
