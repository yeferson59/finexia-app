# Plan — Fondos de inversión (rentabilidad variable)

> **Estado:** Fase 1 implementada (000056 – 000057) · 23 sep 2026; fases 2 – 4 pendientes
> **Alcance:** módulos `market` y `portfolio` (backend), feature nueva `funds` (frontend)
> **Migraciones:** 000056 – 000057 (+ 000058 opcional) · **Fases:** 3 + 1 opcional
> **Relacionado con:** [`PLAN_RENTABILIDAD_EFECTIVO.md`](./PLAN_RENTABILIDAD_EFECTIVO.md)
> y [`PLAN_TASAS_MULTIPLES_EFECTIVO.md`](./PLAN_TASAS_MULTIPLES_EFECTIVO.md).
> Esos planes cubren el dinero que rinde a una **tasa conocida**; este cubre el
> que rinde lo que **resulte**.

## 1. Qué se quiere

Hay dinero que se parece al efectivo pero no rinde a una tasa: un fondo de
inversión colectiva (FIC), un fondo de pensiones voluntarias, un *money market*
de un neobanco. La entidad no promete una tasa. Publica cada día el **valor de
la unidad**, y la rentabilidad sale de cómo se mueve ese valor: «9,8 % E.A.
últimos 30 días», «−1,2 % en el mes». Puede bajar.

Finexia tiene que poder:

1. **Registrar el fondo** como lo muestra la entidad: en unidades con su valor
   de unidad, o, si la app solo muestra el saldo, en pesos.
2. **Actualizar su valor** cuando el usuario quiera (a diario, semanal, con el
   extracto mensual), también con fecha pasada.
3. **Mostrar su rentabilidad** como la muestra la entidad: por periodo (30, 90,
   180 días, año corrido, desde el inicio), en porcentaje y en E.A., y la
   ganancia en dinero.
4. Que todo eso **cuente en los totales** que ya existen: resumen, asignación,
   serie de crecimiento, reportes, MCP.

## 2. Qué hay hoy y por qué no alcanza

| Opción hoy | Qué pasa con un fondo |
|---|---|
| Registrarlo como **efectivo con tasa** | La tasa se fija por adelantado y Finexia causa cada día. Un fondo no tiene tasa: habría que cambiarla cada semana, y un mes negativo no se puede decir |
| Registrarlo como **efectivo sin tasa** y anotar «Intereses» | Un rendimiento negativo no existe como movimiento; `cash_interest` suma unidades a costo cero (000042) y no resta |
| Registrarlo como **ETF** (el alias `fondo` ya lleva ahí, `market/asset.go:174`) | El job de precios le pide cotización al proveedor con el ticker. Un FIC no cotiza en ningún proveedor, así que no hay precio y la posición **vale lo que costó**: ganancia 0, `positionsAtCost` |
| Poner el precio a mano | Solo un admin puede (`PATCH /portfolios/assets/:id/price`), y va a `assets.current_price`, que es compartido |

Además, la app **no guarda historia de precios**: `user_asset_prices` tiene una
fila por usuario y activo, y la serie de crecimiento se arma con
`portfolio_snapshots` ya cerrados. Sin historia no hay «rentabilidad 30 días».

## 3. Decisiones

**D1. Un fondo es una posición de unidades × valor de unidad, no efectivo.**
Así se contabiliza el fondo mismo, así lo reporta la Superfinanciera y así lo
entiende todo lo que ya existe: `quantity` son unidades, `price` es el costo
promedio por unidad, y el valor es unidades × precio actual. Una compra es un
aporte y una venta es un retiro, con su ganancia realizada por costo promedio.
Un mes negativo es solo un valor de unidad más bajo.
*Descartado:* efectivo con «rendimientos» positivos y negativos anotados como
movimientos. Obliga a un tipo nuevo que reste sin ser retiro, a enseñarle a la
serie de crecimiento que no es flujo, y no sirve para quien sí ve unidades.

**D2. Un tipo de activo nuevo: `fund`.** No es `etf` (el job de precios lo
intentaría sincronizar), ni `cash` (se causaría como cuenta), ni `other` (se
pierde en la asignación). Con tipo propio:
- el job de precios lo salta: `syncOneAsset` solo cotiza acciones, ETF, bonos y
  cripto, y `RefreshAssetPrice` ya responde «no se puede cotizar» a lo demás;
- la asignación y los reportes lo muestran como «Fondos»;
- las pantallas saben cuándo ofrecer «Actualizar valor».

