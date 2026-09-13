import { page } from 'vitest/browser';
import { describe, it, expect } from 'vitest';
import { render } from 'vitest-browser-svelte';
import SectorClassification from './sector-classification.svelte';

/*
 * Las dos formas de clasificar un activo son excluyentes, y quien lo hace
 * cumplir es este formulario: solo pinta los campos del modo elegido, y un
 * campo que no está no se envía. Es la única parte de la regla que no está en
 * un schema, así que es la que hay que probar aquí.
 */
describe('sector-classification.svelte', () => {
	it('con una industria única no pinta las casillas del reparto', async () => {
		render(SectorClassification, { sector: 'technology', weights: {} });

		await expect.element(page.getByRole('combobox')).toBeInTheDocument();
		await expect.element(page.getByRole('spinbutton').first()).not.toBeInTheDocument();
	});

	it('pasar a varias industrias quita el desplegable y avisa de lo que se descarta', async () => {
		render(SectorClassification, { sector: 'technology', weights: {} });

		await page.getByRole('radio', { name: 'Varias industrias' }).click();

		await expect.element(page.getByRole('combobox')).not.toBeInTheDocument();
		await expect.element(page.getByLabelText('Tecnología')).toBeInTheDocument();
		await expect.element(page.getByText('Tecnología deja de ser la industria')).toBeInTheDocument();
	});

	it('con reparto abre en varias industrias y enseña la suma y lo que falta', async () => {
		render(SectorClassification, {
			sector: '',
			weights: { technology: 33.1, financials: 13.8 }
		});

		await expect.element(page.getByRole('combobox')).not.toBeInTheDocument();
		// El total se enseña tal cual: no tiene por qué dar 100, y verlo es lo
		// que deja decidir si el resto es la caja del fondo o un olvido.
		await expect.element(page.getByText('46,9%', { exact: true })).toBeInTheDocument();
		await expect.element(page.getByText('Faltan 53,1%')).toBeInTheDocument();
	});

	it('volver a una industria avisa de que el reparto se descarta', async () => {
		render(SectorClassification, { sector: '', weights: { technology: 33.1 } });

		await page.getByRole('radio', { name: 'Una industria' }).click();

		await expect.element(page.getByText('se descarta el reparto')).toBeInTheDocument();
	});

	it('avisa cuando los pesos se pasan de 100', async () => {
		render(SectorClassification, { sector: '', weights: { technology: 80, energy: 80 } });

		await expect.element(page.getByText('Sobran 60,0%')).toBeInTheDocument();
	});
});
