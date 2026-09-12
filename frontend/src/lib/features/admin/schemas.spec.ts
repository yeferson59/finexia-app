import { describe, it, expect } from 'vitest';
import {
	assetCreateSchema,
	assetPriceSchema,
	assetUpdateSchema,
	inviteUserSchema,
	rateCreateSchema,
	rateUpdateSchema,
	rowIdSchema
} from './schemas';

/** Primer mensaje de error, que es el que la action devuelve al formulario. */
function firstError(result: {
	success: boolean;
	error?: { issues: { message: string }[] };
}): string {
	return result.success ? '' : (result.error?.issues[0].message ?? '');
}

describe('inviteUserSchema', () => {
	it('normaliza el rol y recorta los espacios', () => {
		const parsed = inviteUserSchema.parse({
			email: '  nueva@finexia.test ',
			name: ' Nueva ',
			role: ' Admin '
		});

		expect(parsed).toEqual({ email: 'nueva@finexia.test', name: 'Nueva', role: 'admin' });
	});

	it('cae a customer cuando el rol llega vacío (atajo de la lista de espera)', () => {
		expect(inviteUserSchema.parse({ email: 'a@b.test', name: '', role: '' }).role).toBe('customer');
	});

	it('exige el correo antes que nada', () => {
		expect(firstError(inviteUserSchema.safeParse({ email: '  ', name: '', role: 'sin-rol' }))).toBe(
			'El correo es requerido'
		);
	});

	it('rechaza un rol que no existe', () => {
		expect(
			firstError(inviteUserSchema.safeParse({ email: 'a@b.test', name: '', role: 'root' }))
		).toBe('Rol inválido');
	});
});

describe('rowIdSchema', () => {
	it('acepta ids que no son UUID, como los de invitación', () => {
		expect(rowIdSchema.parse('invite-1')).toBe('invite-1');
	});

	it('rechaza un id vacío', () => {
		expect(firstError(rowIdSchema.safeParse('   '))).toBe('ID requerido');
	});
});

describe('assetCreateSchema', () => {
	it('normaliza ticker y moneda a mayúsculas', () => {
		const parsed = assetCreateSchema.parse({
			ticker: ' aapl ',
			name: ' Apple Inc. ',
			assetType: 'stock',
			currency: 'usd',
			exchange: ' NASDAQ ',
			sector: ' technology '
		});

		expect(parsed).toEqual({
			ticker: 'AAPL',
			name: 'Apple Inc.',
			assetType: 'stock',
			currency: 'USD',
			exchange: 'NASDAQ',
			sector: 'technology',
			// Viaja siempre, aunque el formulario no mande ninguno: en el backend
			// un envío sin desglose es la instrucción de borrar el que hubiera.
			sectorWeights: []
		});
	});

	/*
	 * El formulario esconde el desplegable de industria cuando el tipo no la
	 * admite, así que el envío de una cripto no trae el campo. Ausente tiene que
	 * leerse como «sin clasificar» y no romper el alta: es el camino normal de
	 * la mitad del catálogo.
	 */
	it('deja la industria vacía cuando el formulario no manda el campo', () => {
		const parsed = assetCreateSchema.parse({
			ticker: 'BTC-USD',
			name: 'Bitcoin',
			assetType: 'crypto',
			currency: 'USD'
		});

		expect(parsed.sector).toBe('');
	});

	/*
	 * El sector no se valida contra una lista aquí a propósito: el backend lo
	 * normaliza («Tecnología», «Financial Services» y `technology` son el mismo
	 * sector) y rechaza lo que no reconoce. Cerrar el vocabulario en este lado
	 * dejaría fuera grafías que el servidor sí acepta, y una importación manda
	 * justamente esas.
	 */
	it('deja pasar la grafía libre de la industria', () => {
		const parsed = assetCreateSchema.parse({
			ticker: 'JPM',
			name: 'JPMorgan',
			assetType: 'stock',
			currency: 'USD',
			sector: 'Financial Services'
		});

		expect(parsed.sector).toBe('Financial Services');
	});

	it('da el mismo aviso falte el campo que falte', () => {
		const message = 'Ticker, nombre, tipo y moneda son requeridos';
		for (const missing of ['ticker', 'name', 'assetType', 'currency']) {
			const input = { ticker: 'AAPL', name: 'Apple', assetType: 'stock', currency: 'USD' };
			const result = assetCreateSchema.safeParse({ ...input, [missing]: '' });
			expect(firstError(result)).toBe(message);
		}
	});

	it('deja el exchange vacío si no se indica', () => {
		const parsed = assetCreateSchema.parse({
			ticker: 'BTC',
			name: 'Bitcoin',
			assetType: 'crypto',
			currency: 'USD'
		});

		expect(parsed.exchange).toBe('');
	});
});

describe('assetPriceSchema', () => {
	it('acepta un precio positivo escrito como texto', () => {
		expect(assetPriceSchema.parse({ id: 'a1', price: '190.00', currency: 'USD' }).price).toBe(190);
	});

	it('rechaza precios no positivos o no numéricos', () => {
		expect(firstError(assetPriceSchema.safeParse({ id: 'a1', price: '0' }))).toBe(
			'Precio inválido'
		);
		expect(firstError(assetPriceSchema.safeParse({ id: 'a1', price: 'gratis' }))).not.toBe('');
	});

	it('exige el id del activo', () => {
		expect(firstError(assetPriceSchema.safeParse({ id: '', price: '10' }))).toBe(
			'ID de activo requerido'
		);
	});

	it('cae a USD cuando la fila no trae moneda', () => {
		expect(assetPriceSchema.parse({ id: 'a1', price: '10' }).currency).toBe('USD');
	});
});

