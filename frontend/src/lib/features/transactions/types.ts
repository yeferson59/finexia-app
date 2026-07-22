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
	category: null,
	notes: null
};
