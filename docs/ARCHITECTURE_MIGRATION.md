# Plan de migración de arquitectura — Finexia

> **Objetivo:** evolucionar el backend (y de forma secundaria el frontend) desde una
> arquitectura en capas técnicas hacia un **monolito modular organizado por dominios**,
> de forma **incremental** (patrón *strangler fig*), sin big-bang, sin congelar el
> desarrollo de features y manteniendo la suite de tests en verde en todo momento.

---

## 1. Diagnóstico de la arquitectura actual

### 1.1 Cómo está organizado hoy el backend

```
backend/
├── cmd/api/                  # main: crea Fiber, DB, cache, S3, mail y llama a internal.New()
└── internal/
    ├── root.go               # Bootstrap: cablea TODO (repos → services → handlers → routes → schedulers)
    ├── config/               # env, db (pgx), cache (redis), storage (s3)
    ├── entities/             # structs de dominio compartidos por todos
    ├── dtos/                 # DTOs request/response por dominio (auth, portfolio, user, marketing)
    ├── repositories/         # UN struct Repository con *pgxpool.Pool para todos los dominios
    ├── services/             # UN struct Services con todos los casos de uso de la app
    ├── handlers/             # UN struct Handlers que depende de services.Services completo
    ├── routes/               # registro de rutas por área, sobre Handlers/Middlewares globales
    ├── middlewares/          # cors, helmet, jwt, limiter, rbac, recovery, etc.
    ├── scheduler/            # 5 schedulers que reciben services.Services completo
    ├── alphavantage/ finnhub/ yahoo/  # clientes de proveedores de precios
    ├── prices/               # Provider + fallback entre proveedores
    ├── geoip/ mail/ logger/  # infraestructura transversal
    └── migrations/ migrator/ # SQL migrations + runner
```

### 1.2 Qué funciona bien (y hay que conservar)

- ✅ Separación handler → service → repository ya existe y es consistente.
- ✅ Inyección de dependencias manual y explícita (sin frameworks mágicos).
- ✅ Ya hay interfaces orientadas a testing: `services.Repository`, `Mailer`,
  `GeoLocator`, `prices.Provider` (con fakes en tests).
- ✅ Buen patrón en `prices/`: interfaz + implementaciones (`alphavantage`,
  `finnhub`, `yahoo`) + decorador `Fallback`. **Este es el patrón a replicar.**
- ✅ Migraciones SQL versionadas, logger abstraído, config centralizada.

### 1.3 Puntos de dolor (por qué migrar)

| Problema | Evidencia | Consecuencia |
|---|---|---|
| **God interface** | `services/repository.go`: 1 interfaz `Repository` con **~92 métodos** de todos los dominios | Cualquier fake de test debe implementar (o embeber) la interfaz completa → `testsupport_test.go` tiene ~700 líneas |
| **God struct de servicios** | `services.Services` contiene auth, users, portfolios, marketing, import/export, 2FA, snapshots… | Todo consumidor (handlers, middlewares, 5 schedulers) recibe acceso a *todo*; imposible razonar sobre qué usa qué |
| **Archivos gigantes** | `repositories/portfolio.go` (1.166 líneas), `services/transaction_import.go` (913), `services/auth.go` (688) | Merge conflicts frecuentes, difícil navegar y revisar |
| **Acoplamiento por capas técnicas** | Tocar "portfolio" implica tocar `dtos/`, `entities/`, `repositories/`, `services/`, `handlers/`, `routes/` | Los cambios de una feature se dispersan por 6 paquetes |
| **Sin fronteras entre dominios** | `entities/` es un paquete plano compartido; nada impide que marketing importe lógica de auth | Los dominios se contaminan entre sí con el tiempo |
| **Bootstrap monolítico** | `internal/root.go` conoce todos los paquetes | Añadir un dominio = editar el mismo archivo siempre |

### 1.4 Arquitectura objetivo: monolito modular por dominios

Cada dominio se convierte en un **módulo autocontenido** con sus propias capas
internas y una superficie pública mínima. La infraestructura compartida se agrupa
en `platform/`. Regla de oro: **un módulo solo puede importar `platform/` y las
interfaces públicas de otro módulo, nunca sus internals**.

