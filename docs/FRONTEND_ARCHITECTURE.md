# Arquitectura del frontend (SvelteKit)

> Estado final de la migración descrita en
> [`FRONTEND_ARCHITECTURE_MIGRATION.md`](./FRONTEND_ARCHITECTURE_MIGRATION.md),
> cerrada el **2026-07-31**. Este documento describe cómo está organizado el
> frontend **hoy** y qué reglas hay que respetar al añadir código; el de
> migración cuenta cómo se llegó hasta aquí y por qué se tomó cada decisión.

## 1. Idea central

**`routes/` solo orquesta.** Un `+page.server.ts` llama a `lib/api` y devuelve
datos; un `+page.svelte` compone componentes de una feature. Toda la lógica de
negocio vive en `lib/features/<feature>/` y todo el acceso al backend en
`lib/api/`.

```
frontend/src/
├── app.html · app.d.ts · hooks.server.ts   # sesión y guardas globales
├── routes/                                 # SOLO orquestación y composición
│   ├── (legal)/ · auth/ · api/ · sitemap.xml/
│   └── dashboard/                          # portfolios, assets, transactions,
│                                           # platforms, reports, settings,
│                                           # notifications, admin…
└── lib/
    ├── api/          # capa de acceso al backend (server-only)
    │   ├── client.ts     # authedFetch/authedFetchSafe: auth, refresh y redirección
    │   ├── schemas.ts    # contratos HTTP como schemas Zod, espejo de docs/API.md
    │   ├── types.ts      # los tipos, derivados de los schemas con z.infer
    │   └── auth.ts · portfolio.ts · transactions.ts · platforms.ts ·
    │     market.ts · user.ts · marketing.ts
    │
    ├── features/     # un directorio por dominio funcional
    │   ├── admin/ · auth/ · dashboard/ · investments/ · landing/ · legal/ ·
    │   │ notifications/ · platforms/ · portfolio/ · reports/ · settings/ ·
    │   │ transactions/
    │   └── <feature>/
    │       ├── components/     # componentes del dominio
    │       ├── <feature>.ts    # helpers puros, constantes y tipos del dominio
    │       ├── schemas.ts      # schemas Zod de sus formularios
    │       ├── state/          # estado del dominio (runes .svelte.ts), si aplica
    │       └── index.ts        # superficie pública
    │
    ├── ui/           # design system sin dominio (button, card, input, table…)
    ├── server/       # session.ts, testing.ts (server-only transversal)
    └── shared/       # css.ts, format/, finance/, chart/, config/, form.ts,
                      # privacy.svelte.ts
```

Sin aliases propios: todo se importa por `$lib`, el estándar de SvelteKit.

## 2. Capas y reglas de dependencia

Las capas van de abajo arriba. Una capa **solo** puede importar de las que
tiene por debajo:

```
routes/          ← composición de páginas y form actions
   ↑
features/        ← lógica de dominio; una feature NO conoce a otra
   ↑
api/  ui/        ← acceso al backend · design system (no se conocen entre sí)
   ↑
shared/          ← utilidades transversales sin dominio
```

| Regla | Motivo |
|---|---|
| `lib/shared` no importa de `features`, `api` ni `ui` | Es la capa más baja: todos dependen de ella. |
| `lib/ui` no conoce dominios (ni `features` ni `api`) | El design system se reutiliza; recibe todo por props o snippets. |
| `lib/api` no depende de la UI ni de `features` | Es el acceso al backend, no sabe cómo se pinta. |
| Una feature no importa de otra feature | Lo compartido baja a `ui`, `api` o `shared`. Dentro de una feature todo es relativo. |
| `routes/` importa `$lib/features/<x>`, nunca sus internos | El `index.ts` es el contrato de la feature; lo demás puede cambiar sin avisar. |
| `routes/` no importa `lib/api/client` | Los paths y los tipos viven en los módulos de dominio de `lib/api`. |

**Todas fallan el CI**: están implementadas con `no-restricted-imports` por
directorio en [`eslint.config.js`](../frontend/eslint.config.js).

**Única excepción documentada:** la hoja de estilos global de una feature
(`import '$lib/features/landing/landing.css'`) se importa por su ruta, porque
un barrel de JS no puede reexportar un side-effect de CSS.

## 3. Convenciones

- **Contratos del backend en un solo sitio.** `lib/api/schemas.ts` los define
  como schemas Zod, mantenidos a mano contra [`API.md`](./API.md), y
  `lib/api/types.ts` los deriva con `z.infer`: el tipo no puede desalinearse de
  su schema. Ninguna feature redeclara un shape del backend: lo importa de
  `types.ts` y, si le conviene, lo reexporta desde su módulo.
