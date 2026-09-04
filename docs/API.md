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
| DELETE | `/users/waitlist/:id` | admin | Elimina una entrada de la waitlist |
| GET | `/users/me` | usuario | Perfil propio |
| PATCH | `/users/me` | usuario | Actualiza perfil propio (`preferredCurrency`, §2.7) |
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
| GET | `/portfolios/allocation` | usuario | Asignación de activos por categoría (soporta `?currency=`) |
| GET | `/portfolios/holdings` | usuario | Activos consolidados: una fila por activo sumando todos los portfolios (soporta `?currency=`) |
| POST | `/portfolios` | usuario | Crea portfolio |
| POST | `/portfolios/sources` | usuario | Crea plataforma/fuente |
| POST | `/portfolios/entries` | usuario | Crea posición (entry); la categoría sale del activo, no del cuerpo |
| DELETE | `/portfolios/entries/:entryId` | usuario | Elimina una posición **y todas sus transacciones**; devuelve `deletedTransactions` |
| GET | `/portfolios/entries/:entryId/transactions` | usuario | Transacciones de una posición |
| POST | `/portfolios/entries/:entryId/transactions` | usuario | Crea transacción |
| PUT | `/portfolios/transactions/:txnId` | usuario | Actualiza transacción |
| DELETE | `/portfolios/transactions/:txnId` | usuario | Elimina transacción; la posición se recalcula sola y queda en cantidad 0 si era la última |
| GET | `/portfolios/sources` | usuario | Lista plataformas con sus totales (soporta `?currency=`) |
| PATCH | `/portfolios/sources/:id` | usuario | Actualiza plataforma |
| DELETE | `/portfolios/sources/:id` | usuario | Elimina plataforma |
| GET | `/portfolios/assets` | usuario, paginada | Catálogo de assets (curados + los que aportó el llamante; §2.8) |
| PATCH | `/portfolios/assets/:id/price` | admin | Fija precio manual de un asset |
| GET | `/portfolios/growth` | usuario | Crecimiento agregado (`?since=`, `?currency=`) |
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

Junto a ellos viaja `positionsUnconverted`, que cuenta las posiciones cuyos
importes **no** están en `baseCurrency` porque ninguna tasa conectaba su moneda
con ella. Siguen sumadas en los totales con su importe nativo —descartarlas
subestimaría el portfolio, y es la misma decisión que toma `fxConverted` en los
holdings—, así que un valor > 0 significa que `totalCostBase` y
`totalMarketValue` están sumando monedas distintas y hay que decirlo en pantalla
en vez de presentarlos como comparables. No particiona los otros tres: una
posición puede estar valorada con la clave de su dueño y aun así sin convertir.

Hasta la migración 000024 esto no se podía saber: cuando faltaba la tasa la
vista multiplicaba por **1**, de modo que el resumen afirmaba en silencio que un
euro es un dólar. Esa migración también sustituye las búsquedas de tasa en línea
por la función SQL `fx_rate()`, que resuelve el par con la misma regla que la
aplicación (`GetConversionRate`): tasa propia antes que compartida, inversa si
solo se guardó la dirección contraria, y salto por USD si no hay par directo.
Dos capas con la misma regla evitan el caso más confuso — un aviso en una
pantalla contradiciendo un total en otra.

#### Conversión parcial del resumen

Con `?currency=`, cada portfolio se convierte por separado y trae
`displayCurrency` —la moneda en la que están sus totales— y `fxConverted`, que
dice si esa es la que se pidió. Un portfolio cuya moneda base no conecta con la
solicitada **no falla la petición**: se devuelve con sus totales intactos,
`displayCurrency` igual a `baseCurrency` y `fxConverted: false`.

Antes ese caso daba 404 para la lista entera, de modo que un solo portfolio en
una moneda que nadie cotiza vaciaba el panel completo. La contrapartida es que
el cliente tiene que mirar el campo: las filas con `fxConverted: false` se
pueden mostrar —con su propia moneda al lado— pero **no se pueden sumar** con
las demás. El frontend las excluye de los agregados y dice cuántas quedaron
fuera.

