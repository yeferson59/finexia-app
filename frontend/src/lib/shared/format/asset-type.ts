/**
 * Etiquetas y colores del `assetType` de un activo.
 *
 * El vocabulario es el de `market.AssetType` del backend —singular:
 * `stock`, `etf`, `crypto`…—, que es lo que hablan tanto los holdings
 * (`GET /portfolios/:id`) como el reparto del dashboard
 * (`GET /portfolios/allocation`), desde que este último agrupa por
 * `assets.asset_type` en vez de por la columna copiada `portfolio_entries.category`.
 *
 * Vivía duplicado: `features/portfolio` lo tenía bien y `features/dashboard`
 * guardaba una segunda copia tecleada con el vocabulario *plural* de
 * `portfolio.type` (`stocks`, `etfs`, `bonds`). Esa copia dejó de acertar
 * ningún lookup, así que el donut del dashboard pintaba el nombre crudo del
 * backend y mandaba todas las porciones al color de reserva. Una sola tabla es
 * lo que impide que las dos gráficas vuelvan a discrepar sobre la misma
 * posición.
 *
 * No confundir con `portfolio-type.ts`: aquel describe el `type` de un
 * portafolio (plural, y con combinaciones como `stocks_etfs`), que es otra cosa.
 */

/** Clases de activo que devuelve el backend (`market.AssetType`). */
export const ASSET_TYPE_LABELS: Record<string, string> = {
	stock: 'Acciones',
	etf: 'ETFs',
	crypto: 'Cripto',
	bond: 'Bonos',
	cash: 'Efectivo',
	real_estate: 'Inmobiliario',
	commodity: 'Materias primas',
	other: 'Otros'
};

export const ASSET_TYPE_COLORS: Record<string, string> = {
	stock: '#d4912a',
	etf: '#22c97e',
	crypto: '#6b8cef',
	bond: '#b988e0',
	cash: '#8a8780',
	real_estate: '#e0885a',
	commodity: '#e0c15a',
	other: '#5ab4e0'
};

/** Color de reserva de una clase de activo que el backend añada y aquí no esté. */
export const ASSET_TYPE_FALLBACK_COLOR = '#5ab4e0';

/**
 * Etiqueta legible de una clase de activo. Una desconocida conserva su nombre
 * crudo en vez de desaparecer del gráfico.
 */
export function formatAssetType(type: string): string {
	return ASSET_TYPE_LABELS[type] ?? type;
}

/** Color de una clase de activo, con reserva para las que no estén en la tabla. */
export function assetTypeColor(type: string): string {
	return ASSET_TYPE_COLORS[type] ?? ASSET_TYPE_FALLBACK_COLOR;
}
