import { test, expect } from '@playwright/test';
import { login, TEST_PLATFORM_ID } from './helpers';

/**
 * Una plataforma que todavía apuntan posiciones no se puede borrar.
 *
 * El backend se niega —arrastrarlas consigo sería borrar el historial del dueño
 * por un clic— y devuelve un 409 diciendo qué hay que quitar antes. Este flujo
 * es la razón de ser de ese 409: sin él la petición terminaba en un 500 con el
 * motivo sólo en el log del servidor, y el modal se quedaba mudo.
 */
test.describe('plataformas', () => {
	test('rechaza eliminar una plataforma que todavía tiene posiciones', async ({ page }) => {
		await login(page);
		await page.goto(`/dashboard/platforms/${TEST_PLATFORM_ID}`);

		// El de la cabecera abre el modal; el del modal es el que envía.
		await page.getByRole('button', { name: 'Eliminar' }).first().click();
		await expect(page.getByRole('heading', { name: 'Confirmar eliminación' })).toBeVisible();

		await page.locator('form[action="?/delete"]').getByRole('button', { name: 'Eliminar' }).click();

		// El motivo, donde el usuario está mirando, y en términos de lo que puede
		// hacer a continuación.
		await expect(page.getByRole('alert')).toContainText('todavía tiene posiciones registradas');

		// Y nada se borró: seguimos en el detalle de la misma plataforma.
		await expect(page).toHaveURL(new RegExp(TEST_PLATFORM_ID));
		await expect(page.getByRole('heading', { name: 'Broker Demo' })).toBeVisible();
	});
});
