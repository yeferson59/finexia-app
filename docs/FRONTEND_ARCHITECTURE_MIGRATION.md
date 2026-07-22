# Plan de migración de arquitectura — Frontend (SvelteKit)

> **Objetivo:** evolucionar el frontend desde una organización por tipo de archivo
> (`components/`, `routes/` con lógica embebida) hacia una **arquitectura modular por
> features** con una **capa de API tipada por dominio**, de forma **incremental**
> (patrón *strangler fig*), sin big-bang, manteniendo `pnpm check`, lint, unit tests
> y E2E en verde en todo momento.
>
> Este documento desarrolla y reemplaza la "Fase 9" esbozada en
> [`ARCHITECTURE_MIGRATION.md`](./ARCHITECTURE_MIGRATION.md). Ambas migraciones son
> independientes: pueden avanzar en paralelo sin bloquearse, porque el contrato HTTP
> no cambia en ninguna de las dos.

---

## 1. Diagnóstico de la arquitectura actual

### 1.1 Cómo está organizado hoy

```
frontend/src/
├── app.d.ts / app.html / hooks.server.ts   # locals de sesión, guardas globales
├── config/                    # env.ts (vacío), features.ts (feature flags)
├── components/                # FUERA de lib/ (alias $components)
│   ├── ui/                    # design system: button, card, input, pagination…
│   ├── auth/                  # login-register.svelte (1.377 líneas)
│   ├── dashboard/             # sidebar, header, cards del dashboard
│   ├── landing-page/          # hero, faq, footer, metrics… + landing.css
│   └── cookie-notice.svelte
├── lib/
│   ├── server/                # api.ts (authedFetch), session.ts, testing.ts
│   ├── stores/                # investments.svelte.ts, privacy.svelte.ts (runas)
│   ├── utils.ts               # cajón de sastre: formatters, helpers varios
│   ├── seo.ts
│   └── index.ts               # vacío
└── routes/                    # TODA la lógica de página vive aquí
    ├── (legal)/  auth/  api/  sitemap.xml/
    └── dashboard/             # portfolios, investments, transactions, platforms,
                               # reports, settings, notifications, admin/…
```

### 1.2 Qué funciona bien (y hay que conservar)

- ✅ **`lib/server/api.ts` es una joya**: `authedFetch`/`authedFetchSafe` centralizan
  auth, refresh single-flight y redirección a `/auth`. Bien testeado. Es la base de
  la futura capa de API — no se toca su lógica, solo se construye encima.
- ✅ Manejo de sesión centralizado en `hooks.server.ts` + `lib/server/session.ts`,
  con tests.
- ✅ Svelte 5 con runes forzado por config; stores modernos (`.svelte.ts`).
- ✅ Design system incipiente en `components/ui/` con tests unitarios por componente.
- ✅ Tooling completo: svelte-check, eslint + prettier, vitest (unit + browser),
  Playwright E2E, Tailwind 4.
- ✅ Validación con Zod en las form actions de auth.

### 1.3 Puntos de dolor (por qué migrar)

| Problema | Evidencia | Consecuencia |
|---|---|---|
| **Sin capa de API tipada** | 24 archivos de `routes/` llaman `authedFetch(event, '/path')` crudo y parsean JSON a mano | Cada loader re-declara los endpoints; un cambio de path/shape en el backend obliga a cazar strings por todo `routes/` |
| **Contratos duplicados** | `portfolios/[id]/+page.server.ts` define localmente `Holding`, `PortfolioDetail`, `Risk`, `TopTransaction`, `GrowthDataPoint`…; otras rutas redefinen los mismos shapes | Tipos divergen silenciosamente entre páginas; cero fuente de verdad del contrato con el backend |
| **Páginas monolíticas** | `portfolios/[id]/assets/[symbol]/+page.svelte` (**2.014 líneas**), `login-register.svelte` (1.377), `portfolios/[id]/+page.svelte` (1.309), `settings/+page.svelte` (1.146), `transactions/import/+page.svelte` (1.007) | Estado, formateo, markup y lógica de negocio mezclados; imposible testear unitariamente; conflictos de merge constantes |
| **`components/` fuera de `lib/`** | Alias ad-hoc `$components`; SvelteKit solo aplica sus convenciones (server-only, empaquetado) dentro de `lib/` | Dos formas de importar lo mismo; el design system (`ui/`) convive sin frontera con componentes de feature (`dashboard/`, `auth/`) |
| **Lógica compartida sin dueño** | `lib/utils.ts` es un cajón de sastre; `lib/stores/investments.svelte.ts` es estado de la feature investments pero vive en un paquete global | Nadie sabe dónde poner código nuevo → todo acaba en utils o en la página |
| **Validación inconsistente** | Zod se usa en actions de auth y portfolios, pero otras actions parsean `FormData` a mano | Errores de validación con formatos distintos por página |

