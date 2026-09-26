import { page } from 'vitest/browser';
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

describe('period-returns.svelte', () => {
	// Se asigna el estado en vez de llamar a `toggle()`, que lo guarda en
	// `localStorage`: los archivos de prueba comparten origen, y otro que
	// arrancara mientras tanto nacería con los importes tapados.
	afterEach(() => {
		privacy.hidden = false;
	});

	it('shows each window with its rate and its money gain', async () => {
		render(PeriodReturns, { returns: RETURNS, currency: 'USD' });

		await expect.element(page.getByText('1 día')).toBeInTheDocument();
		await expect.element(page.getByText('+0,41%')).toBeInTheDocument();
		await expect.element(page.getByText('+$42.10')).toBeInTheDocument();

		// La semana perdió dinero: el depósito de 500 no se cuenta como ganancia.
		await expect.element(page.getByText('-1,05%')).toBeInTheDocument();
		await expect.element(page.getByText('−$107.90')).toBeInTheDocument();
	});

	it('leaves a window the history does not reach as a dash and says why', async () => {
		render(PeriodReturns, { returns: RETURNS, currency: 'USD' });

		await expect.element(page.getByText('1 año')).toBeInTheDocument();
		await expect.element(page.getByText('—')).toBeInTheDocument();
		await expect.element(page.getByText(/^El historial empieza el 2 de/)).toBeInTheDocument();
	});

	it('masks the money, not the rates, in hidden mode', async () => {
		privacy.hidden = true;
		render(PeriodReturns, { returns: RETURNS, currency: 'USD' });

		await expect.element(page.getByText('+0,41%')).toBeInTheDocument();
		await expect.element(page.getByText('+$42.10')).not.toBeInTheDocument();
		await expect.element(page.getByText('+••••••')).toBeInTheDocument();
	});

	it('renders nothing while no window has a figure', async () => {
		render(PeriodReturns, {
			returns: [{ period: 'ALL', available: false, historyStart: '2026-09-25' }],
			currency: 'USD'
		});

		await expect.element(page.getByText('Rentabilidad por periodo')).not.toBeInTheDocument();
	});
});
