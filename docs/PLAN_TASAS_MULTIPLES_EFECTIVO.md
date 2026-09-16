# Plan — Varias tasas en una plataforma

> **Estado:** Implementado (000046 – 000048) · 16 sep 2026
> **Alcance:** módulo `portfolio` (backend), feature `cash` (frontend)
> **Migraciones:** 000046 – 000048 · **Fases:** 3
> **Parte de:** [`PLAN_RENTABILIDAD_EFECTIVO.md`](./PLAN_RENTABILIDAD_EFECTIVO.md).
> Sustituye su **D1** («dos productos, dos plataformas») y amplía su **D6**
> (lo ganado antes de registrar la tasa).

## 1. Qué se quiere

Una misma entidad paga distinto según dónde y cuánto dinero tengas. Finexia
tiene que poder registrar los tres casos, que se pueden dar a la vez:

1. **Tramos de saldo.** «12 % E.A. hasta $5.000.000 y 8 % sobre lo que pase».
2. **Bolsillos con tasa propia.** La cuenta paga 8 % y la «cajita» paga 10 %,
   y las dos son de la misma plataforma.
3. **Tasa fija por depósito.** Un depósito conserva la tasa del día en que se
   hizo, a un plazo o sin él, aunque la entidad cambie la tasa después.

Y responder bien a esta situación: **el depósito se hizo hace unos días y ya
generó intereses, pero se registra hoy.**

## 2. Qué hay hoy

| Pieza | Estado | Dónde |
|---|---|---|
| Una tasa por plataforma + moneda, versionada por fecha | Existe | `uk_cash_yield_rates_version` (000043) |
| Un saldo `CASH-XXX` por portafolio + plataforma | Existe | `idx_entries_portfolio_asset_source` (000003) |
| Tope remunerado: lo que pasa del tope gana 0 | Existe | `max_balance` (000043), `earningBasis` |
| Tramos con tasas distintas | **Falta** | — |
| Dos productos de una entidad | Solo como dos plataformas | D1 del plan anterior |
| Un depósito con su propia tasa y plazo | **Falta** | — |
| Una tasa que empieza en el pasado | **No se permite** (desde ayer) | `notBeforeGrace` en `cash_rate.go` |
| Un saldo rinde desde que se creó en Finexia | Existe | `OpenedOn` = `pe.created_at` |

### Por qué «dos plataformas» no basta

- Las cifras por plataforma (detalle, reportes, asignación, MCP) parten la
  entidad en dos, y ninguna suma su total.
- El nombre de una plataforma es único por usuario, así que «Cajita» no se
  puede repetir en dos entidades.
- Un depósito a plazo por plataforma llenaría la lista de plataformas con lotes.

## 3. Decisiones

**D1. Un bolsillo es una subcuenta de la cuenta, no una plataforma.** Pertenece
a una plataforma + moneda, y su dinero suma en esa plataforma. Ninguna consulta
por plataforma cambia: la cajita cuenta dentro de la entidad.

**D2. La cuenta principal es `pocket_id NULL`.** Todo lo que existe hoy ya es la
cuenta principal: no hay backfill, y una plataforma sin bolsillos funciona
exactamente igual.

**D3. La clave de una cuenta pasa a ser plataforma + moneda + bolsillo.** Donde
hoy se busca `(source_id, currency)` —tasas, tope y tramos, recálculo, lecturas
de MCP y del resumen semanal— se busca con el bolsillo también.

**D4. Los tramos se escriben por «desde».** La tasa de la versión (`annual_rate`)
rige desde 0 y cada tramo rige desde un saldo. Así la tasa principal sigue
siendo la cifra que publica la entidad, y todo lo que ya la lee no cambia.
**El tope es un tramo al 0 %**: la migración convierte `max_balance` en ese
tramo y quita la columna, así que hay una sola forma de decirlo.

