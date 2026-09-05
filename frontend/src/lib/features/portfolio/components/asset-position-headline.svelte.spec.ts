import { page } from 'vitest/browser';
import { describe, it, expect } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Headline from './asset-position-headline.svelte';
import type { AssetPosition } from '../asset';

function position(overrides: Partial<AssetPosition> = {}): AssetPosition {
	return {
		ticker: 'AAPL',
		name: 'Apple Inc.',
		assetType: 'stock',
		exchange: 'NASDAQ',
		currency: 'USD',
		costCurrency: 'USD',
		baseCurrency: 'USD',
		marketPrice: 214.35,
		totalQty: 42,
		totalCost: 7072.8,
		averageCost: 168.4,
		totalValue: 9002.7,
		gainLoss: 1929.9,
		gainLossPercent: 27.29,
		allocation: 20,
		fxConverted: true,
		...overrides
	};
}

describe('asset-position-headline.svelte', () => {
	// Doce tarjetas se han quedado en una cifra y tres frases; las tres frases
	// son las que llevan lo que decían las tarjetas que no se repetían.
	it('dice la cantidad, el valor y de dónde sale la ganancia', async () => {
		render(Headline, { position: position(), portfolioName: 'Cartera Principal' });

		await expect.element(page.getByText('Tienes 42 acciones')).toBeInTheDocument();
		await expect.element(page.getByText('$9,002.70')).toBeInTheDocument();
		await expect
			.element(page.getByText('+$1,929.90 sobre los $7,072.80 que invertiste (+27,29%)'))
			.toBeInTheDocument();
		// Las tarjetas «precio promedio» y «precio actual» juntas, que es donde
		// significan algo: la una comparada con la otra.
		await expect
			.element(page.getByText('Pagaste $168.40 por acción; hoy cotiza a $214.35.'))
			.toBeInTheDocument();
		// La tarjeta «asignación», situada dentro de su portafolio.
		await expect.element(page.getByText('Es el 20,0% de Cartera Principal.')).toBeInTheDocument();
	});

	/*
	 * El efectivo de una cuenta no gana nada, y salía «+$0.00 (0,00%)» escrito
	 * en verde: un signo y un color que prometen lo que la propia cifra
	 * desmiente.
	 */
	it('no pinta de ganancia una ganancia que redondea a cero', async () => {
		render(Headline, {
			position: position({
				assetType: 'cash',
				totalQty: 9500,
				totalCost: 9500,
				totalValue: 9500,
				averageCost: 1,
				marketPrice: 1,
				gainLoss: 0,
				gainLossPercent: 0
			}),
			portfolioName: 'Reserva'
		});

		await expect
			.element(page.getByText('Vale lo mismo que los $9,500.00 que invertiste.'))
			.toBeInTheDocument();
		// Y con los dos precios iguales, la frase de las unidades solo repetiría
		// que no ha pasado nada.
		await expect.element(page.getByText(/Pagaste/)).not.toBeInTheDocument();
	});

	// Restar un precio de coste en dólares de una cotización en euros da un
	// número que no existe; la ficha los ponía uno al lado del otro sin decirlo.
	it('avisa cuando el coste y la cotización van en monedas distintas', async () => {
		render(Headline, {
			position: position({ ticker: 'MC.FR', currency: 'EUR', costCurrency: 'USD' }),
			portfolioName: 'Cartera Principal'
		});

		await expect
			.element(page.getByText(/La compra se liquidó en USD y el mercado cotiza en EUR/))
			.toBeInTheDocument();
	});

	// Una pérdida es la otra mitad de la frase y tiene que leerse como tal.
	it('escribe una pérdida con su signo', async () => {
		render(Headline, {
			position: position({
				totalValue: 6000,
				gainLoss: -1072.8,
				gainLossPercent: -15.17
			}),
			portfolioName: 'Cartera Principal'
		});

		await expect
			.element(page.getByText('−$1,072.80 sobre los $7,072.80 que invertiste (-15,17%)'))
			.toBeInTheDocument();
	});
});
