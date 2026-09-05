import { expect, test } from '@playwright/test';
import { login } from './helpers';

test.describe('dashboard charts', () => {
	test('la gráfica de crecimiento se recorre con el teclado', async ({ page }) => {
		await login(page);

		const chart = page.getByRole('slider', {
			name: 'Recorrer la gráfica de crecimiento del portafolio'
		});
		await expect(chart).toBeVisible();
		await expect(page.getByText(/Pasa el cursor por la gráfica/)).toBeVisible();

		await chart.focus();
		await chart.press('End');

		// El punto señalado se anuncia y además se pinta en el detalle de arriba.
		await expect(chart).toHaveAttribute('aria-valuetext', /Valor de mercado \$/);
		await expect(page.getByText(/^Capital invertido \$/)).toBeVisible();
		await expect(page.getByText(/^Ganancia [+−]\$/)).toBeVisible();
	});

	test('la gráfica se puede leer en rentabilidad en vez de en dinero', async ({ page }) => {
		await login(page);

		const card = page.locator('section[aria-label="Crecimiento del portafolio"]');
		const units = card.getByRole('group', { name: 'Unidad de la gráfica' });
		await expect(units.getByRole('button', { name: 'Valor' })).toHaveAttribute(
			'aria-pressed',
			'true'
		);

		await units.getByRole('button', { name: '%', exact: true }).click();

		// La serie del fixture crece a base de aportes: en dinero la curva sube y
		// en rentabilidad no, que es justo lo que esta vista viene a enseñar.
		const series = card.getByRole('table', {
			name: /Rentabilidad acumulada y ganancia sobre coste/
		});
		await expect(
			series.getByRole('columnheader', { name: 'Rentabilidad acumulada' })
		).toBeAttached();
		await expect(series.getByRole('columnheader', { name: 'Ganancia sobre coste' })).toBeAttached();
		await expect(series.getByRole('cell').first()).toHaveText(/^[+-]?\d+,\d%$/);

		// Y la cifra del periodo va con las otras métricas, no escondida en el SVG.
		await expect(card.getByText('Rentabilidad real, Todo')).toBeVisible();
	});

	test('la gráfica ofrece sus dos series como tabla para el lector de pantalla', async ({
		page
	}) => {
		await login(page);

		const series = page.getByRole('table', {
			name: /Valor de mercado y capital invertido del portafolio/
		});
		await expect(series.getByRole('columnheader', { name: 'Valor de mercado' })).toBeAttached();
		await expect(series.getByRole('columnheader', { name: 'Capital invertido' })).toBeAttached();

		// Cada fila se identifica con la fecha completa: el eje abrevia a mes y
		// año en los rangos largos, y así varias filas se llamarían igual.
		const firstRow = series.getByRole('rowheader').first();
		await expect(firstRow).toHaveText(/^\d{2} de \p{L}+ de \d{4}$/u);
		await expect(firstRow.locator('time')).toHaveAttribute('datetime', /^\d{4}-\d{2}-\d{2}$/);
	});

	// La tabla accesible de la gráfica se ocultaba con `height: 1px`, que para
	// una tabla es un mínimo y no un máximo: sus filas seguían ocupando sitio y
	// la página terminaba en una franja vacía tan alta como la serie.
	test('la tabla oculta de la gráfica no alarga la página', async ({ page }) => {
		await login(page);
		await page.getByRole('slider').first().waitFor();

		const slack = await page.evaluate(() => {
			const doc = document.documentElement.scrollHeight;
			const last = document.querySelector('section[aria-label="Crecimiento del portafolio"]');
			const bottom = last!.getBoundingClientRect().bottom + window.scrollY;
			return doc - bottom;
		});

		expect(slack).toBeLessThan(400);
	});
});

