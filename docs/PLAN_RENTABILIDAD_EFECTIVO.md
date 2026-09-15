# Plan — Rentabilidad del efectivo (tasas tipo renta fija)

> **Estado:** implementado · 15 sep 2026
> **Alcance:** módulo `portfolio` (backend), feature `cash` (frontend)
> **Migraciones:** 000042 – 000045 · **Fases:** 4

### Estado de implementación

| Fase | Estado |
|---|---|
| 1. La ganancia del efectivo cuenta | Implementada (000042) |
| 2. Registrar tasas | Implementada (000043) |
| 3. Causación automática | Implementada (000044, job `accrue-cash-interest`) |
| 4. Pulido | Implementada (000045) |

Decisiones que cambiaron respecto al plan al implementarlo:

- **Costo promedio (D7).** Solo recorren el costo promedio por fecha los saldos
  de efectivo que tienen intereses. El resto de posiciones conserva la fórmula
  de 000041.
- **Tasas.** Llegan por `GET /portfolios/cash/rates`, no dentro de los saldos.
  Solo cambia la versión más reciente. Borrar un cambio de tasa reabre la
  versión anterior si terminaba justo la víspera.
- **Fechas.** Se aceptan desde ayer en UTC, con un día de margen para quien
  está al oeste de Greenwich.
- **Abono y tope.** La Fase 4 abrió los dos. `posting: monthly` calcula cada día
  igual y abona el mes entero el último día, en una sola transacción ligada a
  cada día que paga; un día pendiente cuenta en el saldo del siguiente, así que
  un año por meses rinde lo mismo que uno por días. Si la cuenta deja de rendir
  antes de cerrar el mes, el job abona lo pendiente el último día que ganó.
  `max_balance` es de la cuenta: sus saldos se reparten el tope en proporción a
  lo que guarda cada uno.
- **Primer día que rinde un saldo.** Un saldo gana desde el día en que se abrió
  (`created_at` en UTC), no desde el primer día de la tasa.
- **Saldo del día en recuperaciones.** Los abonos se cuentan por el día en que
  se ganaron, aunque su transacción quede fechada después (D10).
- **Nombres.** Los archivos se llaman `cash_rate.*` (la tasa) y `cash_interest.*`
  / `postgres_cash_accrual.go` (lo que rinde), no `cash_yield.*`. La tabla sí
  conservó el nombre `cash_yield_rates`. Las tablas de §8, §9, §10 y §12 dicen
  ya los nombres reales.
- **Campos que no se hicieron.** Dos del §9 se quedaron fuera porque nada los
  usa: `annualRatePct` en cada movimiento (el extracto agrupa los abonos y no
  los desglosa por tasa) y `accruals`/`interestEarned` por versión de la tasa
  (basta `accruedThrough`, que es lo que decide si una versión se puede
  corregir). Añadirlos es una consulta más en cada listado; si alguna pantalla
  los pide, ahí es cuando se pagan.
- **Plataforma ajena al recalcular.** `POST /portfolios/cash/interest/recalculate`
  comprueba la propiedad dentro de su misma transacción, como hace
  `CreateCashRate`, y responde `ErrPlatformNotFound` (404): la misma respuesta
  que las demás escrituras de una tasa. Que la plataforma esté inactiva no
  importa — una que dejó de recibir dinero sigue pudiendo corregir su pasado.

## 1. Qué se quiere

Una plataforma que paga rendimiento sobre el saldo (una cuenta de ahorro, una
"cajita", una billetera en dólares con APY) debe poder:

1. **Registrar su tasa**, por ejemplo 9 % E.A., sobre la cuenta de efectivo de
   esa plataforma.
2. **Calcular los intereses sola**, día a día, sin que el usuario los anote.
3. Hacer que esos intereses **suban el saldo y cuenten como ganancia** del
   portafolio donde está el efectivo, en todas las cifras: ganancia,
   rentabilidad, plataformas y crecimiento.

## 2. Qué había antes

> Este apartado es el planteamiento: describe el estado **anterior** a las
> migraciones 000042 – 000045. Lo que hay hoy está en la tabla de arriba.

| Pieza | Estado | Dónde |
|---|---|---|
| El efectivo es una posición (`CASH-COP`, precio 1) por plataforma + portafolio + moneda | Existe | `portfolio/cash.go`, `portfolio/postgres_cash.go` |
| Depósito, retiro e intereses son movimientos (`transfer_in`, `transfer_out`, `cash_interest`) | Existe | migraciones 000040 y 000041 |
| Los intereses anotados cuentan como rentabilidad en la serie de crecimiento, no como aporte | Existe | `postgres_cash_db_test.go:233` |
| Intereses anotados con fecha pasada entran como historia (flujo), no como rendimiento | Existe | `postgres_cash_db_test.go:255` |
| Guardar la tasa de una cuenta | **Falta** | — |
| Calcular y abonar los intereses automáticamente | **Falta** | — |
| Que los intereses aparezcan en la **ganancia** del portafolio | **Falta (modelo)** | `recalculate_avg_cost` en 000041 |

