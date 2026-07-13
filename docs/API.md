# Contrato HTTP del backend — Finexia

> **Propósito:** este documento congela el contrato HTTP actual del backend
> (rutas, métodos, autenticación y convención de respuestas). Es el **contrato
> de no-regresión** de la migración de arquitectura descrita en
> [`ARCHITECTURE_MIGRATION.md`](./ARCHITECTURE_MIGRATION.md): ninguna fase de
> la migración puede cambiar lo aquí descrito. Generado en Fase 0 (2026-07-13)
> a partir de `backend/internal/routes/*.go`.

---

## 1. Convenciones globales

### 1.1 Sobre de respuesta (envelope)

Todas las respuestas JSON comparten el mismo sobre, producido por los helpers
de `internal/handlers/helpers.go` (a replicar idéntico en `platform/httpx`):

**Éxito:**

```json
{
  "success": true,
  "message": "…",
  "details": "…",
  "data": {},
  "timestamp": "2026-07-13T00:00:00Z"
}
```

**Error:**

```json
{
  "success": false,
  "message": "…",
  "details": "…",
  "timestamp": "2026-07-13T00:00:00Z"
}
```

Algunos flujos (auth) añaden un campo `action` con un código estable de
máquina (p. ej. `auth:login:2fa`, `auth:register:disabled`) tanto en éxitos
(`responseSuccessAction`) como en errores (`responseErrorAction`). En los
errores mapeados desde el dominio (`responseFromDomain`) el sobre lleva
`message` + `action` (sin `details`).

### 1.2 Mapeo de errores de dominio a códigos HTTP

`responseFromDomain` mapea el **texto** del error de servicio a un status
(convención congelada; los módulos nuevos deben reproducirla exactamente):

| El mensaje de error contiene… | Status |
|---|---|
| `too many` | `429 Too Many Requests` |
| `failed` o `invalid` | `400 Bad Request` |
| `not found` | `404 Not Found` |
| `already exist`, `already found` o `duplicate` | `409 Conflict` |
| (cualquier otro) | `500 Internal Server Error` |

Otros helpers de estado directo: `responseBadRequest` (400),
`responseUnauthorized` (401), `responseInternalServerError` (500),
`responseStatusOk` (200), `responseSuccess`/`responseSuccessAction`/
`responseErrorAction` (status explícito).

### 1.3 Autenticación

- **Access token:** JWT HS256 en `Authorization: Bearer <token>`; validado por
  el middleware JWT en todo lo que está bajo la sección "requiere sesión".
- **Refresh token:** cookie `refresh_token` (`HttpOnly`, `SameSite=Strict`,
  `Secure` en producción, `Path=/`), emitida en login/2FA-login/refresh y
  rotada en cada refresh.
- **RBAC:** las rutas marcadas *admin* exigen además rol `admin`
  (middleware `RequireAdmin`); la violación responde `403`.

### 1.4 Middlewares globales

Aplicados a toda la app, en orden: `recovery`, `requestid`, `response_time`,
`logger`, `cors`, `helmet`, `limiter` (rate limit global). Las rutas públicas
de auth añaden `AuthLimiter` (rate limit más estricto) y las autenticadas
`UserLimiter` (rate limit por usuario).

### 1.5 Paginación

Las rutas marcadas *paginada* aceptan `?page=` y `?limit=` (middleware
`paginate` de Fiber) y devuelven en `data` un bloque `MetaData`:

```json
{
  "currentPage": 1,
  "<limitKey>": 20,
  "offset": 0,
  "<totalKey>": 42,
  "totalPages": 3,
  "previous": false,
  "next": true
}
```

(`limitKey`/`totalKey` conservan nombres históricos por área, p. ej.
`usersForPage`/`totalUsers`.)

---

## 2. Rutas

### 2.1 Health (público)

| Método | Path | Descripción |
|---|---|---|
| GET | `/health/livez` | Liveness probe |
| GET | `/health/readyz` | Readiness probe |
| GET | `/health/startupz` | Startup probe |

