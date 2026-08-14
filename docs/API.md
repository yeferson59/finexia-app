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

`httpx.FromDomain` resuelve el status **exclusivamente por el tipo del error**
(ver `TECH_DEBT.md` #1). Cada dominio etiqueta sus errores en el origen con un
`httpx.Kind` mediante los helpers `httpx.AsBadRequest` / `AsNotFound` /
`AsConflict` / `AsTooManyRequests` (o `Tagged(kind, err)`); `FromDomain`
recupera ese `Kind` vía `errors.As`, de modo que el status es **independiente
del texto** del mensaje y robusto ante el envoltorio con `%w`:

| Etiqueta del error | Status |
|---|---|
| `AsTooManyRequests` | `429 Too Many Requests` |
| `AsBadRequest` | `400 Bad Request` |
| `AsNotFound` | `404 Not Found` |
| `AsConflict` | `409 Conflict` |
| (sin etiqueta) | `500 Internal Server Error` |

El antiguo mapeo por **substring del mensaje** fue **retirado**: un error sin
etiquetar siempre es 500. Todo error de dominio que deba exponer un status
distinto de 500 se etiqueta en su origen (servicio o repositorio; p. ej. las
violaciones de unique constraint se traducen a un sentinela `AsConflict`). Los
errores realmente internos (fallos de proveedor de precios, de S3, de DB) no se
etiquetan y quedan como 500.

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
| GET | `/portfolios` | usuario | Portfolios del usuario |
| GET | `/portfolios/id` | usuario | **Deprecado** — alias de `GET /portfolios`. Misma respuesta, más las cabeceras `Deprecation: true` y `Link: </portfolios>; rel="successor-version"` |
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
| DELETE | `/portfolios/transactions/:txnId` | usuario | Elimina transacción; la posición se recalcula sola y queda en cantidad 0 si era la última |
| GET | `/portfolios/sources` | usuario | Lista plataformas |
| PATCH | `/portfolios/sources/:id` | usuario | Actualiza plataforma |
| DELETE | `/portfolios/sources/:id` | usuario | Elimina plataforma |
| GET | `/portfolios/assets` | usuario, paginada | Catálogo de assets (curados + los que aportó el llamante; §2.8) |
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

#### Procedencia del precio

Bajo BYO-key la valoración prefiere el precio que trajo la clave del propio
usuario, luego el precio manual del operador, y si no hay ninguno **valora la
posición a su coste de compra** (§2.10). Los tres dan un número de la misma
forma y significan cosas distintas, así que las respuestas dicen cuál sirvieron.

Cada holding de `GET /portfolios/:id` lleva:

| Campo | Valor |
|---|---|
| `priceSource` | `own` — precio de la clave del propio usuario |
| | `manual` — precio de referencia del operador, sin garantía de frescura |
| | `cost` — no hay precio; la posición vale su coste |
| `priceUpdatedAt` | Cuándo se obtuvo `marketPrice`; `null` si no hay |

Con `priceSource: "cost"`, `marketPrice` viene **vacío** —nunca `"0"`— y la
rentabilidad de esa posición es cero por construcción, no por mercado: un
cliente que la pinte como retorno está inventando el dato.

`GET /portfolios/summary` (y `GET /portfolios`) traen el mismo desglose
agregado por portfolio: `positionsPricedOwn`, `positionsPricedManual` y
`positionsAtCost`, que **suman exactamente `totalPositions`**. Sirven para
avisar de que un total solo está parcialmente valorado a mercado. No son
importes: `?currency=` convierte los totales y deja estos tres intactos.

#### Moneda de los holdings

Una posición arrastra hasta tres monedas: la base del portfolio, la de coste
(`costCurrency`, en la que se liquidó la compra) y la del activo (`currency`, en
la que cotiza). Sumar los importes crudos de dos posiciones en monedas distintas
no significa nada, y restar un valor de mercado en EUR de un coste en USD
significa menos todavía. Por eso cada holding de `GET /portfolios/:id` trae sus
totales ya convertidos a `baseCurrency`:

| Campo | Valor |
|---|---|
| `costBasisBase` | `quantity × price` convertido a la moneda base |
| `marketValueBase` | `quantity × marketPrice` convertido a la moneda base (a coste si `priceSource` es `cost`) |
| `fxConverted` | `false` si faltaba la tasa; los dos importes vienen entonces **sin convertir**, en su moneda nativa |

Son los únicos campos del holding que un cliente puede sumar entre posiciones.
`price` y `marketPrice` siguen siendo por unidad y en su moneda nativa: se
convierten importes, nunca precios unitarios, porque redondear a los dos
decimales de la moneda destino destruiría el precio de una fracción de acción o
de una cripto.

Con `fxConverted: false` la posición se muestra, pero cualquier total que la
incluya está mezclando monedas: hay que decirlo en pantalla en vez de sumar. Las
tasas son datos BYO-key (§2.10), así que la ausencia se resuelve sincronizando
con la clave del propio usuario, no reintentando.

### 2.8 Assets (JWT; *admin* donde se indica)

| Método | Path | Acceso | Descripción |
|---|---|---|---|
| POST | `/assets` | usuario | Añade un asset al catálogo |
| POST | `/assets/import` | admin | Import masivo de assets |

`POST /assets` hace dos cosas distintas según quién llame, con el mismo cuerpo
(`ticker`, `name`, `assetType`, `currency`, `exchange` opcional) y la misma
respuesta `201`:

| Llamante | Efecto |
|---|---|
| admin | **Cura** la fila: `isCurated: true`, visible para todos, y sobrescribe los metadatos si el ticker ya existía |
| usuario | **Aporta** la fila: `isCurated: false`, visible solo para quien la aportó, y **nunca** sobrescribe un ticker existente |

Si el ticker ya está en el catálogo, la aportación devuelve la fila existente y
la añade al catálogo del llamante; la respuesta no distingue entre "creada" y
"ya existía", porque esa diferencia solo informaría de lo que ha aportado otro
usuario. El aporte está limitado a **50 activos nuevos por usuario cada 24
horas** (`429` al superarlo).

Esto es lo que reemplaza al viejo `admin`-only: bajo el modelo de proveedores el
catálogo era del operador porque el operador pagaba la cuota. Con BYO-key cada
usuario sincroniza sus propias tenencias con su propia clave, y el import de
transacciones (§2.7) ya creaba filas para cualquier usuario que subiera un
archivo — la puerta solo estaba cerrada por delante.

El catálogo que devuelve `GET /portfolios/assets` está acotado en consecuencia:
las filas curadas más las que ha aportado el llamante. Un admin ve la tabla
entera, porque modera lo que aportan los usuarios.

> Las rutas de sincronización de administración (`POST /assets/sync`,
> `POST /assets/:id/sync`) **fueron retiradas**: los datos de mercado son
> BYO-key y la aplicación ya no tiene clave de proveedor con la que ejecutarlas.
> Su sustituta es `POST /market/sync` (§2.10), con el alcance del llamante.

### 2.9 Exchange rates (JWT; *admin* donde se indica)

| Método | Path | Acceso | Descripción |
|---|---|---|---|
| GET | `/exchange-rates` | admin, paginada | Lista tasas |
| GET | `/exchange-rates/latest` | usuario | Tasas compartidas vigentes, sin paginar |
| POST | `/exchange-rates` | admin | Crea tasa |
| POST | `/exchange-rates/import` | admin | Import masivo |
| POST | `/exchange-rates/refresh` | admin, limitado | Relee el feed público ahora |
| PATCH | `/exchange-rates/:id` | admin | Actualiza tasa |

Cada fila lleva un campo `source` que dice quién la puso:

- `manual` — la escribió un administrador (`POST`, `PATCH`) o vino de una hoja
  importada. Es el valor por defecto de la columna, así que lo llevan también
  todas las filas anteriores a que existiera el feed.
- `dolarapi` — la publicó el feed público de [dolarapi.com](https://dolarapi.com):
  un solo par, **USD → COP a la TRM**, la tasa oficial que publica la
  Superintendencia Financiera.
- `ecb` — la derivó el feed de [referencia del BCE](https://www.ecb.europa.eu/stats/eurofxref/eurofxref-daily.xml),
  que publica cada día hábil cuánto vale un euro en cada divisa. De ahí salen
  **USD ↔ X** para EUR, GBP, CHF, JPY, CAD, AUD, CNY, MXN y BRL.

Los dos feeds se leen juntos porque ninguno cubre lo del otro: el BCE no cotiza
el peso colombiano y dolarapi no cotiza nada más. Que uno esté caído no impide
guardar lo que publicó el otro; el fallo se reporta y los pares del feed caído
conservan su valor anterior.

Las tasas del BCE se anclan al dólar y en las dos direcciones a propósito. Al
dólar porque es la divisa por la que salta la conversión cuando no hay par
directo (`GetConversionRate` prueba directo, inverso y salto por USD), así que
USD ↔ X deja cualquier par de esa lista a dos saltos. En las dos direcciones
porque la vista `portfolio_summary` busca el par tal cual está escrito y no lo
invierte.

> Las tasas que trae la clave de un usuario siguen yendo a
> `user_exchange_rates` y únicamente él las ve; una conversión consulta primero
> la suya y solo después esta tabla. `POST /exchange-rates/sync` y
> `POST /exchange-rates/:id/sync` fueron retiradas por el mismo motivo que las
> de assets.

Lo que sí puede estar aquí, y es la novedad, es una tasa que trajo la
aplicación: el feed no pide clave, publica un dato oficial y no impone
condiciones de redistribución, así que una sola lectura sirve para todos los
usuarios. Eso es justo lo que esta tabla es, y por eso la restricción de la
migración 000018 —que vació esta tabla— no le aplica: lo que no se puede
compartir es lo que pagó la clave de alguien.

El job `public-exchange-rates` la refresca cada hora. Un despliegue recién
creado espera ese intervalo antes de tener tasa, porque la primera ejecución de
un `Every` persistido cae un intervalo por delante; `POST /exchange-rates/refresh`
existe para no esperarla.

Un refresco **sobrescribe** el par que toca, incluida una tasa introducida a
mano. Es la precedencia buscada: la alternativa —respetar las filas manuales—
dejaría el par clavado a un número escrito a mano para siempre, porque no hay
endpoint para borrarlo y devolvérselo al feed. Corregir a mano un par que el
feed cubre es corregirlo hasta el siguiente refresco, y `source` es lo que hace
eso visible.

### 2.10 Datos de mercado — BYO-key (JWT)

Cada usuario aporta su propia clave de proveedor. Todas estas rutas actúan
sobre las claves y tenencias de **quien llama**: el id de usuario sale de los
locals de autenticación, nunca del path, así que no hay variante de admin ni
forma de nombrar la credencial de otro.

Las de escritura llevan un limitador propio, más estrecho que el compartido
(~10/min por usuario): guardar y verificar gastan cuota del proveedor del
usuario.

| Método | Path | Descripción |
|---|---|---|
| GET | `/market/credentials` | Estado de las claves propias |
| PUT | `/market/credentials/:provider` | Guarda una clave (body `{ "apiKey": "…" }`) |
| POST | `/market/credentials/:provider/verify` | Recomprueba una clave guardada |
| DELETE | `/market/credentials/:provider` | Borra una clave |
| POST | `/market/sync` | Sincroniza las tenencias propias con las claves propias |

`:provider` es `finnhub` o `alphavantage`.

**La clave nunca se devuelve**, ni siquiera a su dueño: se guarda cifrada
(cifrado de sobre, `internal/platform/secretbox`) y de ella solo se sirven sus
cuatro últimos caracteres. El objeto de respuesta no tiene campo donde quepa:

```json
{
  "provider": "finnhub",
  "last4": "3f9a",
  "status": "active",
  "lastVerifiedAt": "2026-07-26T09:30:00Z",
  "createdAt": "2026-07-20T11:00:00Z",
  "updatedAt": "2026-07-26T09:30:00Z"
}
```

`status` es `active`, `invalid` (el proveedor rechazó la clave) o
`rate_limited` (la clave sirve, pero su cuota está agotada).

`PUT` verifica la clave contra el proveedor **antes** de guardarla, así que un
400 significa que el proveedor la rechazó. Los errores del proveedor no se
propagan al cuerpo de la respuesta: Alpha Vantage lleva la clave en el query
string y su texto de error puede citarla.

Un proveedor **inalcanzable** (timeout, 5xx, respuesta ilegible) no es lo mismo
que una clave rechazada: da 500, no 400, y no toca el estado almacenado. Marcar
`invalid` en ese caso sacaría la clave de las consultas de sincronización de
forma permanente, así que una caída del proveedor retiraría en silencio una
clave que funciona.

Una clave cuya cuota está agotada **sí se guarda**, con `status: "rate_limited"`:
negarse a guardarla dejaría sin configurar una clave perfectamente válida solo
por la hora a la que el usuario la introdujo.

`POST /market/sync` sincroniza precios **y** tasas de cambio de quien llama, y
devuelve ambos:

```json
{
  "prices": [
    { "assetId": "…", "ticker": "AAPL", "price": "190.55", "source": "finnhub", "fetchedAt": "…" }
  ],
  "rates": [
    { "fromCurrency": "USD", "toCurrency": "COP", "rate": "4100.5", "source": "finnhub", "fetchedAt": "…" }
  ]
}
```

Las tasas van en el mismo viaje porque una posición cotizada en otra moneda no
vale nada sin ellas, y bajo BYO-key no se puede usar la de otro usuario. El
trabajo se corta a los 60 s y devuelve lo que dio tiempo a traer: el sync
espacia sus llamadas para no agotar la cuota personal (13 s entre peticiones a
Alpha Vantage), así que una cartera grande no cabe en una petición HTTP. El
resto lo recoge el job diario.

---

## 3. Reglas de no-regresión

1. Ningún PR de migración añade, elimina ni renombra rutas de este documento.
2. El sobre de respuesta (§1.1) y el mapeo de errores (§1.2) se replican
   byte-a-byte en `platform/httpx`.
3. Cualquier discrepancia detectada entre este documento y el código se
   corrige **en el documento** (el código actual es la fuente de verdad) y se
   anota en el PR.