### 1.4 Arquitectura objetivo: features + capa de API tipada

Regla de oro: **`routes/` solo orquesta** (loaders/actions delgados que llaman a
`lib/api`, páginas que componen componentes de feature). La lógica vive en
`lib/features/<feature>/` y el acceso al backend en `lib/api/`. Las features no se
importan entre sí; lo compartido baja a `lib/ui`, `lib/api` o `lib/shared`.

```
frontend/src/
├── hooks.server.ts             # sin cambios de rol
├── routes/                     # SOLO orquestación y composición
│   └── dashboard/portfolios/[id]/
│       ├── +page.server.ts     # ~20 líneas: llama a lib/api/portfolio y devuelve data
│       └── +page.svelte        # ~100 líneas: compone componentes de la feature
└── lib/
    ├── api/                    # capa de acceso al backend (server-only)
    │   ├── client.ts           # ← lib/server/api.ts (authedFetch, sin cambios de lógica)
    │   ├── types.ts            # contratos compartidos (espejo de docs/API.md)
    │   ├── auth.ts             # login(), register(), refresh(), reset(), invite()…
    │   ├── portfolio.ts        # getPortfolio(), listHoldings(), createEntry()…
    │   ├── market.ts           # assets, exchange rates
    │   ├── user.ts             # perfil, preferencias, avatar, admin
    │   └── marketing.ts        # waitlist
    │
    ├── features/               # un directorio por dominio funcional
    │   ├── auth/
    │   │   ├── components/     # login-form, register-form, two-factor-form…
    │   │   ├── schemas.ts      # schemas Zod de sus formularios
    │   │   └── index.ts        # superficie pública de la feature
    │   ├── portfolio/          # componentes de detalle, holdings, growth, add-entry
    │   ├── transactions/       # listado + wizard de import
    │   ├── platforms/
    │   ├── dashboard/          # widgets: net-worth, allocation, recent-activity…
    │   │   └── state/          # ← lib/stores/investments.svelte.ts
    │   ├── settings/
    │   ├── admin/              # users, assets, exchange-rates
    │   ├── landing/            # ← components/landing-page/
    │   └── legal/              # cookie-notice + páginas legales compartidas
    │
    ├── ui/                     # ← components/ui/ (design system puro, sin dominio)
    ├── server/                 # session.ts, testing.ts (server-only transversal)
    └── shared/                 # seo.ts, config/features.ts, utils repartido en
        ├── format/             #   money.ts, date.ts, percent.ts
        └── privacy.svelte.ts   #   ← lib/stores/privacy.svelte.ts (transversal real)
```

**Decisiones de diseño que acompañan la estructura:**

1. **`lib/api` es la única puerta al backend.** Ningún loader/action llama a
   `authedFetch` directamente ni escribe paths (`'/portfolios/' + id`) fuera de
   `lib/api`. Cada función devuelve datos tipados con los tipos de `lib/api/types.ts`.
2. **Un solo lugar para los contratos.** `lib/api/types.ts` (o `types/` por dominio
   si crece) se mantiene a mano contra `docs/API.md` del backend. Cuando el backend
   publique OpenAPI, se evalúa generación automática (anotar en `TECH_DEBT.md`).