### 2.2 Marketing (público)

| Método | Path | Descripción |
|---|---|---|
| POST | `/marketing/waitlists` | Alta en la waitlist |

### 2.3 Avatar público

| Método | Path | Descripción |
|---|---|---|
| GET | `/users/:id/avatar` | Devuelve el avatar del usuario (S3) |

### 2.4 Auth — público (con `AuthLimiter`)

| Método | Path | Descripción |
|---|---|---|
| POST | `/auth/register` | Registro (403 `auth:register:disabled` si el self-registration está apagado; 409 `auth:register:duplicate` si el email existe) |
| POST | `/auth/login` | Login; 200 con `data.accessToken` + cookie `refresh_token`. Si hay 2FA: 200 con `action=auth:login:2fa` y `data.twoFactorToken`. Email sin verificar: 403 `auth:login:unverified` |
| POST | `/auth/refresh` | Rotación del refresh token (cookie); 401 si falta o es inválido |
| POST | `/auth/2fa/login` | Segundo paso del login 2FA (token pendiente + código TOTP/recovery) |
| GET | `/auth/invitations` | Valida un token de invitación |
| POST | `/auth/invitations/accept` | Acepta invitación fijando contraseña |
| POST | `/auth/password-reset` | Solicita link de reset |
| GET | `/auth/password-reset` | Valida token de reset |
| POST | `/auth/password-reset/confirm` | Confirma reset con nueva contraseña |
| POST | `/auth/verify-email` | (Re)envía link de verificación |
| GET | `/auth/verify-email` | Valida token de verificación |
| POST | `/auth/verify-email/confirm` | Marca el email como verificado |

### 2.5 Auth — requiere sesión (JWT)

| Método | Path | Descripción |
|---|---|---|
| GET | `/auth/2fa` | Estado de 2FA |
| POST | `/auth/2fa/setup` | Inicia enrolamiento 2FA |
| POST | `/auth/2fa/enable` | Confirma y activa 2FA |
| POST | `/auth/2fa/disable` | Desactiva 2FA |
| POST | `/auth/2fa/recovery-codes` | Regenera códigos de recuperación |
| GET | `/auth/session` | Sesión actual + usuario |
| GET | `/auth/sessions` | Lista de sesiones activas |
| DELETE | `/auth/sessions/:id` | Revoca una sesión |
| POST | `/auth/sessions/revoke-others` | Revoca las demás sesiones |
| POST | `/auth/logout` | Cierra la sesión actual |

### 2.6 Users (JWT; *admin* donde se indica)

| Método | Path | Acceso | Descripción |
|---|---|---|---|
| GET | `/users` | admin, paginada | Lista de usuarios |
| POST | `/users` | admin | Crea un usuario |
| GET | `/users/invitations` | admin, paginada | Lista invitaciones |
| POST | `/users/invitations` | admin | Crea invitación |
| POST | `/users/invitations/:id/resend` | admin | Reenvía invitación |
| DELETE | `/users/invitations/:id` | admin | Revoca invitación |
| GET | `/users/waitlist` | admin, paginada | Lista la waitlist |
| GET | `/users/me` | usuario | Perfil propio |
| PATCH | `/users/me` | usuario | Actualiza perfil propio |
| POST | `/users/me/avatar` | usuario | Sube avatar |
| GET | `/users/me/preferences` | usuario | Preferencias propias |
| PATCH | `/users/me/preferences` | usuario | Actualiza preferencias |
| PATCH | `/users/me/password` | usuario | Cambia contraseña |
| GET | `/users/:id` | admin | Usuario por id |
| PATCH | `/users/:id` | admin | Actualiza usuario |
| PATCH | `/users/:id/ban` | admin | Banea/desbanea |
| DELETE | `/users/:id` | admin | Elimina usuario |

### 2.7 Portfolios (JWT)

