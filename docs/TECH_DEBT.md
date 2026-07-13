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