**Alias de importación.** `fondo`, `fondos`, `fund`, `funds` y `mutual fund`
**pasan de `etf` a `fund`**, y se agregan `fic`, `fondo de inversion`,
`fondo de inversion colectiva`, `fondo de pensiones voluntarias`, `fpv` y
`money market`. Un fondo que cotiza en bolsa se sigue diciendo `etf`,
`fondo indexado` o `index fund`, que no cambian.
Por qué: en un archivo de un usuario colombiano, «fondo» es un FIC casi siempre,
y como ETF quedaba valorado a costo para siempre (el proveedor no lo conoce).
El cambio solo afecta a **activos nuevos**: el importador reutiliza la fila del
catálogo cuando el ticker ya existe, así que nada de lo importado cambia de
tipo.

**D3. Dos formas de registrarlo, un solo modelo.** Cada fondo de un usuario
tiene un **modo de seguimiento**, que se elige al crearlo y no cambia:

| Modo | Cuándo | Qué escribe el usuario | Unidades |
|---|---|---|---|
| `units` | El extracto muestra unidades y valor de unidad (FIC de fiduciarias y comisionistas) | Aportes y retiros con unidades y valor de unidad; valores de unidad por fecha | Reales |
| `balance` | La app solo muestra el saldo (bolsillos de inversión de neobancos, FPV) | Aportes y retiros en dinero; saldos por fecha | **Sintéticas** |

**D4. El modo `balance` usa unidades sintéticas (unitización).** Es la técnica
con la que los fondos mismos separan aportes de rendimiento:
- El primer aporte compra a valor de unidad **1**.
- Un aporte o retiro del día *d* opera al valor de unidad de la **última marca
  anterior a *d***: `unidades = monto / VU`.
- Un saldo *S* del día *d* (fin del día, con los movimientos de ese día) fija
  `VU(d) = S / unidades(d)`.

El usuario nunca ve las unidades: ve saldo, aportes, retiros y rendimiento.
Pero por dentro el fondo es igual al del modo `units`, así que la rentabilidad
por periodo, la serie y los totales salen con el mismo código. Y la
rentabilidad resultante es **ponderada por tiempo**: un aporte grande no se
confunde con rendimiento.

**D5. En modo `balance` el hecho es el saldo; las unidades se derivan.** Se
guarda el saldo que escribió el usuario y se recalculan las unidades de cada
movimiento y el valor de unidad de cada marca con una **pasada hacia adelante**
en orden de fecha, cada vez que cambia algo del fondo (un movimiento o una
marca, nuevo, editado o borrado). Así un movimiento corregido no deja unidades
viejas. La pasada corre en Go, en una transacción, con el fondo bloqueado
(`SELECT … FOR UPDATE` sobre `user_funds`).

**D6. El valor de unidad es del usuario, no compartido.** Cada usuario escribe
el suyo, y en modo `balance` es sintético: no tiene sentido compartirlo. La
historia va a una tabla nueva (`fund_marks`) y **la marca más reciente se copia
a `user_asset_prices`** con `source = 'user'`. Con eso no cambian ni la vista
`portfolio_summary`, ni las posiciones, ni el trigger de
`recorded_market_price` (000036): todos ya prefieren `uap.price`. En
procedencia de precio, un fondo con marca cuenta como `own`; sin marcas, como
`cost`, y así debe verse («Sin valor actualizado»).

**D7. La marca es por fondo, no por portafolio.** La clave es
`(user_id, asset_id)`, la misma de `user_asset_prices`. Si el fondo está
repartido en dos portafolios, en modo `balance` el saldo que se escribe es el
**total del extracto**, y `VU = S / unidades de todos los portafolios`. Es la
misma regla que ya usan los tramos del efectivo (D5 del plan de tasas
múltiples).

**D8. Finexia no estima entre marcas.** El valor queda en la última marca hasta
la siguiente. No se proyecta con la rentabilidad reciente: sería presentar una
estimación como si fuera el valor. La interfaz dice «Valor al 30 sep» y avisa
cuando la última marca tiene más de 35 días.

**D9. Una marca con fecha pasada corrige la serie de crecimiento.** Si el
usuario registra el 5 oct el saldo del extracto del 30 sep, la ganancia tiene
que aparecer el 30 sep y no el 5 oct; si no, la gráfica muestra un salto el
día del registro y un mes plano antes. Para poder corregir sin adivinar:

- **El snapshot guarda con qué valoró cada fondo.** El job, en la misma lectura
  con la que arma el snapshot, escribe una fila por posición `fund` en
  `fund_snapshot_values`: unidades, valor de unidad usado, costo por unidad y
  tasa de cambio. Es la misma consulta, así que la fila y el total salen de la
  misma foto de la base.
- **Al escribir, editar o borrar una marca del día *d*,** en la misma
  transacción, cada snapshot de ese fondo con fecha ≥ *d* se revalora:
  `VU nuevo = última marca ≤ fecha del snapshot`, o el costo por unidad
  guardado si no hay ninguna. El snapshot suma
  `unidades × (VU nuevo − VU guardado) × tasa` a `total_value`,
  `total_gain_loss` y la clave `fund` de `allocation`, recalcula
  `total_gain_loss_pct`, y la fila del fondo guarda el VU nuevo.
