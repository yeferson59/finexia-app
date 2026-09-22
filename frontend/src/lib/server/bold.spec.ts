import { describe, expect, it, vi } from 'vitest';

vi.mock('$env/dynamic/private', () => ({ env: {} }));

const { createCheckout, fetchPayment, integritySignature, isOrderId, newOrderId } =
	await import('./bold');

const keys = { apiKey: 'identidad', secretKey: 'secreta' };

describe('integritySignature', () => {
	it('reproduce el ejemplo de la documentación de Bold', () => {
		expect(integritySignature('inv0334', 39400, 'COP', 'kgfq2nN0o52XqnuXZWIN2F')).toBe(
			'620a64c6eab8858d0f96d4f818a1d77be5e9b9eb9dc681f527de1af54fc1b739'
		);
	});
});

describe('newOrderId / isOrderId', () => {
	it('genera ids que cumplen el formato de Bold y el nuestro', () => {
		const id = newOrderId();
		expect(id.length).toBeLessThanOrEqual(60);
		expect(id).toMatch(/^[A-Za-z0-9_-]+$/);
		expect(isOrderId(id)).toBe(true);
	});

	it('rechaza lo que no es una orden nuestra', () => {
		expect(isOrderId(null)).toBe(false);
		expect(isOrderId('inv0334')).toBe(false);
		expect(isOrderId('FNX-APOYO-1/../../x')).toBe(false);
	});
});

describe('createCheckout', () => {
	it('firma la orden con la llave secreta sin exponerla', () => {
		const checkout = createCheckout(keys, 20000, 'https://finexia.me/apoyar');
		expect(checkout).toMatchObject({
			currency: 'COP',
			amount: '20000',
			apiKey: 'identidad',
			redirectionUrl: 'https://finexia.me/apoyar'
		});
		expect(checkout.integritySignature).toBe(
			integritySignature(checkout.orderId, 20000, 'COP', 'secreta')
		);
		expect(JSON.stringify(checkout)).not.toContain('secreta');
	});
});

describe('fetchPayment', () => {
	const orderId = 'FNX-APOYO-1790000000000-0a1b2c3d';

	it('consulta el comprobante con la llave de identidad', async () => {
		const fetch = vi.fn(
			async () => new Response(JSON.stringify({ payment_status: 'APPROVED', total: 20000 }))
		);
		expect(await fetchPayment(fetch, keys, orderId)).toEqual({ status: 'APPROVED', total: 20000 });
		expect(fetch).toHaveBeenCalledWith(
			`https://payments.api.bold.co/v2/payment-voucher/${orderId}`,
			expect.objectContaining({ headers: { Authorization: 'x-api-key identidad' } })
		);
	});

	it('devuelve null si Bold falla o responde algo desconocido', async () => {
		const down = vi.fn(async () => new Response('', { status: 500 }));
		const odd = vi.fn(async () => new Response(JSON.stringify({ payment_status: 'RARO' })));
		const offline = vi.fn(async () => {
			throw new Error('red');
		});
		expect(await fetchPayment(down, keys, orderId)).toBeNull();
		expect(await fetchPayment(odd, keys, orderId)).toBeNull();
		expect(await fetchPayment(offline, keys, orderId)).toBeNull();
	});
});
