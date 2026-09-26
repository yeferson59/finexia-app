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

	test('the create form says when it could not save', async ({ page }) => {
		await login(page);
		await page.goto('/dashboard/portfolios/add');

		await page.fill('#name', 'Retiro');
		await page.locator('input[name="riskId"]').first().check();
		await page.getByRole('button', { name: 'Crear portafolio' }).click();

		// El stub no tiene alta de portafolios. Lo que se comprueba es que el
		// rechazo se ve: la action devolvía `{ success: false }` con un 200 y el
		// formulario no leía `form`, así que no quedaba ni rastro en pantalla.
		await expect(page.getByText('No pudimos crear el portafolio')).toBeVisible();
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

	test('parte la ganancia del portafolio en periodos', async ({ page }) => {
		await login(page);
		await page.goto(`/dashboard/portfolios/${TEST_PORTFOLIO_ID}`);

		const strip = page.getByRole('region', { name: 'Rentabilidad por periodo' });
		await expect(strip.getByRole('tab', { name: /^1 día / })).toBeVisible();
		await expect(strip.getByRole('tab', { name: /^Desde el inicio / })).toBeVisible();

		// La serie del portafolio es la agregada a escala: el porcentaje no cambia
		// con la escala y el dinero sí, en la moneda del portafolio.
		const year = strip.getByRole('tab', { name: /^1 año / });
		await expect(year).toContainText('-1,21%');
		await year.click();
		await expect(strip.getByRole('tabpanel')).toContainText(/perdiste \$[\d,]+\.\d{2},/);
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

	// Doce tarjetas —seis de «Resumen de Posición» y seis de «Información del
	// Activo»— se han quedado en una cifra y tres frases. La mitad de aquellas
	// repetía algo que ya estaba en la misma pantalla.
	test('asset detail states the position once, in the portfolio currency', async ({ page }) => {
		await login(page);
		await page.goto(`/dashboard/portfolios/${TEST_PORTFOLIO_ID}/assets/AAPL`);

		await expect(page.getByRole('heading', { level: 1 })).toContainText('AAPL');
		// La vuelta dice a dónde lleva, en vez de un «Volver» a secas.
		await expect(page.getByRole('link', { name: 'Volver a Cartera Principal' })).toBeVisible();

		const headline = page.locator('section[aria-labelledby="position-value"]');
		await expect(headline).toContainText('Tienes 42 acciones');
		await expect(headline).toContainText('$9,002.70');
		await expect(headline).toContainText('+$1,929.90 sobre los $7,072.80 que invertiste (+27,29%)');
		// Las tarjetas «precio promedio» y «precio actual», comparadas entre sí.
		await expect(headline).toContainText('Pagaste $168.40 por acción; hoy cotiza a $214.35.');
		// La tarjeta «asignación», situada dentro de su portafolio.
		await expect(headline).toContainText('Es el 20,0% de Cartera Principal.');

		// La transacción de compra del fixture aparece en el historial.
		await expect(page.getByRole('rowheader', { name: /Compra/ })).toBeVisible();
	});

	// El precio unitario de un interés se cotiza por debajo del céntimo: con dos
	// decimales salía «$0.00» en la misma fila que su total de $19.95.
	test('asset detail keeps a sub-cent unit price readable', async ({ page }) => {
		await login(page);
		await page.goto('/dashboard/portfolios/11111111-1111-4111-8111-111111111113/assets/USD');

		// Y un activo cuyo valor no se ha movido no lleva ni signo ni color de
		// ganancia. Es efectivo en su propia moneda, así que el dinero se
		// depositó, no se invirtió (`asset-position-headline`).
		await expect(page.getByText('Vale lo mismo que los $9,500.00 que depositaste.')).toBeVisible();

		const row = page.getByRole('row').filter({ hasText: 'Interés' });
		await expect(row).toContainText('$0.0021');
		await expect(row).toContainText('$19.95');
	});

	test('asset detail opens the add-transaction and quick-sell forms', async ({ page }) => {
		await login(page);
		await page.goto(`/dashboard/portfolios/${TEST_PORTFOLIO_ID}/assets/AAPL`);

		await page.getByRole('button', { name: 'Registrar movimiento' }).click();
		await expect(page.getByRole('button', { name: 'Registrar transacción' })).toBeVisible();

		await page.keyboard.press('Escape');
		await expect(page.getByRole('dialog', { name: 'Registrar transacción' })).not.toBeVisible();

		// La venta rápida se abre desde el lote de compra del historial.
		await page.getByRole('button', { name: /^Vender/ }).click();
		await expect(page.getByRole('dialog', { name: 'Vender posición' })).toBeVisible();
		await expect(page.getByRole('button', { name: 'Confirmar Venta Total' })).toBeVisible();
	});

	// Un dividendo llega a la cuenta, así que el alta lo abona al efectivo de la
	// plataforma salvo que se desmarque; los demás tipos no ofrecen la casilla.
	test('asset detail offers to pay a dividend into the platform cash', async ({ page }) => {
		await login(page);
		await page.goto(`/dashboard/portfolios/${TEST_PORTFOLIO_ID}/assets/AAPL`);

		await page.getByRole('button', { name: 'Registrar movimiento' }).click();
		const dialog = page.getByRole('dialog', { name: 'Registrar transacción' });
		const credit = dialog.getByRole('checkbox', { name: /Abonar al efectivo de la plataforma/ });

		await expect(credit).toHaveCount(0);

		await dialog.getByLabel('Tipo').selectOption('dividend');
		await expect(credit).toBeChecked();

		await dialog.getByLabel(/Monto del dividendo/).fill('25');
		await expect(dialog).toContainText(/Suma .*25[.,]00 a tu efectivo en USD/);

		await dialog.getByLabel('Tipo').selectOption('buy');
		await expect(credit).toHaveCount(0);
	});

	// El espejo del dividendo: una compra puede salir del efectivo que la
	// plataforma ya guarda, y solo una compra lo ofrece.
	test('asset detail offers to pay a purchase from the platform cash', async ({ page }) => {
		await login(page);
		await page.goto(`/dashboard/portfolios/${TEST_PORTFOLIO_ID}/assets/AAPL`);

		await page.getByRole('button', { name: 'Registrar movimiento' }).click();
		const dialog = page.getByRole('dialog', { name: 'Registrar transacción' });
		const pay = dialog.getByRole('checkbox', {
			name: /Pagarlo con mi efectivo en esta plataforma/
		});

		// Sin marcar: el dinero llega de fuera salvo que se diga lo contrario, y el
		// selector de cajón no estorba hasta que se marca.
		await expect(pay).not.toBeChecked();
		await expect(dialog.getByLabel('De dónde sale')).toHaveCount(0);
		await expect(dialog).toContainText(/En Cuenta principal tienes .*2.500[.,]00/);

		await dialog.getByLabel('Cantidad').fill('2');
		await dialog.getByLabel('Precio unitario').fill('150');
		await pay.check();
		await expect(dialog).toContainText(/Resta .*300[.,]00 de tu efectivo en USD/);

		// Con más de un sitio de donde sacarlo, se elige cuál: la mayoría del
		// dinero de una cuenta está en un bolsillo, no en el saldo principal.
		const from = dialog.getByLabel('De dónde sale');
		// Por valor, que es el id del bolsillo en las fixtures del stub: la
		// etiqueta lleva el saldo pegado y no es un texto estable.
		await from.selectOption('88888888-8888-4888-8888-888888888888');
		await expect(dialog).toContainText(/En Para acciones tienes .*800[.,]00/);

		// Más de lo que hay: se avisa, sin bloquear —quien decide es el backend—.
		await dialog.getByLabel('Cantidad').fill('100');
		await expect(dialog).toContainText('No alcanza para esta compra');

		await dialog.getByLabel('Tipo').selectOption('dividend');
		await expect(pay).toHaveCount(0);
	});

	// Lo mismo con una venta: lo cobrado, menos la comisión, llega al efectivo.
	test('the quick sell pays the proceeds into the platform cash', async ({ page }) => {
		await login(page);
		await page.goto(`/dashboard/portfolios/${TEST_PORTFOLIO_ID}/assets/AAPL`);

		await page.getByRole('button', { name: /^Vender/ }).click();
		const dialog = page.getByRole('dialog', { name: 'Vender posición' });
		const credit = dialog.getByRole('checkbox', { name: /Abonar al efectivo de la plataforma/ });

		await expect(credit).toBeChecked();
		await expect(dialog).toContainText(/Suma .* a tu efectivo en USD/);

		await credit.uncheck();
		await expect(credit).not.toBeChecked();
	});

	// Borrar una transacción es irreversible y el botón vive en una tabla de
	// filas casi idénticas, así que pasa por una confirmación que dice cuál se
	// va a borrar y qué le ocurre a la posición.
	test('asset detail deletes a transaction after confirming', async ({ page }) => {
		await login(page);
		await page.goto(`/dashboard/portfolios/${TEST_PORTFOLIO_ID}/assets/AAPL`);

		await page
			.getByRole('button', { name: /^Eliminar el / })
			.first()
			.click();

		const dialog = page.getByRole('dialog', { name: 'Eliminar transacción' });
		await expect(dialog).toBeVisible();
		await expect(dialog).toContainText('la cantidad pasa a 0');

		await dialog.getByRole('button', { name: 'Eliminar' }).click();

		// Al confirmar, el diálogo se cierra y la página se recarga con lo que
		// responda el backend.
		await expect(dialog).toBeHidden();
		await expect(page.getByRole('heading', { name: 'Movimientos' })).toBeVisible();
	});

	/* La misma acción con el mismo nombre de punta a punta: la entrada del
	   detalle, el título de la pantalla y el botón que la remata. */
	test('the add-asset flow is called the same from end to end', async ({ page }) => {
		await login(page);
		await page.goto(`/dashboard/portfolios/${TEST_PORTFOLIO_ID}`);

		await page.getByRole('link', { name: 'Añadir activo' }).click();
		await page.waitForURL(`**/dashboard/portfolios/${TEST_PORTFOLIO_ID}/add`);

		await expect(page.getByRole('heading', { level: 1 })).toContainText('Añadir activo');
		await expect(page.getByRole('button', { name: 'Añadir activo' })).toBeVisible();
	});

	/* Crear un activo en el catálogo no es lo mismo que añadirlo al portafolio,
	   y por eso conserva su propio verbo dentro del buscador. */
	test('offers to create an asset the catalogue does not have', async ({ page }) => {
		await login(page);
		await page.goto(`/dashboard/portfolios/${TEST_PORTFOLIO_ID}/add`);

		await page.fill('#asset-search', 'ZZZQ');

		await expect(page.getByText('No hay ningún activo que se llame')).toBeVisible();
		await page.getByRole('button', { name: 'Crear ZZZQ' }).click();

		await expect(page.getByText('Al crearlo queda elegido y sigues con la posición')).toBeVisible();
		await expect(page.getByRole('button', { name: 'Crear activo' })).toBeVisible();
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

		await page.getByRole('button', { name: 'Añadir activo' }).click();

		// The action redirects back to the portfolio detail on success.
		await page.waitForURL(`**/dashboard/portfolios/${TEST_PORTFOLIO_ID}`);
		await expect(page.getByRole('heading', { level: 1 })).toContainText('Cartera Principal');
	});
});