### El hueco que no se ve

`portfolio_summary` (000039) calcula `ganancia = valor − costo`, y el costo sale
de `cantidad × precio promedio`. Para el efectivo, `recalculate_avg_cost`
mantiene el precio promedio en 1 aunque la cantidad incluya los intereses. El
costo crece igual que el valor:

| Caso | Valor | Costo | Ganancia |
|---|---:|---:|---:|
| Depósito $10.000.000 + intereses $71.082 | 10.071.082 | 10.071.082 | **0** |

Los reportes sí muestran +0,71 %, porque Dietz resta los flujos y
`cash_interest` no es un flujo. Pero la tarjeta del portafolio, las plataformas
(`postgres_platform.go:51`) y `total_gain_loss` de los snapshots dicen 0.
Automatizar los intereses sin arreglar esto daría una cifra que el usuario nunca
vería como ganancia. Por eso es la **Fase 1**.

## 3. Cómo va a funcionar

```
Tasa registrada ─▶ Job diario ─▶ Causación ─▶ Abono cash_interest ─▶ Saldo sube ─▶ Snapshot ─▶ Ganancia y rentabilidad
  9 % E.A.         05:30 UTC     por saldo     precio 1, sin flujo                 22:00 UTC
```

No hay que reescribir la rentabilidad. Cada abono es un `cash_interest` igual al
que hoy se anota a mano, y todo lo que ya lo sabe leer sigue funcionando: la
serie de crecimiento, los reportes y el retiro de transacciones (000038).

## 4. Decisiones

**D1. La tasa pertenece a la cuenta: plataforma + moneda.** Es lo que publica la
entidad y lo que aparece en el extracto. Si la cuenta está repartida entre
varios portafolios, cada saldo gana la misma tasa sobre lo suyo. Si una entidad
tiene dos productos con tasas distintas (cuenta y bolsillo), se registran como
dos plataformas.

**D2. Forma canónica: tasa efectiva anual (E.A.), base 365.** Es la convención
de las cuentas de ahorro en Colombia y equivale al APY de las cuentas en
dólares. Se guarda como fracción (`0.090000`). Convertir una tasa nominal es
ayuda del formulario (Fase 4): `annualFromNominal` en `lib/features/cash`.

**D3. Las tasas se versionan por fecha de vigencia.** Si la entidad cambia la
tasa, se crea una versión nueva desde una fecha y la anterior se cierra. Una
tasa que ya generó intereses no se edita; así cada día conserva la tasa con la
que se calculó.

**D4. Causación diaria con capitalización diaria.** El interés de un día se
calcula sobre un saldo que ya incluye los intereses anteriores. Así, en un año,
el rendimiento coincide exactamente con la E.A. El abono es diario por defecto;
el abono mensual (Fase 4) acumula lo causado y lo abona el último día del mes.

**D5. Base del cálculo: saldo al cierre del día.** Se suman las transacciones
con `transaction_date ≤ día`, más lo causado y pendiente de abono si el abono es
mensual. Si hay tope remunerado, el tope se aplica a toda la cuenta y se reparte
entre los portafolios en proporción a su saldo.

**D6. Solo hacia adelante.** La vigencia empieza hoy o después. Lo ganado antes
de registrar la tasa se anota como movimiento de "Intereses", que ya existe. Si
tiene fecha pasada, entra como historia, que es lo correcto: ese dinero ya
estaba en el saldo que el usuario cargó.

**D7. Los intereses son unidades a costo cero en el costo promedio del
efectivo.** Esta es la decisión que ordena todo. Con ella, el precio promedio de
un saldo con intereses baja de 1 (10.000.000 / 10.071.082 = 0,992942) y **todas**
las lecturas que calculan `cantidad × precio promedio` muestran la ganancia sin
cambiar ninguna consulta: resumen, plataformas, snapshots, crecimiento, MCP y
correo semanal. Es la misma lógica de costo promedio que ya usan los demás
activos.
*Alternativa descartada:* una columna aparte de "intereses retenidos". Obligaría
a tocar cada consulta de costo y a mantener dos definiciones de ganancia.

**D8. Idempotencia por construcción.** Hay un `UNIQUE (entry_id, accrual_date)`
en la tabla de causaciones. Correr el job dos veces, tener dos réplicas o hacer
una recuperación nunca abona el mismo día dos veces.

**D9. Retención y redondeo.** La retención en la fuente es opcional (un
porcentaje por tasa) y el abono se registra neto, como pide hoy el formulario.
Cada abono se redondea a los decimales de la moneda y el resto pasa al día
siguiente (`rounding_carry`), así que en un año no se pierde ni un centavo.