3. **Las features exponen `index.ts`** y solo eso se importa desde `routes/`.
   Prohibido `import ... from '$lib/features/portfolio/components/holdings-table.svelte'`
   desde otra feature.
4. **`lib/ui` no conoce dominios**: no importa de `features/` ni de `api/`. Recibe
   todo por props/snippets.
5. **Zod en el borde, siempre**: toda form action valida con un schema de la feature
   (`features/<x>/schemas.ts`); opcionalmente los responses críticos del backend se
   validan en `lib/api` con `z.parse` en dev.
6. **Páginas con presupuesto de tamaño**: una `+page.svelte` no supera ~300 líneas;
   si lo hace, extrae componentes a su feature.
7. **Aliases finales**: solo `$lib` (estándar). `$components` y `$/*` se eliminan al
   final de la migración.

---

## 2. Reglas del proceso de migración

- 🔒 **Nunca se rompe `main`**: cada fase termina con la verificación estándar en
  verde. Si una fase no cabe en un PR razonable, se parte.
- 🔀 **Una feature por PR** (máximo). Los PRs de migración no mezclan refactor con
  features ni bugfixes.
- 🧪 **Primero red de seguridad, después mover código**: no se trocea una página sin
  E2E que cubra su flujo principal.
- 📦 **Convivencia temporal**: `components/` (legacy) y `lib/features/` coexisten
  durante la migración. Está prohibido que código nuevo importe de `$components`.
- 🚫 **Mover ≠ mejorar**: al extraer componentes no se rediseña UI ni se cambia
  comportamiento. Las mejoras detectadas van a `docs/TECH_DEBT.md`.
- 🌐 **Cero cambios visibles**: mismas URLs, mismo HTML renderizado (mismas clases,
  mismo SEO), mismos flujos. Los E2E de Playwright son el contrato de no-regresión.
- 🧭 **Svelte 5 idiomático al mover**: los componentes extraídos usan runes y
  snippets (ya es la convención del proyecto), sin introducir patrones de Svelte 4.

**Verificación estándar al cerrar cada fase** (desde `frontend/`):

```bash
pnpm check          # svelte-check + tsc
pnpm lint           # prettier + eslint
pnpm test:unit -- --run
pnpm test:e2e
```

---

## 3. Checklist de migración

### Fase 0 — Red de seguridad y línea base *(sin mover código)* ✅

- [x] Verificar que CI ejecuta `check`, `lint`, `test:unit` y `test:e2e` del frontend
      en cada PR; si no, configurarlo antes de tocar nada.
      *(`.github/workflows/frontend-ci.yml`)*
- [x] Auditar cobertura E2E actual (solo existe `e2e/landing.e2e.ts`) y añadir smoke
      tests de los flujos críticos que la migración va a tocar
      *(los flujos autenticados corren contra el stub `e2e/mocks/mock-api.mjs`,
      escrito contra `docs/API.md`)*:
  - [x] Login + logout (y redirección de rutas protegidas a `/auth`).
        *(`e2e/auth.e2e.ts`)*
  - [x] Dashboard principal renderiza sus widgets. *(`e2e/dashboard.e2e.ts`)*
  - [x] Detalle de portfolio + añadir entrada. *(`e2e/portfolio.e2e.ts`)*
  - [x] Listado de transacciones + wizard de import (al menos el preview).
        *(`e2e/transactions.e2e.ts`)*
  - [x] Settings: cambio de un dato de perfil. *(`e2e/settings.e2e.ts`)*
  - [x] Admin: listado de usuarios (con rol admin). *(`e2e/admin.e2e.ts`)*
- [x] Capturar snapshot del HTML/SEO de la landing y páginas legales (los E2E ya
      comparan title/meta; ampliar si hace falta) para detectar regresiones de SEO.
      *(`e2e/seo.e2e.ts`: title, description, canonical, OG, robots y
      `X-Robots-Tag` de áreas privadas)*
