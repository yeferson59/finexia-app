# `lib/features` — módulos por dominio funcional

Un directorio por dominio (`auth/`, `portfolio/`, `dashboard/`, `transactions/`,
`platforms/`, `investments/`, `settings/`, `notifications/`, `admin/`,
`reports/`, `landing/`, `legal/`). Aquí vive la lógica
de negocio, los componentes de feature, los schemas Zod y el estado de cada
dominio. `routes/` **solo orquesta**: loaders/actions delgados que llaman a
`lib/api`, páginas que componen componentes de la feature.

## Anatomía de una feature

```
features/<feature>/
├── components/     # componentes propios del dominio
├── <feature>.ts    # helpers puros, constantes y tipos del dominio, si aplica
├── schemas.ts      # schemas Zod de sus formularios
├── state/          # estado del dominio (runes .svelte.ts), si aplica
└── index.ts        # superficie pública: lo único que routes/ y otras capas importan
```

No todo componente entra en `index.ts`: los que solo usa otro componente de la
misma feature (`exchange-rate-note` dentro de `wealth-headline`, `asset-combobox`
dentro de `portfolio-entry-form`, los pasos del historial de transacciones…) se
importan por ruta relativa y no forman parte de la superficie pública. Cuando
uno deja de ser interno porque una segunda feature lo necesita, no se exporta:
baja a `lib/ui`, que es lo que hizo el selector de moneda.

## Reglas de dependencia

- **Una feature no importa de otra feature.** Lo compartido baja a `lib/ui`,
  `lib/api` o `lib/shared`.
- Las features exponen `index.ts` y **solo eso** se importa desde fuera.
  Prohibido `import ... from '$lib/features/portfolio/components/x.svelte'` desde
  otra feature o desde `routes/`.
- Una feature puede importar de `lib/ui`, `lib/api` y `lib/shared`.
- **Los contratos del backend no se redeclaran**: se importan de
  `$lib/api/types` y, si a la feature le conviene, se reexportan desde su
  módulo (`platforms.ts`, `portfolio.ts`). Una copia local acaba divergiendo.
- Toda form action valida con un schema Zod de la feature
  (`features/<x>/schemas.ts`).
- Presupuesto de tamaño: una `+page.svelte` no supera ~300 líneas; ningún
  archivo de producción supera ~500. Si crece, extrae componentes aquí.

> Estado: migración cerrada (Fase 7, 2026-07-31). Son doce features: `admin`,
> `auth`, `dashboard`, `investments`, `landing`, `legal`, `notifications`,
> `platforms`, `portfolio`, `reports`, `settings` y `transactions`.
>
> Las reglas de arriba **fallan el CI** desde `eslint.config.js`, no dependen de
> que alguien las recuerde en la revisión. La foto completa de la arquitectura
> está en `docs/FRONTEND_ARCHITECTURE.md`; el porqué de cada decisión, en
> `docs/FRONTEND_ARCHITECTURE_MIGRATION.md` (retrospectiva del piloto en 3.1,
> notas de `auth` en 4.1, cierre de la Fase 5 en 5.2, reparto de CSS con
> contenedores `:global` en 6.1 y cierre en 7.1).