**D10. La fecha contable del abono nunca cae en la historia.** Se usa
`GREATEST(día causado, fecha del último snapshot del portafolio)`. Si el job
estuvo caído varios días, los abonos atrasados siguen contando como rendimiento
y no como aporte (ver el cálculo de `stretch.opened_on` en
`postgres_snapshot.go`). La fecha real queda en `accrual_date`.

## 5. El cálculo

```
i_d        = (1 + EA)^(1/365) − 1
bruto_D    = base_D × i_d
retención_D = bruto_D × tasa_retención
neto_D     = bruto_D − retención_D
abono_D    = redondear(neto_D + arrastre_{D−1}, decimales de la moneda)
arrastre_D = neto_D + arrastre_{D−1} − abono_D
```

En Go, con la librería que ya está en `go.mod`:

```go
// dailyRate convierte una tasa E.A. a su equivalente diario, base 365.
func dailyRate(ea decimal.Decimal) (decimal.Decimal, error) {
	return compoundinterest.NewRateConversion().
		RateDecimal(ea).
		EffectiveAnnual().
		Annually().
		ToPeriodicAt(compoundinterest.Daily)
}
```

Un test debe fijar que `(1 + i_d)^365 − 1 = EA`, para no depender de cómo
interpreta el builder la frecuencia de origen.

### Ejemplo: $10.000.000 COP al 9 % E.A., abono diario

`i_d = 1,09^(1/365) − 1 = 0,000236131` (0,0236131 % diario)

| Día | Base | Interés del día | Abono | Saldo al cierre |
|---|---:|---:|---:|---:|
| 1 | 10.000.000,00 | 2.361,3115 | 2.361,31 | 10.002.361,31 |
| 2 | 10.002.361,31 | 2.361,8691 | 2.361,87 | 10.004.723,18 |
| 3 | 10.004.723,18 | 2.362,4268 | 2.362,43 | 10.007.085,61 |
| 30 | — | — | Σ ≈ 71.082 | ≈ 10.071.082 |
| 365 | — | — | Σ ≈ 900.000 | ≈ 10.900.000 |

Con una retención del 7 %, el abono del día 1 es de 2.196,02 y el rendimiento
neto queda en ≈ 8,34 % E.A.

## 6. La ganancia en el portafolio (D7 en números)

El costo promedio se recorre en orden de fecha:

- Un depósito suma unidades y costo.
- Un interés suma unidades, pero no costo.
- Un retiro se lleva la parte proporcional del costo.
- Un saldo vaciado vuelve a empezar en 1.

| Movimiento | Unidades | Costo | Costo/unidad | Ganancia |
|---|---:|---:|---:|---:|
| Depósito 10.000.000 | 10.000.000,00 | 10.000.000,00 | 1,000000 | 0 |
| 30 días de intereses | 10.071.082,43 | 10.000.000,00 | 0,992942 | 71.082 |
| Retiro 5.000.000 | 5.071.082,43 | 5.035.290,35 | 0,992942 | 35.792 |
| Retiro del resto | 0 | 0 | 1 (se reinicia) | 0 |
| Depósito 2.000.000 | 2.000.000,00 | 2.000.000,00 | 1,000000 | 0 |

Hay que explicar en la interfaz la diferencia entre las dos cifras:

- **Ganancia** es lo no realizado: baja con los retiros, igual que al vender
  acciones.
- **Rentabilidad** (Dietz, en reportes y crecimiento) cuenta el 100 % del
  interés en el periodo en que se ganó, aunque después se retire.

Las dos son correctas; responden preguntas distintas.

## 7. Modelo de datos

### 000042 — `cash_interest_cost_basis` (Fase 1)

