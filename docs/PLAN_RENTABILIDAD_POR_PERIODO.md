# Plan — Rentabilidad por periodo en el resumen

> **Estado:** implementado · 25 sep 2026
> **Alcance:** módulo `portfolio` (backend), feature `dashboard` (frontend)
> **Migraciones:** ninguna · **Fases:** 4

### Estado de implementación

| Fase | Estado |
|---|---|
| 1. Backend: cálculo y contrato | Implementada |
| 2. Frontend: franja en el resumen del panel | Implementada |
| 3. Detalle de portafolio | Implementada |
| 4. Correo semanal, MCP y manual | Implementada |

Decisiones que cambiaron respecto al plan al implementarlo:

- **Dónde va la franja.** Dentro de `WealthHeadline`, como fila entera bajo la
  cifra y el selector de moneda, no como bloque suelto en la página: es la
  misma ganancia partida en ventanas y la franja del reparto sigue pegada
  debajo del bloque. Solo se pinta si hay portafolios que sumar y al menos una
  ventana con cifra; una cuenta con un solo día de historia no ve siete rayas.
- **Tono.** Porcentaje y dinero se colorean cada uno por su signo, porque
  pueden discrepar (D4). Lo que redondea a cero va neutro.
- **Separador, fuera del componente.** `PeriodReturns` no trae filete ni aire
  propios: la cabecera del panel lo pone arriba y el detalle de portafolio
  abajo, como su gráfica. Los dos preguntan antes `hasTrailingReturns`, que
  sale por el `index.ts` de `dashboard` junto al componente, para no dejar un
  filete sobre un hueco.
- **Detalle de portafolio.** La ruta compone `PeriodReturns` entre la cifra y
  la gráfica, en la moneda base del portafolio. Va en la ruta porque
  `features/portfolio` no puede importar de `features/dashboard`.
- **Nota de cada ventana.** «Desde el 22 de jul de 2026» o «El historial
  empieza el …» va en `title` y, para el lector de pantalla, en un `dd` oculto.
- **Correo semanal.** Cada fila toma el `1W` de su portafolio. El total en
  dinero es la suma de las filas, como el total del correo es la suma de los
  portafolios, así que las filas siguen sumando la cabecera. El porcentaje del
  total es el `1W` de la serie agregada, porque los porcentajes no se suman.
  La plantilla dice ahora «ganancia desde el …» y «sin contar aportes ni
  retiros». El servicio de portafolios ganó dos lecturas para esto,
  `GetTrailingReturns` y `GetPortfolioTrailingReturns`, y perdió
  `GetPortfolioValuesAsOf` (con `PortfolioValuePoint` y su consulta), que solo
  usaba el correo.
- **MCP.** `get_portfolio_growth` devuelve `returns` con `period`, `from`,
  `gain`, `netFlow` y `pct`, con la misma convención que los fondos: sin
  `from`, el historial no llega y no hay cifra.
- **Manual.** Apartado 5.4 nuevo, que explica además por qué «Desde el inicio»
  no coincide con «sobre lo invertido»; el 6.3 y el 13 lo mencionan. Se
  regeneraron solo las capturas que enseñan la franja (03, 05 y 15).

## 1. Qué se quiere

El resumen del panel dice cuánto hay y cuánto se ha ganado **sobre lo
invertido**, desde siempre. Falta la pregunta que más se hace un inversionista:
¿cuánto gané hoy, esta semana, este mes, este año? Es decir, las
rentabilidades por periodo, en dinero y en porcentaje.

## 2. De dónde sale

No hace falta guardar nada nuevo. Todo lo necesario ya existe:

| Pieza | Dónde | Qué aporta |
|---|---|---|
| Snapshot diario por portafolio | `portfolio_snapshots`, job `SyncPortfolioSnapshots` (`snapshot.go`) | El valor de cada día |
| Serie de crecimiento | `GetPortfolioGrowthByUserID` / `ByPortfolioID` (`postgres_snapshot.go`) | Un punto por día, el de **hoy en vivo**, y el `netFlow` de cada tramo |
| Retorno por tramo | `SubperiodReturns` (`growth_metrics.go`) y `periodReturns` (`shared/finance/returns.ts`) | Dietz modificada: un depósito no cuenta como ganancia |
| Resumen del panel | `wealth-headline.svelte` | Hoy solo la ganancia total sobre lo invertido |