#### Moneda de la cuenta

`preferredCurrency` (en `PATCH /users/me`) es la moneda en la que la cuenta ve
sus totales: es lo que usan el panel y el listado de portfolios cuando no se
pide otra cosa, y lo que `?currency=` omitido resuelve en la asignación.

No admite cualquier código ISO, sino los que la aplicación puede **convertir**,
que es una lista bastante más corta:

```
USD, COP, EUR, GBP, CHF, JPY, CAD, AUD, CNY, MXN, BRL
```

Es exactamente lo que publican contra el USD las dos fuentes públicas sin clave
—el feed de referencia del BCE y la TRM de dolarapi (§ Tasas)—, así que la
conversión funciona para cualquier cuenta y no solo para una que haya traído su
propia clave de mercado. Un código fuera de la lista responde **400** diciendo
cuáles valen, en vez de guardarse y quedar en nada: la preferencia decide qué
pares pide el sync, así que una moneda sin fuente detrás dejaría a la cuenta
pidiendo un par inexistente y con cifras que nadie puede convertir.

La misma lista es la que aceptan los `?currency=` del resumen, la asignación y
las transacciones. Ampliarla es un cambio en un solo sitio
(`internal/platform/currency`) en cuanto una fuente publique el par contra USD;
el sync deriva los pares que trae de lo que cada usuario tiene y prefiere
(`GetRequiredCurrencyPairs`), no de una lista fija.

#### Moneda de la asignación

`GET /portfolios/allocation` suma por categoría **todos** los portfolios del
usuario, que pueden estar denominados en monedas distintas, así que siempre
responde convertido: `?currency=` acepta lo mismo que el resumen y, si se omite,
se usa la moneda preferida de la cuenta. No hay modo «sin convertir» — un
reparto porcentual sobre una suma de euros y dólares no significa nada, que es
lo que devolvía antes.

Cada elemento lleva `currency` —la misma para todos, que es lo que hace
comparables los `percent`— y `positionsUnconverted`, con el mismo significado
que en el resumen: posiciones incluidas a valor nominal por no haber tasa.

#### De dónde sale la categoría

La categoría de una posición es el **tipo del activo** (`assets.asset_type`),
traducido al vocabulario plural que devuelve este endpoint (`stocks`, `etfs`,
`bonds`, …). El catálogo es la única fuente: el `category` que llevan los
holdings y las entries se deriva del mismo sitio, así que el donut del panel y
el del portfolio no pueden discrepar.

Antes existía `portfolio_entries.category`, una copia del tipo tomada al crear
la entry que nadie volvía a escribir. Agrupar por ella era lo que hacía que dos
gráficos sobre las mismas posiciones no coincidieran: corregir el tipo de un
activo (un ETF de bonos fichado primero como ETF normal) movía la posición en el
donut del portfolio, que agrupa los holdings por su tipo de activo, y la dejaba
en la porción vieja en el del panel. **La migración 000026 elimina esa columna**
—y su enum, que no tenía otro uso— junto con el campo `category` del cuerpo de
`POST /portfolios/entries`, que solo servía para llenarla.

Ese campo se sigue **aceptando e ignorando**, así que un cliente antiguo que lo
mande no se rompe; lo que ya no ocurre es que un valor inválido devuelva 400,
porque no hay nada que validar. La clase de una posición la decide el activo.

#### Activos consolidados

`GET /portfolios/holdings` contesta «¿cuánto tengo de X?» sin preguntar en qué
portfolio está: una fila por activo, con las unidades sumadas entre portfolios,
el valor de la posición y su peso sobre el total. Ninguna otra vista la
contesta —los holdings de `GET /portfolios/:id` solo suman dentro de su
portfolio y la asignación pliega todo a ocho clases de activo—, así que un
activo repartido entre tres portfolios no tenía una sola fila en ninguna parte.