- [x] Anotar línea base: número de archivos > 500 líneas en `src/` y % de loaders
      que llaman `authedFetch` directo (hoy: 24 archivos). La Fase 7 se valida
      contra estos números. *(→ [`FRONTEND_MIGRATION_BASELINE.md`](./FRONTEND_MIGRATION_BASELINE.md))*
- [x] Confirmar que existe `docs/API.md` (entregable de la Fase 0 del backend); si
      aún no existe, documentar aquí al menos los endpoints que consume el frontend.
      *(existe, generado 2026-07-13; cubre todo lo que consume el frontend)*

### Fase 1 — Fundaciones: `lib/ui`, `lib/api/client` y convenciones *(movimientos mecánicos)* ✅

- [x] Mover `src/components/ui/` → `src/lib/ui/` (componentes + specs), actualizando
      imports a `$lib/ui/...`.
- [x] Mover `src/lib/server/api.ts` → `src/lib/api/client.ts` **sin cambios de
      lógica** (re-export temporal desde la ruta vieja para no tocar 24 archivos en
      este PR; se elimina en Fase 2).
- [x] Crear `src/lib/shared/` y repartir `lib/utils.ts` por tema (`format/money.ts`,
      `format/date.ts`, `css.ts`) manteniendo re-exports temporales desde `utils.ts`.
- [x] Mover `src/config/*` → `src/lib/shared/config/` (el `env.ts` vacío se eliminó;
      se volverá a introducir con validación de env públicas cuando haga falta).
- [x] Crear los esqueletos `src/lib/features/` y `src/lib/api/` con un `README.md`
      corto que resuma las reglas de la sección 1.4 (qué puede importar qué).
- [x] Verificación estándar. *(`pnpm check` 0 errores, `pnpm lint`, 122 unit tests,
      22 E2E — todo en verde)*

### Fase 2 — Capa de API tipada por dominio *(el cambio de mayor impacto)* ✅

- [x] Crear `lib/api/types.ts` con los contratos que hoy están duplicados en los
      loaders (extraerlos de `portfolios/[id]/+page.server.ts`,
      `assets/[symbol]/+page.server.ts`, `admin/*/+page.server.ts`, etc.),
      contrastados contra `docs/API.md`.
- [x] Crear los módulos por dominio, cada uno con funciones tipadas que encapsulan
      path + método + parseo (devuelven `ApiResult<T>`, o la `Response` cruda para
      streams/proxies y los flujos públicos de auth/marketing):
  - [x] `lib/api/auth.ts`
  - [x] `lib/api/portfolio.ts` (portfolios, holdings, entries, snapshots, growth)
  - [x] `lib/api/transactions.ts` (listado + import preview/commit + export)
  - [x] `lib/api/platforms.ts`
  - [x] `lib/api/market.ts` (assets, exchange rates)
  - [x] `lib/api/user.ts` (perfil, preferencias, avatar, admin de usuarios + sesiones/2FA)
  - [x] `lib/api/marketing.ts` (waitlist)
- [x] Migrar los loaders/actions/endpoints para consumir `lib/api/<dominio>` y
      **borrar sus interfaces locales** (un commit por área de rutas: auth,
      dashboard raíz, portfolios, transactions, platforms, admin, settings, api/).
- [x] Eliminar el re-export temporal de `lib/server/api.ts`; verificar que
      `authedFetch` solo se importa desde `lib/api/*`:
      `grep -rn "server/api" src/routes/` sale vacío.
- [x] Mover los tests de `api.spec.ts` junto a `lib/api/` (`client.spec.ts`) y
      añadir tests de los módulos de dominio (`domains.spec.ts`: paths correctos,
      métodos, propagación de errores).
- [x] Verificación estándar + E2E completo. *(`pnpm check` 0 errores, `pnpm lint`,
      134 unit tests, 22 E2E — todo en verde)*

### Fase 3 — Feature piloto: `landing` *(la más aislada; valida el patrón)* ✅

- [x] Crear `lib/features/landing/` y mover `components/landing-page/*` (incluido
      `landing.css`) con `index.ts` como superficie pública.
- [x] Actualizar `routes/+page.svelte` para importar desde
      `$lib/features/landing`.
