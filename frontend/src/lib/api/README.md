# `lib/api` — capa de acceso al backend (server-only)

Única puerta al backend HTTP. Ningún loader/action de `routes/` debe llamar a
`authedFetch` directamente ni construir paths (`'/portfolios/' + id`) fuera de
aquí.

## Contenido

- `client.ts` — `authedFetch` / `authedFetchSafe`: auth, refresh single-flight y
  redirección a `/auth`. Además `apiUrl()` (única lectura de `env.BASE_API`),
  `apiRequest`/`apiRequestSafe` y el tipo `ApiResult<T>` (vista tipada y plana
  del response: `ok`, `status`, `success`, `data`, `message`, `details`,
  `action`) sobre los que se construyen los módulos de dominio.
- `types.ts` — contratos compartidos con el backend, espejo de `docs/API.md`.
  Única fuente de verdad de los shapes de la API (`ApiEnvelope`, `PageMeta`,
  `Paginated`, `PortfolioSummary`, `Holding`, `Transaction`, `Asset`, …).
- Módulos por dominio (`auth.ts`, `portfolio.ts`, `transactions.ts`,
  `platforms.ts`, `market.ts`, `user.ts`, `marketing.ts`): funciones tipadas que
  encapsulan `path + método + parseo` y devuelven `ApiResult<T>` (o la `Response`
  cruda para streams/proxies y los flujos públicos de `auth`/`marketing`).

## Convención de retorno

- **Lecturas** (GET): `ApiResult<T>` vía `apiRequestSafe` (degradan con
  `ok: false` si el backend no responde; un 401 sigue redirigiendo a `/auth`).
- **Comandos** (POST/PATCH/PUT/DELETE): `ApiResult<T>` vía `apiRequest`.
- **Streams/proxies** (exports XLSX, import preview/commit, combobox de assets) y
  **flujos públicos** (`auth`, `marketing`): devuelven la `Response` cruda.

## Reglas de dependencia

- `lib/api` **no** importa de `lib/features` ni de `lib/ui`.
- Puede importar de `lib/server` (sesión) y de `lib/shared`.
- Cada función devuelve datos tipados con los tipos de `types.ts`.
- Validación Zod opcional de los responses críticos en dev.

> Estado: Fase 2 completa — capa de API tipada por dominio. Todos los
> loaders/actions de `routes/` consumen estos módulos; ninguno importa
> `$lib/server/api` (eliminado) ni construye paths/`BASE_API` a mano.
