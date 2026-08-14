import { expect, test } from '@playwright/test';
import { TEST_PORTFOLIO_ID, login } from './helpers';

test.describe('portfolio detail', () => {
	test('renders the portfolio with its holdings', async ({ page }) => {
		await login(page);
		await page.goto(`/dashboard/portfolios/${TEST_PORTFOLIO_ID}`);

		await expect(page.getByRole('heading', { level: 1 })).toContainText('Cartera Principal');
		await expect(page.getByText('AAPL').first()).toBeVisible();
		await expect(page.getByText('Moderado').first()).toBeVisible();
	});

	test('asset detail shows the position summary and its transactions', async ({ page }) => {
		await login(page);
		await page.goto(`/dashboard/portfolios/${TEST_PORTFOLIO_ID}/assets/AAPL`);

		await expect(page.getByRole('heading', { level: 1 })).toContainText('AAPL');
		await expect(page.getByRole('heading', { name: 'Resumen de Posición' })).toBeVisible();
		await expect(page.getByRole('heading', { name: 'Información del Activo' })).toBeVisible();

		// La transacción de compra del fixture aparece en el historial.
		await expect(page.getByText('Compra').first()).toBeVisible();
	});

	test('asset detail opens the add-transaction and quick-sell forms', async ({ page }) => {
		await login(page);
		await page.goto(`/dashboard/portfolios/${TEST_PORTFOLIO_ID}/assets/AAPL`);

		await page.getByRole('button', { name: '+ Agregar' }).click();
		await expect(page.getByRole('button', { name: 'Registrar transacción' })).toBeVisible();

		// La venta rápida se abre desde el lote de compra del historial.
		await page.getByRole('button', { name: 'Vender' }).click();
		await expect(page.getByText('Vender desde compra')).toBeVisible();
		await expect(page.getByRole('button', { name: 'Confirmar Venta Total' })).toBeVisible();
	});

	// Borrar una transacción es irreversible y el botón vive en una tabla de
	// filas casi idénticas, así que pasa por una confirmación que dice cuál se
	// va a borrar y qué le ocurre a la posición.
	test('asset detail deletes a transaction after confirming', async ({ page }) => {
		await login(page);
		await page.goto(`/dashboard/portfolios/${TEST_PORTFOLIO_ID}/assets/AAPL`);

		await page.getByRole('button', { name: 'Eliminar transacción' }).first().click();

		const dialog = page.getByRole('dialog', { name: 'Eliminar transacción' });
		await expect(dialog).toBeVisible();
		await expect(dialog).toContainText('la cantidad pasa a 0');

		await dialog.getByRole('button', { name: 'Eliminar' }).click();

		// Al confirmar, el diálogo se cierra y la página se recarga con lo que
		// responda el backend.
		await expect(dialog).toBeHidden();
		await expect(page.getByRole('heading', { name: 'Historial de Transacciones' })).toBeVisible();
	});

	test('adds an entry through the add-asset form', async ({ page }) => {
		await login(page);
		await page.goto(`/dashboard/portfolios/${TEST_PORTFOLIO_ID}/add`);

		// Platform.
		await page.selectOption('select[name="platformId"]', { label: 'Broker Demo' });

		// Asset combobox: focus triggers the suggestion fetch (via /api/assets).
		await page.click('#asset-search');
		const suggestion = page.getByRole('option').filter({ hasText: 'AAPL' });
		await expect(suggestion).toBeVisible();
		// La opción se elige en `mousedown` —antes del blur del input, para que la
		// lista no se cierre antes de tiempo— y eso la saca del DOM en mitad del
		// `click`: Playwright reintenta y se queda esperando a un elemento que ya
		// no existe. Se dispara el mismo evento que escucha el componente.
		await suggestion.dispatchEvent('mousedown');
		await expect(suggestion).toBeHidden();

		// Purchase details; the date picker defaults to today.
		await page.fill('input[name="quantity"]', '5');
		await page.fill('input[name="purchasePrice"]', '100');

		await page.getByRole('button', { name: 'Agregar Activo' }).click();

		// The action redirects back to the portfolio detail on success.
		await page.waitForURL(`**/dashboard/portfolios/${TEST_PORTFOLIO_ID}`);
		await expect(page.getByRole('heading', { level: 1 })).toContainText('Cartera Principal');
	});
});