- [x] Mover `components/cookie-notice.svelte` a `lib/features/legal/` (lo usan
      landing y páginas legales).
- [x] Verificar contra el snapshot SEO de Fase 0 que el HTML no cambió.
      *(`e2e/seo.e2e.ts` + `e2e/landing.e2e.ts` en verde; mismas clases, mismo
      JSON-LD, mismo `<head>`)*
- [x] **Retrospectiva del piloto**: ajustar en este documento cualquier decisión
      (naming, index.ts, ubicación de css) antes de replicar el patrón.
      *(→ sección 3.1)*

#### 3.1 Retrospectiva del piloto `landing` (patrón a replicar)

Decisiones validadas con la feature piloto, que las Fases 4–6 deben seguir:

- **Estructura**: los componentes viven en `features/<feature>/components/*.svelte`;
  el `index.ts` en la raíz de la feature es la única superficie pública. Los
  componentes se mueven con `git mv` para conservar el historial.
- **`index.ts` = barrel de re-exports nombrados**:
  `export { default as Header } from './components/header.svelte';`. `routes/`
  importa desestructurando (`import { Header, Hero } from '$lib/features/landing'`),
  nunca la ruta interna del `.svelte`.
- **CSS global como excepción explícita a la regla "solo `index.ts`"**: un barrel
  de JS no puede reexportar un side-effect de CSS. La hoja `landing.css` se queda
  en la raíz de la feature (`features/landing/landing.css`) y `routes/` la importa
  por su ruta (`import '$lib/features/landing/landing.css'`). Es el único import de
  una feature que no pasa por `index.ts`; se documenta aquí para no tratarlo como
  violación de frontera en la Fase 7.
- **Reutilización entre áreas vía `index.ts`**: `Brand` y `Footer` los consume
  también `routes/(legal)/+layout.svelte`. Que una ruta importe el `index.ts` de
  otra feature es correcto (la regla que se blinda en Fase 7 es *feature ↛
  feature*, no *route ↛ feature*). No se duplicaron componentes ni se creó un
  paquete compartido prematuro.
- **`cookie-notice` → `features/legal/`**: aunque hoy es un solo componente, se le
  dio su propia feature porque es el hogar natural de las páginas legales
  compartidas; su `index.ts` exporta `CookieNotice`.
- **Cero cambios de markup/estilos**: solo se movieron archivos y se reescribieron
  imports; los E2E de landing y SEO pasan sin tocar snapshots.

### Fase 4 — Feature `auth` *(la más sensible)* ✅

- [x] Crear `lib/features/auth/` y **trocear `login-register.svelte` (1.377 líneas)**
      en componentes: `login-form`, `register-form`, `two-factor-challenge`,
      `invite-only-notice` y `password-input` (el campo con toggle de visibilidad,
      antes duplicado 3×), orquestados por el contenedor `login-register`; más los
      formularios `forgot-password-form`, `reset-password-form`, `verify-email-panel`
      y `accept-invite-form` que hoy vivían en sus páginas.
- [x] Centralizar los schemas Zod de auth en `features/auth/schemas.ts`
      (login, 2FA, registro, forgot/reset, accept-invite, verify/resend), antes
      repartidos por las actions.
- [x] Adelgazar las páginas de `routes/auth/**`: las actions validan con los
      schemas de la feature y llaman a `lib/api/auth`; el markup solo compone los
      componentes de la feature (las 4 páginas standalone quedan en ~12 líneas).
- [x] Migrar `login-register.svelte.spec.ts` a specs por componente extraído
      (`login-register` contenedor, `login-form`, `register-form`,
      `invite-only-notice`, `password-input`).
- [x] Verificación estándar + E2E de auth completo. *(`pnpm check` 0 errores,
      `pnpm lint`, 141 unit tests, 22 E2E — todo en verde)*

#### 4.1 Notas de la migración de `auth`