**D5. Los tramos se aplican a la cuenta entera, como hoy el tope.** Se calcula el
interés de la cuenta sobre lo que guardan todos sus saldos y cada saldo se lleva
su parte en proporción a lo que guarda. Con un solo tramo al 0 %, el resultado
es idéntico al tope actual.

**D6. Un depósito a tasa fija es un bolsillo de tipo `fixed`.** Tiene un solo
depósito (el de apertura), una sola versión de tasa que no admite versiones
nuevas, y un vencimiento opcional. Reutiliza todo lo que ya funciona: causación
diaria, costo promedio a cero para los intereses (D7 del plan anterior) y la
serie de crecimiento.
*Alternativa descartada:* modelarlo como un activo `bond` (el alias `cdt` ya
lleva ahí). Un bono tiene precio de mercado y cupones; un CDT o una cajita a
plazo es efectivo que rinde a una tasa conocida.

**D7. Al vencer, el dinero vuelve solo a la cuenta principal.** El job hace un
retiro del depósito y un depósito en la cuenta principal del mismo portafolio,
los dos con fecha de vencimiento. Como los flujos se compensan, la rentabilidad
no cambia.

**D8. Un depósito a tasa fija se puede abrir con fecha pasada.** Ahí no se
anotan intereses a mano, así que no hay nada que contar dos veces: al crearlo,
Finexia calcula de una vez los días que faltan hasta ayer. Los abonos se fechan
con D10 del plan anterior y cuentan como rendimiento. **Así se resuelve «lo abrí
hace días».**
En la cuenta principal y en los bolsillos flexibles sigue valiendo D6 del plan
anterior: lo que la entidad ya pagó se anota como movimiento de Intereses,
porque la cifra del extracto es exacta y el cálculo de Finexia es una
estimación.

**D9. El abono `at_maturity` es opcional; lo normal es diario.** Con abono diario
el valor del portafolio sube cada día. Con `at_maturity` (calcula cada día y
abona el último) el saldo cuadra con un extracto que solo muestra el capital,
pero el valor salta el día del vencimiento.

**D10. El libro guarda la tasa efectiva del día.** Con tramos no hay una sola
tasa, así que `cash_interest_accruals.annual_rate` guarda la que resultó ese
día: `(1 + bruto/saldo)^365 − 1`. `balance_basis` pasa a ser siempre lo que
guardaba el saldo, y ya no su parte del tope.

## 4. El cálculo con tramos

Con `H` = lo que guarda la cuenta entera al cierre, `desde_0 = 0` y
`i(r) = (1 + r)^(1/365) − 1`:

```
bruto_cuenta = Σ_k  max(0, min(H, desde_{k+1}) − desde_k) × i(r_k)
bruto_saldo  = bruto_cuenta × base_saldo / H
```

La retención, el arrastre del redondeo y el abono mensual siguen igual
(§5 del plan anterior).

### Ejemplo: 12 % hasta $5.000.000, 8 % desde ahí, saldo de $8.000.000

| Tramo | Saldo en el tramo | Tasa diaria | Interés del día |
|---|---:|---:|---:|
| Desde $0 al 12 % | 5.000.000 | 0,0310538 % | 1.552,69 |
| Desde $5.000.000 al 8 % | 3.000.000 | 0,0210874 % | 632,62 |
| **Total** | 8.000.000 | — | **2.185,31** |

Si todo el saldo ganara el 12 % serían 2.484,30. La tasa efectiva del día que
guarda el libro es **10,48 % E.A.** Si la cuenta está repartida en $6.000.000 y
$2.000.000 entre dos portafolios, a cada uno le tocan 1.638,98 y 546,33.

## 5. Depósitos a tasa fija

```
Abrir (fecha real) ─▶ Job calcula ─▶ Abonos diarios ─▶ Vence ─▶ Retiro del depósito + depósito en la principal
 capital + tasa       días que faltan  (o al vencer)    job      mismo día, flujo neto 0
```

- **Qué se rechaza:** depósitos, retiros e intereses a mano en el bolsillo; una
  versión nueva de su tasa; pausar la tasa. Todo eso se hace con **cancelar**.
