/**
 * Feature `funds` — superficie pública.
 *
 * Los fondos de inversión (`routes/dashboard/funds`): fondos de inversión
 * colectiva, de pensiones voluntarias, money market. No rinden a una tasa como
 * el efectivo: la entidad publica el valor de la unidad y la rentabilidad es
 * cómo se movió. Son posiciones de unidades cuyo valor escribe el dueño desde
 * el extracto, como marcas por fecha (000057).
 *
 * Se siguen por unidades (el extracto trae unidades y valor de unidad) o por
 * saldo (la app solo enseña el saldo, y las unidades las lleva Finexia por
 * dentro, 000058).
 *
 * `fund-list` los enseña como tarjetas, `fund-create-form` registra uno con su
 * primera compra, `fund-mark-form` anota el valor de unidad —o el saldo— de un
 * día y lista los que ya hay, y `fund-movement-form` anota los aportes y
 * retiros de un fondo por saldo. `fund-performance` enseña la rentabilidad por
 * periodo, el dinero y la gráfica (`fund-chart`, interno, con la geometría de
 * `chart.ts`); `fund-marks-paste`, interno de `fund-mark-form`, lee la tabla que
 * se pega desde un extracto (`performance.ts`). `fund-link`, interno de
 * `fund-mark-form`, enlaza un fondo por unidades al que publica la
 * Superintendencia Financiera (000059), y `public-fund-picker` busca en ese
 * catálogo, también desde `fund-create-form`. `funds.ts` aporta los helpers puros y reexporta los contratos
 * `Fund` y `FundMark` de `$lib/api/types`.
 */
export { default as FundList } from './components/fund-list.svelte';
export { default as FundCreateForm } from './components/fund-create-form.svelte';
export { default as FundMarkForm } from './components/fund-mark-form.svelte';
export { default as FundMovementForm } from './components/fund-movement-form.svelte';
export { default as FundPerformanceView } from './components/fund-performance.svelte';

export * from './funds';
export * from './schemas';
export * from './performance';
