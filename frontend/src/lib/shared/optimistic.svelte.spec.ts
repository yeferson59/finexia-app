import { describe, it, expect, vi } from 'vitest';
import type { ActionResult } from '@sveltejs/kit';

const applyAction = vi.fn();
const invalidateAll = vi.fn();
vi.mock('$app/forms', () => ({ applyAction }));
vi.mock('$app/navigation', () => ({ invalidateAll }));

const { OptimisticList, optimisticSubmit, submitError, syncing } =
	await import('./optimistic.svelte');

type Row = { id: string; name: string };

const server: Row[] = [
	{ id: 'a', name: 'uno' },
	{ id: 'b', name: 'dos' }
];

describe('OptimisticList — lo pendiente sobre la lista del servidor', () => {
	it('pone arriba una fila nueva y la retira al deshacer', () => {
		const list = new OptimisticList<Row>();
		const undo = list.add({ id: 'p', name: 'nueva' });

		expect(list.view(server).map((r) => r.id)).toEqual(['p', 'a', 'b']);
		expect(list.isPending('p')).toBe(true);
		expect(list.delta).toBe(1);

		undo();
		expect(list.view(server)).toEqual(server);
		expect(list.delta).toBe(0);
	});

	it('cambia una fila sin tocar la del servidor', () => {
		const list = new OptimisticList<Row>();
		const undo = list.patch('b', { name: 'editada' });

		expect(list.view(server)[1]).toEqual({ id: 'b', name: 'editada' });
		expect(server[1].name).toBe('dos');
		expect(list.isPending('b')).toBe(true);

		undo();
		expect(list.view(server)[1].name).toBe('dos');
	});

	it('un deshacer viejo no pisa una edición posterior de la misma fila', () => {
		const list = new OptimisticList<Row>();
		const first = list.patch('a', { name: 'primera' });
		list.patch('a', { name: 'segunda' });

		first();
		expect(list.view(server)[0].name).toBe('segunda');
	});

	it('esconde una fila borrada y la devuelve si el borrado falla', () => {
		const list = new OptimisticList<Row>();
		const undo = list.remove('a');

		expect(list.view(server).map((r) => r.id)).toEqual(['b']);
		expect(list.delta).toBe(-1);

		undo();
		expect(list.view(server).map((r) => r.id)).toEqual(['a', 'b']);
	});
});

describe('submitError — cuándo un envío no se hizo', () => {
	it('acepta un éxito', () => {
		expect(submitError({ type: 'success', status: 200, data: { success: true } }, 'x')).toBeNull();
		expect(submitError({ type: 'success', status: 200 }, 'x')).toBeNull();
	});

	it('lee el motivo de un fail()', () => {
		expect(submitError({ type: 'failure', status: 400, data: { error: 'Sin saldo' } }, 'x')).toBe(
			'Sin saldo'
		);
		expect(submitError({ type: 'failure', status: 400 }, 'x')).toBe('x');
	});

	it('trata un 200 con success:false como rechazo', () => {
		expect(
			submitError({ type: 'success', status: 200, data: { success: false, error: 'No' } }, 'x')
		).toBe('No');
		expect(submitError({ type: 'success', status: 200, data: { success: false } }, 'x')).toBe('x');
	});

	it('un error inesperado usa el texto de reserva', () => {
		expect(submitError({ type: 'error', error: new Error('boom') }, 'x')).toBe('x');
	});
});

describe('optimisticSubmit — pintar al pulsar, confirmar después', () => {
	type Input = Parameters<ReturnType<typeof optimisticSubmit>>[0];
	type Callback = (opts: { result: ActionResult }) => Promise<void>;

	/* Lo que `use:enhance` le pasa al enviar; solo se lee `formData`. */
	function send(submit: ReturnType<typeof optimisticSubmit>, result: ActionResult) {
		const callback = submit({ formData: new FormData() } as Input) as Callback;
		return callback({ result });
	}

	it('pinta antes de la respuesta y lo retira cuando llega lo real', async () => {
		invalidateAll.mockClear();
		const undo = vi.fn();
		const apply = vi.fn(() => undo);
		const onSuccess = vi.fn();
		const submit = optimisticSubmit({ fallbackError: 'x', apply, onSuccess });

		const callback = submit({ formData: new FormData() } as Input) as Callback;
		expect(apply).toHaveBeenCalledOnce();
		expect(syncing.active).toBe(true);

		await callback({ result: { type: 'success', status: 200, data: { success: true } } });
		expect(onSuccess).toHaveBeenCalledOnce();
		expect(invalidateAll).toHaveBeenCalledOnce();
		expect(undo).toHaveBeenCalledOnce();
		expect(syncing.active).toBe(false);
	});

	it('un rechazo deshace lo pintado, no recarga y dice el motivo', async () => {
		invalidateAll.mockClear();
		const undo = vi.fn();
		const onError = vi.fn();
		const submit = optimisticSubmit({ fallbackError: 'x', apply: () => undo, onError });

		await send(submit, { type: 'failure', status: 409, data: { error: 'Tiene posiciones' } });
		expect(undo).toHaveBeenCalledOnce();
		expect(onError).toHaveBeenCalledWith('Tiene posiciones');
		expect(invalidateAll).not.toHaveBeenCalled();
	});

	it('con syncForm pasa el resultado a `form`, pero nunca un error', async () => {
		applyAction.mockClear();
		const submit = optimisticSubmit({ fallbackError: 'x', syncForm: true, apply: () => {} });

		const failure: ActionResult = { type: 'failure', status: 400, data: { banError: 'No' } };
		await send(submit, failure);
		expect(applyAction).toHaveBeenCalledWith(failure);

		applyAction.mockClear();
		await send(submit, { type: 'error', error: new Error('boom') });
		expect(applyAction).not.toHaveBeenCalled();
		// Nadie lo enseña en la pantalla: lo dice la cabecera.
		expect(syncing.error).toBe('x');
		syncing.dismiss();
	});
});
