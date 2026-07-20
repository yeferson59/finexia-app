# Línea base — Fase 0 de la migración del frontend

> Snapshot tomado el **2026-07-20** sobre `frontend/src/`, como exige la Fase 0 de
> [`FRONTEND_ARCHITECTURE_MIGRATION.md`](./FRONTEND_ARCHITECTURE_MIGRATION.md).
> La Fase 7 se valida contra estos números: **cero** loaders con `authedFetch`
> crudo y **ningún** archivo de producción > ~500 líneas.

## 1. Archivos de producción > 500 líneas (13)

Comando:

```bash
find src -type f \( -name '*.svelte' -o -name '*.ts' \) \
  ! -name '*.spec.ts' ! -name '*.test.ts' -exec wc -l {} + \
  | awk '$1>500 && $2!="total"' | sort -rn
```

| Líneas | Archivo |
|---:|---|
| 2014 | `src/routes/dashboard/portfolios/[id]/assets/[symbol]/+page.svelte` |
| 1377 | `src/components/auth/login-register.svelte` |
| 1309 | `src/routes/dashboard/portfolios/[id]/+page.svelte` |
| 1146 | `src/routes/dashboard/settings/+page.svelte` |
| 1007 | `src/routes/dashboard/transactions/import/+page.svelte` |
| 996 | `src/routes/dashboard/portfolios/[id]/add/+page.svelte` |
| 782 | `src/routes/dashboard/platforms/[id]/+page.svelte` |
| 714 | `src/routes/dashboard/investments/[id]/+page.svelte` |
| 671 | `src/routes/dashboard/admin/assets/+page.svelte` |
| 607 | `src/routes/dashboard/admin/exchange-rates/+page.svelte` |
| 604 | `src/routes/dashboard/portfolios/add/+page.svelte` |
| 595 | `src/routes/dashboard/admin/users/+page.svelte` |
| 568 | `src/components/landing-page/hero.svelte` |

## 2. Acceso crudo al backend desde `routes/`

- **24 de 34** archivos de servidor en `src/routes/` (71 %) importan
  `authedFetch`/`authedFetchSafe` de `$lib/server/api` directamente
  (`grep -rln "lib/server/api" src/routes`). La Fase 2 debe dejar este número
  en **0** (todo pasa por `lib/api/<dominio>`).
- Además, **7** archivos de `src/routes/` construyen URLs con `env.BASE_API` a
  mano (las actions de `auth/**`, `api/waitlist` y el logout de
  `dashboard/+page.server.ts`); también deben migrar a `lib/api` en la Fase 2.

## 3. Red de seguridad creada en esta fase

- **CI**: `.github/workflows/frontend-ci.yml` ejecuta `check`, `lint`,
  `test:unit` y `test:e2e` en cada PR que toque `frontend/**`.
- **E2E**: la suite pasó de 1 archivo (`landing.e2e.ts`) a 8. Los flujos
  autenticados corren contra un stub del contrato HTTP (`docs/API.md`) en
  `frontend/e2e/mocks/mock-api.mjs`, arrancado por Playwright junto a la app
  (`BASE_API` apunta al stub):
  - `auth.e2e.ts` — login, logout, credenciales inválidas y redirección de
    rutas protegidas a `/auth`.
  - `dashboard.e2e.ts` — el dashboard renderiza sus widgets con datos.
  - `portfolio.e2e.ts` — detalle de portfolio + añadir entrada (formulario
    completo con combobox de activos).
  - `transactions.e2e.ts` — listado + wizard de import hasta el preview.
  - `settings.e2e.ts` — cambio de nombre de perfil y secciones de seguridad.
  - `admin.e2e.ts` — listado de usuarios con rol admin y rechazo a no-admin.
  - `seo.e2e.ts` — snapshot de title/description/canonical/OG/robots de la
    landing y páginas legales, y `X-Robots-Tag` de las áreas privadas.
- **Contrato**: `docs/API.md` existe (Fase 0 del backend, 2026-07-13) y cubre
  todos los endpoints que consume el frontend; el stub de e2e se escribió
  contra ese documento.