El panel ya pide la serie completa (`getAggregateGrowth(event, { currency })`
sin `period`), así que los periodos pueden viajar en esa misma respuesta sin
una petición más.

## 3. Decisiones

**D1. Periodos.** `1D` (1 día), `1W` (7 días), `1M` (1 mes), `3M` (3 meses),
`YTD` (en lo que va del año), `1Y` (1 año) y `ALL` (desde el inicio).

**D2. Ancla.** Cada periodo se mide desde el **último punto de la serie con
fecha ≤ objetivo**, y hasta el último punto (hoy, en vivo). Los objetivos, con
`hoy` la fecha del último punto:

| Periodo | Objetivo |
|---|---|
| `1D` | hoy − 1 día |
| `1W` | hoy − 7 días |
| `1M` | hoy − 1 mes (`AddDate`) |
| `3M` | hoy − 3 meses |
| `YTD` | 31 de diciembre del año anterior: el cierre desde el que se mide el año |
| `1Y` | hoy − 1 año |
| `ALL` | el primer punto |

Si el job se saltó días, el ancla queda antes del objetivo; la respuesta lleva
la fecha real (`from`) para que nadie compare contra un día que no fue.

**D3. Historial corto.** Si no hay punto en o antes del objetivo, el periodo va
`available: false`. No se publica «1 año» con la cifra de dos meses. `ALL`
está disponible desde dos puntos. `YTD` sigue la misma regla: una cuenta
abierta este año no tiene cierre del anterior, y su cifra es la de `ALL`.

**D4. Dos cifras por periodo.**

- **Ganancia en dinero** = `V_fin − V_ancla − Σ netFlow` de los tramos
  `(ancla, fin]`. Lo ganado o perdido sin contar lo que se metió o sacó.
- **Porcentaje** = rentabilidad ponderada por tiempo: `Π(1 + rᵢ) − 1` sobre los
  mismos tramos, con `rᵢ` de `SubperiodReturns`. Es la cifra de la vista `%` de
  la gráfica y del centro de reportes; el panel no puede decir otra.
- Pueden tener signo distinto en casos raros (mucho dinero entró justo antes de
  una caída). Es la misma discrepancia que el centro de reportes ya explica.
- Si ningún tramo de la ventana tiene base positiva (la cuenta estuvo vacía),
  no hay porcentaje: `returnPct` se omite en vez de publicar un 0 % que nadie
  midió. La ganancia en dinero sí se publica.

**D5. Se calcula en el backend.** Una sola fuente de verdad, que además pueden
usar el MCP (`get_portfolio_growth`) y el correo semanal. El frontend solo lo
enseña. La aritmética es la de `growth_metrics.go`, que ya coincide con
`shared/finance/returns.ts`.

**D6. Siempre sobre la serie entera.** `GET /portfolios/growth?period=1M`
sigue devolviendo solo el último mes de puntos, pero los periodos se calculan
sobre la historia completa: si no, `period=1M` dejaría `1Y` sin ancla. El
servicio pide la serie completa y recorta los puntos en Go. El CTE de flujos ya
recorre todas las transacciones sin importar `since`, así que el coste extra
son unas filas de totales; recortar en Go da los mismos puntos que el filtro
SQL (`fecha >= since`), porque el `netFlow` de cada punto no depende de la
ventana.

**D7. Límites que se documentan, no se arreglan aquí.**

- La serie agregada convierte a la moneda pedida con la tasa **de hoy** en
  todas las fechas (moneda constante), así que la variación cambiaria no
  aparece como rentabilidad.
- Una posición valorada a coste (sin precio de mercado) gana 0 %; el aviso
  `MarketKeyNotice` ya lo cubre.
- `1D` compara el punto en vivo con el snapshot anterior, que el job escribe
  una vez al día en UTC. Por eso se llama «1 día» y no «Hoy».

## 4. Fase 1 — Backend

**4.1 Función pura** — `backend/internal/portfolio/trailing_returns.go`

