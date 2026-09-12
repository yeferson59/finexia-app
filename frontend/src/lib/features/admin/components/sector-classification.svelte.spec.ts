import { page } from 'vitest/browser';
import { describe, it, expect } from 'vitest';
import { render } from 'vitest-browser-svelte';
import SectorClassification from './sector-classification.svelte';

/*
 * Las dos formas de clasificar un activo son excluyentes, y quien lo hace
 * cumplir es este formulario: desactiva la que no está en uso, y un campo
 * desactivado no se envía. Es la única parte de la regla que no está en un
 * schema, así que es la que hay que probar aquí.
 */
describe('sector-classification.svelte', () => {
	it('desactiva el desglose cuando hay una industria única', async () => {
		render(SectorClassification, { sector: 'technology', weights: {} });

		await page.getByText('Desglose por industrias').click();

		await expect.element(page.getByLabelText('Tecnología')).toBeDisabled();
		await expect
			.element(page.getByText('vuelve la de arriba a «Sin clasificar»', { exact: false }))
			.toBeInTheDocument();
	});

	it('desactiva la industria única cuando hay desglose, y suma los pesos', async () => {
		render(SectorClassification, {
			sector: '',
			weights: { technology: 33.1, financials: 13.8 }
		});

		// Por rol y no por etiqueta: el `<label>` lleva dentro el «(opcional)».
		await expect.element(page.getByRole('combobox')).toBeDisabled();
		// El total se enseña tal cual: no tiene por qué dar 100, y verlo es lo
		// que deja decidir si el resto es la caja del fondo o un olvido.
		await expect.element(page.getByText('Suma: 46,9%')).toBeInTheDocument();
	});

	it('avisa cuando los pesos se pasan de 100', async () => {
		render(SectorClassification, { sector: '', weights: { technology: 80, energy: 80 } });

		await expect
			.element(page.getByText('por encima de 100 %', { exact: false }))
			.toBeInTheDocument();
	});
});
