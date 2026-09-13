import { page } from 'vitest/browser';
import { describe, it, expect, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import ModalHarness from './modal.test-harness.svelte';

const dialogEl = () => document.querySelector('dialog');

describe('modal.svelte', () => {
	// Un `<dialog>` cerrado sigue en el DOM: si el contenido viviera dentro sin
	// guardia, cada pantalla montaría todos sus formularios en cada carga.
	it('no monta el contenido mientras `open` es falso', async () => {
		render(ModalHarness, { open: false, title: 'Nuevo activo' });

		expect(dialogEl()?.open).toBe(false);
		await expect.element(page.getByText('contenido del modal')).not.toBeInTheDocument();
	});

	it('abre el diálogo y lo nombra con su título', async () => {
		render(ModalHarness, { open: true, title: 'Nuevo activo' });

		await expect.element(page.getByRole('dialog', { name: 'Nuevo activo' })).toBeVisible();
		await expect.element(page.getByText('contenido del modal')).toBeVisible();
	});

	// `showModal()` deja inerte el resto de la página; sin él el fondo seguía
	// siendo tabulable por detrás del formulario.
	it('se abre en modo modal, no como diálogo suelto', async () => {
		render(ModalHarness, { open: true, title: 'Nuevo activo' });

		expect(dialogEl()?.matches(':modal')).toBe(true);
	});

	it('avisa al padre cuando se pulsa la X', async () => {
		const onClose = vi.fn();
		render(ModalHarness, { open: true, title: 'Nuevo activo', onClose });

		await page.getByRole('button', { name: 'Cerrar' }).click();

		expect(onClose).toHaveBeenCalled();
	});

	// Lo trae el `<dialog>` nativo: las copias anteriores no cerraban con Escape,
	// y el `close` que dispara al hacerlo es el que devuelve el estado al padre.
	it('avisa al padre cuando el diálogo se cierra solo', async () => {
		const onClose = vi.fn();
		render(ModalHarness, { open: true, title: 'Nuevo activo', onClose });

		dialogEl()?.dispatchEvent(new Event('close'));

		expect(onClose).toHaveBeenCalled();
	});

	it('bloquea el scroll de la página que queda detrás', async () => {
		render(ModalHarness, { open: true, title: 'Nuevo activo' });

		expect(document.body.style.overflow).toBe('hidden');
	});

	it('pinta la descripción cuando se le pasa una', async () => {
		render(ModalHarness, {
			open: true,
			title: 'Importar activos',
			description: 'Se admite .csv, .xlsx y .xls.'
		});

		await expect.element(page.getByText('Se admite .csv, .xlsx y .xls.')).toBeVisible();
	});

	// La descripción es la mitad de lo que se decide —«la usan todas las
	// cuentas»—, así que el lector la anuncia junto al título al abrir.
	it('describe el diálogo con su descripción', async () => {
		render(ModalHarness, {
			open: true,
			title: 'Invitar a alguien',
			description: 'Le enviaremos un enlace de un solo uso.'
		});

		await expect
			.element(page.getByRole('dialog', { name: 'Invitar a alguien' }))
			.toHaveAccessibleDescription('Le enviaremos un enlace de un solo uso.');
	});

	it('sin descripción no apunta a un elemento que no existe', async () => {
		render(ModalHarness, { open: true, title: 'Nuevo activo' });

		expect(dialogEl()?.hasAttribute('aria-describedby')).toBe(false);
	});

	// El tono es lo que pinta el filete de rojo en las confirmaciones de borrado.
	it('expone el tono para que el estilo lo distinga', async () => {
		render(ModalHarness, { open: true, title: 'Eliminar transacción', tone: 'danger' });

		expect(dialogEl()?.dataset.tone).toBe('danger');
	});
});