```
backend/
├── cmd/api/main.go                 # sin cambios de rol: solo arranca la app
└── internal/
    ├── app/                        # composition root (reemplaza a root.go)
    │   └── app.go                  # cablea módulos y registra rutas/schedulers
    │
    ├── platform/                   # "shared kernel" técnico (sin lógica de negocio)
    │   ├── config/                 # ← internal/config
    │   ├── logger/                 # ← internal/logger
    │   ├── database/               # pool pgx, helpers de tx
    │   ├── cache/                  # storage redis
    │   ├── objectstore/            # cliente S3
    │   ├── mail/                   # ← internal/mail
    │   ├── geoip/                  # ← internal/geoip
    │   ├── httpx/                  # middlewares genéricos, helpers de respuesta/errores
    │   └── marketdata/             # ← prices/ + alphavantage/ + finnhub/ + yahoo/
    │
    ├── auth/                       # módulo: login, sesiones, refresh, 2FA, password reset,
    │   │                           #         verificación de email, invitaciones
    │   ├── module.go               # New(deps) → *Module; Routes(r fiber.Router)
    │   ├── domain.go               # entidades del dominio (User… ver nota de `identity`)
    │   ├── service.go / service_*.go
    │   ├── repository.go           # interfaz PEQUEÑA definida aquí (el consumidor la define)
    │   ├── postgres.go             # implementación pgx de esa interfaz
    │   ├── handler.go
    │   ├── dto.go
    │   └── *_test.go               # tests del módulo viven con el módulo
    │
    ├── user/                       # perfil, preferencias, avatar, administración
    ├── portfolio/                  # portfolios, entries, transacciones, plataformas,
    │   │                           # snapshots, import/export  (subdividir en archivos por sub-área)
    ├── market/                     # assets, exchange rates, sincronización de precios
    ├── marketing/                  # waitlist
    ├── notification/               # alertas de actividad/seguridad, resumen semanal
    │
    ├── scheduler/                  # runner genérico de cron; cada módulo aporta sus jobs
    ├── migrations/                 # sin cambios (esquema DB sigue siendo global)
    └── migrator/
```

**Decisiones de diseño que acompañan la estructura:**

1. **Interfaces definidas por el consumidor y pequeñas.** `auth.Repository` declara
   solo los ~25 métodos que auth necesita; `portfolio.Repository` los suyos. La god
   interface de 92 métodos desaparece repartida entre módulos.
2. **Cada módulo expone un tipo `Module`** con constructor `New(...)` que recibe solo
   las dependencias que usa, y un método `Routes(router)` que registra sus endpoints.
3. **Comunicación entre módulos por interfaces públicas**, definidas en el módulo
   consumidor. Ej.: `portfolio` necesita saber si un user existe → define
   `type UserReader interface { GetUserByID(...) }` y `app/` le inyecta el servicio
   de `user`. Nunca se importa el repositorio de otro módulo.
4. **Entidades viven en su módulo.** `entities.Portfolio` → `portfolio.Portfolio`.
   Para tipos realmente compartidos (User/Account/Session, usados por auth, user,
   portfolio y notification) crear un paquete mínimo `internal/identity` con solo
   esos structs — es la única excepción permitida al reparto por módulos.
5. **Los DTOs viven junto a su handler** (`auth/dto.go`), no en un árbol paralelo.
6. **Transacciones**: `platform/database` expone un helper `WithinTx(ctx, fn)` para
   los casos (import masivo) donde un service necesita atomicidad multi-repositorio.

---

## 2. Reglas del proceso de migración

- 🔒 **Nunca se rompe `main`**: cada fase termina con `go build ./... && go test ./...`
  y `golangci-lint run` en verde. Si una fase no cabe en un PR razonable, se parte.
- 🔀 **Un módulo por PR** (máximo). Los PRs de migración no mezclan refactor con
  features ni bugfixes.
- 🧪 **Primero red de seguridad, después mover código.** No se extrae un módulo cuyo
  comportamiento no esté cubierto por tests.