- **El contrato se comprueba en desarrollo.** Los módulos de dominio pasan su
  schema a `apiRequest`/`apiRequestSafe`, que en `dev` valida el `data` recibido
  y avisa en consola con la ruta y el campo que falla. Nunca cambia el
  resultado, y `import.meta.env.DEV` lo deja fuera del build de producción. Así
  un backend que se sale del contrato se nota donde ocurre, y no tres
  componentes más abajo como un `undefined`.
- **`ApiResult<T>` en toda la capa de API.** Las funciones de dominio
  encapsulan path + método + parseo; devuelven la `Response` cruda solo para
  streams/proxies y para los flujos públicos de auth/marketing.
- **Superficie pública por `index.ts`.** No todo componente entra: los que solo
  usa otro componente de la misma feature se importan en relativo y no forman
  parte del contrato.
- **Zod en el borde.** Toda form action que reciba campos valida con un schema
  de `features/<x>/schemas.ts`. Dos excepciones, ambas por no tener nada que
  parsear: los archivos subidos (avatar, imports de CSV), que se comprueban por
  tipo y tamaño, y el `logout` del dashboard, que no lee ningún campo del
  formulario —solo cookies—.
- **Svelte 5 idiomático.** Runes (`$state`, `$derived`, `$props`, `$effect`) y
  snippets; el modo runes está forzado por configuración.
- **Presupuesto de tamaño.** Ningún archivo de producción supera ~500 líneas y
  ninguna `+page.svelte` supera ~300. Si crece, se extraen componentes a su
  feature.
- **CSS.** Scoped por componente, que es el modo por defecto de Svelte. Cuando
  varios componentes de una feature comparten chrome, las reglas viven en el
  componente contenedor y se declaran como `.contenedor :global(.clase)`, que
  Svelte compila con el hash del contenedor: alcanzan a los hijos sin escapar
  del bloque. Una hoja global (`landing.css`, `auth-forms.css`) solo cuando las
  clases son exclusivas de esa área.

## 4. Pruebas

| Nivel | Dónde | Qué cubre |
|---|---|---|
| Unit (node) | `*.spec.ts` junto al módulo | Helpers puros, schemas, capa de API, sesión. |
| Componente (browser) | `*.svelte.spec.ts` | Render y comportamiento de componentes de `ui` y de features. |
| E2E (Playwright) | `frontend/e2e/*.e2e.ts` | Flujos completos contra el stub `e2e/mocks/mock-api.mjs`, escrito contra `docs/API.md`. |
| Contrato del stub | `e2e/mocks/contract.spec.ts` | Que las fixtures del stub cumplan los schemas de `lib/api`, para que la suite E2E no pase en verde sobre formas que el backend no envía. |

```bash
cd frontend
pnpm check          # svelte-check + tsc
pnpm lint           # prettier + eslint (incluye las reglas de frontera)
pnpm check:arch     # presupuesto de tamaño, `routes/` y restos del legacy
pnpm test:unit -- --run
pnpm test:e2e
```

`.github/workflows/frontend-ci.yml` ejecuta las cinco en cada PR que toque
`frontend/**`.

**`check:arch`** (`scripts/check-architecture.mjs`) cubre lo que ESLint no puede
expresar: el presupuesto de tamaño, que `routes/` no declare schemas Zod ni
llame al cliente HTTP, que no reaparezcan los caminos del legacy
(`$components`, `$lib/utils`, `$lib/stores`) y que cada feature tenga su
`index.ts` con los componentes en `components/`. Eran una foto en
`FRONTEND_MIGRATION_BASELINE.md`; ahora fallan el CI.

## 5. Dónde poner código nuevo

| Si es… | Va a… |
|---|---|
| Una llamada nueva al backend | `lib/api/<dominio>.ts` + su schema en `lib/api/schemas.ts` (el tipo sale solo) |
| Un componente de un dominio | `lib/features/<feature>/components/` (y al `index.ts` solo si lo usa `routes/` u otra área) |
| Un helper puro de un dominio | `lib/features/<feature>/<feature>.ts` con su spec |
| La validación de un formulario | `lib/features/<feature>/schemas.ts` |
| Un componente reutilizable sin dominio | `lib/ui/` |
| Un formateador o utilidad transversal | `lib/shared/` |
| Una página | `routes/`: loader delgado + composición |