Es la asignación un nivel más fino y comparte con ella todas sus reglas: mismo
`?currency=` (omitido, la moneda de la cuenta), mismo orden de precios —el del
usuario, luego el manual del catálogo, luego el coste—, misma elección de la
moneda junto con el precio y mismo tratamiento de una posición sin tasa, que se
cuenta a valor nominal y se declara. Los `percent` de las dos respuestas salen
del mismo cálculo sobre las mismas posiciones, así que las dos gráficas no
pueden discrepar.

Lo que solo trae esta:

- `quantity`, las unidades. Solo significan algo dentro de su fila: sumar
  acciones con bitcoins no da nada, y por eso el peso se mide en dinero.
- `portfolios`, en cuántos aparece el activo.
- `marketPrice`, el precio por unidad **en la moneda del activo** (`currency`),
  no en `displayCurrency`: es lo que cotiza, no lo que se convirtió. Llega
  **vacío** cuando `priceSource` es `cost` — cada entry pagó el suyo y ningún
  número representa al activo. Vacío no es cero.

Las entries en cantidad 0 quedan fuera: una posición vendida entera no es algo
que el usuario tenga.

#### Moneda de la serie de crecimiento

`GET /portfolios/growth` agrega los snapshots de **todos** los portfolios, cada
uno guardado en su propia moneda base, así que convierte por la misma razón que
la asignación: `?currency=` acepta la misma lista y, omitido, manda la moneda
preferida de la cuenta. `summary.currency` dice en cuál quedó la serie y cada
punto lleva `portfoliosUnconverted`, los portfolios sumados esa fecha a valor
nominal por no haber tasa.

La conversión usa la tasa de hoy en todas las fechas: la app guarda la última
tasa por par, no una serie histórica. Los puntos pasados quedan por tanto
reexpresados a la tasa actual —la presentación habitual a moneda constante— que
es el límite honesto de lo que se guarda; la alternativa era sumarlos crudos.

`GET /portfolios/:id/growth` es un solo portfolio y por tanto una sola moneda:
no convierte nada y `portfoliosUnconverted` es siempre 0.

#### Moneda de las plataformas

`GET /portfolios/sources` devuelve por plataforma un `totalValue`, que es el
coste de sus posiciones —cantidad × coste medio ponderado— sumado sobre entries
que pueden estar liquidadas en monedas distintas. Convierte por la misma razón
que la asignación: `?currency=` acepta la misma lista y, omitido, manda la
moneda preferida de la cuenta.

Cada plataforma lleva `displayCurrency` —la moneda en la que quedó su
`totalValue`— y `positionsUnconverted`, con el mismo significado que en el
resumen: posiciones incluidas a valor nominal por no haber tasa.

Hasta entonces `totalValue` salía de un `SUM` sobre la columna sin mirar
`cost_currency`, así que una plataforma con una compra en pesos y tres en
dólares devolvía sus importes nominales sumados: un número en ninguna moneda, e
inflado en cuanto entraba una de unidad menor. El cliente, además, no tenía cómo
saberlo — lo pintaba con un «$» fijo.

#### Métricas por plataforma

`GET /portfolios/sources` devuelve las plataformas **ordenadas por lo invertido,
de mayor a menor** (empates por nombre, para que el orden no cambie entre
lecturas). El orden por fecha de creación no decía nada de la cuenta: la
plataforma abierta más tarde no es la que importa.

Cada fila lleva, toda en `displayCurrency`:

| Campo | Valor |
|---|---|
| `totalValue` | Lo invertido: cantidad × coste medio ponderado |
| `marketValue` | Lo que vale hoy, sobre **las mismas posiciones** y en la misma moneda |
| `gainLoss` | `marketValue − totalValue` |
| `gainLossPct` | Esa diferencia sobre lo invertido, en % |
| `percent` | Parte del total invertido de la cuenta que vive en esta plataforma |
| `investments` | Posiciones **abiertas** |
| `assets` / `portfolios` | Activos y portafolios distintos sobre los que se reparten |
| `positionsUnconverted` | Posiciones contadas a valor nominal por falta de tasa **en cualquiera de las dos patas** |
| `positionsPricedOwn` / `positionsPricedManual` / `positionsAtCost` | De dónde salió `marketValue`; suman `investments` |