```go
type TrailingPeriod string // "1D","1W","1M","3M","YTD","1Y","ALL"

type TrailingReturn struct {
    Period     TrailingPeriod
    Available  bool
    From, To   time.Time
    StartValue decimal.Decimal
    EndValue   decimal.Decimal
    NetFlow    decimal.Decimal
    Gain       decimal.Decimal
    Rate       decimal.Decimal
    HasRate    bool
}

func BuildTrailingReturns(points []GrowthPoint) []TrailingReturn
```

Reutiliza `SubperiodReturns`, `growthDecimal` y `returns.ChainReturns`.

**4.2 Servicio** — `GetPortfolioGrowth` y `GetPortfolioGrowthByID` piden la
serie completa, calculan `GrowthSummary.Trailing` sobre ella y recortan los
puntos al `period` pedido.

**4.3 Contrato** — `GrowthResponseDTO` gana `returns`:

```json
"returns": [
  { "period": "1D", "available": true, "from": "2026-09-24", "to": "2026-09-25",
    "startValue": "10250.00", "endValue": "10292.10", "netFlow": "0.00",
    "gain": "42.10", "returnPct": "0.41" },
  { "period": "1Y", "available": false, "historyStart": "2026-03-02" }
]
```

Importes y porcentaje a dos decimales, como el resto de `summary`. Siempre van
los siete periodos, en ese orden.

**4.4 Pruebas** — `trailing_returns_test.go`, sin base de datos:

- un depósito dentro de la ventana no suma a la ganancia ni al porcentaje;
- un retiro no cuenta como pérdida;
- historial corto → `available=false`;
- hueco de snapshots → el ancla es anterior y `From` lo dice;
- `YTD` el 1 de enero y una cuenta abierta este año;
- ventana sin base positiva → `HasRate=false`;
- invariante: `ALL.Rate == BuildGrowthMetrics(points).TotalReturn`;
- invariante: la ganancia es la suma de las de cada tramo.

Y en el servicio y el handler: los periodos salen de la serie entera aunque se
pida `period=1M`, y el JSON lleva la forma de 4.3.

**4.5 Documentación** — sección «Rentabilidad por periodo» en `docs/API.md`.

## 5. Fase 2 — Frontend: resumen del panel

- **Contrato.** `returns` opcional en `portfolioGrowthSchema` y en
  `PortfolioGrowth`: un backend anterior no lo manda y la franja no aparece.
- **Lógica pura.** `features/dashboard/trailing.ts` (+ `trailing.spec.ts`):
  etiquetas («1 día», «7 días», «1 mes», «3 meses», «En el año», «1 año»,
  «Desde el inicio»), tono y el texto de lo no disponible («El historial
  empieza el 2 mar 2026»).
- **Componente.** `features/dashboard/components/period-returns.svelte`, bajo
  `WealthHeadline`:

  ```
  1 día      7 días     1 mes      3 meses    En el año   1 año   Desde el inicio
  +0,41 %    −1,20 %    +2,85 %    +6,10 %    +9,40 %     —       +13,02 %
  +$42.100   −$124.000  +$285.300  +$590.000  +$880.000           +$1.150.000
  ```

  Importes con `privacy.money(...)`, porcentaje con `formatSignedPercent`,
  verde y rojo con `--green`/`--red`, cifras tabulares, grilla de 2–3 columnas
  en móvil sin scroll horizontal, `<dl>` o tabla con `<caption>` para lectores
  de pantalla. Cambia sola con el selector de moneda.
- **Pruebas.** `period-returns.svelte.spec.ts`; `returns` en
  `e2e/mocks/mock-api.mjs` y un caso de Playwright. `pnpm check`, `pnpm lint`,
  `pnpm check:arch`, `pnpm test:unit -- --run`.

## 6. Fase 3 — Detalle de portafolio

`GET /portfolios/:id/growth` ya trae `returns` desde la Fase 1; la página
`routes/dashboard/portfolios/[id]` reutiliza el componente.

## 7. Fase 4 — Lo que esto deja arreglar

- **Correo semanal** (`notification/weekly_summary.go`). Hoy compara el valor
  con el de hace siete días (`applyChange` + `ROI`), así que **un depósito sale
  como ganancia de la semana**. Pasa a usar `1W` de `BuildTrailingReturns`.
- **MCP.** `returns` en la salida de `get_portfolio_growth`.
- **Manual de usuario.** Sección nueva y PDF regenerado (`pnpm check:manual`).
