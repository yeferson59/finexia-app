# `lib/shared` — utilidades transversales sin dominio

Código compartido que no pertenece a ninguna feature ni al design system.
Sustituye al antiguo cajón de sastre `lib/utils.ts`, repartido por tema.

## Contenido

- `css.ts` — `cn()`, combinación condicional de clases.
- `currency.ts` — `SUPPORTED_CURRENCIES` y `resolveDisplayCurrency()`: en qué
  monedas puede expresar dinero la app, copia de la lista que valida el backend.
- `format/money.ts` — `formatCurrency()`, `currencySymbol()`.
- `format/date.ts` — `formatCalendarDate()`, `todayLocalDateString()`.
- `format/asset-type.ts` — etiquetas y colores de `market.AssetType` (`stock`,
  `etf`, …), la clase de un activo. Los tenían duplicados el detalle de
  portafolio y el donut del dashboard, y la segunda copia hablaba el
  vocabulario _plural_ de `portfolio.type`, con lo que no acertaba ninguna
  clave. Una sola tabla mantiene el mismo color para la misma clase en las dos
  gráficas.
- `format/portfolio-type.ts` — etiquetas del `type` de un portafolio (plural y
  con combinaciones, `stocks_etfs`). Es otro vocabulario: no confundir con el
  anterior.
- `format/percent.ts` — `formatPercent()` y `formatSignedPercent()`, con la coma
  decimal de es-CO: los porcentajes se escapaban con `toFixed`, que escribe un
  punto, y en una misma tarjeta convivían «+12.35%» y «$1.234,50».
- `finance/returns.ts` — la aritmética de rentabilidad: retornos por tramo con
  Dietz modificada, encadenado, rentabilidad ponderada por tiempo, anualización
  y la curva acumulada que dibuja la gráfica en porcentaje. Vivió en
  `features/reports`, que la estrenó; bajó aquí cuando el dashboard y el detalle
  de portafolio necesitaron la misma cifra, porque dos copias del cálculo
  acaban dando dos respuestas distintas para lo mismo. La entrada
  (`ReturnSeriesPoint`) se declara estructuralmente para no importar de
  `lib/api`.
- `config/features.ts` — feature flags (`investments`, `selfRegistration`).
- `flash.svelte.ts` — `flash()`, acuse temporal que se retira solo. Estaba
  copiado en tres pantallas del dashboard y las tres copias dejaban el
  `setTimeout` vivo al desmontar y compartían reloj entre dos acuses seguidos.
- `form.ts` — reparto del `form` de una página entre sus secciones
  (`actionSucceeded`, `actionError`, `actionData`), que usan ajustes y
  notificaciones.
- `privacy.svelte.ts` — modo oculto: enmascara importes en todo el dashboard.
  Es estado transversal de verdad (lo consumen widgets del dashboard, el
  detalle de portfolio y el de plataforma), por eso vive aquí y no en una
  feature.

Cada módulo se importa por su ruta (`$lib/shared/format/date`): no hay barrel.
El antiguo `lib/utils.ts` —el cajón de sastre que repartió la Fase 1— se
eliminó en la Fase 7, igual que `lib/stores/`.

## Reglas de dependencia

- `lib/shared` **no** importa de `lib/features`, `lib/api` ni `lib/ui`: es la
  capa más baja y todos pueden depender de ella.