- **Los flujos no se tocan:** una marca no es un movimiento de dinero.
- **Solo se corrige lo que tiene fila.** Los snapshots anteriores a 000057 no
  tienen fondos (el tipo no existía), y una posición que entra con fecha pasada
  no está en los snapshots de antes de su alta: ahí no hay nada que revalorar.

Así la serie queda como si el usuario hubiera registrado la marca el día del
extracto, y la **rentabilidad del fondo** (D10) coincide con lo que muestra la
gráfica.

**D10. La rentabilidad se calcula como la publica la entidad.** Con `VU` al
final (`t`, la última marca) y al inicio (`s`, la última marca en o antes de
`t − n` días):

```
periodo = VU_t / VU_s − 1
E.A.    = (VU_t / VU_s)^(365 / días) − 1      (días = t − s, reales)
```

- Periodos: 30, 90, 180 y 365 días, año corrido y desde el inicio.
- Si no hay marca tan atrás, el periodo va `null`: no se extrapola.
- La E.A. solo se muestra con 28 días o más; con menos, anualizar exagera.
- En modo `units`, con los valores de unidad reales, la cifra es comparable con
  la que publica el fondo. En modo `balance` es la del dinero del usuario, ya
  neta de comisión y retención (el saldo ya viene neto).

La **ganancia en dinero** es la de siempre: no realizada
(`valor − costo`) + realizada en los retiros (costo promedio).

**D11. Comisiones y retención ya están dentro del valor de unidad.** No se
anotan aparte. Lo que sí se anota es lo que cobra un retiro (penalidad por
pacto de permanencia, GMF): va en `fees` del retiro y cuenta como pérdida.

**D12. Los movimientos de un fondo `balance` solo se escriben por sus rutas.**
`CreateTransaction`, `UpdateTransaction` y `DeleteTransaction` rechazan una
posición en modo `balance` (`ErrFundBalanceManaged`, 409), porque una compra
escrita a mano con unidades rompería D5. Es el mismo patrón que
`ErrCashPocketFixed`. Un fondo `units` sí usa las rutas de siempre (compra y
venta con unidades y precio) y agrega solo las marcas.

**D13. El alta escribe el precio antes que los movimientos.** Un fondo con
historia llega con ganancia. Si los aportes se insertan antes que la marca, el
trigger de 000036 los registra a costo y toda esa ganancia se vuelve
rendimiento del día del alta (el problema que 000036 resolvió). La creación
escribe `user_asset_prices` primero; el trigger toma el valor actual.

**D14. Pagar con el efectivo de la plataforma funciona igual que hoy.** Un
aporte puede salir del saldo de efectivo y un retiro volver a él, con la misma
opción `settlement` de las compras y ventas (`postgres_cash_link.go`). Así
«saqué del bolsillo y metí al fondo» no cambia la rentabilidad.

## 4. El cálculo en modo `balance`

Fondo creado el 1 jul 2026, registrado solo con saldos de la app:

| Fecha | Evento | VU usado | Unidades | Total unidades | VU de la marca |
|---|---|---:|---:|---:|---:|
| 1 jul | Aporte 10.000.000 | 1 (primer aporte) | 10.000.000,00000000 | 10.000.000,00000000 | — |
| 31 jul | Saldo 10.080.000 | — | — | 10.000.000,00000000 | 1,00800000 |
| 15 ago | Aporte 5.000.000 | 1,00800000 | 4.960.317,46031746 | 14.960.317,46031746 | — |
| 31 ago | Saldo 15.110.000 | — | — | 14.960.317,46031746 | 1,01000531 |
| 20 sep | Retiro 2.000.000 | 1,01000531 | −1.980.187,60911267 | 12.980.129,85120479 | — |
| 30 sep | Saldo 13.050.000 | — | — | 12.980.129,85120479 | 1,00538285 |

Rentabilidad por mes (ponderada por tiempo):

| Mes | Del periodo | E.A. |
|---|---:|---:|
| Julio (30 días) | +0,8000 % | +10,18 % |
| Agosto (31 días) | +0,1989 % | +2,37 % |
| Septiembre (30 días) | **−0,4577 %** | **−5,43 %** |
| Desde el inicio | +0,5383 % | — |

En dinero: aportó 15.000.000, retiró 2.000.000 y tiene 13.050.000, así que
ganó **50.000**: 14.559,90 realizados en el retiro (costo promedio
1,00265252 por unidad) y 35.440,10 sin realizar.

Las dos cifras dicen cosas distintas y las dos se muestran: el 0,54 % es cómo
le fue al fondo; los 50.000 son cuánto ganó el usuario con sus aportes.

### Por qué el aporte opera a la marca anterior