- 📦 **Convivencia temporal**: durante la migración coexisten `internal/services`
  (legacy) y los módulos nuevos. El bootstrap cablea ambos hasta que el legacy quede
  vacío. Está prohibido que un módulo nuevo importe `internal/services`.
- 🚫 **No se reescribe lógica al mover** (mover ≠ mejorar). Los rediseños de lógica
  se anotan en `docs/TECH_DEBT.md` y se hacen en PRs posteriores.
- 🌐 **La API HTTP no cambia**: mismas rutas, mismos contratos JSON. Los tests E2E
  del frontend (`playwright`) sirven como verificación de no-regresión.

**Verificación estándar al cerrar cada fase** (desde `backend/`):

```bash
go build ./...
go vet ./...
go test ./... -count=1
golangci-lint run
```

---

## 3. Checklist de migración

### Fase 0 — Red de seguridad y línea base *(sin mover código)*

- [x] Ejecutar y guardar la línea base: `go test ./... -coverprofile=baseline.out` y
      anotar el % de cobertura por paquete en este documento.

  **Línea base de cobertura (2026-07-13, `go test ./... -count=1 -coverprofile=baseline.out`):**

  | Paquete | Cobertura |
  |---|---|
  | `internal/alphavantage` | 84.8% |
  | `internal/config` | 73.1% |
  | `internal/dtos/portfolio` | 100.0% |
  | `internal/entities` | 65.6% |
  | `internal/finnhub` | 84.4% |
  | `internal/geoip` | 81.8% |
  | `internal/handlers` | 16.0% |
  | `internal/mail` | 37.0% |
  | `internal/middlewares` | 17.1% |
  | `internal/prices` | 100.0% |
  | `internal/services` | 80.6% |
  | `internal/yahoo` | 86.2% |
  | `pkg/helpers` | 93.3% |
  | `cmd/api`, `internal` (root), `internal/logger`, `internal/migrator`, `internal/repositories`, `internal/routes`, `internal/scheduler` | 0.0% |
  | `internal/dtos/{auth,marketing,user}`, `pkg/dtos` | sin tests (solo structs) |
  | **Total** | **42.6%** |

  > Nota: `repositories` (requiere Postgres real), `routes`, `scheduler` y los
  > bootstrap (`cmd/api`, `internal/root.go`) no tienen tests unitarios; su
  > comportamiento queda cubierto indirectamente por los tests de handlers y
  > por los E2E del frontend.
- [x] Verificar que CI ejecuta build + tests + lint del backend en cada PR; si no,
      configurarlo antes de tocar nada. → No existía: añadido
      `.github/workflows/backend-ci.yml` (build + vet + test con cobertura +
      golangci-lint en cada PR/push que toque `backend/`).
- [x] Identificar rutas HTTP sin ningún test (comparar `routes/*.go` contra los
      tests de handlers) y añadir al menos un test de humo por ruta crítica
      (auth login/refresh, CRUD de portfolio, import/export). → El CRUD de
      portfolio ya estaba cubierto (`handlers/portfolio_test.go`); añadidos
      `handlers/auth_test.go` (login, refresh, register) y
      `handlers/export_import_test.go` (export XLSX, import preview e import).
- [x] Documentar el contrato HTTP actual (métodos, paths, códigos de estado) en
      `docs/API.md` — servirá de contrato de no-regresión.
- [x] Congelar convención de errores HTTP actual (helpers de `handlers/helpers.go`)
      para replicarla idéntica en `platform/httpx`. → Congelada en `docs/API.md`
      §1.1–§1.2 (sobre de respuesta + mapeo substring→status de `responseFromDomain`).
- [x] Crear `docs/TECH_DEBT.md` para anotar mejoras detectadas durante la migración
      que NO se harán en los PRs de migración.

### Fase 1 — Crear `platform/` (shared kernel) *(solo movimientos mecánicos)*

