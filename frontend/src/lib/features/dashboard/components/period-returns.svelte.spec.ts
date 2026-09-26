import { page, userEvent } from 'vitest/browser';
import { afterEach, describe, it, expect } from 'vitest';
import { render } from 'vitest-browser-svelte';
import PeriodReturns from './period-returns.svelte';
import { privacy } from '$lib/shared/privacy.svelte';
import type { TrailingReturn } from '$lib/api/types';

const RETURNS: TrailingReturn[] = [
	{
		period: '1D',
		available: true,
		historyStart: '2026-03-02',
		from: '2026-09-24',
		to: '2026-09-25',
		startValue: '10250.00',
		endValue: '10292.10',
		netFlow: '0.00',
		gain: '42.10',
		returnPct: '0.41'
	},
	{
		period: '1W',
		available: true,
		historyStart: '2026-03-02',
		from: '2026-09-18',
		to: '2026-09-25',
		startValue: '9900.00',
		endValue: '10292.10',
		netFlow: '500.00',
		gain: '-107.90',
		returnPct: '-1.05'
	},
	{ period: '1Y', available: false, historyStart: '2026-03-02' }
];

const reading = () => page.getByRole('tabpanel');

describe('period-returns.svelte', () => {
	// Se asigna el estado en vez de llamar a `toggle()`, que lo guarda en
	// `localStorage`: los archivos de prueba comparten origen, y otro que
	// arrancara mientras tanto nacería con los importes tapados.
	afterEach(() => {
		privacy.hidden = false;
	});

	it('shows every window with its rate on the rule', async () => {
		render(PeriodReturns, { returns: RETURNS, currency: 'USD' });

		await expect.element(page.getByRole('tab', { name: '1 día +0,41%' })).toBeInTheDocument();
		await expect.element(page.getByRole('tab', { name: '7 días -1,05%' })).toBeInTheDocument();
		await expect.element(page.getByRole('tab', { name: '1 año —' })).toBeInTheDocument();
	});

	it('reads the chosen window in money, from the day it starts', async () => {
		render(PeriodReturns, { returns: RETURNS, currency: 'USD' });

		// Sin mes en la serie, abre en la primera ventana con cifra.
		await expect
			.element(page.getByRole('tab', { name: /^1 día/ }))
			.toHaveAttribute('aria-selected', 'true');
		await expect.element(reading()).toMatchTextContent(/^Desde el 24 de .+ ganaste \$42\.10, sin/);

		// La semana perdió dinero: el depósito de 500 no se cuenta como ganancia.
		await page.getByRole('tab', { name: /^7 días/ }).click();
		await expect.element(reading()).toMatchTextContent(/perdiste \$107\.90, sin contar/);
	});

	it('moves along the rule with the arrow keys', async () => {
		render(PeriodReturns, { returns: RETURNS, currency: 'USD' });

		await page.getByRole('tab', { name: /^1 día/ }).click();
		await userEvent.keyboard('{ArrowLeft}');

		const year = page.getByRole('tab', { name: /^1 año/ });
		await expect.element(year).toHaveAttribute('aria-selected', 'true');
		await expect.element(year).toHaveFocus();
		// Una ventana a la que el historial no llega dice desde cuándo lo hay.
		await expect
			.element(reading())
			.toMatchTextContent(/^El historial empieza el 2 de .+ todavía no hay cifra/);
	});

	it('masks the money, not the rates, in hidden mode', async () => {
		privacy.hidden = true;
		render(PeriodReturns, { returns: RETURNS, currency: 'USD' });

		await expect.element(page.getByText('+0,41%')).toBeInTheDocument();
		await expect.element(reading()).not.toMatchTextContent('$42.10');
		await expect.element(reading()).toMatchTextContent(/ganaste ••••••,/);
	});

	it('renders nothing while no window has a figure', async () => {
		render(PeriodReturns, {
			returns: [{ period: 'ALL', available: false, historyStart: '2026-09-25' }],
			currency: 'USD'
		});

		await expect.element(page.getByRole('tablist')).not.toBeInTheDocument();
	});
});