```sql
-- El costo promedio de un saldo de efectivo, recorrido en orden: el interés
-- entra como unidades a costo cero, un retiro se lleva su parte proporcional
-- del costo y un saldo vaciado vuelve a empezar en 1.
CREATE OR REPLACE FUNCTION cash_entry_avg_cost(p_entry_id UUID)
RETURNS NUMERIC
LANGUAGE plpgsql
STABLE
AS $$
DECLARE
  v_units NUMERIC := 0;
  v_cost  NUMERIC := 0;
  t       RECORD;
BEGIN
  FOR t IN
    SELECT type, quantity, price, COALESCE(fx_rate, 1) AS fx_rate
    FROM transactions
    WHERE entry_id = p_entry_id
    ORDER BY transaction_date, created_at
  LOOP
    IF t.type IN ('buy', 'transfer_in') THEN
      v_units := v_units + t.quantity;
      v_cost  := v_cost + t.quantity * t.price * t.fx_rate;
    ELSIF t.type = 'cash_interest' THEN
      v_units := v_units + t.quantity;
    ELSIF t.type IN ('sell', 'transfer_out') AND v_units > 0 THEN
      v_cost  := v_cost * GREATEST(1 - t.quantity / v_units, 0);
      v_units := GREATEST(v_units - t.quantity, 0);
    END IF;
  END LOOP;

  RETURN CASE WHEN v_units > 0 THEN v_cost / v_units ELSE 1 END;
END;
$$;

-- Lo que antes se reconocía por pe.price = 1: un saldo que la app lleva uno a
-- uno. Ahora el precio promedio baja con los intereses, así que la prueba se
-- hace sobre las transacciones.
CREATE OR REPLACE FUNCTION cash_entry_at_par(p_entry_id UUID)
RETURNS BOOLEAN
LANGUAGE sql
STABLE
AS $$
  SELECT NOT EXISTS (
    SELECT 1
    FROM transactions t
    JOIN portfolio_entries pe ON pe.id = t.entry_id
    WHERE t.entry_id = p_entry_id
      AND t.type IN ('buy', 'transfer_in', 'sell', 'transfer_out')
      AND (t.price <> 1 OR t.fx_rate <> 1 OR t.currency <> pe.cost_currency)
  );
$$;
```

- `recalculate_avg_cost`: la **cantidad** se sigue calculando como en 000041
  (suma sin orden, que es lo que validan los chequeos de saldo en Go). Solo el
  **precio** de las posiciones con `asset_type = 'cash'` pasa a
  `cash_entry_avg_cost(entry_id)`. La rama de los demás activos no cambia.
- Backfill: `UPDATE portfolio_entries SET price = cash_entry_avg_cost(id)` para
  las posiciones de efectivo que tengan algún `cash_interest`. Se actualiza
  `portfolio_entries` directamente para no disparar los triggers de
  `transactions`, como hacen 000036 y 000038.
- Down: restaura la función de 000041 y recalcula el precio de esas posiciones
  con su fórmula.

### 000043 — `cash_yield_rates` (Fase 2)

```sql
CREATE TYPE cash_interest_posting AS ENUM ('daily', 'monthly');

CREATE TABLE IF NOT EXISTS cash_yield_rates (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  source_id        UUID NOT NULL REFERENCES investment_sources(id) ON DELETE CASCADE,
  currency         CHAR(3) NOT NULL,
  -- Fracción: 0.090000 es 9 % E.A.
  annual_rate      NUMERIC(9, 6) NOT NULL CHECK (annual_rate > 0 AND annual_rate <= 1),
  withholding_rate NUMERIC(5, 4) NOT NULL DEFAULT 0 CHECK (withholding_rate >= 0 AND withholding_rate < 1),
  max_balance      NUMERIC(20, 8) CHECK (max_balance IS NULL OR max_balance > 0),
  posting          cash_interest_posting NOT NULL DEFAULT 'daily',
  effective_from   DATE NOT NULL,
  ended_on         DATE CHECK (ended_on IS NULL OR ended_on >= effective_from),
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT uk_cash_yield_rates_version UNIQUE (source_id, currency, effective_from)
);
```

Tasa vigente para un día D: la fila con el mayor `effective_from ≤ D` de esa
plataforma y moneda, siempre que `ended_on IS NULL OR ended_on ≥ D`. Pausar es
poner `ended_on`; reanudar es crear una versión nueva. La propiedad se comprueba
con `investment_sources.user_id`, como ya hace `CreateCashMovement`.

### 000044 — `cash_interest_accruals` (Fase 3)

```sql
CREATE TABLE IF NOT EXISTS cash_interest_accruals (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  entry_id       UUID NOT NULL REFERENCES portfolio_entries(id) ON DELETE CASCADE,
  rate_id        UUID REFERENCES cash_yield_rates(id) ON DELETE SET NULL,
  accrual_date   DATE NOT NULL,
  -- Copias: la historia no depende de que la tasa siga existiendo.
  annual_rate    NUMERIC(9, 6)  NOT NULL,
  balance_basis  NUMERIC(20, 8) NOT NULL,
  gross_amount   NUMERIC(20, 8) NOT NULL,
  withholding    NUMERIC(20, 8) NOT NULL,
  net_amount     NUMERIC(20, 8) NOT NULL,
  rounding_carry NUMERIC(20, 8) NOT NULL DEFAULT 0,
  -- pending: causado sin abonar (abono mensual)
  -- posted:  abonado; si transaction_id quedó NULL, el usuario borró el abono
  -- carried: el abono redondeado fue 0 y todo pasó al arrastre
  status         TEXT NOT NULL CHECK (status IN ('pending', 'posted', 'carried')),
  transaction_id UUID REFERENCES transactions(id) ON DELETE SET NULL,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT uk_cash_accrual_day UNIQUE (entry_id, accrual_date)
);

CREATE INDEX IF NOT EXISTS idx_cash_accruals_transaction ON cash_interest_accruals(transaction_id);
```

