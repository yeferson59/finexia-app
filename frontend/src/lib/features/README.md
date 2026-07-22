# `lib/features` — módulos por dominio funcional

Un directorio por dominio (`auth/`, `portfolio/`, `dashboard/`, `transactions/`,
`platforms/`, `settings/`, `admin/`, `landing/`, `legal/`). Aquí vive la lógica
de negocio, los componentes de feature, los schemas Zod y el estado de cada
dominio. `routes/` **solo orquesta**: loaders/actions delgados que llaman a
`lib/api`, páginas que componen componentes de la feature.

## Anatomía de una feature

```
features/<feature>/
├── components/     # componentes propios del dominio
├── schemas.ts      # schemas Zod de sus formularios
├── state/          # estado del dominio (runes .svelte.ts), si aplica
└── index.ts        # superficie pública: lo único que routes/ y otras capas importan
```

## Reglas de dependencia

- **Una feature no importa de otra feature.** Lo compartido baja a `lib/ui`,
  `lib/api` o `lib/shared`.
- Las features exponen `index.ts` y **solo eso** se importa desde fuera.
  Prohibido `import ... from '$lib/features/portfolio/components/x.svelte'` desde
  otra feature o desde `routes/`.
- Una feature puede importar de `lib/ui`, `lib/api` y `lib/shared`.
- Toda form action valida con un schema Zod de la feature
  (`features/<x>/schemas.ts`).
- Presupuesto de tamaño: una `+page.svelte` no supera ~300 líneas; ningún
  archivo de producción supera ~500. Si crece, extrae componentes aquí.

> Estado: Fase 4 completada — `landing`, `legal` y `auth` ya migradas. El resto
> de features se pueblan en las Fases 5–6 siguiendo el patrón validado en la
> retrospectiva del piloto (sección 3.1 de
> `docs/FRONTEND_ARCHITECTURE_MIGRATION.md`; ver también las notas de `auth` en
> la sección 4.1).