`percent` es lo que hace legible el orden: «la más grande» dice poco hasta que
es «la más grande, y tiene el 62% del dinero». Es participación sobre el coste,
no sobre el valor de mercado, así que responde dónde se puso el dinero y no
dónde resultó que creció.

Viene en `0` cuando la plataforma se lee sola —la relectura tras un `PATCH`—:
una participación necesita el conjunto, e inventar un 100% para una fila sería
peor que no decir nada. `gainLoss` sí viaja ahí, porque no lo necesita.

El precio de cada posición y la moneda en la que cotiza se eligen juntos, igual
que en `/portfolios/holdings`: una posición sin precio de mercado se valora a su
propio coste, que es un importe en la moneda de coste y no en la del activo.

##### Qué cuenta como posición, y qué no

Solo las que tienen `quantity > 0`, el mismo filtro que aplican
`/portfolios/holdings` y la asignación: una posición vendida del todo no es algo
que la plataforma tenga. Nunca aportó a los importes —cantidad cero multiplica y
desaparece— pero sí se contaba en `investments` y, si su moneda no tenía tasa,
en `positionsUnconverted`: una plataforma vaciada hace años informaba posiciones
que ya no existían y un aviso de cambio sobre un importe de cero.

##### Por qué hacen falta los tres contadores de precio

Una posición sin precio de mercado se valora **a su propio coste**, que es justo
contra lo que se la compara, así que aporta exactamente cero a `gainLoss`. Un
`gainLoss` de cero es entonces lo que informa una plataforma plana y también una
que nadie ha valorado, y `positionsAtCost` es lo único que las separa. Se toman
sobre la posición y no sobre el activo, así que los tres suman `investments`.

##### Qué no incluye lo invertido

`totalValue` es el coste de lo que **sigue en cartera**: `quantity` es el neto
que queda tras las ventas, así que una posición vendida a medias arrastra la
mitad de su coste. No lleva comisiones —el coste medio que escribe el trigger de
`recalculate_avg_cost` es un precio, no un desembolso— ni lo que una venta
realizó; para el dinero que entró y salió de verdad está el flujo de caja de
`transaction_cash_flow` (000027).

#### Eliminar una plataforma que todavía tiene posiciones

`DELETE /portfolios/sources/:id` **se niega** con un `409` cuando quedan
posiciones que la apuntan, y `details` dice cuántas. No es una política elegida
aquí, es la única respuesta honesta que permite el esquema:
`portfolio_entries.source_id` es `NOT NULL` y su clave ajena dice
`ON DELETE SET NULL` (migración 000003), dos reglas que se contradicen en cuanto
se borra una plataforma con algo encima. Postgres intentaba anular la columna, el
`NOT NULL` lo rechazaba, y al cliente le llegaba un `500` con el motivo sólo en
el log del servidor.

Arrastrar las posiciones con la plataforma tampoco es la salida: son el
historial de operaciones del dueño, y un clic en un botón de borrar no es
consentimiento para borrarlo. Así que la petición se rechaza, no se destruye
nada, y se dice qué hay que quitar antes.

El bloqueo cuenta **todas** las entradas, incluidas las vendidas del todo:
siguen apuntando a la fila y seguirían rompiendo el borrado. Por eso ese número
puede ser mayor que el `investments` que informa el listado, que sólo cuenta
posiciones abiertas — una plataforma que enseña «0 posiciones» puede aun así
rechazar el borrado, y el mensaje lo dice.

Las otras dos respuestas no cambian: `404` si la plataforma no es de quien la
pide, `200` cuando no queda nada que la apunte.

#### Eliminar una posición no es eliminar una transacción

`DELETE /portfolios/transactions/:txnId` quita una operación y deja que la
posición se recalcule con las que queden (queda en cantidad 0 si era la última).
`DELETE /portfolios/entries/:entryId` quita la posición entera: la fila es el
padre de su historial y `fk_transactions_entry` cascadea, así que se lleva cada
compra, venta y dividendo registrados contra ella.