La relación con la transacción vive en esta tabla, así que `transactions` no
cambia. Borrar un abono automático deja la causación con `transaction_id NULL`,
y el `UNIQUE` impide que el job vuelva a crearlo.

## 8. Backend, archivo por archivo

Todo queda dentro del módulo `portfolio`, que ya es dueño del efectivo. No hay
dependencias nuevas entre módulos, así que `arch_test.go` sigue pasando.

| Archivo | Cambio | Fase |
|---|---|---|
| `migrations/000042_*.sql` | Costo promedio del efectivo y `cash_entry_at_par` | 1 |
| `portfolio/postgres_cash.go` | Cambiar `pe.price = 1` (líneas 284 y 336) por `cash_entry_at_par(pe.id)` | 1 |
| `portfolio/cash.go` | Actualizar el comentario del modelo: el interés ya no deja el precio en 1 | 1 |
| `migrations/000043_*.sql` | Tabla de tasas | 2 |
| `portfolio/cash_rate.go` | `CashRate`, `CashRateInput.Validate`, `InterestPosting`, errores (`ErrInvalidCashRate` 400, `ErrCashRateInUse`/`ErrCashRateNotLatest`/`ErrCashRateOverlaps` 409, `ErrCashRateNotFound` 404) | 2 |
| `portfolio/cash_interest.go` | `CashRateVersion`, `PendingDays`, `dailyRate`, `accrueDay` (funciones puras) | 3 |
| `portfolio/dto_cash_rate.go` | DTOs de creación, edición, pausa y recálculo | 2 |
| `portfolio/repository.go` | Interfaces `CashRateStore` y `CashAccrualStore`, embebidas en el repositorio | 2–3 |
| `portfolio/postgres_cash_rate.go` | CRUD de tasas. `Create` cierra la versión vigente en la misma transacción | 2 |
| `portfolio/service_cash_rate.go` | Casos de uso de las tasas | 2 |
| `portfolio/service_cash_interest.go` | `AccrueCashInterest(ctx, through)` y `RecalculateCashInterest` | 3–4 |
| `portfolio/handler_cash_rate.go` | Handlers con el mismo patrón que `handler_cash.go` | 2 |
| `portfolio/module.go` | Rutas `/cash/rates…` junto a las de `/cash`, antes de `/:id` | 2 |
| `migrations/000044_*.sql` | Tabla de causaciones | 3 |
| `portfolio/postgres_cash_accrual.go` | Objetivos por día; causación de un saldo en una transacción con `FOR UPDATE` | 3 |
| `portfolio/cash_interest_job.go` | `CashInterestJob` (`Name() = "accrue-cash-interest"`) | 3 |
| `app/app.go` | `sched.Register(portfolio.NewCashInterestJob(…), scheduler.DailyAt{Hour: 5, Minute: 30}, scheduler.WithStore(persistent))` | 3 |
| `portfolio/cash.go` + `postgres_cash_accrual.go` | `CashBalance` suma `interestEarned`, `interestThisMonth`, `interestThisMonthValue`, `lastAccrualDate` (y `pendingInterest` en la Fase 4); `CashMovement` suma `automatic` | 3 |
| `migrations/000045_*.sql` | Saldo al cierre como función, con lo causado y sin abonar | 4 |
| `portfolio/cash_interest.go` | `CashDay`, `creditsOn`, `earningBasis`, `creditHeld`, `CashAccrualFilter`, `RecalculateCashInterestInput` | 4 |
| `portfolio/postgres_cash_accrual.go` | `GetHeldCashInterest`, `PostHeldCashInterest`, `ClearCashInterest` | 4 |
| `mcp/tools_portfolio.go` | Tool `get_cash_accounts` | 4 |
| `notification/weekly_summary.go` | Bloque de efectivo del resumen semanal | 4 |
| `docs/API.md` (Efectivo), `docs/MANUAL_DE_USUARIO.md` (9.5 y 9.6) | Documentación | 4 |

### Por qué 05:30 UTC

A las 05:30 UTC ya terminó en Colombia el día que se causa (00:30 hora local).
Además, el abono queda registrado mucho antes del snapshot de las 22:00 UTC, así
que cae en el tramo que abrió el snapshot anterior y cuenta como rendimiento.
Los días son UTC, igual que `snapshotDay`.

### El job, en pseudocódigo