El 15 ago el fondo ya había ganado algo desde el 31 jul, pero nadie lo dijo.
Operar el aporte a 1,008 atribuye esa ganancia a todas las unidades, también a
las nuevas. El **valor total** queda exacto en la siguiente marca
(`S / unidades`); lo único aproximado es el reparto del rendimiento entre antes
y después del aporte, y en un fondo de renta fija a la vista eso es del orden
de centavos. Quien quiera exactitud escribe el saldo del día anterior al
aporte (el formulario lo ofrece: «¿Cuánto tenías justo antes?»).

## 5. El modo `units`

Es un fondo con unidades reales. El flujo es el de cualquier activo:

- **Aporte:** compra de `unidades` a `valor de unidad` (los dos del extracto).
- **Retiro:** venta de unidades al valor de unidad del día.
- **Marca:** valor de unidad por fecha, sin tocar las unidades.

Ejemplo: 1.000 unidades compradas a 12.345,678901 el 1 sep; el 30 sep el valor
de unidad es 12.431,220000. Valor: 12.431.220; ganancia 85.541,10; periodo
+0,6929 %, **E.A. +9,08 %** (29 días).

## 6. Modelo de datos

### 000056 — `fund_asset_type` (Fase 1)

```sql
ALTER TYPE asset_type ADD VALUE IF NOT EXISTS 'fund';
```

Sola, porque un valor que añade `ADD VALUE` no se usa en la misma transacción
(como 000048). El down no lo quita: sacar un valor de un enum obliga a
reconstruir el tipo, y un valor sin usar no estorba.

### 000057 — `funds` (Fase 1; `balance` se abre en la Fase 2)

```sql
CREATE TYPE fund_tracking AS ENUM ('units', 'balance');

-- Un fondo de un usuario: cómo lo sigue. La fila existe para todo activo
-- `fund` que el usuario tenga; sin ella un fondo importado se trata como
-- `units`.
CREATE TABLE IF NOT EXISTS user_funds (
  user_id    UUID NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
  asset_id   UUID NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
  tracking   fund_tracking NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (user_id, asset_id)
);

-- Lo que valía el fondo un día. En `units` el hecho es unit_value; en
-- `balance` el hecho es balance y unit_value se deriva (D5).
CREATE TABLE IF NOT EXISTS fund_marks (
  user_id    UUID NOT NULL,
  asset_id   UUID NOT NULL,
  mark_date  DATE NOT NULL,
  unit_value NUMERIC(20, 8) NOT NULL CHECK (unit_value > 0),
  balance    NUMERIC(20, 8)          CHECK (balance IS NULL OR balance > 0),
  notes      VARCHAR(500),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (user_id, asset_id, mark_date),
  FOREIGN KEY (user_id, asset_id) REFERENCES user_funds(user_id, asset_id) ON DELETE CASCADE
);

-- Con qué valoró el snapshot cada posición `fund` (D9). Sin esto no se puede
-- saber cuánto corregir cuando llega una marca con fecha pasada.
CREATE TABLE IF NOT EXISTS fund_snapshot_values (
  entry_id      UUID NOT NULL REFERENCES portfolio_entries(id) ON DELETE CASCADE,
  snapshot_date DATE NOT NULL,
  portfolio_id  UUID NOT NULL REFERENCES portfolios(id) ON DELETE CASCADE,
  asset_id      UUID NOT NULL REFERENCES assets(id)     ON DELETE CASCADE,
  units         NUMERIC(20, 8)  NOT NULL,
  unit_value    NUMERIC(20, 8)  NOT NULL,  -- el que se usó (o el corregido)
  unit_cost     NUMERIC(20, 8)  NOT NULL,  -- pe.price: el valor sin marcas
  fx_rate       NUMERIC(24, 10) NOT NULL,  -- moneda del activo → base
  PRIMARY KEY (entry_id, snapshot_date)
);

CREATE INDEX IF NOT EXISTS idx_fund_snapshot_values_asset
  ON fund_snapshot_values(asset_id, snapshot_date);
```

- `balance > 0`: un saldo en cero no define valor de unidad. Un retiro total se
  escribe como retiro (`all: true`), no como marca.
- La PK sirve la lectura que más se hace («la marca más reciente en o antes
  de…») con un escaneo hacia atrás del índice.
- Down: borra las dos tablas y el tipo, y las filas `source = 'user'` de
  `user_asset_prices` de activos `fund` (sin ellas vuelven a valer su costo).

### 000058 — `fund_public_values` (Fase 4, opcional)

```sql
-- Valores de unidad publicados (datos abiertos de la Superfinanciera). No son
-- datos de un proveedor con licencia, así que son compartidos, como
-- assets.current_price (000018).
CREATE TABLE IF NOT EXISTS fund_public_values (
  asset_id   UUID NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
  value_date DATE NOT NULL,
  unit_value NUMERIC(20, 8) NOT NULL CHECK (unit_value > 0),
  PRIMARY KEY (asset_id, value_date)
);
```

## 7. Backend, archivo por archivo

