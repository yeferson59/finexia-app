import { expect, test } from '@playwright/test';
import { login } from './helpers';

test.describe('reports', () => {
	test('opens with what the account returned and over how long', async ({ page }) => {
		await login(page);
		await page.goto('/dashboard/reports');

		// La cifra de cabecera es la rentabilidad del periodo, no el saldo: la
		// serie del fixture crece a base de aportes y el saldo mentiría.
		const headline = page.getByRole('region', { name: 'Lo que rindió tu dinero' });
		await expect(headline.locator('.amount')).toHaveText(/^[+-]?\d+,\d%$/);
		// El periodo del fixture: de junio de 2025 a julio de 2026.
		await expect(headline.getByText(/Del 1 de junio de 2025 al 28 de julio de 2026/)).toBeVisible();
		await expect(headline.getByText(/Hoy la cuenta vale/)).toBeVisible();

		// Y cuando la rentabilidad y la ganancia sobre coste se separan —aquí lo
		// hacen— la cabecera dice por qué, en vez de dejar dos cifras que parecen
		// contradecirse.
		await expect(headline.getByText(/Las dos cifras no dicen lo mismo/)).toBeVisible();
	});

	test('lays the monthly returns out as one matrix, a year per row', async ({ page }) => {
		await login(page);
		await page.goto('/dashboard/reports');

		const matrix = page.getByRole('table', { name: /Rentabilidad de cada mes/ });
		// Una fila por año: el fixture cubre de junio de 2025 a julio de 2026.
		await expect(matrix.getByRole('row', { name: /^2026 / })).toBeVisible();
		await expect(matrix.getByRole('row', { name: /^2025 / })).toBeVisible();

		// La celda se lee con su mes y su año por las cabeceras de la tabla, sin
		// necesidad de un `aria-label` por celda.
		const january = matrix
			.getByRole('row', { name: /^2026 / })
			.getByRole('cell')
			.first();
		await expect(january).toHaveText(/^\+\d+,\d%$/);

		// Los meses sin dato lo dicen en vez de dejar la celda muda.
		await expect(
			matrix
				.getByRole('row', { name: /^2025 / })
				.getByRole('cell')
				.first()
		).toContainText('sin dato');

		// El total del año cierra su fila: era el «acumulado del año» de cada
		// tarjeta, repetido tantas veces como años.
		await expect(matrix.getByRole('columnheader', { name: 'Total' })).toBeVisible();

		// Y el pie deja dicho, una sola vez, que las cifras son rendimiento y no
		// saldo, y qué marca el asterisco de un mes incompleto.
		await expect(page.getByText(/no cuenta como\s+rentabilidad/)).toHaveCount(1);
		await expect(page.getByText(/Un asterisco marca el mes/)).toBeVisible();
	});

	test('publishes each risk measure next to what it measures', async ({ page }) => {
		await login(page);
		await page.goto('/dashboard/reports');

		const movement = page.getByRole('region', { name: 'Cómo se movió' });

		// Lo que mide cada cifra se lee: antes vivía en un `title` que en un móvil
		// no se abre.
		const drawdown = movement.getByRole('row', { name: /Máxima caída/ });
		await expect(drawdown).toContainText('La peor bajada desde un máximo');

		for (const label of ['Mejor mes', 'Peor mes', 'Volatilidad anualizada', 'Ratio de Sharpe']) {
			await expect(movement.getByRole('rowheader', { name: label })).toBeVisible();
		}

		// El mes del máximo va aparte de su cifra, no pegado con un punto medio.
		await expect(movement.getByRole('row', { name: /Mejor mes/ })).toContainText(/de 20\d\d/);

		// Con setenta puntos de historial no queda ninguna medida sin calcular.
		await expect(page.getByText('N/A')).toHaveCount(0);

		// El Sharpe se publica en gris y con su reparo al lado: es un cociente
		// estimado, y en verde se leía como un sello de calidad.
		const sharpe = movement.getByRole('row', { name: /Ratio de Sharpe/ });
		await expect(sharpe.locator('.value')).toHaveClass(/neutral/);
		await expect(sharpe.getByText(/margen de error/)).toBeVisible();

		// Las seis cifras que subieron a la cabecera no se repiten aquí.
		for (const gone of ['Rentabilidad del periodo', 'Valor actual', 'Periodo cubierto']) {
			await expect(movement.getByRole('rowheader', { name: gone })).toHaveCount(0);
		}
	});

	test('projects the accumulated return from zero, with the money in a table', async ({ page }) => {
		await login(page);
		await page.goto('/dashboard/reports');

		const projection = page.getByRole('region', { name: 'Si el ritmo se mantiene' });
		await expect(
			projection.getByRole('heading', { name: 'Si el ritmo se mantiene' })
		).toBeVisible();
		await expect(
			page.getByText('La proyección necesita al menos seis meses de historial.')
		).toHaveCount(0);

		// La tabla es la gráfica en cifras, y es visible: la versión oculta para
		// lectores de pantalla dejaba fuera a quien no puede leer un SVG diminuto.
		const table = page.getByRole('table', { name: /Valor proyectado/ });
		await expect(table.getByRole('rowheader', { name: 'Valor proyectado' })).toBeVisible();
		await expect(table.getByRole('rowheader', { name: 'Acumulado desde hoy' })).toBeVisible();
		// Cinco años, y el primero es hoy con un acumulado de cero.
		await expect(table.getByRole('columnheader')).toHaveCount(5);
		await expect(table.getByRole('columnheader', { name: /hoy/ })).toBeVisible();
	});

	test('never lets the page scroll sideways on a phone', async ({ page }) => {
		await page.setViewportSize({ width: 390, height: 844 });
		await login(page);
		await page.goto('/dashboard/reports');

		// La matriz y la tabla de la proyección se desplazan dentro de su carril;
		// la página, no. El `sin dato` de cada celda es un `.sr-only` en posición
		// absoluta y, si su carril no está posicionado, se escapa del recorte y
		// arrastra el ancho de scroll de todo el documento: el titular y la prosa
		// se iban de lado con él.
		const { doc, viewport } = await page.evaluate(() => ({
			doc: document.documentElement.scrollWidth,
			viewport: document.documentElement.clientWidth
		}));

		expect(doc).toBe(viewport);

		// Y los dos carriles sí se desplazan, que es lo que hace legible una
		// matriz de trece columnas en un móvil.
		const matrix = page.getByRole('region', { name: /tabla desplazable/ }).first();
		const scrolls = await matrix.evaluate((el) => el.scrollWidth > el.clientWidth);

		expect(scrolls).toBe(true);
	});

	test('offers the downloadable reports with what each one holds', async ({ page }) => {
		await login(page);
		await page.goto('/dashboard/reports');

		const downloads = page.getByRole('region', { name: 'Llévate los datos' });
		await expect(downloads.getByRole('heading', { name: 'Resumen mensual' })).toBeVisible();
		await expect(downloads.getByText('Cada compra, venta y dividendo')).toBeVisible();

		// Cada enlace dice qué archivo baja: tres «Descargar» idénticos no se
		// distinguían con un lector de pantalla.
		await expect(
			downloads.getByRole('link', { name: 'Descargar Resumen mensual en XLSX' })
		).toHaveAttribute('href', /\/dashboard\/reports\/download\?type=summary/);
		await expect(
			downloads.getByRole('link', { name: 'Descargar Riesgo y volatilidad en XLSX' })
		).toBeVisible();
	});
});