```go
func (s *service) AccrueCashInterest(ctx context.Context, through time.Time) (int, []error) {
	// Un objetivo por saldo: efectivo a la par, con una tasa vigente
	// para (source_id, cost_currency) en algún día sin causar
	// entre max(effective_from, última causación + 1) y through.
	targets, err := s.repo.CashAccrualTargets(ctx, through)
	if err != nil {
		return 0, []error{err}
	}

	for _, t := range targets {
		for day := t.From; !day.After(t.Through); day = day.AddDate(0, 0, 1) {
			// En una transacción: bloquea el saldo, lee base + arrastre
			// + pendientes, calcula con accrueDay (Go puro), inserta la
			// causación (ON CONFLICT DO NOTHING) y, si toca abonar,
			// el cash_interest con fecha GREATEST(day, último snapshot).
			err := s.repo.AccrueCashEntry(ctx, t, day, accrueDay)
			// Registra y cuenta los errores por saldo, como SyncPortfolioSnapshots.
		}
	}
}
```

### Base del día

```sql
SELECT COALESCE(SUM(CASE
         WHEN type IN ('buy', 'transfer_in', 'cash_interest') THEN quantity
         WHEN type IN ('sell', 'transfer_out')                THEN -quantity
         ELSE 0 END), 0)
FROM transactions
WHERE entry_id = $1 AND transaction_date <= $2;
-- + SUM(net_amount) de las causaciones 'pending' con accrual_date < $2 (abono mensual)
-- Con tope: base × LEAST(1, max_balance / saldo total de la cuenta en ese día)
```

## 9. API

| Método y ruta | Qué hace |
|---|---|
| `GET /portfolios/cash/rates` | Tasas del usuario, vigentes e históricas, con `latest` y `accruedThrough` |
| `POST /portfolios/cash/rates` | Crea una tasa o una versión nueva. Si hay una vigente, la cierra el día anterior |
| `PUT /portfolios/cash/rates/:rateId` | Edita una tasa **sin** causaciones; si ya tiene, responde 409 |
| `POST /portfolios/cash/rates/:rateId/end` | Pausa la tasa desde `endedOn` |
| `DELETE /portfolios/cash/rates/:rateId` | Borra una tasa sin causaciones |
| `GET /portfolios/cash` | Cada saldo trae `interestEarned`, `interestThisMonth`, `interestThisMonthValue`, `pendingInterest` y `lastAccrualDate` |
| `GET /portfolios/cash/movements` | Cada movimiento trae `automatic` |
| `POST /portfolios/cash/interest/recalculate` | Rehace los días de una cuenta desde `from` |

```json
POST /portfolios/cash/rates
{
  "sourceId": "3f7c…",
  "currency": "COP",
  "annualRatePct": "9.00",
  "withholdingPct": "0",
  "maxBalance": null,
  "posting": "daily",
  "effectiveFrom": "2026-09-15"
}
```

**Validaciones** (en `CashYieldRateInput.Validate`, con un mensaje por regla,
como `CashMovementInput`):

- La tasa es mayor que 0 % y como máximo 100 %.
- La retención está entre 0 % y 100 % (sin llegar a 100).
- `effectiveFrom` es hoy (UTC) o una fecha posterior.
- La moneda está en `currency.IsSupported`.
- La plataforma pertenece al usuario y está activa.
- `posting` vale `daily` o `monthly` (`monthly` solo desde la Fase 4).
- `maxBalance`, si viene, es mayor que 0.

## 10. Frontend

| Archivo | Cambio |
|---|---|
| `lib/api/schemas/cash.ts` | `cashRateSchema`; `cashBalanceSchema` suma `interestEarned`, `interestThisMonth`, `interestThisMonthValue`, `pendingInterest`, `lastAccrualDate`; `cashMovementSchema` suma `automatic` |
| `lib/api/cash.ts` | `getRates`, `createRate`, `updateRate`, `endRate`, `deleteRate`, `recalculateInterest` |
| `lib/features/cash/rates.ts` | `formatAnnualRate` ("9,25 % E.A."), `dailyRateFromAnnual`, `annualFromNominal`, `projectInterest`, `cashAccountRate`, `describeCashAccountRate`, `cashYield`, `cashRateErrorMessage` |
| `lib/features/cash/interest.ts` | `groupAutomaticInterest`, `groupCashLedgerByMonth`: los abonos del libro, agrupados |
| `lib/features/cash/schemas.ts` | `cashRateCreateSchema`/`UpdateSchema`/`EndSchema`/`DeleteSchema`, `cashRecalculateSchema`, `toCashRateBody` |
| `components/cash-rate-form.svelte` (nuevo) | Tasa E.A. (%), vigente desde (hoy), abono, y en "Opciones avanzadas" la retención y el tope. Proyección en vivo sobre el saldo actual |
| `components/cash-accounts.svelte` | Chip "9 % E.A." o enlace "Agregar tasa" en cada cuenta, lo ganado este mes y lo calculado sin abonar |
| `components/cash-summary.svelte` | "Intereses del mes" y "tasa promedio ponderada" del efectivo |
| `components/cash-movements.svelte` | Los abonos automáticos se agrupan por cuenta y mes ("Intereses automáticos · 14 abonos · + $ 33.071"), se pueden expandir y llevan la marca "Automático" |
| `routes/dashboard/cash/+page.server.ts` | Acciones `createRate`, `updateRate`, `endRate`, `deleteRate`, `recalculateInterest` |
| Posiciones del portafolio | En las filas de efectivo se oculta el costo promedio (ahora sería 0,99) y la ganancia se rotula "Intereses" |

