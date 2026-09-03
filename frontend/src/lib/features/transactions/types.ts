/**
 * Tipos del wizard de importación de transacciones. Espejo del contrato que
 * devuelven los endpoints `import/preview` e `import/commit`.
 */

export interface ImportMapping {
	date: number | null;
	type: number | null;
	ticker: number | null;
	assetName: number | null;
	quantity: number | null;
	price: number | null;
	fees: number | null;
	currency: number | null;
	/** Tasa a la que esa fila liquidó en la moneda de la cuenta. */
	fxRate: number | null;
	category: number | null;
	notes: number | null;
}

export interface ImportRow {
	rowNumber: number;
	raw: string[];
	date: string;
	type: string;
	ticker: string;
	assetName: string;
	quantity: string;
	price: string;
	fees: string;
	currency: string;
	fxRate: string;
	/** La de la cuenta, que sale de los valores por defecto, no de la fila. */
	costCurrency: string;
	category: string;
	notes: string;
	valid: boolean;
	errors: string[];
}

export interface ImportPreview {
	sheets: string[];
	sheet: string;
	headerRow: number;
	headers: string[];
	suggestedMapping: ImportMapping;
	missingFields: string[];
	totalRows: number;
	validRows: number;
	invalidRows: number;
	rows: ImportRow[];
}

export interface ImportResult {
	totalRows: number;
	imported: number;
	skipped: number;
	errors: { row: number; message: string }[];
}

export interface ImportDefaults {
	type: string;
	currency: string;
	/**
	 * Moneda de la cuenta: en la que el bróker debitó, y en la que quedará el
	 * coste de toda posición que abra esta importación.
	 *
	 * Va en los valores por defecto y no en una columna porque un extracto es
	 * una cuenta: su moneda de liquidación es del archivo, no de la fila. Vacía
	 * significa «la misma que la de cada fila», que es lo que hacía toda
	 * importación anterior a la tasa por transacción.
	 */
	costCurrency: string;
	category: string;
	dateFormat: string;
}

export type ImportStep = 'upload' | 'map' | 'done';

/** Portafolio/plataforma seleccionables como destino de la importación. */
export interface ImportPortfolioOption {
	id: string;
	name: string;
	baseCurrency: string;
	isDefault?: boolean;
}

export interface ImportPlatformOption {
	id: string;
	name: string;
}

export const emptyMapping: ImportMapping = {
	date: null,
	type: null,
	ticker: null,
	assetName: null,
	quantity: null,
	price: null,
	fees: null,
	currency: null,
	fxRate: null,
	category: null,
	notes: null
};
