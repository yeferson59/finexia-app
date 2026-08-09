import { page } from 'vitest/browser';
import { describe, it, expect } from 'vitest';
import { render } from 'vitest-browser-svelte';
import ExchangeRateNote from './exchange-rate-note.svelte';
import type { ExchangeRate } from '$lib/api/types';

function rate(overrides: Partial<ExchangeRate> = {}): ExchangeRate {
	return {
		id: 'r1',
		fromCurrency: 'USD',
		toCurrency: 'COP',
		rate: '3157.43',
		rateDate: '2026-08-09T00:00:00Z',
		source: 'dolarapi',
		createdAt: '2026-08-09T18:01:11Z',
		...overrides
	};
}

describe('exchange-rate-note.svelte', () => {
	it('shows the rate, its source and the day it is from', async () => {
		render(ExchangeRateNote, { rate: rate() });

		await expect.element(page.getByText('1 USD = 3.157,43 COP')).toBeInTheDocument();
		// dolarapi publishes the TRM, which is the name the reader can check
		// against their bank statement.
		await expect.element(page.getByText(/TRM/)).toBeInTheDocument();
		await expect.element(page.getByText(/2026/)).toBeInTheDocument();
	});

	it('names a hand-entered rate as such', async () => {
		render(ExchangeRateNote, { rate: rate({ source: 'manual' }) });

		await expect.element(page.getByText(/manualmente/)).toBeInTheDocument();
	});

	// The rate comes from a third-party feed, so "no rate" is a state the
	// dashboard reaches on an ordinary bad day. Showing nothing is the point:
	// the alternative is a figure nobody can reproduce.
	it('renders nothing when there is no rate', async () => {
		render(ExchangeRateNote, { rate: null });

		await expect.element(page.getByText(/COP/)).not.toBeInTheDocument();
	});

	it.each(['0', '-3157.43', 'n/a'])('renders nothing for the unusable rate %s', async (bad) => {
		render(ExchangeRateNote, { rate: rate({ rate: bad }) });

		await expect.element(page.getByText(/=/)).not.toBeInTheDocument();
	});
});