Por eso la respuesta trae cuerpo, cosa rara en un `DELETE`:

```json
{ "success": true, "data": { "deletedTransactions": 11 } }
```

El número no es deducible por quien llama —pidió borrar una posición, no once
operaciones—, y es la diferencia entre confirmar lo que pasó y suponerlo. Una
posición que no exista o que sea de otro usuario devuelve el mismo `404`: la
pertenencia se impone en el `WHERE`, no con una lectura previa.

#### Errores de cliente: el motivo viaja en `details`

Un 4xx generado por `FromDomain` incluye el texto del error de dominio en
`details`, además del `message` genérico del handler y del `action`. Es lo que
convierte «no se pudo registrar el activo» en «USD no se convierte en sí misma a
1.0638», que es la diferencia entre corregir un campo y reintentar a ciegas
contra el mismo rechazo. Un 5xx **no** lo incluye: ese texto está escrito para
quien lee los logs y puede arrastrar una consulta o el nombre de una restricción.

#### Resumen de crecimiento: crecimiento no es rendimiento

`summary` lleva dos lecturas que no hay que confundir. `initialValue`,
`currentValue` y `totalGrowthPct` comparan el **valor** del primer snapshot con
el del último, así que abrir un portfolio o añadir una posición cuenta como
crecimiento. `gainLoss` y `gainLossPct` son el beneficio del último punto
—mercado menos capital invertido—, que es el rendimiento. Las dos discrepan a
propósito y una cartera en pérdidas puede tener `totalGrowthPct` positivo.

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

#### Moneda de una transacción: `currency` + `fxRate`

Una operación también arrastra dos monedas, y por una razón distinta a la de los
holdings: el activo cotiza en una y la cuenta liquida en otra. Comprar LVMH
(`MC.FR`, cotiza en EUR) desde una cuenta en dólares deja una confirmación con
un precio de 606,60 EUR, una tasa de 1,0638 y un cargo de 15,55 USD.

Cada transacción guarda las tres piezas:

| Campo | Valor |
|---|---|
| `price` | En `currency`: la moneda en la que se ejecutó la operación, tal como la imprime el bróker |
| `currency` | Moneda de la operación. Por omisión, la de coste de la posición |
| `fxRate` | Cuántos `costCurrency` valía 1 `currency` **el día de la operación**. Por omisión `1` |
| `costCurrency` | La de la posición (`portfolio_entries.cost_currency`), repetida aquí para que `fxRate` se pueda interpretar sin releer la entry |
| `fees` | La comisión, en `feesCurrency` |
| `feesCurrency` | `currency` o `costCurrency`, nunca una tercera. Por omisión `currency` |

El coste real es `price × fxRate`, y es lo que el trigger de coste medio guarda
en `portfolio_entries.price`. Guardar la tasa —en vez de convertir al leer— es lo
que hace que el resultado sea el del titular de la cuenta y no el del mercado
local: `exchange_rates` guarda una fila por par (§migración 000014), así que una
conversión hecha al leer aplica la tasa de hoy a las dos patas y las cancela, y
lo que queda en pantalla es la rentabilidad en EUR con el símbolo del dólar.

Dos combinaciones se rechazan con `400` en vez de guardarse:

- `fxRate` distinto de 1 cuando `currency` y `costCurrency` coinciden: una moneda
  no se convierte en sí misma, y el número escala el coste de la posición sin que
  nada aguas abajo pueda detectarlo.
- `fxRate` ausente cuando difieren. La app **no** rellena el hueco con la tasa
  actual: para una compra de hace un año esa es justo la tasa equivocada, y el
  error sería invisible. El cliente propone la tasa y el titular la corrige
  contra su confirmación.
- `feesCurrency` que no sea ni `currency` ni `costCurrency`. La fila lleva una
  sola tasa, así que una comisión en una tercera moneda solo se alcanzaría
  inventando una segunda.

Un cliente que no mande ninguno de estos campos sigue funcionando igual que
antes: la operación se registra en la moneda de la posición, con tasa 1 y la
comisión en esa misma moneda.

