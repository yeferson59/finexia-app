# `lib/api` — capa de acceso al backend (server-only)

Única puerta al backend HTTP. Ningún loader/action de `routes/` debe llamar a
`authedFetch` directamente ni construir paths (`'/portfolios/' + id`) fuera de
aquí.

## Contenido

- `client.ts` — `authedFetch` / `authedFetchSafe`: auth, refresh single-flight y
  redirección a `/auth`. Movido desde `lib/server/api.ts` **sin cambios de
  lógica**. Es la base sobre la que se construyen los módulos de dominio.
- (Fase 2) `types.ts` — contratos compartidos con el backend, espejo de
  `docs/API.md`. Única fuente de verdad de los shapes de la API.
- (Fase 2) módulos por dominio (`auth.ts`, `portfolio.ts`, `transactions.ts`,
  `platforms.ts`, `market.ts`, `user.ts`, `marketing.ts`): funciones tipadas que
  encapsulan `path + método + parseo` y devuelven tipos de `types.ts`.

## Reglas de dependencia

- `lib/api` **no** importa de `lib/features` ni de `lib/ui`.
- Puede importar de `lib/server` (sesión) y de `lib/shared`.
- Cada función devuelve datos tipados con los tipos de `types.ts`.
- Validación Zod opcional de los responses críticos en dev.

> Estado: Fase 1 — solo existe `client.ts` (más un re-export temporal en
> `lib/server/api.ts`). Los módulos de dominio y `types.ts` llegan en la Fase 2.