- [x] Crear `internal/platform/` y mover, en commits separados y con `gofmt`/imports
      actualizados:
  - [x] `internal/config` → `internal/platform/config`
  - [x] `internal/logger` → `internal/platform/logger`
  - [x] `internal/mail` → `internal/platform/mail`
  - [x] `internal/geoip` → `internal/platform/geoip`
  - [x] `internal/prices` + `alphavantage` + `finnhub` + `yahoo` →
        `internal/platform/marketdata` (subpaquetes: `marketdata/alphavantage`, etc.;
        paquete renombrado `prices` → `marketdata`)
- [x] Dividir `platform/config`: separar carga de env (`env.go`) de los
      constructores de infraestructura (`db.go`, `cache.go`, `storage.go`) en
      `platform/database`, `platform/cache`, `platform/objectstore`.
- [x] Crear `internal/platform/httpx` con:
  - [x] Helpers de respuesta/error extraídos de `handlers/helpers.go` (los handlers
        legacy pasan a delegar en ellos, sin duplicar).
  - [x] Los middlewares genéricos (recovery, requestid, response_time, logger, cors,
        helmet, limiter) desde `internal/middlewares`. Los middlewares con lógica de
        negocio (`jwt`, `rbac`) se quedan donde están: migrarán al módulo `auth`.
        → `UserLimiter` se queda también en `middlewares` (su key depende del
        local `LocalUserID` del JWT), delegando en `httpx.KeyedRateLimiter`.
- [x] Añadir `platform/database.WithinTx(ctx, pool, fn)` (helper de transacciones)
      con test.
- [x] Verificación estándar + revisar que ningún paquete `platform/*` importe
      `entities`, `services`, `repositories` ni `handlers` (el kernel no conoce el
      negocio): `grep -rn "finexia-app/internal/\(services\|entities\|repositories\|handlers\)" internal/platform/` debe salir vacío. → Verificado (grep vacío,
      incluyendo también `middlewares`, `routes`, `dtos` y `scheduler`).

### Fase 2 — Módulo piloto: `marketing` *(el dominio más pequeño; valida el patrón)*

- [x] Crear `internal/marketing/` con la estructura estándar de módulo:
  - [x] `domain.go`: mover `entities.Waitlist*` (lo que aplique de `entities/marketing.go`).
  - [x] `dto.go`: mover `dtos/marketing/waitlist.go`.
  - [x] `repository.go`: definir interfaz solo con los métodos que marketing usa hoy
        (extraerlos de la god interface `services.Repository`). → 1 método
        (`SaveWaitlistEmail`).
  - [x] `postgres.go`: mover la implementación desde `repositories/marketing.go`.
  - [x] `service.go`: mover la lógica desde `services/marketing.go`; depende de su
        `Repository` + interfaz local `marketing.Mailer` con solo
        `SendWaitlistConfirmation`.
  - [x] `handler.go` + `module.go` con `New(...)` y `Routes(router fiber.Router)`
        replicando exactamente las rutas de `routes/marketing.go`.
  - [x] Mover/adaptar los tests correspondientes al paquete del módulo
        (`TestSaveWaitlistEmail` + tests de la ruta; fakes locales de ~25 líneas).
- [x] Cablear el módulo en el bootstrap (`internal/root.go` por ahora): construir
      `marketing.New(...)` y llamar a `mod.Routes(...)`; borrar el registro legacy.
- [x] Eliminar el código muerto legacy: métodos de marketing en `services`,
      `repositories/marketing.go`, `handlers/marketing.go`, `routes/marketing.go`,
      `dtos/marketing/` y sus métodos en la interfaz `services.Repository`
      (también `SendWaitlistConfirmation` salió de `services.Mailer`).
- [x] Confirmar contra `docs/API.md` que las rutas y respuestas no cambiaron
      (`POST /marketing/waitlists` y `GET /users/waitlist` intactos).
