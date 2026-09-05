import { expect, test } from '@playwright/test';
import { TEST_PORTFOLIO_ID, login } from './helpers';

/*
 * El listado: una fila por portafolio, ordenadas de mayor a menor, con la barra
 * que reparte cada una entre el capital que se puso y lo que ha ganado.
 */
test.describe('portfolio list', () => {
	test('ranks the portfolios and totals them at the foot', async ({ page }) => {
		await login(page);
		await page.goto('/dashboard/portfolios');

		await expect(page.getByRole('heading', { level: 1 })).toContainText('Portafolios');

		// De mayor a menor valor, que es lo que hace legible la escalera.
		const names = page.locator('tbody th a');
		await expect(names).toHaveText(['Cartera Principal', 'Reserva', 'Cripto']);

		// La descripción que escribió el dueño, no la etiqueta del tipo.
		await expect(page.locator('tbody tr').first()).toContainText(
			'Acciones y ETFs a largo plazo, 5 posiciones'
		);
		await expect(page.locator('tbody tr').first()).toContainText('Moderado');

		// El total vive al pie de su columna, no en una tarjeta encima.
		const foot = page.locator('tfoot tr');
		await expect(foot).toContainText('3 portafolios, 10 posiciones abiertas');
		await expect(foot).toContainText('$89,406.10');
		await expect(foot).toContainText('+14,02%');
	});

	test('opens a portfolio from its name', async ({ page }) => {
		await login(page);
		await page.goto('/dashboard/portfolios');

		await page.getByRole('link', { name: 'Cripto' }).click();
		await expect(page.getByRole('heading', { level: 1 })).toContainText('Cripto');
	});

	test('sends someone with nothing to the create form', async ({ page }) => {
		await login(page);
		await page.goto('/dashboard/portfolios');

		await page.getByRole('link', { name: 'Crear portafolio' }).click();
		await page.waitForURL('**/dashboard/portfolios/add');
	});
});

test.describe('portfolio detail', () => {
	test('renders the portfolio with its holdings', async ({ page }) => {
		await login(page);
		await page.goto(`/dashboard/portfolios/${TEST_PORTFOLIO_ID}`);

		await expect(page.getByRole('heading', { level: 1 })).toContainText('Cartera Principal');
		await expect(page.getByText('AAPL').first()).toBeVisible();
		await expect(page.getByText('Riesgo moderado, 5 activos')).toBeVisible();

		// El valor sale una vez, en la cifra de arriba, con el capital del que
		// viene. Antes volvía a salir en el centro del donut y en dos tarjetas.
		const headline = page.locator('section[aria-labelledby="market-value"]');
		await expect(headline).toContainText('$45,035.10');
		await expect(headline).toContainText('sobre los $37,150.50 que invertiste');
		await expect(headline).toContainText('+21,22%');
	});

	// Cuatro tarjetas —mejor activo, peor activo, concentración y el donut por
	// tipo— eran lecturas de esta misma lista. Ahora la lista está ordenada y
	// dos frases dicen lo que costaba encontrar en ella.
	test('the positions list carries what the stat cards used to', async ({ page }) => {
		await login(page);
		await page.goto(`/dashboard/portfolios/${TEST_PORTFOLIO_ID}`);

		const positions = page.locator('section[aria-labelledby="positions-title"]');
		await expect(positions).toContainText('Acciones 54,4%, ETFs 45,6%.');
		await expect(positions).toContainText('NVDA es la que más ha rendido');
		await expect(positions).toContainText('CSPX, la que menos');
		await expect(positions).toContainText('La mayor operación registrada aquí: NVDA');

		// De mayor a menor peso: la primera fila es la posición dominante.
		await expect(positions.locator('tbody th').first()).toContainText('VWCE');
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

		await page.keyboard.press('Escape');
		await expect(page.getByRole('dialog', { name: 'Registrar transacción' })).not.toBeVisible();

		// La venta rápida se abre desde el lote de compra del historial.
		await page.getByRole('button', { name: 'Vender' }).click();
		await expect(page.getByRole('dialog', { name: 'Vender posición' })).toBeVisible();
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