`feesCurrency` es un campo aparte porque la comisión no se cobra siempre del
mismo lado que la ejecución: la confirmación que motivó todo esto cotiza el
llenado en 606,60 EUR y la comisión en 0,00 USD. Multiplicar las dos por la
misma tasa es correcto para una sola de ellas, así que el coste de la posición
convierte el precio por `fxRate` y la comisión solo si se cobró en `currency`.

#### Importación de un extracto en dos monedas

El wizard de import (§2.6) toma las mismas tres piezas repartidas entre la fila
y el archivo, porque un extracto es **una** cuenta:

| Dónde | Campo | Valor |
|---|---|---|
| Columna mapeable | `fxRate` | La tasa de esa fila. Sin mapear, todas las filas valen 1 |
| Columna mapeable | `currency` | La moneda de la fila, como hasta ahora |
| `defaults.costCurrency` | — | La moneda de la cuenta, una para todo el archivo. Vacía significa «la misma de cada fila», que es lo que producía toda importación anterior |

Con `costCurrency` puesta, una fila en otra moneda **sin** tasa se marca
inválida en el preview con el motivo, en vez de importarse convertida a la tasa
de hoy. La comisión de una fila importada va en la moneda de esa fila: es lo que
significa la columna «Moneda» del archivo para los importes que etiqueta.

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

### 2.11 MCP — Model Context Protocol (JWT)

| Método | Path | Descripción |
|---|---|---|
| POST | `/mcp` | Endpoint JSON-RPC del servidor MCP |
| GET/DELETE | `/mcp` | 405 (el transporte va en modo *stateless*) |

**No sigue el sobre de §1.1**: es la única ruta de la API que no lo hace,
porque el cuerpo es JSON-RPC 2.0, que trae su propio sobre (`result` / `error`)
y lo define el protocolo. El resto de convenciones sí aplican: mismo bearer
token que todo lo demás, mismo limitador por usuario y un 401 —ese sí, en el
sobre normal— cuando el token falta o no vale.

La petición necesita `Content-Type: application/json` y un `Accept` que ofrezca
`application/json` **y** `text/event-stream`, aunque la respuesta siempre sea
JSON: lo exige la especificación del transporte *streamable HTTP*.

Las *tools* son todas de solo lectura y responden siempre con los datos de
**quien llama** — el id de usuario sale de los locals de autenticación y no hay
argumento con el que nombrar a otro:

| Tool | Devuelve |
|---|---|
| `list_portfolios` | Carteras con coste, valor de mercado y resultado |
| `get_holdings` | Posiciones consolidadas por activo, sumadas entre carteras |
| `get_allocation` | Reparto por categoría de activo, todo en una moneda |
| `list_recent_transactions` | Últimas transacciones (`limit`, máx. 200) |
| `get_portfolio_growth` | Serie de valor desde los snapshots (`period`: `1M`/`3M`/`6M`/`1Y`) |
| `list_platforms` | Plataformas con lo que se tiene en cada una |
| `search_assets` | Catálogo de activos (`query`, `limit`, máx. 100) |
| `list_exchange_rates` | Tasas compartidas más recientes |

Las que aceptan `currency` toman un ISO 4217 de la lista de §2.9; omitirlo
reporta en la moneda preferida de la cuenta. Un código no soportado vuelve como
*tool error* —no como error de protocolo— con la lista de los aceptados.

Configuración de un cliente MCP:

```json
{
  "mcpServers": {
    "finexia": {
      "type": "http",
      "url": "https://finexia.me/mcp",
      "headers": { "Authorization": "Bearer <access token>" }
    }
  }
}
```

---

## 3. Reglas de no-regresión

1. Ningún PR de migración añade, elimina ni renombra rutas de este documento.
2. El sobre de respuesta (§1.1) y el mapeo de errores (§1.2) se replican
   byte-a-byte en `platform/httpx`.
3. Cualquier discrepancia detectada entre este documento y el código se
   corrige **en el documento** (el código actual es la fuente de verdad) y se
   anota en el PR.