describe('assetUpdateSchema', () => {
	/** Ficha completa tal como la manda el formulario de edición. */
	const full = {
		id: 'a1',
		ticker: ' aapl ',
		name: ' Apple Inc. ',
		assetType: 'stock',
		currency: 'usd',
		exchange: ' NASDAQ ',
		sector: 'technology',
		isCurated: 'on',
		price: ' 190.50 '
	};

	it('normaliza los mismos campos que el alta y conserva el id', () => {
		expect(assetUpdateSchema.parse(full)).toEqual({
			id: 'a1',
			ticker: 'AAPL',
			name: 'Apple Inc.',
			assetType: 'stock',
			currency: 'USD',
			exchange: 'NASDAQ',
			sector: 'technology',
			sectorWeights: [],
			isCurated: true,
			price: '190.50'
		});
	});

	/*
	 * El desglose es la otra forma de clasificar, y la que hace falta para un
	 * fondo de mercado ancho. Las dos son excluyentes: el formulario desactiva
	 * la que no está en uso, y esto es la red por debajo.
	 */
	it('acepta un desglose por industrias en lugar de una industria única', () => {
		const parsed = assetUpdateSchema.parse({
			...full,
			sector: '',
			sectorWeights: [
				{ sector: 'technology', weight: '33.1' },
				{ sector: 'financials', weight: '13.8' }
			]
		});

		expect(parsed.sectorWeights).toEqual([
			{ sector: 'technology', weight: 33.1 },
			{ sector: 'financials', weight: 13.8 }
		]);
	});

	it('rechaza llevar industria única y desglose a la vez', () => {
		expect(
			firstError(
				assetUpdateSchema.safeParse({
					...full,
					sector: 'technology',
					sectorWeights: [{ sector: 'energy', weight: '40' }]
				})
			)
		).toBe('Elige una industria única o un desglose por industrias, no las dos');
	});

	// Quedarse corto está bien —la ficha de un fondo deja unas décimas en caja y
	// el reparto normaliza—, pasarse no: eso no es una transcripción incompleta
	// sino una equivocada.
	it('acepta un desglose que no llega a 100 y rechaza el que se pasa', () => {
		const short = assetUpdateSchema.safeParse({
			...full,
			sector: '',
			sectorWeights: [{ sector: 'technology', weight: '97.3' }]
		});
		expect(short.success).toBe(true);

		expect(
			firstError(
				assetUpdateSchema.safeParse({
					...full,
					sector: '',
					sectorWeights: [
						{ sector: 'technology', weight: '80' },
						{ sector: 'energy', weight: '80' }
					]
				})
			)
		).toBe('Los pesos del desglose suman más de 100 %');
	});

	// Reclasificar un activo a cripto esconde el desplegable, y entonces el
	// envío no trae `sector`: el backend lo lee como vacío y le quita la
	// industria, que es lo correcto —una cripto no puede tener una— y no un
	// descuido del formulario.
	it('lee la industria ausente como una desclasificación', () => {
		const { sector, ...withoutSector } = full;
		void sector;

		expect(assetUpdateSchema.parse(withoutSector).sector).toBe('');
	});

	// El checkbox no manda nada cuando está desmarcado, así que `null` es la
	// forma en que un formulario dice «quítalo del catálogo compartido».
	it('lee el checkbox ausente como una despublicación explícita', () => {
		expect(assetUpdateSchema.parse({ ...full, isCurated: null }).isCurated).toBe(false);
		expect(assetUpdateSchema.parse({ ...full, isCurated: undefined }).isCurated).toBe(false);
	});

	it('admite el precio en blanco, que deja el guardado como está', () => {
		expect(assetUpdateSchema.parse({ ...full, price: '  ' }).price).toBe('');
	});

	it('conserva el precio como texto, con sus decimales de cola', () => {
		expect(assetUpdateSchema.parse({ ...full, price: '190.00' }).price).toBe('190.00');
	});

	it('rechaza un precio que no es un número positivo', () => {
		expect(firstError(assetUpdateSchema.safeParse({ ...full, price: '0' }))).toBe(
			'Precio inválido'
		);
		expect(firstError(assetUpdateSchema.safeParse({ ...full, price: 'gratis' }))).toBe(
			'Precio inválido'
		);
	});

	it('exige el id de la fila que se está editando', () => {
		expect(firstError(assetUpdateSchema.safeParse({ ...full, id: '  ' }))).toBe('ID requerido');
	});

	it('da el mismo aviso del alta falte el campo que falte', () => {
		for (const missing of ['ticker', 'name', 'assetType', 'currency']) {
			const result = assetUpdateSchema.safeParse({ ...full, [missing]: '' });
			expect(firstError(result)).toBe('Ticker, nombre, tipo y moneda son requeridos');
		}
	});
});

describe('rateCreateSchema', () => {
	it('normaliza las dos monedas a mayúsculas', () => {
		const parsed = rateCreateSchema.parse({
			fromCurrency: 'usd',
			toCurrency: ' cop ',
			rate: '4000'
		});

		expect(parsed).toEqual({ fromCurrency: 'USD', toCurrency: 'COP', rate: '4000' });
	});

	it('da el mismo aviso falte el campo que falte', () => {
		expect(
			firstError(rateCreateSchema.safeParse({ fromCurrency: 'USD', toCurrency: '', rate: '4000' }))
		).toBe('Moneda origen, destino y tasa son requeridos');
	});
});

describe('rateUpdateSchema', () => {
	it('rechaza una tasa no positiva', () => {
		expect(firstError(rateUpdateSchema.safeParse({ id: 'r1', rate: '-2' }))).toBe('Tasa inválida');
	});

	it('exige el id de la tasa', () => {
		expect(firstError(rateUpdateSchema.safeParse({ id: '', rate: '4000' }))).toBe(
			'ID de tasa requerido'
		);
	});
});