test.describe('el reparto del patrimonio', () => {
	test('lee el mismo total por plataforma, por portafolio y por tipo', async ({ page }) => {
		await login(page);

		const where = page.getByRole('region', { name: 'Dónde está' });

		// Abre por el primer corte con datos, que en el fixture son las plataformas.
		await expect(where.getByRole('tab', { name: 'Plataforma' })).toHaveAttribute(
			'aria-selected',
			'true'
		);
		await expect(where.getByRole('link', { name: 'Broker Demo' })).toBeVisible();
		await expect(where.getByText('5 posiciones abiertas')).toBeVisible();

		// Cada fila dice su participación también para el lector de pantalla: la
		// barra es la única forma visual de leerla.
		await expect(where.getByText(/^\d+,\d% del total$/).first()).toBeAttached();

		await where.getByRole('tab', { name: 'Portafolio' }).click();
		await expect(where.getByRole('link', { name: 'Cartera Principal' })).toBeVisible();
		await expect(where.getByRole('link', { name: 'Cripto' })).toBeVisible();
		await expect(where.getByRole('link', { name: 'Reserva' })).toBeVisible();
		await expect(where.getByText('Acciones & ETFs')).toBeVisible();

		// El reparto por clase de activo no tiene rendimiento que enseñar —el
		// endpoint no lo devuelve—, así que la última columna cambia de sentido.
		await where.getByRole('tab', { name: 'Tipo de activo' }).click();
		await expect(where.getByRole('columnheader', { name: 'Del total' })).toBeVisible();
		await expect(where.getByRole('rowheader', { name: 'Acciones', exact: true })).toBeVisible();
	});

	test('las pestañas del reparto se recorren con las flechas', async ({ page }) => {
		await login(page);

		const where = page.getByRole('region', { name: 'Dónde está' });
		const first = where.getByRole('tab', { name: 'Plataforma' });

		await first.focus();
		await first.press('ArrowRight');

		await expect(where.getByRole('tab', { name: 'Portafolio' })).toBeFocused();
		await expect(where.getByRole('tab', { name: 'Portafolio' })).toHaveAttribute(
			'aria-selected',
			'true'
		);
	});
});

test.describe('dashboard', () => {
	test('renders the main widgets with data from the backend', async ({ page }) => {
		await login(page);

		// La cifra, una sola vez y en grande.
		await expect(page.getByRole('heading', { name: 'Patrimonio total', level: 1 })).toBeVisible();
		await expect(page.getByText(/sobre lo invertido/)).toBeVisible();

		await expect(page.getByRole('region', { name: 'Dónde está' })).toBeVisible();
		await expect(page.locator('section[aria-label="Crecimiento del portafolio"]')).toBeVisible();

		// Movimientos alimentados por los fixtures de la API.
		await expect(page.getByRole('heading', { name: 'Movimientos' })).toBeVisible();
		await expect(page.getByText('AAPL').first()).toBeVisible();
	});

	// La cabecera no decía en qué sección estaba el usuario; ahora sale del mismo
	// listado que pinta el menú, así que los dos no pueden discrepar.
	test('la cabecera nombra la sección abierta', async ({ page }) => {
		await login(page);

		// `banner` es la barra superior del panel; las páginas traen además su
		// propio `<header>` con el título, y `header` a secas casa con los dos.
		await expect(page.getByRole('banner').getByText('Resumen')).toBeVisible();

		await page.getByRole('link', { name: 'Plataformas' }).click();
		await expect(page.getByRole('banner').getByText('Plataformas')).toBeVisible();
	});

	// Enseñar la tasa es el motivo de que la aplicación la traiga: una cifra en
	// pesos que nadie puede cuadrar con su banco no sirve de mucho.
	test('enseña la tasa con la que convierte al cambiar de moneda', async ({ page }) => {
		await login(page);

		// En dólares no hay conversión que enseñar, así que la línea no está.
		await expect(page.getByText(/^1\s*USD\s*=/)).toBeHidden();

		await page.getByLabel('Moneda de visualización').selectOption('COP');

		await expect(page.getByText('1 USD = 4.123,46 COP')).toBeVisible();
		await expect(page.getByText(/TRM · dolarapi\.com/)).toBeVisible();
	});
});