| Archivo | Cambio | Fase |
|---|---|---|
| `migrations/000056_*.sql`, `000057_*.sql` | Tipo `fund`, `user_funds`, `fund_marks` | 1 |
| `market/asset.go` | `Fund AssetType = "fund"` en la lista y en `IsValid`; alias nuevos (D2) | 1 |
| `market/service_sync.go` | `Fund` no llega aquí; si llega, `errAssetTypeUnsupported` como hoy | 1 |
| `portfolio/fund.go` (nuevo) | `Fund`, `FundTracking`, `FundMark`; validaciones; errores (`ErrFundNotFound` 404, `ErrFundBalanceManaged` 409, `ErrFundNoUnits` 409, `ErrInvalidFundMark` 400) | 1 |
| `portfolio/postgres_fund.go` (nuevo) | CRUD de marcas; `syncFundPrice` (la marca más reciente a `user_asset_prices`, o borrarla si no quedan) en la misma transacción | 1 |
| `portfolio/service_fund.go` (nuevo) | Crear fondo `units` (activo contribuido, `user_funds`, posición, compra, marca — D13), marcas | 1 |
| `portfolio/handler_fund.go`, `dto_fund.go` (nuevos) | Rutas de §8 | 1–3 |
| `portfolio/postgres_snapshot.go` | `GetAllPortfolioSummaryRows` trae las posiciones `fund` en la misma consulta; `UpsertPortfolioSnapshot` escribe `fund_snapshot_values` en la misma transacción | 1 |
| `portfolio/postgres_fund.go` | `restateFundSnapshots(asset, desde)`: revalora los snapshots de D9 en la transacción de la marca | 1 |
| `mcp/dto.go` | `fund` en los enums de las descripciones | 1 |
| `portfolio/fund_units.go` (nuevo) | `replay(events) → []unitsAt, []markValue`: la pasada de D4–D5, pura y sin base de datos | 2 |
| `portfolio/postgres_fund.go` | `RecomputeBalanceFund`: bloquear, leer movimientos y marcas, `replay`, escribir `quantity`/`price` de cada transacción y `unit_value` de cada marca, `syncFundPrice` | 2 |
| `portfolio/service_fund.go` | Aporte, retiro (`all`), saldo; alta `balance` con saldo inicial | 2 |
| `portfolio/postgres_transaction.go` | Guarda de D12 en crear, editar y borrar (junto a `requireTypeAllowed`) | 2 |
| `portfolio/fund_performance.go` (nuevo) | Periodos, E.A., ganancia realizada y no realizada (D10) | 3 |
| `portfolio/postgres_entry.go`, `mcp/tools_portfolio.go` | `valuedOn` (fecha de la última marca) en las filas de fondo; herramienta `get_funds` (rentabilidad por periodo) | 3 |
| `notification/weekly_summary.go` | Fondos con marca en la semana: variación del periodo | 3 |
| `market/fund_public_job.go` (nuevo) | Job diario que lee el dataset de rentabilidades de FIC y llena `fund_public_values` y `assets.current_price` de los FIC del catálogo | 4 |
| `docs/API.md`, `docs/MANUAL_DE_USUARIO.md` | Documentación | 1–3 |

## 8. API

Todas bajo `/portfolios`, antes de `/:id`, como las de `/cash`.

| Método y ruta | Qué hace | Fase |
|---|---|---|
| `GET /portfolios/funds` | Fondos del usuario: modo, unidades, valor, costo, `valuedOn` | 1 |
| `POST /portfolios/funds` | Crea un fondo (activo + posición + apertura) | 1 (`units`), 2 (`balance`) |
| `GET /portfolios/funds/:assetId/marks` | Historia de marcas | 1 |
| `POST /portfolios/funds/:assetId/marks` | Marca `{date, unitValue}` (`units`) o `{date, balance}` (`balance`). Misma fecha = reemplaza | 1–2 |
| `DELETE /portfolios/funds/:assetId/marks/:date` | Borra una marca | 1 |
| `POST /portfolios/funds/:assetId/marks/bulk` | Varias marcas de una vez (pegar la tabla del extracto) | 3 |
| `POST /portfolios/funds/:assetId/contributions` | Aporte en modo `balance` `{portfolioId, date, amount, balanceBefore?, settlement?}` | 2 |
| `POST /portfolios/funds/:assetId/withdrawals` | Retiro en modo `balance` `{portfolioId, date, amount, fees?, all?, settlement?}` | 2 |
| `PUT/DELETE /portfolios/funds/movements/:txnId` | Editar o borrar un aporte/retiro `balance` (recalcula) | 2 |
| `GET /portfolios/funds/:assetId/performance` | Periodos, E.A., ganancia, serie de valor de unidad | 3 |

