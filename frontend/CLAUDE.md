# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Contexto

`frontend/` es la app SvelteKit (Svelte 5, runes forzadas, Tailwind 4,
adapter-node) del monorepo Finexia; `backend/` es la API en Go y `docs/` la
documentación común. El frontend es la **única puerta al backend**: el backend
vive en una red privada y solo se alcanza a través de esta app.

Documentación, comentarios y UI están en español (es-CO). Sigue esa convención
al escribir código y mensajes nuevos.

Lecturas obligadas antes de tocar código: `docs/FRONTEND_ARCHITECTURE.md` (las
reglas que fallan el CI) y `docs/API.md` (contrato del backend). Además hay un
README por capa en `src/lib/api/`, `src/lib/features/` y `src/lib/shared/`.

## Comandos

```sh
pnpm dev                    # vite dev (espera el backend en $BASE_API, .env)
pnpm build && pnpm preview
pnpm check                  # svelte-kit sync + svelte-check + tsc
pnpm lint                   # prettier --check + eslint (incluye reglas de frontera)
pnpm format                 # prettier --write
pnpm check:arch             # presupuesto de tamaño, routes/ y restos del legacy
pnpm check:manual           # el PDF del manual corresponde al Markdown vigente
pnpm test:unit -- --run     # vitest (proyectos `client` y `server`)
pnpm test:e2e               # playwright contra el stub e2e/mocks/mock-api.mjs
```

CI (`.github/workflows/frontend-ci.yml`) corre esas seis comprobaciones en cada
PR que toque `frontend/**`. Desde la raíz, `make run` levanta backend (docker
db + cache + API) y frontend a la vez.

### Un solo test

```sh
pnpm exec vitest run --project server src/lib/shared/format/money.spec.ts
pnpm exec vitest run --project client src/lib/ui/button.svelte.spec.ts
pnpm exec vitest run --project server -t 'nombre del caso'
pnpm exec playwright test e2e/portfolio.e2e.ts -g 'nombre del caso'
```

Los dos proyectos de vitest se separan por nombre de archivo, no por carpeta:
`*.svelte.spec.ts` corre en Chromium (`client`), el resto en node (`server`).
`expect.requireAssertions` está activo: un test sin aserciones falla. Playwright
levanta el stub y hace `build` + `preview` por su cuenta (~3 min el primer
arranque); `CHROMIUM_EXECUTABLE_PATH` apunta a un Chromium ya instalado si el
entorno no puede descargarlo.

## Arquitectura

Capas de abajo arriba, cada una importa **solo** de las inferiores:

```
routes/            composición de páginas y form actions
  ↑
features/          lógica de dominio; una feature NO conoce a otra
  ↑
api/   ui/         acceso al backend · design system (no se conocen entre sí)
  ↑
shared/            utilidades transversales sin dominio
```

`routes/` solo orquesta: un `+page.server.ts` llama a un módulo de dominio de
`lib/api` y devuelve datos; un `+page.svelte` compone componentes de una
feature. Nada de `authedFetch`, schemas Zod ni paths del backend en `routes/`.

Estas reglas no son convención: `no-restricted-imports` por directorio en
`eslint.config.js` y `scripts/check-architecture.mjs` las hacen fallar el CI.
`check:arch` cubre además el presupuesto de tamaño (≤500 líneas por archivo de
producción, ≤300 por `+page.svelte`), que cada feature tenga `index.ts` con sus
componentes en `components/`, y que no reaparezcan `$components`, `$lib/utils`
ni `$lib/stores`.

Única excepción documentada a la superficie pública: la hoja de estilos global
de una feature (`import '$lib/features/landing/landing.css'`) se importa por
ruta, porque un barrel de JS no reexporta un side-effect de CSS.

No hay aliases propios: todo se importa por `$lib`.

### `lib/api` (server-only)

- `schemas/` define los contratos del backend como schemas Zod mantenidos a
  mano contra `docs/API.md`. **Es la fuente de verdad de los shapes**;
  `types.ts` los deriva con `z.infer` y es de donde todo el mundo importa
  tipos. Ninguna feature redeclara un shape del backend.