- **Cancelar antes de tiempo** termina la tasa la víspera, abona lo pendiente y
  mueve el saldo a la cuenta principal. La **penalidad** va como comisión del
  retiro: cuenta como pérdida, no como aporte.
- **Borrar** quita el depósito entero (el depósito, los intereses, la tasa y el
  bolsillo). Es para algo que se anotó mal.

### Ejemplo: $10.000.000 al 10 % E.A., 90 días, abierto el 1 sep 2026 y registrado el 15 sep

| Momento | Qué pasa | Cifra |
|---|---|---:|
| 15 sep, al guardar | Se calculan del 1 al 14 sep (14 días) | + 36.624,23 |
| Cada mañana | Un día más, sobre el saldo con intereses | ≈ 2.620 al día |
| 30 nov, vencimiento | 90 días ganados en total (del 1 sep al 29 nov) | + 237.794,68 |
| 30 nov, el job | Mueve $10.237.794,68 a la cuenta principal | flujo neto 0 |

Con 4 % de retención, lo ganado neto en los 90 días es ≈ $228.176.

## 6. Modelo de datos

### 000046 — `cash_rate_tiers` (Fase 1)

```sql
CREATE TABLE IF NOT EXISTS cash_yield_rate_tiers (
  rate_id      UUID NOT NULL REFERENCES cash_yield_rates(id) ON DELETE CASCADE,
  -- Desde qué saldo de la cuenta rige. El tramo desde 0 es annual_rate de la
  -- versión, así que aquí solo van los de arriba.
  from_balance NUMERIC(20, 8) NOT NULL CHECK (from_balance > 0),
  -- Fracción, como annual_rate. 0 es un tope: desde ahí no rinde.
  annual_rate  NUMERIC(9, 6)  NOT NULL CHECK (annual_rate >= 0 AND annual_rate <= 1),
  PRIMARY KEY (rate_id, from_balance)
);

INSERT INTO cash_yield_rate_tiers (rate_id, from_balance, annual_rate)
SELECT id, max_balance, 0 FROM cash_yield_rates WHERE max_balance IS NOT NULL;

ALTER TABLE cash_yield_rates DROP COLUMN IF EXISTS max_balance;
```

Down: restaura `max_balance` desde el primer tramo al 0 % y borra la tabla. Los
demás tramos se pierden, y así debe decirlo el comentario de la migración.

### 000047 — `cash_pockets` (Fase 2)

```sql
CREATE TYPE cash_pocket_kind AS ENUM ('flexible', 'fixed');  -- 'fixed' se abre en la Fase 3

CREATE TABLE IF NOT EXISTS cash_pockets (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  source_id  UUID NOT NULL REFERENCES investment_sources(id) ON DELETE CASCADE,
  currency   CHAR(3) NOT NULL,
  name       VARCHAR(100) NOT NULL,
  kind       cash_pocket_kind NOT NULL DEFAULT 'flexible',
  opened_on  DATE NOT NULL,
  matures_on DATE CHECK (matures_on IS NULL OR (kind = 'fixed' AND matures_on > opened_on)),
  closed_on  DATE CHECK (closed_on IS NULL OR closed_on >= opened_on),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT uk_cash_pockets_name UNIQUE (source_id, currency, name)
);

-- Un bolsillo con saldos no se borra por debajo de ellos.
ALTER TABLE portfolio_entries ADD COLUMN IF NOT EXISTS pocket_id UUID REFERENCES cash_pockets(id);
ALTER TABLE cash_yield_rates  ADD COLUMN IF NOT EXISTS pocket_id UUID REFERENCES cash_pockets(id) ON DELETE CASCADE;

-- La cuenta principal conserva su clave; un bolsillo es único por portafolio.
DROP INDEX IF EXISTS idx_entries_portfolio_asset_source;
CREATE UNIQUE INDEX idx_entries_portfolio_asset_source
  ON portfolio_entries(portfolio_id, asset_id, COALESCE(source_id::TEXT, ''))
  WHERE pocket_id IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_entries_portfolio_pocket
  ON portfolio_entries(portfolio_id, pocket_id)
  WHERE pocket_id IS NOT NULL;

ALTER TABLE cash_yield_rates DROP CONSTRAINT IF EXISTS uk_cash_yield_rates_version;
CREATE UNIQUE INDEX IF NOT EXISTS uk_cash_yield_rates_version
  ON cash_yield_rates(source_id, currency, COALESCE(pocket_id::TEXT, ''), effective_from);
```