```json
POST /portfolios/funds
{
  "portfolioId": "…",
  "sourceId": "3f7c…",
  "name": "Fiducuenta",
  "currency": "COP",
  "tracking": "balance",
  "opening": {
    "date": "2026-01-10",
    "amount": "12000000",
    "currentBalance": "12640000",
    "currentBalanceDate": "2026-09-22"
  }
}
```

- `opening` en `units`: `{date, units, unitValue}` y opcionalmente
  `{currentUnitValue, currentUnitValueDate}`.
- `opening` en `balance`: «aporté en total `amount` desde `date` y hoy vale
  `currentBalance`». Es el alta rápida: un solo aporte y una marca. Quien quiera
  la historia completa agrega después aportes, retiros y marcas con fecha.
- El activo es **contribuido** (000021): privado, `asset_type = 'fund'`,
  `exchange` = nombre de la plataforma, ticker generado (`FND-` + 8 hex) para
  no chocar con `idx_assets_ticker_exchange`.

```json
GET /portfolios/funds/:assetId/performance
{
  "tracking": "balance",
  "currency": "COP",
  "valuedOn": "2026-09-30",
  "unitValue": "1.00538285",
  "units": "12980129.85120479",
  "value": "13050000.00",
  "invested": "15000000.00",
  "withdrawn": "2000000.00",
  "realizedGain": "14559.90",
  "unrealizedGain": "35440.10",
  "periods": [
    { "key": "30d", "from": "2026-08-31", "to": "2026-09-30", "days": 30, "pct": "-0.4577", "eaPct": "-5.43" },
    { "key": "90d", "from": null, "to": "2026-09-30", "days": null, "pct": null, "eaPct": null },
    { "key": "ytd", "from": null, "to": "2026-09-30", "days": null, "pct": null, "eaPct": null },
    { "key": "inception", "from": "2026-07-01", "to": "2026-09-30", "days": 91, "pct": "0.5383", "eaPct": "2.18" }
  ],
  "series": [{ "date": "2026-07-01", "unitValue": "1" }, { "date": "2026-07-31", "unitValue": "1.008" }]
}
```

En modo `balance` el «desde el inicio» parte de la marca implícita del primer
aporte (VU = 1). El año corrido necesita una marca al 31 dic o antes; un fondo
abierto en julio no la tiene, y por eso va `null` y no se confunde con el
«desde el inicio».

## 9. Frontend

| Archivo | Cambio | Fase |
|---|---|---|
| `lib/shared/format/asset-type.ts` | `fund: 'Fondos'` y su color | 1 |
| `lib/features/portfolio/components/asset-create-inline.svelte`, `lib/features/transactions/transactions.ts`, `lib/features/admin/admin.ts` | `fund` en las listas de tipo | 1 |
| `lib/api/schemas/funds.ts`, `lib/api/funds.ts` (nuevos) | Esquemas y cliente | 1–3 |
| `lib/features/funds/` (nueva) | `funds.ts`, `schemas.ts`, `performance.ts` (formato de periodos) y sus `.spec.ts` | 1–3 |
| `routes/dashboard/funds/` (nueva) | Tarjeta por fondo: valor, «Valor al 30 sep», ganancia, rentabilidad 30 días E.A.; aviso si la marca tiene más de 35 días | 1 |
| `components/fund-create-form.svelte` | Paso 1: «¿Tu extracto muestra unidades?» (Sí → `units`, No → `balance`). Paso 2: la apertura de §8 | 1–2 |
| `components/fund-mark-form.svelte` | «Actualizar valor»: fecha + valor de unidad o saldo. Muestra la variación contra la marca anterior antes de guardar | 1–2 |
| `components/fund-movement-form.svelte` | Aporte y retiro en modo `balance`, con «¿Cuánto tenías justo antes?» plegado y «Retirar todo» | 2 |
| `components/fund-performance.svelte` | Tabla de periodos y gráfica del valor de unidad (en `balance`: «Índice de rentabilidad», base 100) | 3 |
| `components/fund-marks-paste.svelte` | Pegar columnas fecha / valor del extracto | 3 |
| Menú del dashboard | «Fondos» junto a «Efectivo»; desde Efectivo, un enlace «¿Tu dinero rinde variable? Regístralo como fondo» | 1 |

En **Inversiones**, un fondo aparece como cualquier posición; su fila enlaza a
su página en Fondos en lugar de ofrecer «Actualizar precio».

## 10. Fases

1. **Fondos por unidades.** 000056–000057, tipo `fund`, marcas, precio del
   usuario, exclusión del job de precios, pantalla de Fondos básica.
   *Listo cuando:* creo un FIC con 1.000 unidades a 12.345,678901, registro
   12.431,22 el 30 sep y la posición vale 12.431.220 con ganancia 85.541,10
   en el resumen, la asignación y MCP; y el job de precios no lo toca.
