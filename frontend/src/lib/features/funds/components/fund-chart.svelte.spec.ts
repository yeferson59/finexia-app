import { page, userEvent } from 'vitest/browser';
import { describe, it, expect, beforeAll } from 'vitest';
import { render } from 'vitest-browser-svelte';
import FundChart from './fund-chart.svelte';

// The plan's example: the index of a fund followed by balance, July to September.
const points = [
	{ date: '2026-07-01', value: 100 },
	{ date: '2026-07-31', value: 100.8 },
	{ date: '2026-08-31', value: 101.0005305 },
	{ date: '2026-09-30', value: 100.53828551 }
];

const props = {
	points,
	caption: 'Índice de rentabilidad del fondo',
	formatValue: (v: number) => v.toFixed(4),
	formatAxis: (v: number) => v.toFixed(2),
	formatDate: (iso: string) => iso
};

describe('fund-chart.svelte', () => {
	// The app's tokens live in the global stylesheet, which a component test
	// does not load; the chart reads them for its colours.
	beforeAll(() => {
		const style = document.createElement('style');
		style.textContent = `:root { --bg: #08090a; --border: rgba(255,255,255,.07); --border-strong: rgba(255,255,255,.12);
			--text: #eceae5; --text-muted: #9f9992; --text-dim: #807b74; --amber: #d4912a; --amber-light: #e8a535;
			--font-mono: monospace; } body { background: #08090a; }`;
		document.head.append(style);
	});

	it('draws the series and keeps every point in a table for screen readers', async () => {
		render(FundChart, props);

		expect(
			document.querySelector('polyline.line')?.getAttribute('points')?.split(' ')
		).toHaveLength(4);
		await expect
			.element(page.getByRole('table', { name: 'Índice de rentabilidad del fondo' }))
			.toBeInTheDocument();
		expect(document.querySelectorAll('tbody tr')).toHaveLength(4);
	});

	it('walks the points from the keyboard and shows the one it is on', async () => {
		render(FundChart, props);

		const slider = page.getByRole('slider', { name: 'Recorrer la gráfica del fondo' });
		await slider.click();
		await userEvent.keyboard('{End}');

		await expect.element(slider).toHaveAttribute('aria-valuetext', '2026-09-30: 100.5383');
		await expect.element(page.getByText('100.5383').first()).toBeVisible();

		await userEvent.keyboard('{ArrowLeft}');
		await expect.element(slider).toHaveAttribute('aria-valuetext', '2026-08-31: 101.0005');

		expect(document.querySelector('line.cursor')).not.toBeNull();
	});
});
