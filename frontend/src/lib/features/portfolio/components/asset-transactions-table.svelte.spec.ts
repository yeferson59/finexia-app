import { page } from 'vitest/browser';
import { describe, it, expect, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Table from './asset-transactions-table.svelte';
import type { Transaction } from '$lib/api/types';
import type { TxnMeta } from '../asset';

function txn(overrides: Partial<Transaction> = {}): Transaction {
	return {
		id: '66666666-6666-4666-8666-666666666601',
		entryId: '55555555-5555-4555-8555-555555555555',
		type: 'buy',
		quantity: '12',
		price: '186.90',
		currency: 'USD',
		fees: '1.20',
		transactionDate: '2026-02-20T00:00:00Z',
		notes: '',
		createdAt: '2026-02-20T00:00:00Z',
		...overrides
	};
}

const meta: TxnMeta = { total: 1, page: 1, limit: 20, totalPages: 1 };
const formatAmount = (value: number, code: string) => `${code} ${value.toFixed(2)}`;

const handlers = { onEdit: () => {}, onToggleSell: () => {}, onDelete: () => {} };

describe('asset-transactions-table.svelte', () => {
	// Era una rejilla de `<div>` con párrafos: el contenido llegaba al lector de
	// pantalla como una cadena de cifras sin nombre de columna.
	it('expone las columnas como una tabla de verdad', async () => {
		// Ancha: por debajo de 760px la fila se pliega y las cabeceras se
		// esconden a propósito, que es justo lo que este caso no mide.
		await page.viewport(1280, 800);
		render(Table, { transactions: [txn()], txnMeta: meta, formatAmount, ...handlers });

		await expect.element(page.getByRole('columnheader', { name: 'Cantidad' })).toBeInTheDocument();
		await expect.element(page.getByRole('columnheader', { name: 'Total' })).toBeInTheDocument();
		// El tipo es la cabecera de su fila, que es lo que la nombra.
		await expect.element(page.getByRole('rowheader')).toHaveTextContent('Compra');
	});

	// El total es lo que la cuenta pagó, no lo que la operación cotizó: la tasa
	// lleva precio y cantidad a la moneda de la cuenta.
	it('convierte el total a la moneda de la cuenta y dice la tasa', async () => {
		render(Table, {
			transactions: [txn({ currency: 'EUR', costCurrency: 'USD', fxRate: '1.08' })],
			txnMeta: meta,
			formatAmount,
			...handlers
		});

		await expect.element(page.getByText('USD 2422.22')).toBeInTheDocument();
		await expect.element(page.getByText('EUR × 1.08')).toBeInTheDocument();
	});

	// Los botones eran iconos de 13px cuyo único texto vivía en el `aria-label`,
	// en una tabla de filas casi idénticas y con un borrado irreversible.
	it('nombra qué fila borra cada botón', async () => {
		const onDelete = vi.fn();
		render(Table, { transactions: [txn()], txnMeta: meta, formatAmount, ...handlers, onDelete });

		const button = page.getByRole('button', { name: 'Eliminar el Compra del 20 de feb de 2026' });
		await button.click();

		expect(onDelete).toHaveBeenCalledOnce();
	});

	// Vender parte de un lote solo tiene sentido sobre lo que entró: un
	// dividendo no es una posición que se pueda deshacer.
	it('solo ofrece vender sobre un lote de entrada', async () => {
		const { rerender } = await render(Table, {
			transactions: [txn()],
			txnMeta: meta,
			formatAmount,
			...handlers
		});

		await expect.element(page.getByRole('button', { name: /^Vender/ })).toBeInTheDocument();

		await rerender({ transactions: [txn({ type: 'dividend' })] });
		await expect.element(page.getByRole('button', { name: /^Vender/ })).not.toBeInTheDocument();
	});

	// La nota se cortaba en «Dividendo trim…», que es justo la parte que no se
	// puede adivinar.
	it('escribe la nota entera bajo el tipo', async () => {
		render(Table, {
			transactions: [txn({ notes: 'Dividendo trimestral de la cuenta principal' })],
			txnMeta: meta,
			formatAmount,
			...handlers
		});

		await expect
			.element(page.getByText('Dividendo trimestral de la cuenta principal'))
			.toBeInTheDocument();
	});

	// Una ficha sin movimientos es una invitación a registrar el primero, no una
	// pantalla en blanco.
	it('invita a registrar el primer movimiento cuando no hay ninguno', async () => {
		render(Table, {
			transactions: [],
			txnMeta: { ...meta, total: 0 },
			formatAmount,
			...handlers
		});

		await expect
			.element(page.getByText(/Aún no has registrado ningún movimiento/))
			.toBeInTheDocument();
	});
});
