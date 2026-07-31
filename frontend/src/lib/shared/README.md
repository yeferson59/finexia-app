# `lib/shared` — utilidades transversales sin dominio

Código compartido que no pertenece a ninguna feature ni al design system.
Sustituye al antiguo cajón de sastre `lib/utils.ts`, repartido por tema.

## Contenido

- `css.ts` — `cn()`, combinación condicional de clases.
- `format/money.ts` — `formatCurrency()`.
- `format/date.ts` — `formatCalendarDate()`, `todayLocalDateString()`.
- `config/features.ts` — feature flags (`investments`, `selfRegistration`).
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
