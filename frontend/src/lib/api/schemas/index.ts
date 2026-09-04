/**
 * Contratos HTTP del backend, como schemas Zod.
 *
 * **Fuente de verdad de los shapes de la API**: los tipos de `types.ts` se
 * derivan de aquí con `z.infer`, así que un contrato se escribe una sola vez.
 * Se mantienen a mano contra `docs/API.md`.
 *
 * Que sean schemas y no interfaces tiene un motivo: en desarrollo, la capa de
 * API valida con ellos lo que responde el backend (ver `client.ts`), de modo
 * que una divergencia deja un aviso en consola en vez de propagarse como un
 * `undefined` tres componentes más abajo. En producción no se ejecutan.
 *
 * Era un solo `schemas.ts` que se pasó del presupuesto de 500 líneas que
 * comprueba `check:arch`. Se repartió por dominio siguiendo las secciones que
 * ya tenía dentro; este barril mantiene `$lib/api/schemas` como la única puerta
 * de entrada, así que ningún sitio que los importa tuvo que cambiar.
 */
export * from './pagination';
export * from './portfolio';
export * from './transactions';
export * from './platforms';
export * from './market';
export * from './user';