- **CSS scoped → dos estrategias según el reparto:** al trocear un componente con
  estilos scoped, el markup movido a un hijo pierde esos estilos (Svelte los
  aísla por componente). Se resolvió así:
  - Las clases compartidas por varios sub-formularios de `login-register`
    (`.form-content`, `.error-server`, `.form-switch`, `.switch-link`,
    `.resend-link`, `.forgot-link`, `.form-footer`) —todas exclusivas de auth y
    sin colisión con otros estilos globales— bajaron a un `auth-forms.css` global
    importado una sola vez por el contenedor (patrón `landing.css`). Las variables
    `--gold-primary`/`--text-secondary`/`--error-color` se heredan por DOM desde
    `main.auth-container`, así que los hijos resuelven `var(...)` sin cambios.
  - Las clases propias de un único formulario (`.consent*`, `.two-factor-*`,
    `.invite-only-*`, `.password-*`) y **las de nombre genérico** (`.error-message`,
    y todo el chrome de las páginas standalone: `.wrap`, `.card`, `.brand`,
    `.title`, `.eyebrow`…) se mantuvieron **scoped** en su componente para evitar
    colisiones globales (p. ej. `.error-message` también existe scoped en
    `portfolios/add`). Las páginas standalone se movieron **verbatim** a su
    componente, preservando exactamente el scope actual.
- **Estado de formulario dentro de cada hijo:** cada sub-formulario deriva sus
  errores del prop `form` con `$derived(parseErrors(...))` y posee su propio
  estado de campos/envío; el contenedor solo decide qué formulario mostrar y
  gestiona el toggle login/registro. `parseErrors` se extrajo a `utils.ts`.
- **`svelte:head` se queda en la página:** las páginas standalone conservan su
  `<svelte:head>` (título + `robots noindex`) como orquestación de página; el
  componente de feature solo aporta el `<main>` del formulario.

### Fase 5 — Features del área de inversión: `portfolio`, `dashboard`, `transactions`, `platforms`

*(una feature = un PR; orden sugerido de menor a mayor riesgo)*

- [ ] `lib/features/dashboard/`:
  - [ ] Mover `components/dashboard/*` (sidebar, header, net-worth-card,
        asset-allocation, portfolio-growth, portfolio-overview, recent-activity,
        currency-toggle).
  - [ ] Mover `lib/stores/investments.svelte.ts` → `features/dashboard/state/`
        (o a `features/portfolio` si el piloto de Fase 3 sugirió otra cosa) con su spec.
- [ ] `lib/features/platforms/`: extraer componentes de
      `routes/dashboard/platforms/**` (782 líneas la página de detalle).
- [ ] `lib/features/portfolio/`:
  - [ ] Trocear `portfolios/[id]/assets/[symbol]/+page.svelte` (**2.014 líneas**) —
        el peor archivo del frontend — en componentes de feature (cabecera del
        asset, gráfico, historial de transacciones, formularios de compra/venta…).
  - [x] Trocear `portfolios/[id]/+page.svelte` (1.309 → 358) y
        `portfolios/[id]/add/+page.svelte` (604 → 13). *(Rama 1 de portfolio.)*
        - Componentes: `portfolio-summary-cards`, `portfolio-stats-cards`,
          `allocation-donut`, `holdings-table`, `portfolio-edit-form`,
          `portfolio-add-form`. `portfolio.ts` centraliza helpers puros
          (`groupHoldings`, `computeTypeBreakdown`, `computeDonutSegments`,
          `formatPct`), constantes (`PORTFOLIO_TYPES`, `ASSET_TYPE_*`) y tipos,
          con su `portfolio.spec.ts`.
        - `formatCurrency` (usa `privacy` + `baseCurrency`) se crea en la página
          y se pasa como prop a los componentes que lo necesitan.
        - `portfolio-growth` se sigue importando de `$components/dashboard`
          (la migración de dashboard va en otra rama; no es feature↛feature).
        - *(Verificación: `pnpm check` 0 errores/0 warnings, `pnpm lint`, 146
          unit tests, 22 E2E incluido `portfolio.e2e.ts` — todo en verde.)*
  - [ ] Trocear `investments/*` (714 + 474) reutilizando los mismos componentes.
