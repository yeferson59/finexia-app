import { page } from 'vitest/browser';
import { describe, it, expect } from 'vitest';
import { render } from 'vitest-browser-svelte';
import DeletePosition from './asset-delete-position.svelte';
import type { Holding } from '$lib/api/types';

function holding(overrides: Partial<Holding> = {}): Holding {
	return {
		id: '9f4c1b7a-2d3e-4f50-8a61-b2c3d4e5f607',
		assetId: '7c3e8b5d-1f2a-4b6c-9d0e-3a4b5c6d7e8f',
		ticker: 'MC.FR',
		name: 'LVMH Moet Hennessy Louis Vuitton SE',
		assetType: 'stock',
		exchange: 'PAR',
		currency: 'EUR',
		quantity: '0.0241',
		price: '645.30108',
		marketPrice: '429.45',
		costCurrency: 'USD',
		category: 'stocks',
		entryDate: '2024-12-05T00:00:00Z',
		notes: '',
		...overrides
	};
}

const formatAmount = (value: number, code: string) => `${code} ${value.toFixed(2)}`;

describe('asset-delete-position.svelte', () => {
	// El cascade es la parte que quien borra no puede ver venir: la posición es
	// el padre de su historial. Dar el número es la diferencia entre «se va la
	// posición» y «se van once operaciones».
	it('nombra cuántas transacciones se van con la posición', async () => {
		render(DeletePosition, {
			portfolioId: '3f1c1c5e-1f5a-4f1e-9c2a-9a1d0b2f7e11',
			entries: [holding()],
			transactionsCount: 11,
			formatAmount
		});

		await page.getByRole('button', { name: 'Eliminar' }).click();

		await expect.element(page.getByText('11 transacciones')).toBeInTheDocument();
		// La frase exacta del diálogo: la zona de peligro de detrás dice algo
		// parecido, y un localizador que casa con las dos falla por ambiguo.
		await expect.element(page.getByText('Esta acción no se puede deshacer')).toBeInTheDocument();
	});

	// Con varias entradas, `transactionsCount` cuenta las del ticker entero y no
	// las de la que se va a borrar. Dar esa cifra sería peor que no darla.
	it('no da una cifra que no describe la entrada que se borra', async () => {
		render(DeletePosition, {
			portfolioId: '3f1c1c5e-1f5a-4f1e-9c2a-9a1d0b2f7e11',
			entries: [
				holding(),
				holding({ id: '1a2b3c4d-5e6f-4708-9a0b-1c2d3e4f5061', costCurrency: 'EUR' })
			],
			transactionsCount: 11,
			formatAmount
		});

		const buttons = page.getByRole('button', { name: 'Eliminar' });
		await buttons.first().click();

		await expect
			.element(page.getByText('todas las transacciones de esta posición'))
			.toBeInTheDocument();
	});

	// Una posición por plataforma: el mismo ticker en dos brókers son dos filas
	// y se borran por separado, así que el botón se repite.
	it('ofrece un borrado por entrada', async () => {
		render(DeletePosition, {
			portfolioId: '3f1c1c5e-1f5a-4f1e-9c2a-9a1d0b2f7e11',
			entries: [
				holding(),
				holding({ id: '1a2b3c4d-5e6f-4708-9a0b-1c2d3e4f5061', costCurrency: 'EUR' })
			],
			transactionsCount: 11,
			formatAmount
		});

		await expect.element(page.getByRole('button', { name: 'Eliminar' }).nth(1)).toBeInTheDocument();
	});
});
