import { page } from 'vitest/browser';
import { describe, it, expect } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Stat from './stat.svelte';

describe('stat.svelte', () => {
	it('renders the label and value', async () => {
		render(Stat, { label: 'Patrimonio', value: '$12,500' });

		await expect.element(page.getByText('Patrimonio')).toBeInTheDocument();
		await expect.element(page.getByText('$12,500')).toBeInTheDocument();
	});

	it('renders a numeric value and an inline unit', async () => {
		render(Stat, { label: 'Rendimiento', value: 15.2, unit: 'anual' });

		await expect.element(page.getByText('15.2')).toBeInTheDocument();
		await expect.element(page.getByText('anual')).toBeInTheDocument();
	});

	it('applies the tone modifier to the value', async () => {
		render(Stat, { label: 'Cambio', value: '+3.4%', tone: 'positive' });

		const value = document.querySelector('.stat-value') as HTMLElement;
		expect(value.classList.contains('stat-positive')).toBe(true);
	});

	it('right-aligns when requested', async () => {
		render(Stat, { label: 'Total', value: '10', align: 'right' });

		expect(document.querySelector('.stat')?.classList.contains('stat-right')).toBe(true);
	});
});