- `client.ts` expone `authedFetch`/`authedFetchSafe` (refresh single-flight
  compartido con `hooks.server.ts`, redirección a `/auth` solo cuando el
  refresh token se rechaza), `apiUrl()` —única lectura de `env.BASE_API`— y
  `apiRequest`/`apiRequestSafe`, que en `dev` validan el `data` recibido contra
  su schema y avisan por consola sin cambiar el resultado.
- Retorno: lecturas y comandos devuelven `ApiResult<T>`; solo streams/proxies
  (exports, import de CSV, combobox de assets) y los flujos públicos de
  `auth`/`marketing` devuelven la `Response` cruda.
- `proxy.ts` tiene la lista **cerrada** de rutas del backend alcanzables desde
  fuera (`/mcp`, su OAuth, avatares por id). `hooks.server.ts` las reenvía sin
  cookies, antes de la sesión y del check CSRF.

### Sesión y CSRF

`hooks.server.ts` resuelve la sesión solo bajo `/dashboard`, `/auth` y `/oauth`
(y les pone `X-Robots-Tag: noindex`). El check CSRF de SvelteKit está apagado en
`svelte.config.js` (`trustedOrigins: ['*']`) porque no puede eximir al proxy: lo
sustituye `$lib/server/csrf`, que corre fuera de `dev`.

### Dónde va el código nuevo

| Si es…                                 | Va a…                                                             |
| -------------------------------------- | ----------------------------------------------------------------- |
| Una llamada nueva al backend           | `lib/api/<dominio>.ts` + schema en `lib/api/schemas/<dominio>.ts` |
| Un componente de un dominio            | `lib/features/<feature>/components/` (al `index.ts` solo si sale) |
| Un helper puro de un dominio           | `lib/features/<feature>/<feature>.ts` con su spec                 |
| La validación de un formulario         | `lib/features/<feature>/schemas.ts`                               |
| Un componente reutilizable sin dominio | `lib/ui/`                                                         |
| Un formateador o utilidad común        | `lib/shared/`                                                     |
| Una página                             | `routes/`: loader delgado + composición                           |

Toda form action que lea campos valida con Zod en el borde. Excepciones: las
subidas de archivo (avatar, CSV), que se comprueban por tipo y tamaño, y el
`logout`, que no lee formulario.

CSS scoped por componente; cuando varios comparten chrome, las reglas viven en
el contenedor como `.contenedor :global(.clase)`. Una hoja global solo cuando
las clases son exclusivas de esa área.

## Pruebas

| Nivel                 | Dónde                        | Qué cubre                                                  |
| --------------------- | ---------------------------- | ---------------------------------------------------------- |
| Unit (node)           | `*.spec.ts` junto al módulo  | Helpers puros, schemas, capa de API, sesión                |
| Componente (Chromium) | `*.svelte.spec.ts`           | Render y comportamiento de `ui` y de features              |
| E2E                   | `e2e/*.e2e.ts`               | Flujos completos contra `e2e/mocks/mock-api.mjs`           |
| Contrato del stub     | `e2e/mocks/contract.spec.ts` | Que las fixtures del stub cumplan los schemas de `lib/api` |

Al añadir una fixture al stub, valídala en `contract.spec.ts`: si no, la suite
E2E puede pasar en verde sobre formas que el backend nunca envía. Los ids y
credenciales de prueba están en `e2e/helpers.ts`;
`$lib/server/testing.ts` trae un `Cookies` en memoria para los tests de sesión.

## Blog y manual

- Publicar un artículo es añadir `src/content/blog/<slug>.md` (el nombre es la
  URL) con frontmatter `title`, `description`, `date`, `tags`; `draft: true` lo
  deja visible solo en `pnpm dev`. Si falta un campo, el build falla. Detalles
  en `docs/BLOG.md`.
- El manual de usuario es `docs/MANUAL_DE_USUARIO.md`; su PDF se regenera con
  `pnpm manual:build` (capturas: `pnpm manual:shots`). Si tocas el manual sin
  regenerar, `pnpm check:manual` falla en CI por la huella guardada en
  `src/lib/features/guide/manual-meta.ts`.