- [x] **Retrospectiva del piloto**: ajustar en este documento cualquier decisión de
      estructura que el piloto haya demostrado incómoda, antes de replicar el patrón.

  **Retrospectiva (2026-07-13):**

  1. **Registro de rutas de módulos durante la convivencia.** Hasta que exista
     `internal/app` (Fase 3), los módulos se registran vía la interfaz
     `routes.Module` (`Routes(fiber.Router)`) que el bootstrap pasa a `routes.New`.
     El orden importa: los módulos con rutas públicas se registran **antes** del
     `Use(JWT)` global. `app/` heredará esta responsabilidad en Fase 3.
  2. **Entidades compartidas entre un módulo y el legacy.** `Waitlist` la lee y
     actualiza también el flujo de invitaciones (`ListWaitlist`,
     `SetWaitlistInvited`). Patrón validado: la entidad vive en su módulo y el
     legacy la importa (`marketing.Waitlist`) — la dirección permitida es
     legacy → módulo, nunca al revés. Esos dos métodos siguen en la god
     interface, anotados para migrar con invitations en Fase 4.
  3. **El patrón de módulo no necesitó ajustes**: estructura, interfaces por
     consumidor y tests locales funcionaron tal como estaban diseñados. Se
     replica sin cambios en las fases siguientes.

### Fase 3 — Composition root: `internal/app`

- [x] Crear `internal/app/app.go` con un tipo `App` que:
  - [x] Recibe las dependencias de infraestructura (pool, cache, s3, mail, logger, env)
        vía `app.Deps`.
  - [x] Construye módulos migrados + el "módulo legacy" (services/handlers/routes
        actuales) y registra rutas de ambos. → También absorbe la construcción del
        `fiber.App` (sonic, validador, trust proxy) que vivía en `main.go`.
  - [x] Es el único lugar que arranca schedulers (`startSchedulers`).
- [x] Adelgazar `cmd/api/main.go` para que solo cree infraestructura y llame a `app.New(...).Run(ctx)`.
- [x] Eliminar `internal/root.go` (su contenido queda absorbido por `app`).
- [x] Verificación estándar + smoke test: sin docker en el entorno de CI, el smoke
      manual del frontend se hace en el PR; lo cubre además un test de arranque
      (`internal/app/app_test.go`) que compone la App real (pgx es lazy) y verifica
      health, la ruta del módulo marketing y el gate JWT de las rutas protegidas.
      `Run` quedó separado en `wire()` + `Listen` para poder testear la composición.

### Fase 4 — Módulo `auth` *(el más sensible: máximo cuidado, cero cambios de lógica)*

> **Desviación planificada (2026-07-15):** la fase se ejecuta en **3 PRs** y el
> orden original de sub-áreas no compila por commits: `Login`⇄2FA
> (`getTwoFactor`/`createTwoFactorPending` y `CompleteTwoFactorLogin`→`issueSession`)
> y `Register`→`issueEmailVerification` obligan a mover el núcleo junto.
> **PR A** = identity + núcleo (sesiones/login/refresh + 2FA + verificación de
> email) + middlewares + cleanup job; **PR B** = password reset; **PR C** =
> invitaciones + waitlist.

- [x] Crear `internal/identity/` con los structs compartidos entre módulos:
      `User`, `Account`, `Session`, `RefreshToken` (desde `entities/auth.go` y
      `entities/user.go`). Solo datos, sin lógica. → `Role` incluido; `User` sin
      los slices `Sources`/`Portfolios` (verificado: nunca se pueblan) y sin
      back-references; `ComparePassword` pasó a helper privado del módulo.
- [x] Crear `internal/auth/` y migrar por sub-áreas (ver nota de desviación):
  - [x] Sesiones + login/registro/refresh (`services/auth.go`, `services/session.go`,
        `repositories/auth.go`, `handlers/auth.go`). *(PR A)*
  - [x] 2FA (`services/two_factor.go`, `repositories/two_factor.go`, `handlers/two_factor.go`). *(PR A)*
  - [x] Verificación de email (`services/email_verification.go`, `repositories/verification.go`,
        `handlers/email_verification.go`). *(PR A)*
  - [x] Password reset (`services/password_reset.go`, `repositories/password_reset.go`,
        `handlers/password_reset.go`). *(PR B)* → `PasswordResetStore` +
        `Mailer += SendPasswordReset`; el módulo tiene su propio
        `sendPasswordChangedAlert` (el legacy conserva su copia para
        `ChangeMyPassword` hasta Fase 5); borrados `entities/auth.go` y el
        paquete `dtos/auth` completo; la god interface baja de 60 a 57 métodos.
  - [x] Invitaciones (`services/invitation.go`, `repositories/invitation.go`,
        `handlers/invitation.go`). *(PR C)* → `InvitationStore` + `WaitlistStore`
        (implementado por el **módulo marketing**, que ganó
        `ListWaitlist`/`SetWaitlistInvited`: la tabla waitlist queda 100% en su
        módulo y auth solo importa el tipo público `marketing.Waitlist`);
        las 5 rutas admin viven en `Module.AdminRoutes` con guards inline
        (`RequireAuth` + `UserLimiter` + `RequireAdmin`), registradas antes del
        gate global; DTOs `InviteUser`/`AcceptInvitation` salieron de `dtos/user`.
