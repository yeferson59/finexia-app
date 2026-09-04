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
- `schemas/` — los contratos como schemas Zod, mantenidos a mano contra
  `docs/API.md`. **Fuente de verdad de los shapes de la API.** Repartidos por
  dominio (`pagination`, `portfolio`, `transactions`, `platforms`, `market`,
  `user`) desde que el archivo único se pasó del presupuesto de 500 líneas que
  comprueba `check:arch`; se importan siempre por la carpeta
  (`$lib/api/schemas`), que es la puerta de entrada.
- `types.ts` — los tipos, derivados de esos schemas con `z.infer`
  (`PageMeta`, `PortfolioSummary`, `Holding`, `Transaction`, `Asset`, …), más el
  sobre genérico `ApiEnvelope`. Es el módulo del que todo el mundo importa los
  tipos; los schemas solo los necesita quien valide en tiempo de ejecución.
- Módulos por dominio (`auth.ts`, `portfolio.ts`, `transactions.ts`,
  `platforms.ts`, `market.ts`, `user.ts`, `marketing.ts`): funciones tipadas que
  encapsulan `path + método + parseo` y devuelven `ApiResult<T>` (o la `Response`
  cruda para streams/proxies y los flujos públicos de `auth`/`marketing`).

## Validación del contrato en desarrollo

Cada lectura pasa su schema a `apiRequest`/`apiRequestSafe`. En `dev` —y solo
ahí: `import.meta.env.DEV` lo elimina del build— se valida el `data` recibido y
se avisa por consola con la ruta y el campo que falla. El resultado no cambia:
es un aviso, no un corte. Sin esto, un backend que se sale del contrato no
rompe nada hasta que un `undefined` llega a un componente.

Las fixtures del stub de e2e se validan contra estos mismos schemas en
`e2e/mocks/contract.spec.ts`.

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