2. **Fondos por saldo.** Unidades sintéticas, pasada hacia adelante, aportes y
   retiros por sus rutas, guarda en las rutas genéricas, alta rápida.
   *Listo cuando:* el ejemplo de §4 da exactamente esas unidades, esos valores
   de unidad y esa ganancia; y editar el aporte del 15 ago recalcula todo lo
   que viene después.
3. **Rentabilidad.** Endpoint de rendimiento, periodos y E.A., gráfica, pegar
   marcas, MCP y correo semanal.
   *Listo cuando:* septiembre del ejemplo de §4 muestra −0,4577 % y −5,43 %
   E.A., y 90 días va vacío porque no hay historia.
4. **(Opcional) Valor de unidad automático.** Catálogo curado de FIC con su
   valor de unidad diario desde datos abiertos (§12).

Van en ese orden porque la Fase 2 es la Fase 1 con unidades derivadas, y la
Fase 3 solo lee lo que las dos primeras guardan.

**Lo que cambió respecto al plan al implementar la Fase 1:**

- **El job de precios no se tocó.** `syncOneAsset` ya salta en silencio todo lo
  que no es acción, ETF, bono o cripto, y `RefreshAssetPrice` ya responde
  `ErrAssetNotQuotable`. Excluir `fund` de `GetHeldAssetIDs` no ahorraba nada.
- **Un fondo que llega sin `CreateFund` se adopta.** Un archivo con categoría
  «fondo» o un activo creado desde el formulario de posiciones crea un `fund`
  sin fila en `user_funds`. La lista lo incluye como fondo por unidades, y la
  primera marca le crea la fila (`adoptHeldFund`).
- **`DELETE /portfolios/funds/:assetId`**, que no estaba: deja de seguir un fondo
  que ningún portafolio guarda, y borra el activo si nadie más lo usa. Con
  posiciones responde 409; se borran donde toda posición.
- **`createPortfolioEntryTx`**: el cuerpo de `CreatePortfolioEntry`, extraído
  para abrir la posición dentro de la transacción que crea el activo y la marca.
- **Reparto por industria:** un `fund` sin clasificar cuenta como
  `unclassified` (como un ETF), no como `not_applicable`: un fondo tiene
  industrias, aunque nadie se las haya puesto.
- **Frontend:** página propia `/dashboard/funds` («Fondos» en el menú, junto a
  Efectivo), no una pestaña del efectivo. El panel de precio de un activo manda
  a Fondos en vez de ofrecer un proveedor.
- **Pendiente de la Fase 1:** `docs/MANUAL_DE_USUARIO.md` (obliga a regenerar el
  PDF) y ver la pantalla en la app.

## 11. Pruebas

- **Unitarias:**
  - `replay` con el ejemplo de §4, cifra por cifra.
  - `replay`: primer aporte a VU 1; aporte y marca el mismo día (el aporte va
    primero); marca con unidades en cero (`ErrFundNoUnits`); retiro mayor que
    lo que hay; `all: true`; `balanceBefore` crea la marca de la víspera.
  - Periodos: marca exacta, marca anterior más cercana, sin historia (`null`),
    menos de 28 días (sin E.A.), rendimiento negativo.
  - Validaciones de marcas y de la apertura, un caso por regla.
- **DB:**
  - La marca más reciente llega a `user_asset_prices` y `portfolio_summary`
    cuenta la posición como `own`; borrar todas la devuelve a `cost`.
  - El snapshot guarda la fila de cada fondo; una marca con fecha pasada
    revalora los snapshots desde esa fecha (total, ganancia, `allocation`) y no
    los anteriores; borrarla los devuelve al valor anterior o al costo.
  - Alta con historia: los aportes quedan con `recorded_market_price` al valor
    actual (D13) y la serie no muestra la ganancia vieja como rendimiento.
  - Fondo `balance` en dos portafolios: VU con las unidades de los dos.
  - Editar o borrar un movimiento o una marca recalcula unidades, costo
    promedio y VU.
  - Las rutas genéricas rechazan una posición `balance` (409).
  - Aporte con `settlement` desde el efectivo: la rentabilidad no cambia.
- **Handlers:** 400 por validación, 404 con fondos ajenos, 409 con las guardas.
- **Frontend:** esquemas, formato de periodos (E.A. oculta con menos de 28
  días, `null` como «—»), formulario de alta por modo.

## 12. Fase 4: valor de unidad automático (opcional)