**Ojo con el índice parcial.** Postgres solo lo usa en un `ON CONFLICT` que
repita su predicado. Los tres que hay tienen que pasar a
`ON CONFLICT (portfolio_id, asset_id, COALESCE(source_id::TEXT, '')) WHERE pocket_id IS NULL`:
`postgres_entry.go:220`, `postgres_transaction.go:635` y `postgres_cash.go:273`.
Si falta uno, esa escritura falla con «no unique or exclusion constraint
matching».

`DeletePlatform` ya rechaza una plataforma con posiciones, así que la cascada
solo se lleva bolsillos vacíos.

### 000048 — `cash_fixed_deposits` (Fase 3)

```sql
ALTER TYPE cash_interest_posting ADD VALUE IF NOT EXISTS 'at_maturity';

CREATE INDEX IF NOT EXISTS idx_cash_pockets_maturing
  ON cash_pockets(matures_on) WHERE kind = 'fixed' AND closed_on IS NULL;
```

Un valor que añade `ADD VALUE` no se puede usar en la misma transacción, así que
esta migración solo lo añade. El down solo quita el índice: sacar un valor de un
enum obliga a reconstruir el tipo y a reescribir como otra cosa una tasa que ya
se abonaba al vencer, y un valor sin usar no estorba.

## 7. Backend, archivo por archivo