**Textos del formulario:**

- *Tasa efectiva anual (E.A.):* "La que publica la entidad. Si te dan un APY en
  dólares, es la misma cifra."
- *Vigente desde:* "Los intereses se calculan desde este día. Lo que ganaste
  antes, anótalo como movimiento de intereses."
- *Proyección:* "Con tu saldo de hoy, $ 10.000.000: ≈ $ 2.361 al día · $ 71.082
  en 30 días · $ 900.000 en un año."

**Pruebas:** `cash.spec.ts` (proyección, formato, agrupación), `schemas.spec.ts`
y un spec de componente para la proyección del formulario. Para las capturas: el
stub de e2e no sirve `/portfolios/cash`; hay que usar un proxy con `vite dev`.

## 11. Fases

1. **La ganancia del efectivo cuenta.** Migración 000042, `cash_entry_at_par` y
   ocultar el costo promedio en las filas de efectivo. *Valor inmediato:* los
   intereses que hoy se anotan a mano ya aparecen como ganancia.
   *Listo cuando:* depositar 10.000.000 y anotar 71.082 de intereses muestra una
   ganancia de 71.082 en el portafolio y en la plataforma.
2. **Registrar tasas.** Migración 000043, dominio, CRUD de la API, formulario,
   chip y proyección. Todavía no se causa nada.
   *Listo cuando:* se crea, versiona, pausa y borra una tasa, y la cuenta
   muestra la proyección.
3. **Causación automática.** Migración 000044, job, lecturas extendidas y
   agrupación en movimientos.
   *Listo cuando:* al día siguiente de registrar la tasa aparece el abono, el
   saldo sube y la ganancia y la rentabilidad lo reflejan.
4. **Pulido.** Abono mensual, conversor de nominal a E.A., tope remunerado,
   "recalcular desde una fecha", tool de MCP, intereses en el resumen semanal,
   `API.md` y manual.
   *Cómo quedó:*
   - Migración 000045: `cash_entry_balance_at_close(entry_id, día)`, que además
     de las transacciones cuenta lo causado y sin abonar, y el índice parcial de
     las causaciones pendientes.
   - `POST /portfolios/cash/interest/recalculate` borra los abonos automáticos de
     una cuenta desde un día y los vuelve a causar. La ventana se ensancha hacia
     atrás hasta el primer día de un abono que cruce la fecha: medio abono
     mensual no se puede deshacer. Borrar un `cash_interest` no toca la serie de
     crecimiento (000038), así que lo que se vuelve a abonar entra como
     rendimiento igual que la primera vez.
   - El conversor de nominal a E.A. vive en el formulario
     (`annualFromNominal`), junto a la retención y el tope, en «Opciones
     avanzadas».
   - Tool de MCP `get_cash_accounts`: los saldos con la tasa vigente de su
     cuenta, lo ganado y lo pendiente.
   - El resumen semanal trae un bloque de efectivo: el total, lo ganado en el
     mes, la tasa media ponderada de los saldos que rinden y cuántos no rinden.
     La media deja fuera los saldos sin tasa y los cuenta aparte.

## 12. Pruebas de backend

- **Unitarias (`cash_rate_test.go`, `cash_interest_test.go`):**
  - `dailyRate` cumple `(1 + i_d)^365 − 1 = EA` para 0,5 %, 9 % y 13,5 %.
  - `accrueDay` con retención, con tope, con base 0 y con arrastre (que un año
    de abonos redondeados sume lo mismo que el cálculo sin redondeo).
  - `Validate`: un caso por regla.
- **DB (`postgres_cash_db_test.go`, `postgres_cash_gain_db_test.go`, `postgres_cash_rate_db_test.go`, `postgres_cash_accrual_db_test.go`, `postgres_cash_posting_db_test.go`):**
  - La ganancia del resumen y de la plataforma incluye los intereses.
  - Un retiro reduce la ganancia en proporción.
  - Vaciar la cuenta y volver a depositar deja la ganancia en 0.
  - `lockCashEntry` encuentra el saldo aunque su precio ya no sea 1.
  - Causar el mismo día dos veces deja un solo abono.
  - Un abono causado cuenta como rendimiento: `NetFlow` no cambia.
  - Una recuperación de varios días usa la fecha contable ajustada y sigue
    contando como rendimiento.
  - Un cambio de tasa a mitad de mes usa la tasa de cada día.
  - Una tasa pausada no causa.
  - Un abono automático borrado no se vuelve a crear.
  - Una cuenta repartida con tope reparte el tope en proporción al saldo.
  - Una cuenta sin tasa se comporta exactamente como hoy.