La Superintendencia Financiera publica en datos abiertos la rentabilidad y el
valor de unidad diario de todos los FIC:
[Rentabilidades de los Fondos de Inversión Colectiva (FIC)](https://www.datos.gov.co/Hacienda-y-Cr-dito-P-blico/Rentabilidades-de-los-Fondos-de-Inversi-n-Colectiv/qhpu-8ixx)
(dataset `qhpu-8ixx`, API Socrata). Con eso:

- Un **catálogo curado** de FIC: un activo `fund` por fondo y tipo de
  participación (cada uno tiene su propio valor de unidad), `exchange = 'SFC'`,
  ticker con el código del fondo.
- Un job diario, con el patrón de `public_rates_job.go`, que escribe
  `fund_public_values` y `assets.current_price`.
- Un fondo `units` que se enlaza a un FIC del catálogo **no necesita marcas**:
  la valoración ya prefiere `uap.price` y cae a `current_price`. Si el usuario
  escribe una marca, la suya manda.
- La rentabilidad por periodo lee `fund_public_values` cuando el fondo no tiene
  marcas propias.
- El modo `balance` no se beneficia: sus unidades son sintéticas.

**Por verificar antes de comprometerla:** columnas exactas, con qué retraso se
publica, límites de la API, y que el tipo de participación sea identificable
(el extracto del usuario tiene que decirle cuál es el suyo).

## 13. Casos borde y riesgos

| Caso | Qué pasa | Mitigación |
|---|---|---|
| Marca vieja (el usuario no actualiza) | El valor se queda quieto | D8: «Valor al …» y aviso a los 35 días; el correo semanal lo recuerda |
| Marca con fecha pasada | Los snapshots desde esa fecha quedaban con el valor viejo | D9: se revaloran en la misma transacción |
| El job lee el total y otra transacción escribe una marca antes de que guarde | El snapshot de ese día queda con el VU anterior | Total y fila del fondo salen de la misma consulta, así que son coherentes entre sí; la próxima marca de esa fecha o anterior lo corrige |
| Aporte entre dos marcas (`balance`) | El reparto del rendimiento alrededor del aporte es aproximado | §4: el total es exacto; `balanceBefore` lo vuelve exacto |
| Retiro total con un monto distinto a unidades × VU | Lo recibido es el hecho | `all: true` vende todas las unidades a `monto / unidades` |
| Fondo `balance` editado desde la pantalla de posiciones | Rompería las unidades derivadas | D12: 409 en las rutas genéricas |
| Historia larga (años de marcas) | La pasada recorre todo | Es lineal; un fondo tiene cientos de filas, no miles. Si crece, recalcular desde el primer evento cambiado |
| `recorded_market_price` después de recalcular | Es de inserción: guarda el VU del alta aunque las unidades cambien | Error acotado (el valor por unidad del alta); se documenta en la migración como en 000036 |
| El proveedor tiene un ticker igual | El job pisaría el precio | Tipo `fund` excluido del job, y ticker generado `FND-…` |
| Un fondo indexado importado como «fondo» | Queda `fund` y no sincroniza precio | D2: `etf`, `fondo indexado` e `index fund` siguen en ETF; lo ya importado no cambia |
| Fondo en moneda extranjera | Igual que cualquier posición | `currency` del activo y el `fx_rate` de siempre |
| Rentabilidad con retención | El VU ya es neto | D11 |

**Fuera de alcance:** estimar el valor entre marcas; revalorar snapshots
anteriores a 000057; cambiar el modo de un fondo después de crearlo
(se crea otro); importar el extracto en PDF.

## 14. Criterios de aceptación

Fase 1:

- [x] Creo «FIC Renta Fija» en modo unidades con 1.000 unidades a 12.345,678901; registro 12.431,22 al 30 sep y la posición vale 12.431.220 con ganancia 85.541,10. *(TestFundOpensPricedByItsMark)*
- [x] La posición cuenta como precio propio en el resumen; sin marcas, como costo. *(TestFundOpensPricedByItsMark, TestFundPriceFollowsTheLatestMark)*
- [x] El job de precios y «Actualizar precio» no tocan el fondo. *(no hubo que cambiarlos: ver §10)*
- [ ] La asignación muestra «Fondos» aparte de efectivo y ETF. *(etiqueta y color puestos; falta verlo en la app)*
- [x] Registro hoy una marca con fecha de hace 10 días y la gráfica de crecimiento sube ese día, no hoy. *(TestFundLateMarkRestatesSnapshots)*
- [x] Un archivo con categoría «fondo» crea un `fund`; uno con «etf» o «fondo indexado» sigue creando ETF. *(TestNormalizeAssetTypeFunds)*

Fase 2:

- [ ] Registro el ejemplo de §4 solo con montos y saldos, y el fondo vale 13.050.000 con 50.000 de ganancia (14.559,90 realizada).
- [ ] Corrijo el aporte del 15 ago y todo lo que viene después se recalcula.
- [ ] No puedo editar ese fondo desde la pantalla de transacciones.
- [ ] Un fondo que ya tenía ganancia al darlo de alta no infla la serie de crecimiento.

Fase 3:

- [ ] Septiembre del ejemplo muestra −0,4577 % (−5,43 % E.A.) y 90 días aparece como «—».
- [ ] Pego 30 valores de unidad del extracto y la gráfica los muestra.
- [ ] MCP responde la rentabilidad a 30 días de un fondo.