- [x] Definir la interfaz local de persistencia + `auth/postgres.go`. → En vez de
      una sola interfaz de ~45 métodos (violaría el criterio de ≤~30), se definió
      **un store por sub-área** (`AccountStore`, `SessionStore`, `RefreshTokenStore`,
      `TwoFactorStore`, `VerificationStore`; ninguno >11 métodos) agrupados en
      `auth.Stores`, todos implementados por un único `*PostgresRepository`.
- [x] Mover los middlewares `jwt` y `rbac` a `auth/middleware.go`; el módulo expone
      `RequireAuth()` / `RequireRole(...)` para que `app` y otros módulos protejan rutas.
      → Constantes `auth.Local*` públicas; el legacy (`handlers`, `UserLimiter`) las
      importa (dirección legacy→módulo). El gate global de `routes.Init()` usa
      `r.auth.RequireAuth()`.
- [x] Interfaces locales para efectos secundarios: `auth.Mailer` (SecurityAlert,
      EmailVerification, PasswordReset, Invitation), `auth.GeoLocator`.
- [x] Job de limpieza (`scheduler/auth_cleanup.go`): movido a `auth/cleanup_job.go`
      y registrado desde `app.startSchedulers` (runner genérico en Fase 7).
- [x] Migrar los tests del núcleo (`auth_test.go`, `two_factor_test.go`, `security_test.go`,
      `email_verification_test.go`, `handlers/auth_test.go`, `middlewares/rbac_test.go`)
      con fakes locales — el testsupport del módulo solo implementa los 5 stores
      (~40 hooks) frente a la god interface de 91 métodos del legacy.
      *(password reset/invitaciones en PR B/C)*
- [x] Eliminar el código legacy correspondiente y purgar la interfaz `services.Repository`.
      → PR A: borrados services/repos/handlers/middlewares del núcleo (91 → 60
      métodos); `ChangeMyPassword` legacy delega en el módulo vía
      `services.AuthService`. PR B: password reset (60 → 57); borrados
      `entities/auth.go` y `dtos/auth`. PR C: invitaciones + waitlist
      (57 → **49**); borrados `routes/auth.go`, `entities/invitation.go`,
      `services/token_helpers.go` y `AuthLimiter`; el `Mailer` legacy queda
      solo con ActivityAlert/SecurityAlert/WeeklySummary.