- [ ] `lib/features/transactions/`: trocear `transactions/import/+page.svelte`
      (1.007) en pasos del wizard (upload, preview, commit) + listado.
- [ ] Cada PR: verificación estándar + E2E del flujo correspondiente.

### Fase 6 — Features restantes: `settings`, `admin`

- [ ] `lib/features/settings/`: trocear `settings/+page.svelte` (1.146 líneas) en
      secciones (perfil, seguridad/2FA, sesiones, preferencias, notificaciones).
- [ ] `lib/features/admin/`: extraer componentes de `admin/users`, `admin/assets`,
      `admin/exchange-rates` (595–671 líneas cada una); las tablas/paginación
      comunes bajan a `lib/ui` si no tienen dominio.
- [ ] Verificación estándar + E2E de settings y admin.

### Fase 7 — Demolición del legacy y blindaje de fronteras

- [ ] Verificar que `src/components/` quedó vacío y **borrarlo**; eliminar el alias
      `$components` (y evaluar eliminar `$/*`) de `svelte.config.js`.
- [ ] Eliminar `lib/utils.ts` y `lib/stores/` si quedaron reducidos a re-exports;
      `lib/index.ts` vacío también se borra.
- [ ] Blindar fronteras con ESLint (`import/no-restricted-paths` o
      `eslint-plugin-boundaries`):
  - [ ] `lib/ui` no importa de `lib/features` ni `lib/api`.
  - [ ] Una feature no importa de otra feature.
  - [ ] `routes/` no importa internals de features (solo `index.ts`) ni
        `lib/api/client` directo (solo módulos de dominio).
- [ ] Comparar contra la línea base de Fase 0: cero loaders con `authedFetch`
      crudo, y ningún `.svelte`/`.ts` de producción > ~500 líneas.
- [ ] Actualizar `frontend/README.md` (o crear `docs/FRONTEND_ARCHITECTURE.md`) con
      la estructura final y las reglas de dependencia.
- [ ] Revisión final de `docs/TECH_DEBT.md`.

---

## 4. Riesgos y mitigaciones

| Riesgo | Mitigación |
|---|---|
| Regresión visual/funcional al trocear páginas gigantes | E2E por flujo ANTES de trocear (Fase 0); extraer componentes sin cambiar markup ni clases; PRs pequeños por página |
| La capa de API cambia sutilmente el manejo de errores/redirects | `client.ts` se mueve sin tocar su lógica; los módulos de dominio solo envuelven paths y tipos; los specs existentes de `api.spec.ts` siguen pasando |
| Tipos de `lib/api/types.ts` divergen del backend real | Se contrastan contra `docs/API.md` (contrato compartido con la migración del backend); validación Zod opcional en dev; generación desde OpenAPI como deuda futura |
| Conflictos con features en desarrollo | Una feature por PR; no migrar un área con una feature abierta encima |
| "Mover" se convierte en "rediseñar UI" | Regla explícita de la sección 2; mejoras a `docs/TECH_DEBT.md` |
| Las features se acoplan entre sí con el tiempo | `index.ts` como única superficie pública + reglas ESLint de fronteras en Fase 7 fallando el CI |

## 5. Criterios de éxito

- [ ] `routes/` contiene solo orquestación: ningún loader declara interfaces de la
      API ni llama a `authedFetch` directamente.
- [ ] Existe una única fuente de verdad de los contratos del backend
      (`lib/api/types.ts`) alineada con `docs/API.md`.
- [ ] Ningún archivo de producción supera ~500 líneas; ninguna `+page.svelte`
      supera ~300.
- [ ] `src/components/` y el alias `$components` no existen; todo vive bajo `$lib`.
- [ ] Toda form action valida con un schema Zod de su feature.
- [ ] Las reglas de dependencia (ui ↛ features, feature ↛ feature, routes → solo
      `index.ts`) están automatizadas en ESLint y fallan el CI.
- [ ] Suite completa en verde: `pnpm check && pnpm lint && pnpm test`.