| Archivo | Cambio | Fase |
|---|---|---|
| `migrations/000046_*.sql` | Tramos, conversión del tope | 1 |
| `portfolio/cash_rate.go` | `RateTier`; `CashRate.Tiers` en lugar de `MaxBalance`; `Validate`: tramos ascendentes, sin repetir, `from > 0`, tasa entre 0 y 100, 10 como máximo | 1 |
| `portfolio/cash_interest.go` | `CashRateVersion.Tiers`; `accountGross(H)` en lugar de `earningBasis`; `accrueDay` reparte por la base; `EffectiveAnnualRate(H)` exportada para MCP y el correo | 1 |
| `portfolio/postgres_cash_rate.go` | Escribir y leer tramos en la transacción de la versión (`Create`, `Update`) | 1 |
| `portfolio/postgres_cash_accrual.go` | Leer los tramos con la versión; `AccountHeld` siempre que haya tramos; guardar la tasa efectiva (D10) | 1 |
| `portfolio/dto_cash_rate.go` | `tiers: [{fromBalance, annualRatePct}]` en lugar de `maxBalance` | 1 |
| `mcp/tools_portfolio.go`, `mcp/dto.go` | `tiers` en la fila de la cuenta | 1 |
| `notification/weekly_summary.go` | La media ponderada usa la tasa efectiva sobre el saldo de la cuenta | 1 |
| `migrations/000047_*.sql` | Bolsillos, `pocket_id`, índices parciales | 2 |
| `portfolio/cash_pocket.go` (nuevo) | `CashPocket`, `CashPocketKind`, validaciones, errores (`ErrCashPocketNotFound` 404, `ErrCashPocketNotEmpty` 409) | 2 |
| `portfolio/postgres_cash_pocket.go` (nuevo) | CRUD de bolsillos | 2 |
| `portfolio/postgres_entry.go`, `postgres_transaction.go`, `postgres_cash.go` | `WHERE pocket_id IS NULL` en los `ON CONFLICT` | 2 |
| `portfolio/postgres_cash.go` | `lockCashEntry` y `CreateCashMovement` con el bolsillo; `CashBalance` con `pocketId`, `pocketName` y `pocketKind`; `Move` (dos movimientos en una transacción) | 2 |
| `portfolio/postgres_cash_accrual.go` | `cashAccountHeld`, `cashAccountScope`, `GetCashAccrualTargets`, el control de versión de `AccrueCashInterestDay` y `ClearCashInterest`, con el bolsillo | 2 |
| `portfolio/postgres_cash_rate.go` | La versión más reciente y `accruedThrough`, por bolsillo | 2 |
| `portfolio/cash_interest.go` | `CashAccrualFilter` distingue «todos» de «la cuenta principal» | 2 |
| `mcp/tools_portfolio.go`, `notification/weekly_summary.go` | Clave `(source, currency, pocket)` | 2 |
| Listado de posiciones (`postgres_entry.go`) | El nombre del bolsillo en las filas de efectivo | 2 |
| `migrations/000048_*.sql` | `at_maturity`, índice de vencimientos | 3 |
| `portfolio/cash_pocket.go` | `NewFixedDepositInput.Validate`: `openedOn` desde hace 5 años como mucho y no después de hoy; `maturesOn` después de `openedOn` y no antes de ayer; `at_maturity` exige vencimiento | 3 |
| `portfolio/cash_interest.go` | `creditsOn` para `at_maturity`; `OpenedOn` de un `fixed` = `opened_on` del bolsillo | 3 |
| `portfolio/postgres_cash_pocket.go` | `OpenFixedDeposit` (bolsillo, saldo, depósito y versión con `ended_on = matures_on − 1`, en una transacción); `CloseCashPocket`; `MatureCashPockets` | 3 |
| `portfolio/postgres_cash.go`, `postgres_transaction.go`, `postgres_cash_rate.go` | Rechazar escrituras a mano y versiones nuevas sobre un `fixed` (`ErrCashPocketFixed` 409) | 3 |
| `portfolio/service_cash_interest.go` | Abrir un depósito calcula hasta ayer; el job procesa los vencimientos después de causar | 3 |
| `docs/API.md`, `docs/MANUAL_DE_USUARIO.md` (9.6), `PLAN_RENTABILIDAD_EFECTIVO.md` (D1 sustituida) | Documentación | 1–3 |

## 8. API

| Método y ruta | Qué hace | Fase |
|---|---|---|
| `POST/PUT /portfolios/cash/rates` | Aceptan `tiers` en lugar de `maxBalance` | 1 |
| `GET /portfolios/cash/rates` | Cada versión trae `tiers` | 1 |
| `GET /portfolios/cash/pockets` | Bolsillos del usuario, abiertos y cerrados | 2 |
| `POST /portfolios/cash/pockets` | Crea un bolsillo flexible `{sourceId, currency, name}` | 2 |
| `PUT /portfolios/cash/pockets/:pocketId` | Renombra | 2 |
| `DELETE /portfolios/cash/pockets/:pocketId` | Borra un flexible sin movimientos, o un `fixed` entero | 2–3 |
| `POST /portfolios/cash/movements` | Acepta `pocketId` (null = cuenta principal) | 2 |
| `POST /portfolios/cash/movements/move` | Mueve entre la principal y un bolsillo flexible, dentro de un portafolio | 2 |
| `POST /portfolios/cash/rates`, `/interest/recalculate` | Aceptan `pocketId` | 2 |
| `POST /portfolios/cash/deposits` | Abre un depósito a tasa fija | 3 |
| `POST /portfolios/cash/pockets/:pocketId/close` | Cancela un depósito `{closesOn, penalty}` | 3 |

Las rutas van junto a las de `/cash`, antes de `/:id`, como las demás.

```json
POST /portfolios/cash/deposits
{
  "portfolioId": "…",
  "sourceId": "3f7c…",
  "currency": "COP",
  "name": "CDT 90 días",
  "amount": "10000000",
  "openedOn": "2026-09-01",
  "maturesOn": "2026-11-30",
  "annualRatePct": "10",
  "withholdingPct": "4",
  "posting": "daily",
  "tiers": []
}
```