- [ ] Verificación estándar + E2E de auth del frontend (login, registro, forgot/reset
      password, verify email, accept invite). → build+vet+test+lint en verde y
      greps de frontera vacíos en los 3 PRs; **E2E manual pendiente antes de
      mergear** (login, registro+verify, refresh, sesiones, 2FA, forgot/reset,
      invite/resend/revoke/accept, waitlist admin, change password).

  **Retrospectiva (PRs A/B/C, 2026-07-15):**

  1. **Sub-áreas acopladas**: el checklist original suponía commits
     independientes por sub-área; el ciclo login⇄2FA y Register→verificación
     obligaron a mover el núcleo en un solo PR (3 commits: identity → módulo
     sin cablear → cableado+purga). Password reset e invitaciones sí son
     independientes.
  2. **Stores por sub-área en vez de interfaz única**: mantiene todas las
     interfaces pequeñas y los fakes de test triviales sin perder el único
     `PostgresRepository`.
  3. **Orden de registro con grupo protegido propio**: a diferencia de
     `marketing` (solo rutas públicas), auth tiene un `Use(RequireAuth)` local
     al grupo `/auth`; las rutas públicas legacy que quedan en `routes/auth.go`
     deben registrarse **antes** que el módulo para que sus handlers terminales
     precedan a ese middleware en el stack de Fiber.
  4. **`entities.TwoFactorRecoveryCode` no se portó**: no tenía ningún uso.
  5. **Dependencia módulo→módulo validada** (PR C): auth consume la waitlist vía
     su interfaz local `WaitlistStore`, implementada por `marketing.Service`
     e inyectada por `app` — primera dependencia entre módulos del plan, sin
     importar internals. La ruta `GET /users/waitlist` quedó en
     `auth.AdminRoutes` (dashboard de invitaciones); evaluar moverla a
     marketing en Fase 8 (TECH_DEBT #10).
  6. **Guards admin inline** (PR C): `AdminRoutes` encadena
     `RequireAuth`+`UserLimiter`+`RequireAdmin` por ruta (nunca `group.Use`)
     para no filtrar middlewares a las rutas legacy `/users/*` ni duplicar el
     rate limit.

### Fase 5 — Módulo `user`

- [ ] Crear `internal/user/`: perfil, preferencias, avatar (S3), administración
      (listar/ban/CRUD) desde `services/user.go`, `repositories/user.go`, `handlers/user.go`,
      `dtos/user/`.
- [ ] Definir `user.Repository` local + `postgres.go`.
- [ ] Interfaz local para el object store (solo `Put/Get/Delete` de avatares) en vez
      de pasar `*s3.Client` crudo al service.
- [ ] Exponer `user.Service` como interfaz pública mínima para otros módulos
      (`GetUserByID`, `GetUsersWithWeeklySummary`…), consumida vía interfaces
      definidas en cada consumidor.
- [ ] Migrar tests, eliminar legacy, verificación estándar.

### Fase 6 — Módulo `portfolio` *(el más grande: dividir agresivamente por archivos)*

- [ ] Crear `internal/portfolio/` con el código repartido por sub-área en archivos
      (no en subpaquetes, para evitar ciclos):
  - [ ] `portfolio.go` / `service_portfolio.go`: CRUD de portfolios, risks, summary.
  - [ ] `entry.go`, `transaction.go`: entries y transacciones.
  - [ ] `platform_source.go`: plataformas/fuentes de inversión.
  - [ ] `snapshot.go`: snapshots (desde `services/portfolio_snapshot.go`).
  - [ ] `import.go`, `export.go`: import/export masivo (desde `services/bulk_import.go`,
        `transaction_import.go`, `asset_import.go`, `handlers/import.go`, `handlers/export.go`);
        usar `platform/database.WithinTx` para la atomicidad.
- [ ] **Trocear `repositories/portfolio.go` (1.166 líneas)** en `postgres_portfolio.go`,
      `postgres_entry.go`, `postgres_transaction.go`, `postgres_snapshot.go`, etc.,
      todos implementando `portfolio.Repository`.
- [ ] Dependencias hacia otros módulos vía interfaces locales:
      `portfolio.AssetReader` / `portfolio.RateReader` (implementadas por `market`),
      `portfolio.UserReader` (implementada por `user`).
- [ ] Mover el job de snapshots a `portfolio/snapshot_job.go`.
- [ ] Migrar la montaña de tests (portfolio_service_test, portfolio_test,
      transaction_import_test, etc.) con fakes locales.
- [ ] Eliminar legacy, verificación estándar + E2E de dashboard/investments.

### Fase 7 — Módulos `market` y `notification` + scheduler genérico

- [ ] Crear `internal/market/`: assets, exchange rates y sincronización de precios
      (`services/asset_sync.go`, `services/exchange_rate*.go`, `repositories/exchange_rate.go`,
      `handlers/asset.go`, `handlers/exchange_rate.go`). Depende de
      `platform/marketdata.Provider`.
- [ ] Crear `internal/notification/`: resumen semanal y alertas
      (`services/weekly_summary*`, envíos de `mail` disparados por otros módulos).
      Los módulos publican necesidades vía interfaces locales (p. ej.
      `auth.SecurityNotifier`) implementadas por `notification`.
- [ ] Refactorizar `internal/scheduler/` a un runner genérico:
  - [ ] `type Job interface { Name() string; Run(ctx) error }` + `Scheduler.Register(job, schedule)`.
  - [ ] Cada módulo expone sus jobs (`auth.CleanupJob`, `portfolio.SnapshotJob`,
        `market.PriceSyncJob`, `market.RateSyncJob`, `notification.WeeklySummaryJob`).
  - [ ] `app` registra todos los jobs; desaparecen los 5 schedulers ad-hoc que
        reciben `services.Services` completo.
- [ ] Eliminar legacy, verificación estándar.

### Fase 8 — Demolición del legacy y blindaje de fronteras

- [ ] Verificar que `internal/services`, `internal/repositories`, `internal/handlers`,
      `internal/routes`, `internal/dtos`, `internal/entities` y `internal/middlewares`
      quedaron vacíos y **borrarlos**.
- [ ] Blindar las fronteras con lint (elegir una):
  - [ ] `depguard`/`importas` en `.golangci.yml` prohibiendo que `platform/*` importe
        módulos de dominio y que un módulo importe internals de otro, **o**
  - [ ] un test de arquitectura (`internal/app/arch_test.go`) que recorra los imports
        con `go/packages` y falle ante violaciones.
- [ ] Actualizar `docs/API.md` y crear `docs/ARCHITECTURE.md` con la descripción de
      la arquitectura final + reglas de dependencia (diagrama incluido).
- [ ] Comparar cobertura contra la línea base de Fase 0: no debe haber bajado.
- [ ] Revisión final de `docs/TECH_DEBT.md`: priorizar lo anotado durante la migración.

### Fase 9 — Frontend

La migración del frontend tiene su propio plan detallado (diagnóstico, arquitectura
objetivo por features + capa de API tipada, checklist por fases, riesgos y criterios
de éxito) en [`FRONTEND_ARCHITECTURE_MIGRATION.md`](./FRONTEND_ARCHITECTURE_MIGRATION.md).
Es independiente de las fases 0–8: puede avanzar en paralelo porque el contrato HTTP
no cambia en ninguna de las dos migraciones. El único punto de contacto es
`docs/API.md` (Fase 0 de ambos planes), que actúa como contrato compartido.

---

## 4. Riesgos y mitigaciones

| Riesgo | Mitigación |
|---|---|
| Regresiones en auth (dominio crítico) | Fase 4 va después del piloto; migración por sub-áreas en commits pequeños; E2E de auth obligatorio antes de mergear |
| La migración se eterniza y quedan dos arquitecturas conviviendo | Orden de fases pensado para que cada PR elimine su parte del legacy; la Fase 8 tiene como entregable el borrado físico de los paquetes viejos |
| Conflictos con features en desarrollo | Un módulo por PR y coordinar: no migrar un dominio mientras tenga una feature abierta |
| "Mover" se convierte en "reescribir" | Regla explícita de la sección 2; los rediseños van a `docs/TECH_DEBT.md` |
| Ciclos de imports entre módulos | Interfaces definidas en el consumidor + `identity` como único paquete de tipos compartidos; blindaje con lint en Fase 8 |

## 5. Criterios de éxito

- [ ] Cero paquetes "por capa técnica" globales: no existen `services/`, `handlers/`,
      `repositories/` planos.
- [ ] Ninguna interfaz de más de ~30 métodos; ningún archivo de producción > ~500 líneas.
- [ ] Un módulo nuevo se añade creando un paquete y registrándolo en `app/` (un solo
      punto de cableado).
- [ ] Los tests de un módulo solo definen fakes de las interfaces de ese módulo.
- [ ] Las reglas de dependencia están automatizadas (lint o arch-test) y fallan el CI.
- [ ] API HTTP idéntica a la documentada en Fase 0 (E2E en verde).