- **Handlers (`handler_cash_rate_test.go`):** 400 en cada validación, 404 con
  una plataforma ajena y 409 al editar una tasa con causaciones.

## 13. Casos borde y riesgos

| Caso | Qué pasa | Mitigación |
|---|---|---|
| Se agrega un depósito con fecha pasada después de causar esos días | Los intereses de esos días no se recalculan solos | "Recalcular desde" (Fase 4): borra los abonos automáticos desde una fecha (000038 los retira sin afectar la serie) y los vuelve a causar |
| El job estuvo caído | Recupera todos los días pendientes | D10: la fecha contable se ajusta al último snapshot |
| La entidad cambia la tasa | — | Se crea una versión nueva desde hoy |
| Salto en la línea de ganancia el día de la migración 000042 | Los snapshots viejos guardaron ganancia 0 para el efectivo | Avisarlo en el changelog. La rentabilidad (Dietz) no salta, porque el valor no cambia |
| Precisión de `price NUMERIC(20,8)` | Error de costo menor a $1 en saldos de $100.000.000 | Aceptable; queda documentado en la migración |
| "Ganancia" y "Rentabilidad" no coinciden después de un retiro | Es esperado (§6) | Tooltip en la interfaz |
| Posición `CASH-USD` comprada con pesos a una tasa de cambio | `cash_entry_at_par` la excluye, igual que hoy | Sin cambios |

**Fuera de alcance (anotar en `TECH_DEBT.md`):** la rama de `recalculate_avg_cost`
para acciones promedia todas las compras de la historia. Si se vende todo y se
vuelve a comprar, el costo mezcla el lote viejo con el nuevo. D7 corrige esto
solo para el efectivo.

## 14. Criterios de aceptación

Cada uno con la prueba que lo fija (`internal/portfolio`, contra un Postgres con
las migraciones aplicadas).

- [x] Registro 9 % E.A. en la cuenta en COP de una plataforma, desde hoy, y veo la tasa y la proyección en la cuenta. — `TestCashRateNewVersionEndsTheOneBefore`; la proyección, `rates.spec.ts`.
- [x] Al día siguiente hay un movimiento "Intereses automáticos" de ≈ $ 2.361 sobre $ 10.000.000 y el saldo subió. — `TestCashAccrualCreditsADayOfInterest`.
- [x] La ganancia del portafolio y la de la plataforma suben en esa cifra; la rentabilidad del reporte también; el aporte neto no. — `TestCashInterestIsGainInThePortfolioAndThePlatform`, `TestCashAccrualCaughtUpStillCountsAsReturn`.
- [x] Correr el job dos veces el mismo día no duplica el abono. — `TestCashAccrualComputesADayOnce`.
- [x] Cambiar la tasa crea una versión nueva; los días anteriores conservan la tasa anterior. — `TestCashAccrualSkipsTheDaysWithoutARate`, `TestCashRateThatEarnedInterestKeepsItsPast`.
- [x] Pausar la tasa detiene los abonos desde esa fecha. — `TestCashAccrualSkipsTheDaysWithoutARate`, `TestCashRatePauseAndResume`.
- [x] Borrar un abono automático no lo vuelve a crear. — `TestCashAccrualDoesNotCreditADeletedInterestAgain`.
- [x] Retirar todo y volver a depositar deja la ganancia en 0. — `TestCashEmptiedBalanceStartsOver`.
- [x] Una cuenta sin tasa se comporta exactamente como hoy. — `TestCashWithoutARateIsUntouched`.

De la Fase 4:

- [x] Una tasa de abono mensual calcula cada día y abona el mes en un solo movimiento el último día, y rinde lo mismo que una diaria. — `TestCashAccrualPostsAMonthOnItsLastDay`.
- [x] Una cuenta que deja de rendir a mitad de mes cobra lo pendiente el último día que ganó. — `TestCashAccrualCreditsWhatAPausedRateHolds`.
- [x] Con tope, las cuentas de una plataforma se lo reparten en proporción a su saldo, y por debajo del tope nada cambia. — `TestCashAccrualSharesTheCapBetweenBalances`, `TestCashAccrualUnderTheCapEarnsOnEverything`.
- [x] Recalcular desde una fecha rehace los días sobre el saldo de ahora, y arrastra el mes entero si el abono lo cruzaba. — `TestRecalculateCashInterestRedoesTheDays`, `TestRecalculateCashInterestClearsAWholeMonthlyCredit`.