## 9. Frontend

| Archivo | Cambio | Fase |
|---|---|---|
| `lib/api/schemas/cash.ts`, `lib/api/cash.ts` | `tiers`; bolsillos; depósitos | 1–3 |
| `lib/features/cash/rates.ts` | `projectInterest` con tramos; `describeCashAccountRate`: «12 % E.A. hasta $ 5.000.000 · 8 % después» | 1 |
| `components/cash-rate-form.svelte` | «Tramos» en Opciones avanzadas en lugar de «Tope»: filas *desde saldo → tasa*, con el atajo «Solo paga hasta» (un tramo al 0 %) | 1 |
| `components/cash-accounts.svelte` | Debajo de cada cuenta, sus bolsillos: nombre, tasa y, en un depósito, «Tasa fija · vence 30 nov» | 2–3 |
| `components/cash-pocket-form.svelte` (nuevo) | Crear y renombrar un bolsillo | 2 |
| `components/cash-movement-form.svelte` | Selector de «Bolsillo» cuando la cuenta tiene alguno (por defecto, la principal); acción «Mover» | 2 |
| `components/cash-deposit-form.svelte` (nuevo) | Monto, **fecha de apertura** («Si lo abriste hace días, pon esa fecha: Finexia calcula lo que ya ganó»), plazo en días o fecha de vencimiento, tasa, retención, abono. Muestra lo que rinde al vencer | 3 |
| `routes/dashboard/cash/+page.server.ts` | Acciones de bolsillos, mover, depósitos y cancelar | 2–3 |

## 10. Fases

1. **Tramos.** 000046, dominio, API, formulario y lecturas.
   *Listo cuando:* una cuenta con 12 % hasta $5M y 8 % después abona 2.185,31
   sobre $8M, y un tope viejo sigue rindiendo igual después de migrar.
2. **Bolsillos.** 000047, clave con el bolsillo en tasas y causación,
   movimientos por bolsillo y «Mover».
   *Listo cuando:* la cuenta al 8 % y la cajita al 10 % de la misma plataforma
   causan cada una a su tasa, y la plataforma suma las dos.
3. **Depósitos a tasa fija.** 000048, abrir con fecha pasada, vencimiento,
   cancelar y bloqueos.
   *Listo cuando:* un depósito abierto hace 14 días muestra al guardarlo los
   14 días ganados, y al vencer su saldo pasa a la principal sin mover la
   rentabilidad.

**Lo que se movió de sitio al implementar la Fase 3.** `CloseCashPocket` acabó
siendo dos escrituras y no una —`EndFixedDeposit` y `SettleFixedDeposit`— porque
entre las dos hay que causar los días que el depósito todavía debía: a la tasa
que acaba de terminar, y con el dinero aún dentro. El job hace lo mismo al
vencer, con `MatureCashPockets` como segunda mitad.

Las fases van en ese orden porque cada una se apoya en la anterior: los
bolsillos solo cambian la clave de lo que la Fase 1 ya calcula, y los depósitos
son un tipo de bolsillo.

## 11. Pruebas

- **Unitarias:**
  - `accountGross` con uno, dos y tres tramos, con un tramo al 0 %, y con el
    saldo justo en el límite de un tramo.
  - Un tramo al 0 % da lo mismo que el `earningBasis` de hoy (la prueba de que
    migrar el tope no cambia nada).
  - `Validate` de tramos y de depósitos: un caso por regla.
  - `creditsOn` con `at_maturity`.
  - `PendingDays` de un `fixed` abierto en el pasado.
