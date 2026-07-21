import { page } from 'vitest/browser';
import { describe, it, expect } from 'vitest';
import { render } from 'vitest-browser-svelte';
import ProgressBar from './progress-bar.svelte';

describe('progress-bar.svelte', () => {
	it('exposes the value through progressbar aria attributes', async () => {
		render(ProgressBar, { value: 42, ariaLabel: 'Carga' });

		const bar = page.getByRole('progressbar', { name: 'Carga' });
		await expect.element(bar).toHaveAttribute('aria-valuenow', '42');
		await expect.element(bar).toHaveAttribute('aria-valuemin', '0');
		await expect.element(bar).toHaveAttribute('aria-valuemax', '100');
	});

	it('clamps values above 100 down to 100', async () => {
		render(ProgressBar, { value: 150 });
		await expect.element(page.getByRole('progressbar')).toHaveAttribute('aria-valuenow', '100');
	});

	it('clamps negative values up to 0', async () => {
		render(ProgressBar, { value: -20 });
		await expect.element(page.getByRole('progressbar')).toHaveAttribute('aria-valuenow', '0');
	});

	it('renders the optional caption label', async () => {
		render(ProgressBar, { value: 60, label: '60% completado' });
		await expect.element(page.getByText('60% completado')).toBeInTheDocument();
	});
});
