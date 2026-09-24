/**
 * Feature `funds` — superficie pública.
 *
 * Los fondos de inversión (`routes/dashboard/funds`): fondos de inversión
 * colectiva, de pensiones voluntarias, money market. No rinden a una tasa como
 * el efectivo: la entidad publica el valor de la unidad y la rentabilidad es
 * cómo se movió. Son posiciones de unidades cuyo valor escribe el dueño desde
 * el extracto, como marcas por fecha (000057).
 *
 * `fund-list` los enseña como tarjetas, `fund-create-form` registra uno con su
 * primera compra y `fund-mark-form` anota el valor de unidad de un día y lista
 * los que ya hay. `funds.ts` aporta los helpers puros y reexporta los contratos
 * `Fund` y `FundMark` de `$lib/api/types`.
 */
export { default as FundList } from './components/fund-list.svelte';
export { default as FundCreateForm } from './components/fund-create-form.svelte';
export { default as FundMarkForm } from './components/fund-mark-form.svelte';

export * from './funds';
export * from './schemas';