- **DB:**
  - La migración de un tope a un tramo rinde lo mismo que antes.
  - Tramos en una cuenta repartida entre portafolios.
  - Dos bolsillos de una plataforma causan cada uno a su tasa, y
    `GetPlatforms` suma los dos.
  - Un bolsillo y la cuenta principal en el mismo portafolio no chocan en el
    índice.
  - Los tres `ON CONFLICT` siguen funcionando para la cuenta principal.
  - «Mover» deja `NetFlow` en 0.
  - Un depósito abierto hace 14 días causa esos 14 días al crearse, con abonos
    que cuentan como rendimiento.
  - El vencimiento mueve capital e intereses y la rentabilidad no cambia.
  - Correr el vencimiento dos veces no mueve dos veces.
  - Cancelar con penalidad resta la penalidad como pérdida.
  - Un movimiento a mano o una versión nueva sobre un `fixed` se rechaza
    (también desde `CreateTransaction`).
- **Handlers:** 400 por validación, 404 con bolsillos ajenos y 409 con los
  bloqueos.

## 12. Casos borde y riesgos

| Caso | Qué pasa | Mitigación |
|---|---|---|
| Registro tarde una cuenta **flexible** que ya me pagó | El cálculo no va hacia atrás | D6 del plan anterior: se anota lo pagado. El formulario de tasa lo recuerda si el saldo es anterior a «Rige desde» |
| Registro tarde un **depósito a tasa fija** | Finexia calcula los días que faltan | D8. En él no se anotan intereses a mano |
| Tope creado antes de 000046 | Se convierte en un tramo al 0 % | Prueba de equivalencia; el down lo devuelve |
| Retención con tramos | Se aplica al bruto total | — |
| Un `fixed` en varios portafolios | No tiene sentido para un lote | Un depósito es de un solo portafolio |
| Escribir sobre un `fixed` desde la pantalla de posiciones | Se saltaría la regla de un solo depósito | Guarda en `CreateTransaction`, `Update` y `Delete`, como `requireTypeAllowed` |
| Borrar una de las dos patas de «Mover» | El saldo de un lado queda descuadrado | Aviso en la interfaz. Sin columna de enlace; se paga si hace falta |
| El job estuvo caído el día del vencimiento | Se procesa en la corrida siguiente con fecha `matures_on` | Las dos patas se compensan, también como historia |
| Depósito abierto hace años | Muchos días que calcular al guardarlo | `openedOn` hasta 5 años atrás |
| `at_maturity` en un plazo largo | El valor salta al vencer | D9: por defecto es diario |
| Índice parcial sin su predicado en un `ON CONFLICT` | La escritura falla | Los tres sitios están listados en §6 y tienen prueba |
| Clientes de `maxBalance` | Dejan de recibirlo | Solo lo usan el frontend y MCP, y los dos cambian en la Fase 1 |

**Fuera de alcance:** renovación automática de un depósito al vencer (se abre
otro con el monto que llegó); tasas que dependen de la antigüedad del dinero
dentro de una cuenta flexible; y la retención con umbral en UVT.

## 13. Criterios de aceptación

Fase 1:

- [x] Registro 12 % E.A. con un tramo al 8 % desde $5.000.000, y sobre $8.000.000 el abono del día es 2.185,31.
- [x] Una cuenta con tope antes de la migración abona lo mismo después.
- [x] La cuenta muestra «12 % E.A. hasta $ 5.000.000 · 8 % después» y la proyección usa los tramos.

Fase 2:

- [x] Creo la cajita «Viajes» en Nu (COP), le doy 10 % y muevo $2.000.000 desde la cuenta principal al 8 %: cada una causa a su tasa.
- [x] La plataforma Nu suma la cuenta y la cajita; la rentabilidad no cambia al mover.
- [x] Una plataforma sin bolsillos se comporta exactamente como hoy.

Fase 3:

- [x] Abro hoy (15 sep) un depósito de $10.000.000 al 10 % E.A. con fecha 1 sep a 90 días, y al guardarlo tiene 36.624,23 de intereses.
- [x] No puedo depositar, retirar, anotar intereses ni cambiar la tasa del depósito.
- [x] El 30 nov el saldo del depósito pasa a la cuenta principal y la rentabilidad no se mueve.
- [x] Cancelarlo antes con una penalidad la cuenta como pérdida.
