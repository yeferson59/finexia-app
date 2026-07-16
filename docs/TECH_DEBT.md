# Deuda técnica — Finexia backend

> Registro de mejoras detectadas durante la migración de arquitectura
> ([`ARCHITECTURE_MIGRATION.md`](./ARCHITECTURE_MIGRATION.md)) que **NO** se
> hacen en los PRs de migración (regla: mover ≠ mejorar). Cada entrada indica
> dónde se detectó y qué se propone. Priorizar al cerrar la Fase 8.

| # | Detectado en | Descripción | Propuesta |
|---|---|---|---|
| 1 | Fase 0 | `responseFromDomain` mapea errores por **substring del mensaje** y el orden de los `if` hace que un error como `"failed: portfolio not found"` devuelva 400 en vez de 404. Contrato frágil: cambiar el texto de un error cambia el status HTTP. | Migrar a errores tipados/sentinela por dominio (`errors.Is`) y mapear tipo→status en `platform/httpx`. Hacerlo **después** de la migración, replicando primero la convención tal cual (ver `docs/API.md` §1.2). |
| 2 | Fase 0 | `handlers.GetParamID` nunca devuelve error: la firma `(string, error)` es ruido y sus llamadores manejan un error imposible. | Eliminar el retorno de error (o el helper completo) al mover cada handler a su módulo. |
| 3 | Fase 0 | La ruta `GET /portfolios/id` (lista de portfolios del usuario) tiene un path atípico; lo esperable sería `GET /portfolios`. | Es un cambio de contrato HTTP: coordinar con frontend fuera de la migración (alias + deprecación). |
| 4 | Fase 0 | `repositories/` (0% cobertura) solo se prueba indirectamente; no hay tests de integración contra Postgres real. | Evaluar tests de integración con testcontainers para las implementaciones `postgres.go` de cada módulo tras la Fase 8. |
| 5 | Fase 0 | Los schedulers (`internal/scheduler/`, 0% cobertura) no tienen tests; cada uno duplica el patrón ticker+run. | La Fase 7 introduce el runner genérico `Job`/`Scheduler.Register`; añadir tests al runner en esa fase y dejar los jobs como funciones puras testeables. |
| 6 | Fase 4 (PR A) | `auth.Module.RequireAuth` valida el token con un `context.Context` capturado en el arranque (`Deps.Ctx`), herencia del legacy `middlewares.New(ctx, …)`. | Usar el contexto de la request (`c.Context()`/`c.RequestCtx()`) en `TokenProcessorFunc` y eliminar `Deps.Ctx`. |
| 7 | Fase 4 (PR A) | Duplicaciones temporales por la convivencia legacy↔módulo: `auth/postgres.go::createUser` (copia de `repositories/user.go::CreateUser`, se unifica en Fase 5), `sendPasswordChangedAlert`+`truncate`/`sanitizeIP` en `services/user.go` vs. sus originales en el módulo (mueren en Fase 5), `nextRunTime` copiado en `auth/cleanup_job.go` (Fase 7). (`services/token_helpers.go` ya fue eliminado en PR C.) | Cada copia tiene dueño y fase de eliminación; verificar su borrado al cerrar esas fases. |
| 8 | Fase 4 (PR A) | El módulo auth recibe `*config.Env` completo aunque solo usa ~11 campos. | Extraer un `auth.Config` propio poblado por `app` cuando el resto de módulos exista y se vea el patrón común. |
| 9 | Fase 4 (PR A) | `GetUserByID`/`GetUserByEmail` viven en `auth.AccountStore` pero pertenecen al dominio user (tablas users/roles). | En Fase 5, `auth` debe consumirlos vía una interfaz local implementada por el servicio del módulo `user`. |
| 10 | Fase 4 (PR C) | `GET /users/waitlist` (lectura admin de la waitlist) se sirve desde `auth.AdminRoutes` porque forma parte del dashboard de invitaciones, aunque el dato es del módulo marketing. | Evaluar en Fase 8 mover la ruta a marketing cuando exista un guard admin compartido entre módulos. |
