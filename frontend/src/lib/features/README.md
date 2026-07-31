# `lib/features` — módulos por dominio funcional

Un directorio por dominio (`auth/`, `portfolio/`, `dashboard/`, `transactions/`,
`platforms/`, `investments/`, `settings/`, `admin/`, `landing/`, `legal/`). Aquí vive la lógica
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
misma feature (`currency-toggle` dentro de `net-worth-card`, `asset-combobox`
dentro de `portfolio-entry-form`, los pasos del historial de transacciones…) se
importan por ruta relativa y no forman parte de la superficie pública.

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

> Estado: Fases 3 a 6 completas — las diez features están migradas (`landing`,
> `legal`, `auth`, `dashboard`, `platforms`, `portfolio`, `transactions`,
> `investments`, `settings` y `admin`). Queda la Fase 7: borrar los alias
> heredados, blindar las fronteras con ESLint y trocear
> `dashboard/reports/+page.svelte`, la única página que no entró en ninguna
> feature. Las features siguen el patrón validado en la retrospectiva del piloto
> (sección 3.1 de `docs/FRONTEND_ARCHITECTURE_MIGRATION.md`; ver también las
> notas de `auth` en la 4.1, el cierre de la Fase 5 en la 5.2 y el reparto de
> CSS con contenedores `:global` de la 6.1).