| Método | Path | Acceso | Descripción |
|---|---|---|---|
| GET | `/portfolios/risks` | usuario | Catálogo de niveles de riesgo |
| GET | `/portfolios/id` | usuario | Portfolios del usuario |
| GET | `/portfolios/summary` | usuario | Resumen (soporta `?currency=`) |
| GET | `/portfolios/transactions` | usuario | Transacciones recientes |
| POST | `/portfolios/transactions/import/preview` | usuario | Preview del import (multipart `file`, `sheet`, `mapping`, `defaults`) |
| POST | `/portfolios/transactions/import` | usuario | Import masivo (además `portfolioId`, `sourceId`; `mapping` obligatorio) |
| GET | `/portfolios/allocation` | usuario | Asignación de activos |
| POST | `/portfolios` | usuario | Crea portfolio |
| POST | `/portfolios/sources` | usuario | Crea plataforma/fuente |
| POST | `/portfolios/entries` | usuario | Crea posición (entry) |
| GET | `/portfolios/entries/:entryId/transactions` | usuario | Transacciones de una posición |
| POST | `/portfolios/entries/:entryId/transactions` | usuario | Crea transacción |
| PUT | `/portfolios/transactions/:txnId` | usuario | Actualiza transacción |
| GET | `/portfolios/sources` | usuario | Lista plataformas |
| PATCH | `/portfolios/sources/:id` | usuario | Actualiza plataforma |
| DELETE | `/portfolios/sources/:id` | usuario | Elimina plataforma |
| GET | `/portfolios/assets` | usuario, paginada | Catálogo de assets |
| PATCH | `/portfolios/assets/:id/price` | admin | Fija precio manual de un asset |
| GET | `/portfolios/growth` | usuario | Crecimiento agregado (`?since=`) |
| GET | `/portfolios/export/summary` | usuario | XLSX `resumen-mensual.xlsx` |
| GET | `/portfolios/export/transactions` | usuario | XLSX `transacciones.xlsx` |
| GET | `/portfolios/export/risk` | usuario | XLSX `riesgo-volatilidad.xlsx` |
| PATCH | `/portfolios/:id` | usuario | Actualiza portfolio |
| GET | `/portfolios/:id/top-transaction` | usuario | Mayor transacción |
| GET | `/portfolios/:id/growth` | usuario | Crecimiento del portfolio |
| GET | `/portfolios/:id/assets/:symbol/transactions` | usuario, paginada | Transacciones por asset |
| GET | `/portfolios/:id` | usuario | Portfolio por id |

Los exports responden `200` con cuerpo binario XLSX y
`Content-Disposition: attachment; filename="…"` (sin sobre JSON).

### 2.8 Assets (JWT + admin)

| Método | Path | Descripción |
|---|---|---|
| POST | `/assets` | Crea asset |
| POST | `/assets/import` | Import masivo de assets |
| POST | `/assets/sync` | Sincroniza precios de todos los assets |
| POST | `/assets/:id/sync` | Sincroniza un asset |

### 2.9 Exchange rates (JWT; *admin* donde se indica)

| Método | Path | Acceso | Descripción |
|---|---|---|---|
| GET | `/exchange-rates` | usuario, paginada | Lista tasas |
| POST | `/exchange-rates` | admin | Crea tasa |
| POST | `/exchange-rates/import` | admin | Import masivo |
| POST | `/exchange-rates/sync` | admin | Sincroniza todas |
| POST | `/exchange-rates/:id/sync` | admin | Sincroniza una |
| PATCH | `/exchange-rates/:id` | admin | Actualiza tasa |

---

## 3. Reglas de no-regresión

1. Ningún PR de migración añade, elimina ni renombra rutas de este documento.
2. El sobre de respuesta (§1.1) y el mapeo de errores (§1.2) se replican
   byte-a-byte en `platform/httpx`.
3. Cualquier discrepancia detectada entre este documento y el código se
   corrige **en el documento** (el código actual es la fuente de verdad) y se
   anota en el PR.
